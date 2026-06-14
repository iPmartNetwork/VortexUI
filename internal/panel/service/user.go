package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NodeOps is the slice of the hub the user service needs to push live user
// changes to the fleet. *hub.Hub satisfies it (asserted in wiring), and tests
// supply a fake so provisioning logic is exercised without real nodes.
type NodeOps interface {
	AddUser(ctx context.Context, nodeID uuid.UUID, inboundTag string, u *domain.User) error
	RemoveUser(ctx context.Context, nodeID uuid.UUID, inboundTag string, userID uuid.UUID) error
}

// UserService creates and manages service users, keeping the database and the
// live cores in sync.
type UserService struct {
	users port.UserRepository
	nodes NodeOps
	now   func() time.Time
}

// NewUserService wires the user service.
func NewUserService(users port.UserRepository, nodes NodeOps) *UserService {
	return &UserService{users: users, nodes: nodes, now: time.Now}
}

// CreateUserInput describes a new user. Credentials and the subscription token
// are generated server-side, never accepted from the client.
type CreateUserInput struct {
	Username      string
	Note          string
	DataLimit     int64
	ExpireAt      *time.Time
	DeviceLimit   int
	ResetStrategy domain.ResetStrategy
	InboundIDs    []uuid.UUID
	OnHold        bool
}

// Create persists a user, binds it to inbounds, and provisions it on the live
// cores. The user is returned even if provisioning partially fails (it is safely
// in the DB and a later Sync reconciles), with the provisioning error wrapped so
// the caller can surface a warning.
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	if in.Username == "" {
		return nil, errors.New("username is required")
	}
	creds, err := newCredentials()
	if err != nil {
		return nil, err
	}
	tok, err := randToken()
	if err != nil {
		return nil, err
	}

	now := s.now()
	status := domain.UserStatusActive
	if in.OnHold {
		status = domain.UserStatusOnHold
	}
	reset := in.ResetStrategy
	if reset == "" {
		reset = domain.ResetNone
	}

	u := &domain.User{
		ID:            uuid.New(),
		Username:      in.Username,
		Status:        status,
		Note:          in.Note,
		DataLimit:     in.DataLimit,
		ExpireAt:      in.ExpireAt,
		DeviceLimit:   in.DeviceLimit,
		ResetStrategy: reset,
		Proxies:       creds,
		SubToken:      tok,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("persist user: %w", err)
	}
	if len(in.InboundIDs) > 0 {
		if err := s.users.SetInbounds(ctx, u.ID, in.InboundIDs); err != nil {
			return u, fmt.Errorf("bind inbounds: %w", err)
		}
	}
	if err := s.provision(ctx, u); err != nil {
		return u, fmt.Errorf("user created but provisioning incomplete: %w", err)
	}
	return u, nil
}

// UpdateUserInput is the mutable subset of a user (PUT semantics: these fields
// are replaced wholesale). Credentials, username, and sub-token are immutable
// here by design — rotating them is a separate, deliberate operation.
type UpdateUserInput struct {
	Note          string
	Status        domain.UserStatus
	DataLimit     int64
	ExpireAt      *time.Time
	DeviceLimit   int
	ResetStrategy domain.ResetStrategy
}

// Update applies metadata changes to a user and persists them. Quota/expiry are
// accounting fields enforced elsewhere, so no live core change is needed; a
// status flip to disabled de-provisions, and re-enabling re-provisions.
func (s *UserService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	wasActive := u.Status != domain.UserStatusDisabled

	u.Note = in.Note
	u.Status = in.Status
	u.DataLimit = in.DataLimit
	u.ExpireAt = in.ExpireAt
	u.DeviceLimit = in.DeviceLimit
	if in.ResetStrategy != "" {
		u.ResetStrategy = in.ResetStrategy
	}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}

	nowDisabled := u.Status == domain.UserStatusDisabled
	switch {
	case wasActive && nowDisabled:
		s.deprovision(ctx, u)
	case !wasActive && !nowDisabled:
		if err := s.provision(ctx, u); err != nil {
			return u, err
		}
	}
	return u, nil
}

// deprovision best-effort removes a user from every node it is bound to.
func (s *UserService) deprovision(ctx context.Context, u *domain.User) {
	inbounds, err := s.users.InboundsFor(ctx, u.ID)
	if err != nil {
		return
	}
	for _, in := range inbounds {
		_ = s.nodes.RemoveUser(ctx, in.NodeID, in.Tag, u.ID)
	}
}

// Delete removes a user from the live cores and then the database.
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	inbounds, err := s.users.InboundsFor(ctx, id)
	if err != nil {
		return err
	}
	for _, in := range inbounds {
		// Best-effort de-provision; a dead node must not block deletion.
		_ = s.nodes.RemoveUser(ctx, in.NodeID, in.Tag, u.ID)
	}
	return s.users.Delete(ctx, id)
}

// Sync re-applies a user to every node it is bound to, used after edits or to
// reconcile a node that was offline.
func (s *UserService) Sync(ctx context.Context, id uuid.UUID) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.provision(ctx, u)
}

// provision pushes the user to the live core on each inbound's node, collecting
// per-node failures so one unreachable node does not hide the others.
func (s *UserService) provision(ctx context.Context, u *domain.User) error {
	inbounds, err := s.users.InboundsFor(ctx, u.ID)
	if err != nil {
		return err
	}
	var errs []error
	for _, in := range inbounds {
		if err := s.nodes.AddUser(ctx, in.NodeID, in.Tag, u); err != nil {
			errs = append(errs, fmt.Errorf("node %s inbound %s: %w", in.NodeID, in.Tag, err))
		}
	}
	return errors.Join(errs...)
}

// --- credential generation (server-side only) ---

func newCredentials() (domain.UserCredentials, error) {
	trojan, err := randHex(16)
	if err != nil {
		return domain.UserCredentials{}, err
	}
	ss, err := randHex(16)
	if err != nil {
		return domain.UserCredentials{}, err
	}
	return domain.UserCredentials{
		VMessUUID:    uuid.New(),
		VLESSUUID:    uuid.New(),
		TrojanPass:   trojan,
		ShadowsocksP: ss,
		SSMethod:     "aes-128-gcm",
	}, nil
}

func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
