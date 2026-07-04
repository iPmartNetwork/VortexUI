package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- fakes ---

type fakeAdminRepo struct {
	admin *domain.Admin
	role  *domain.Role
}

func (f *fakeAdminRepo) Create(context.Context, *domain.Admin) error { return nil }
func (f *fakeAdminRepo) GetByUsername(_ context.Context, u string) (*domain.Admin, error) {
	if f.admin != nil && f.admin.Username == u {
		return f.admin, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeAdminRepo) GetByID(context.Context, uuid.UUID) (*domain.Admin, error) {
	return f.admin, nil
}
func (f *fakeAdminRepo) Update(context.Context, *domain.Admin) error { return nil }
func (f *fakeAdminRepo) GetRole(context.Context, uuid.UUID) (*domain.Role, error) {
	return f.role, nil
}

type fakeUserRepo struct {
	created  *domain.User
	bound    []uuid.UUID
	inbounds []domain.Inbound
}

func (f *fakeUserRepo) Create(_ context.Context, u *domain.User) error { f.created = u; return nil }
func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (*domain.User, error) {
	if f.created == nil {
		return nil, domain.ErrNotFound
	}
	return f.created, nil
}
func (f *fakeUserRepo) GetBySubToken(context.Context, string) (*domain.User, error) {
	return f.created, nil
}
func (f *fakeUserRepo) Update(context.Context, *domain.User) error { return nil }
func (f *fakeUserRepo) Delete(context.Context, uuid.UUID) error    { return nil }
func (f *fakeUserRepo) List(context.Context, port.UserFilter) ([]*domain.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) AddUsedTraffic(context.Context, uuid.UUID, int64) error         { return nil }
func (f *fakeUserRepo) AddUsedTrafficBatch(context.Context, map[uuid.UUID]int64) error { return nil }
func (f *fakeUserRepo) SetInbounds(_ context.Context, _ uuid.UUID, ids []uuid.UUID) error {
	f.bound = ids
	return nil
}
func (f *fakeUserRepo) InboundsFor(context.Context, uuid.UUID) ([]domain.Inbound, error) {
	return f.inbounds, nil
}
func (f *fakeUserRepo) PrimaryInboundProtocols(context.Context, []uuid.UUID) (map[uuid.UUID]string, error) {
	return map[uuid.UUID]string{}, nil
}
func (f *fakeUserRepo) StatsForAdmin(context.Context, uuid.UUID) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, nil
}
func (f *fakeUserRepo) StatsByStatusForAdmin(context.Context, uuid.UUID) (map[string]int64, error) {
	return nil, nil
}
func (f *fakeUserRepo) TopUsersForAdmin(context.Context, uuid.UUID, int32) ([]domain.ResellerTopUser, error) {
	return nil, nil
}
func (f *fakeUserRepo) CountExpiringSoonForAdmin(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}
func (f *fakeUserRepo) CountCreatedSinceForAdmin(context.Context, uuid.UUID, time.Time) (int64, error) {
	return 0, nil
}

type fakeNodeOps struct {
	added   []string
	removed []string
}

func (f *fakeNodeOps) AddUser(_ context.Context, nodeID uuid.UUID, tag string, _ *domain.User) error {
	f.added = append(f.added, nodeID.String()+"/"+tag)
	return nil
}
func (f *fakeNodeOps) RemoveUser(_ context.Context, nodeID uuid.UUID, tag string, _ uuid.UUID) error {
	f.removed = append(f.removed, nodeID.String()+"/"+tag)
	return nil
}

// --- tests ---

func newAdmin(t *testing.T, username, password string, totp bool) *domain.Admin {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	a := &domain.Admin{ID: uuid.New(), Username: username, PasswordHash: hash, Sudo: true}
	if totp {
		secret, _, _ := auth.GenerateTOTP("VortexUI", username)
		a.TOTPSecret = secret
		a.TOTPEnabled = true
	}
	return a
}

func TestAuthLogin(t *testing.T) {
	admin := newAdmin(t, "root", "hunter2", false)
	repo := &fakeAdminRepo{admin: admin}
	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	svc := NewAuthService(repo, iss)
	ctx := context.Background()

	token, err := svc.Login(ctx, LoginInput{Username: "root", Password: "hunter2"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if claims, err := iss.Verify(token); err != nil || claims.AdminID != admin.ID {
		t.Fatalf("issued token invalid: %v", err)
	}

	if _, err := svc.Login(ctx, LoginInput{Username: "root", Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("wrong password err = %v, want ErrInvalidCredentials", err)
	}
	if _, err := svc.Login(ctx, LoginInput{Username: "ghost", Password: "x"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("unknown user err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthLoginRequiresTOTPWhenEnabled(t *testing.T) {
	admin := newAdmin(t, "root", "pw", true)
	svc := NewAuthService(&fakeAdminRepo{admin: admin}, auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour))
	ctx := context.Background()

	if _, err := svc.Login(ctx, LoginInput{Username: "root", Password: "pw"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Error("login without TOTP should fail when 2FA enabled")
	}
}

func TestUserUpdateMetadataAndDisableDeprovisions(t *testing.T) {
	nodeID := uuid.New()
	uid := uuid.New()
	repo := &fakeUserRepo{
		created:  &domain.User{ID: uid, Username: "alice", Status: domain.UserStatusActive, DataLimit: 100},
		inbounds: []domain.Inbound{{NodeID: nodeID, Tag: "vless-ws"}},
	}
	ops := &fakeNodeOps{}
	svc := NewUserService(repo, ops)
	exp := time.Now().Add(24 * time.Hour)

	// Plain metadata edit while staying active: no de-provision.
	u, err := svc.Update(context.Background(), uid, UpdateUserInput{
		Status: domain.UserStatusActive, DataLimit: 500, ExpireAt: &exp, Note: "vip",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if u.DataLimit != 500 || u.Note != "vip" || u.ExpireAt == nil {
		t.Errorf("fields not applied: %+v", u)
	}
	if len(ops.removed) != 0 {
		t.Errorf("active edit should not de-provision, got %v", ops.removed)
	}

	// Flip to disabled: must de-provision from every bound node.
	if _, err := svc.Update(context.Background(), uid, UpdateUserInput{Status: domain.UserStatusDisabled}); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if len(ops.removed) != 1 || ops.removed[0] != nodeID.String()+"/vless-ws" {
		t.Errorf("disable did not de-provision: %v", ops.removed)
	}
}

func TestUserServiceCreateGeneratesCredsAndProvisions(t *testing.T) {
	nodeID := uuid.New()
	inID := uuid.New()
	repo := &fakeUserRepo{inbounds: []domain.Inbound{{ID: inID, NodeID: nodeID, Tag: "vless-ws"}}}
	ops := &fakeNodeOps{}
	svc := NewUserService(repo, ops)

	u, err := svc.Create(context.Background(), CreateUserInput{
		Username: "alice", DataLimit: 1 << 30, InboundIDs: []uuid.UUID{inID},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Server-generated identity.
	if u.Proxies.VMessUUID == uuid.Nil || u.Proxies.VLESSUUID == uuid.Nil {
		t.Error("expected generated vmess/vless uuids")
	}
	if u.SubToken == "" || u.Proxies.TrojanPass == "" {
		t.Error("expected generated sub token and trojan password")
	}
	if repo.created == nil || len(repo.bound) != 1 {
		t.Error("user not persisted/bound")
	}
	// Provisioned on the node behind its inbound.
	if len(ops.added) != 1 || ops.added[0] != nodeID.String()+"/vless-ws" {
		t.Errorf("provisioning = %v, want [%s/vless-ws]", ops.added, nodeID)
	}
}

func TestUserResetUsageReactivatesAndReprovisions(t *testing.T) {
	nodeID := uuid.New()
	uid := uuid.New()
	repo := &fakeUserRepo{
		created: &domain.User{
			ID: uid, Username: "alice", Status: domain.UserStatusLimited,
			DataLimit: 100, UsedTraffic: 150,
		},
		inbounds: []domain.Inbound{{NodeID: nodeID, Tag: "vless-ws"}},
	}
	ops := &fakeNodeOps{}
	svc := NewUserService(repo, ops)

	u, err := svc.ResetUsage(context.Background(), uid)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if u.UsedTraffic != 0 || u.LastReset == nil {
		t.Errorf("usage not reset: used=%d lastReset=%v", u.UsedTraffic, u.LastReset)
	}
	if u.Status != domain.UserStatusActive {
		t.Errorf("status = %q, want active after reset", u.Status)
	}
	// Was limited (inactive) -> now active: must re-provision on the node.
	if len(ops.added) != 1 || ops.added[0] != nodeID.String()+"/vless-ws" {
		t.Errorf("re-provision = %v, want [%s/vless-ws]", ops.added, nodeID)
	}
}

func TestUserResetUsageActiveUserDoesNotReprovision(t *testing.T) {
	uid := uuid.New()
	repo := &fakeUserRepo{
		created:  &domain.User{ID: uid, Username: "bob", Status: domain.UserStatusActive, DataLimit: 1000, UsedTraffic: 10},
		inbounds: []domain.Inbound{{NodeID: uuid.New(), Tag: "t"}},
	}
	ops := &fakeNodeOps{}
	svc := NewUserService(repo, ops)

	u, err := svc.ResetUsage(context.Background(), uid)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if u.UsedTraffic != 0 {
		t.Errorf("used = %d, want 0", u.UsedTraffic)
	}
	// Already active -> no redundant provisioning.
	if len(ops.added) != 0 {
		t.Errorf("active user should not be re-provisioned, got %v", ops.added)
	}
}

func TestUserRevokeSubTokenRotatesToken(t *testing.T) {
	uid := uuid.New()
	repo := &fakeUserRepo{created: &domain.User{ID: uid, Username: "alice", SubToken: "old-token"}}
	svc := NewUserService(repo, &fakeNodeOps{})

	u, err := svc.RevokeSubToken(context.Background(), uid)
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if u.SubToken == "old-token" || u.SubToken == "" {
		t.Errorf("sub token not rotated: %q", u.SubToken)
	}
}

func TestSubscriptionBuildForUserResolvesByID(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	nodeID := uuid.New()
	uid := uuid.New()
	user := &domain.User{ID: uid, Username: "alice", SubToken: "tok", Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	userRepo := &fakeUserRepo{
		created:  user,
		inbounds: []domain.Inbound{{ID: uuid.New(), NodeID: nodeID, Tag: "in", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true}},
	}
	nodeRepo := &mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
		nodeID: {ID: nodeID, Name: "n1", Address: "1.1.1.1:50051", LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true}},
	}}
	svc := NewSubscriptionService(userRepo, nodeRepo, nil)
	svc.now = func() time.Time { return now }

	res, err := svc.BuildForUser(context.Background(), uid)
	if err != nil {
		t.Fatalf("build for user: %v", err)
	}
	if res.User.ID != uid || len(res.Proxies) != 1 {
		t.Errorf("unexpected result: user=%s proxies=%d", res.User.ID, len(res.Proxies))
	}
}

// fakeOnlineQuerier returns per-node online maps for LiveConnections tests.
type fakeOnlineQuerier struct {
	byNode map[uuid.UUID]map[string]int
	errs   map[uuid.UUID]bool
}

func (f *fakeOnlineQuerier) OnlineStats(_ context.Context, nodeID uuid.UUID) (map[string]int, error) {
	if f.errs[nodeID] {
		return nil, errors.New("node unreachable")
	}
	return f.byNode[nodeID], nil
}

func (f *fakeOnlineQuerier) OnlineIPs(_ context.Context, nodeID uuid.UUID, _ string) (map[string]int64, error) {
	if f.errs[nodeID] {
		return nil, errors.New("node unreachable")
	}
	return nil, nil
}

func TestUserLiveConnectionsSumsAcrossNodesAndSkipsErrors(t *testing.T) {
	uid := uuid.New()
	nodeA, nodeB, nodeDown := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeUserRepo{
		created: &domain.User{ID: uid, Username: "alice"},
		inbounds: []domain.Inbound{
			{NodeID: nodeA, Tag: "a1"},
			{NodeID: nodeA, Tag: "a2"}, // same node: queried once
			{NodeID: nodeB, Tag: "b1"},
			{NodeID: nodeDown, Tag: "d1"}, // unreachable: skipped
		},
	}
	svc := NewUserService(repo, &fakeNodeOps{})

	// Without a querier wired: not tracked.
	if n, tracked, err := svc.LiveConnections(context.Background(), uid); err != nil || tracked || n != 0 {
		t.Errorf("untracked = (%d,%v,%v), want (0,false,nil)", n, tracked, err)
	}

	svc.SetOnlineQuerier(&fakeOnlineQuerier{
		byNode: map[uuid.UUID]map[string]int{
			nodeA: {uid.String(): 2},
			nodeB: {uid.String(): 3},
		},
		errs: map[uuid.UUID]bool{nodeDown: true},
	})
	n, tracked, err := svc.LiveConnections(context.Background(), uid)
	if err != nil || !tracked {
		t.Fatalf("tracked query failed: tracked=%v err=%v", tracked, err)
	}
	if n != 5 {
		t.Errorf("live connections = %d, want 5 (2+3, node queried once, down node skipped)", n)
	}
}

func TestNodeServiceLogs(t *testing.T) {
	id := uuid.New()
	repo := &fakeNodeRepo{created: &domain.Node{ID: id}}
	svc := NewNodeService(repo, &fakeRegistrar{})

	// No querier wired -> error.
	if _, err := svc.Logs(context.Background(), id, 10); err == nil {
		t.Error("expected error without a log querier")
	}

	svc.SetLogQuerier(fakeLogQuerier{lines: []string{"l1", "l2"}})
	lines, err := svc.Logs(context.Background(), id, 10)
	if err != nil || len(lines) != 2 {
		t.Errorf("logs = %v err=%v, want 2 lines", lines, err)
	}
}

type fakeLogQuerier struct{ lines []string }

func (f fakeLogQuerier) Logs(context.Context, uuid.UUID, int) ([]string, error) {
	return f.lines, nil
}
