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

// Resyncer rebuilds a node's complete desired config (including WireGuard peers)
// and pushes it to the live core. *SyncService satisfies it. It is wired
// optionally so a nil resyncer leaves the lightweight AddUser/RemoveUser path in
// place for nodes that do not host a WireGuard inbound.
type Resyncer interface {
	Resync(ctx context.Context, nodeID uuid.UUID) error
}

// UserService creates and manages service users, keeping the database and the
// live cores in sync.
type UserService struct {
	users  port.UserRepository
	nodes  NodeOps
	now    func() time.Time
	pub    events.Publisher
	online OnlineQuerier
	resync Resyncer
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

// SetResyncer wires the full node-config resync path (the SyncService). It is
// used to provision WireGuard peers, which are only computed during a full
// Resync. A nil resyncer leaves the lightweight AddUser/RemoveUser path in place
// for every node.
func (s *UserService) SetResyncer(r Resyncer) {
	if r != nil {
		s.resync = r
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

// ActiveWithDeviceLimit returns active users that have a positive device limit —
// the candidates for account-sharing checks. Paginated internally.
func (s *UserService) ActiveWithDeviceLimit(ctx context.Context) ([]*domain.User, error) {
	out := []*domain.User{}
	const page = 200
	for offset := 0; ; offset += page {
		batch, total, err := s.users.List(ctx, port.UserFilter{Status: domain.UserStatusActive, Limit: page, Offset: offset})
		if err != nil {
			return nil, err
		}
		for _, u := range batch {
			if u.DeviceLimit > 0 {
				out = append(out, u)
			}
		}
		if offset+page >= total || len(batch) == 0 {
			break
		}
	}
	return out, nil
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
	AdminID       *uuid.UUID // creator (reseller ownership)
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
		AdminID:       in.AdminID,
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
	// InboundIDs replaces the user's inbound bindings. nil leaves them unchanged;
	// an empty (non-nil) slice clears all bindings.
	InboundIDs []uuid.UUID
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

	// Rebind inbounds if requested: drop the user from its old inbounds on the
	// live cores, repoint the bindings, then re-provision onto the new set.
	if in.InboundIDs != nil {
		old, _ := s.users.InboundsFor(ctx, u.ID)
		if err := s.users.SetInbounds(ctx, u.ID, in.InboundIDs); err != nil {
			return u, fmt.Errorf("rebind inbounds: %w", err)
		}
		for _, ib := range old {
			_ = s.nodes.RemoveUser(ctx, ib.NodeID, ib.Tag, u.ID)
		}
		// Rebuild WireGuard peers on any node that previously hosted a WG inbound
		// for this user: bindings are now updated, so a resync drops the user from
		// nodes they were unbound from. New WG nodes are handled by provision.
		s.resyncWireGuardNodes(ctx, old)
		if u.Status != domain.UserStatusDisabled {
			_ = s.provision(ctx, u)
		}
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
	// Rebuild WireGuard peers on affected nodes so the rendered server config
	// reflects the change (the lightweight RemoveUser path does not touch peers).
	s.resyncWireGuardNodes(ctx, inbounds)
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
	// After the user (and its bindings) are gone, rebuild WireGuard peers on any
	// node that hosted a WG inbound for them so they drop out of the server config.
	s.resyncWireGuardNodes(ctx, inbounds)
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

// provision pushes the user to the live core on each inbound's node. Nodes that
// host a WireGuard inbound among the user's bindings are fully resynced (the
// only path that computes WireGuard peers), while nodes without a WG inbound
// keep the lightweight per-inbound AddUser hot-add path so xray hot-add is not
// regressed. Per-node failures are collected so one unreachable node does not
// hide the others.
func (s *UserService) provision(ctx context.Context, u *domain.User) error {
	inbounds, err := s.users.InboundsFor(ctx, u.ID)
	if err != nil {
		return err
	}
	byNode, order := groupInboundsByNode(inbounds)
	var errs []error
	for _, nodeID := range order {
		ins := byNode[nodeID]
		if s.resync != nil && nodeHasWireGuard(ins) {
			// Full resync rebuilds the whole node config including WG peers (and
			// re-applies every other inbound/user on the node).
			if err := s.resync.Resync(ctx, nodeID); err != nil {
				errs = append(errs, fmt.Errorf("node %s resync: %w", nodeID, err))
			}
			continue
		}
		for _, in := range ins {
			if err := s.nodes.AddUser(ctx, in.NodeID, in.Tag, u); err != nil {
				errs = append(errs, fmt.Errorf("node %s inbound %s: %w", in.NodeID, in.Tag, err))
			}
		}
	}
	return errors.Join(errs...)
}

// groupInboundsByNode buckets inbounds by their node id, returning the buckets
// plus the node ids in first-seen order for deterministic iteration.
func groupInboundsByNode(inbounds []domain.Inbound) (map[uuid.UUID][]domain.Inbound, []uuid.UUID) {
	byNode := map[uuid.UUID][]domain.Inbound{}
	var order []uuid.UUID
	for _, in := range inbounds {
		if _, ok := byNode[in.NodeID]; !ok {
			order = append(order, in.NodeID)
		}
		byNode[in.NodeID] = append(byNode[in.NodeID], in)
	}
	return byNode, order
}

// nodeHasWireGuard reports whether any inbound in the slice is a WireGuard one.
func nodeHasWireGuard(inbounds []domain.Inbound) bool {
	for _, in := range inbounds {
		if in.Protocol == domain.ProtoWireGuard {
			return true
		}
	}
	return false
}

// resyncWireGuardNodes best-effort rebuilds the full config of every node that
// hosts a WireGuard inbound among the given inbounds. Used after a binding change
// so the WireGuard peer set (which is only computed during a full Resync) is
// rebuilt — adding newly-bound users and dropping unbound ones. A nil resyncer
// makes this a no-op.
func (s *UserService) resyncWireGuardNodes(ctx context.Context, inbounds []domain.Inbound) {
	if s.resync == nil {
		return
	}
	byNode, order := groupInboundsByNode(inbounds)
	for _, nodeID := range order {
		if nodeHasWireGuard(byNode[nodeID]) {
			_ = s.resync.Resync(ctx, nodeID)
		}
	}
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
