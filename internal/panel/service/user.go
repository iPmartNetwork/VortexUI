package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NodeOps is the slice of the hub the user service needs to push live user
// changes to the fleet. *hub.Hub satisfies it (asserted in wiring), and tests
// supply a fake so provisioning logic is exercised without real nodes.
type NodeOps interface {
	AddUser(ctx context.Context, nodeID uuid.UUID, inboundTag string, u *domain.User) error
	RemoveUser(ctx context.Context, nodeID uuid.UUID, inboundTag string, userID uuid.UUID) error
}

// OnlineQuerier reports a node's live per-user connection counts (keyed by the
// user's stats email == UUID). *hub.Hub satisfies it.
type OnlineQuerier interface {
	OnlineStats(ctx context.Context, nodeID uuid.UUID) (map[string]int, error)
	OnlineIPs(ctx context.Context, nodeID uuid.UUID, userID string) (map[string]int64, error)
}

// UserService creates and manages service users, keeping the database and the
// live cores in sync.
type UserService struct {
	users  port.UserRepository
	nodes  NodeOps
	now    func() time.Time
	pub    events.Publisher
	online OnlineQuerier
}

// NewUserService wires the user service.
func NewUserService(users port.UserRepository, nodes NodeOps) *UserService {
	return &UserService{users: users, nodes: nodes, now: time.Now, pub: events.Nop{}}
}

// SetPublisher wires an event publisher (so create/delete emit user.created /
// user.deleted). A nil publisher leaves the no-op default in place.
func (s *UserService) SetPublisher(p events.Publisher) {
	if p != nil {
		s.pub = p
	}
}

// SetOnlineQuerier wires the source of live per-node connection counts (the
// hub). Without it, LiveConnections reports "not tracked".
func (s *UserService) SetOnlineQuerier(q OnlineQuerier) {
	if q != nil {
		s.online = q
	}
}

// LiveConnections sums a user's current live connections across every node it is
// bound to. The returned bool is false when no online source is wired (so the
// caller can distinguish "zero connections" from "not tracked"). Unreachable
// nodes are skipped rather than failing the whole query.
func (s *UserService) LiveConnections(ctx context.Context, id uuid.UUID) (int, bool, error) {
	if s.online == nil {
		return 0, false, nil
	}
	inbounds, err := s.users.InboundsFor(ctx, id)
	if err != nil {
		return 0, false, err
	}
	email := id.String()
	total := 0
	seen := map[uuid.UUID]bool{}
	for _, in := range inbounds {
		if seen[in.NodeID] {
			continue
		}
		seen[in.NodeID] = true
		m, err := s.online.OnlineStats(ctx, in.NodeID)
		if err != nil {
			continue // node down / not connected: skip its contribution
		}
		total += m[email]
	}
	return total, true, nil
}

// OnlineIP is one source IP currently online for a user, with its last-seen time.
type OnlineIP struct {
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"last_seen"`
}

// OnlineIPList returns the distinct source IPs a user is currently connected
// from across every node it is bound to (deduplicated, keeping the most recent
// last-seen per IP). The bool is false when no online source is wired. A high
// count is a strong signal of account sharing. Unreachable nodes are skipped.
func (s *UserService) OnlineIPList(ctx context.Context, id uuid.UUID) ([]OnlineIP, bool, error) {
	if s.online == nil {
		return nil, false, nil
	}
	inbounds, err := s.users.InboundsFor(ctx, id)
	if err != nil {
		return nil, false, err
	}
	email := id.String()
	latest := map[string]int64{}
	seen := map[uuid.UUID]bool{}
	for _, in := range inbounds {
		if seen[in.NodeID] {
			continue
		}
		seen[in.NodeID] = true
		m, err := s.online.OnlineIPs(ctx, in.NodeID, email)
		if err != nil {
			continue
		}
		for ip, ts := range m {
			if ts > latest[ip] {
				latest[ip] = ts
			}
		}
	}
	out := make([]OnlineIP, 0, len(latest))
	for ip, ts := range latest {
		out = append(out, OnlineIP{IP: ip, LastSeen: time.Unix(ts, 0)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastSeen.After(out[j].LastSeen) })
	return out, true, nil
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
	s.pub.Publish(events.Event{Type: events.UserCreated, UserID: u.ID.String(), Username: u.Username})
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
	if err := s.users.Delete(ctx, id); err != nil {
		return err
	}
	s.pub.Publish(events.Event{Type: events.UserDeleted, UserID: u.ID.String(), Username: u.Username})
	return nil
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

// ResetUsage zeroes a user's used traffic immediately (a manual reset, distinct
// from the scheduled Resetter) and recomputes status. If the reset lifts a
// quota-driven block, the user is re-provisioned on the live cores. Returns the
// updated user; a provisioning failure is returned as a non-nil warning error
// alongside the (already durable) user.
func (s *UserService) ResetUsage(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := s.now()
	wasActive := u.IsActive(now)

	u.UsedTraffic = 0
	u.LastReset = &now
	u.Status = u.DerivedStatus(now)
	u.UpdatedAt = now
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	if !wasActive && u.Status == domain.UserStatusActive {
		if err := s.provision(ctx, u); err != nil {
			return u, fmt.Errorf("usage reset but re-provisioning incomplete: %w", err)
		}
	}
	return u, nil
}

// RevokeSubToken rotates a user's subscription token, invalidating the old
// subscription URL and issuing a new one. Protocol credentials are unchanged, so
// the live cores need no update — only the public subscription link changes.
func (s *UserService) RevokeSubToken(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tok, err := randToken()
	if err != nil {
		return nil, err
	}
	u.SubToken = tok
	u.UpdatedAt = s.now()
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
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
