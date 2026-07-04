package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// Minimal fakes to drive the HTTP stack end-to-end without storage.

type fakeAdminRepo struct{ admin *domain.Admin }

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
func (f *fakeAdminRepo) Update(context.Context, *domain.Admin) error              { return nil }
func (f *fakeAdminRepo) GetRole(context.Context, uuid.UUID) (*domain.Role, error) { return nil, nil }
func (f *fakeAdminRepo) List(context.Context) ([]*domain.Admin, error) {
	return []*domain.Admin{f.admin}, nil
}
func (f *fakeAdminRepo) Delete(context.Context, uuid.UUID) error        { return nil }
func (f *fakeAdminRepo) CountSudo(context.Context) (int, error)         { return 2, nil }
func (f *fakeAdminRepo) CreateRole(context.Context, *domain.Role) error { return nil }
func (f *fakeAdminRepo) ListRoles(context.Context) ([]*domain.Role, error) {
	return []*domain.Role{}, nil
}
func (f *fakeAdminRepo) UpdateRole(context.Context, *domain.Role) error { return nil }
func (f *fakeAdminRepo) DeleteRole(context.Context, uuid.UUID) error    { return nil }
func (f *fakeAdminRepo) SetInbounds(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (f *fakeAdminRepo) ListInboundIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (f *fakeAdminRepo) CountInboundAccess(context.Context, uuid.UUID, []uuid.UUID) (int64, error) {
	return 0, nil
}
func (f *fakeAdminRepo) SetPlans(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (f *fakeAdminRepo) ListPlanIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (f *fakeAdminRepo) CountPlanAccess(context.Context, uuid.UUID, []uuid.UUID) (int64, error) {
	return 0, nil
}
func (f *fakeAdminRepo) SetNodes(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (f *fakeAdminRepo) ListNodeIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (f *fakeAdminRepo) CountNodeAccess(context.Context, uuid.UUID, []uuid.UUID) (int64, error) {
	return 0, nil
}

type fakeUserRepo struct {
	listed   []*domain.User
	subUser  *domain.User
	getUser  *domain.User
	inbounds []domain.Inbound
}

func (f *fakeUserRepo) Create(context.Context, *domain.User) error { return nil }
func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (*domain.User, error) {
	if f.getUser != nil {
		return f.getUser, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeUserRepo) GetBySubToken(_ context.Context, tok string) (*domain.User, error) {
	if f.subUser != nil && f.subUser.SubToken == tok {
		return f.subUser, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeUserRepo) Update(context.Context, *domain.User) error { return nil }
func (f *fakeUserRepo) Delete(context.Context, uuid.UUID) error    { return nil }
func (f *fakeUserRepo) List(context.Context, port.UserFilter) ([]*domain.User, int, error) {
	return f.listed, len(f.listed), nil
}
func (f *fakeUserRepo) AddUsedTraffic(context.Context, uuid.UUID, int64) error         { return nil }
func (f *fakeUserRepo) AddUsedTrafficBatch(context.Context, map[uuid.UUID]int64) error { return nil }
func (f *fakeUserRepo) SetInbounds(context.Context, uuid.UUID, []uuid.UUID) error      { return nil }
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

type fakeNodeRepo struct{ node *domain.Node }

func (f *fakeNodeRepo) Create(context.Context, *domain.Node) error { return nil }
func (f *fakeNodeRepo) GetByID(context.Context, uuid.UUID) (*domain.Node, error) {
	if f.node == nil {
		return nil, domain.ErrNotFound
	}
	return f.node, nil
}
func (f *fakeNodeRepo) Update(context.Context, *domain.Node) error   { return nil }
func (f *fakeNodeRepo) Delete(context.Context, uuid.UUID) error      { return nil }
func (f *fakeNodeRepo) List(context.Context) ([]*domain.Node, error) { return nil, nil }
func (f *fakeNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	return nil
}

type fakeTrafficRepo struct{ points []domain.TrafficPoint }

func (f *fakeTrafficRepo) WriteBatch(context.Context, []domain.TrafficPoint) error { return nil }
func (f *fakeTrafficRepo) UsageSeries(context.Context, uuid.UUID, port.SeriesQuery) ([]domain.TrafficPoint, error) {
	return f.points, nil
}
func (f *fakeTrafficRepo) TotalSeries(context.Context, port.SeriesQuery) ([]domain.TrafficPoint, error) {
	return f.points, nil
}

type nopNodeOps struct{}

func (nopNodeOps) AddUser(context.Context, uuid.UUID, string, *domain.User) error { return nil }
func (nopNodeOps) RemoveUser(context.Context, uuid.UUID, string, uuid.UUID) error { return nil }

func newTestServer(t *testing.T) (http.Handler, *auth.Issuer, *domain.Admin) {
	t.Helper()
	hash, _ := auth.HashPassword("pw")
	admin := &domain.Admin{ID: uuid.New(), Username: "root", PasswordHash: hash, Sudo: true}
	adminRepo := &fakeAdminRepo{admin: admin}

	// A subscribable user bound to one VLESS inbound on one node.
	nodeID := uuid.New()
	subUser := &domain.User{
		ID: uuid.New(), Username: "alice", SubToken: "tok123", DataLimit: 1 << 30, UsedTraffic: 42,
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()},
	}
	userRepo := &fakeUserRepo{
		listed:   []*domain.User{{ID: uuid.New(), Username: "u1"}},
		subUser:  subUser,
		getUser:  &domain.User{ID: uuid.New(), Username: "editable", Status: domain.UserStatusActive, DataLimit: 100},
		inbounds: []domain.Inbound{{ID: uuid.New(), NodeID: nodeID, Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws", Security: domain.SecurityTLS, Enabled: true}},
	}
	nodeRepo := &fakeNodeRepo{node: &domain.Node{ID: nodeID, Name: "de1", Address: "9.9.9.9:50051"}}

	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	authSvc := service.NewAuthService(adminRepo, iss)
	userSvc := service.NewUserService(userRepo, nopNodeOps{})
	subSvc := service.NewSubscriptionService(userRepo, nodeRepo, nil)
	adminSvc := service.NewAdminService(adminRepo, userRepo)
	trafficRepo := &fakeTrafficRepo{points: []domain.TrafficPoint{
		{Time: time.Now().Add(-24 * time.Hour), Up: 100, Down: 200},
		{Time: time.Now(), Up: 50, Down: 75},
	}}
	router := NewRouter(Deps{
		Handlers:  &Handlers{Auth: authSvc, Users: userSvc, Sub: subSvc, Admins: adminSvc, Repo: userRepo, Traffic: trafficRepo},
		Issuer:    iss,
		PanelAuth: &auth.PanelAuth{JWT: iss},
		Auth:      authSvc,
	})
	return router, iss, admin
}

func do(t *testing.T, h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// fakeLimiter denies once a key exceeds the per-call limit, like the real one.
type fakeLimiter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (f *fakeLimiter) Allow(_ context.Context, key string, limit int, _ time.Duration) (bool, time.Duration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counts == nil {
		f.counts = map[string]int{}
	}
	f.counts[key]++
	if f.counts[key] > limit {
		return false, time.Minute, nil
	}
	return true, 0, nil
}

func TestLoginRateLimited(t *testing.T) {
	hash, _ := auth.HashPassword("pw")
	admin := &domain.Admin{ID: uuid.New(), Username: "root", PasswordHash: hash, Sudo: true}
	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	authSvc := service.NewAuthService(&fakeAdminRepo{admin: admin}, iss)
	router := NewRouter(Deps{
		Handlers:  &Handlers{Auth: authSvc},
		Issuer:    iss,
		PanelAuth: &auth.PanelAuth{JWT: iss},
		Auth:      authSvc,
		Limiter:   &fakeLimiter{},
	})

	// The login route is limited to 10/min; the 11th attempt from one IP is 429.
	for i := 0; i < 10; i++ {
		if rec := do(t, router, http.MethodPost, "/api/login", "", `{"username":"root","password":"nope"}`); rec.Code == http.StatusTooManyRequests {
			t.Fatalf("throttled too early at attempt %d", i+1)
		}
	}
	rec := do(t, router, http.MethodPost, "/api/login", "", `{"username":"root","password":"nope"}`)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("11th attempt code = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("429 should carry a Retry-After header")
	}
}

func TestLoginFlow(t *testing.T) {
	h, _, _ := newTestServer(t)

	// Wrong password -> 401.
	if rec := do(t, h, http.MethodPost, "/api/login", "", `{"username":"root","password":"nope"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad login code = %d, want 401", rec.Code)
	}

	// Correct password -> 200 + token.
	rec := do(t, h, http.MethodPost, "/api/login", "", `{"username":"root","password":"pw"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var resp loginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.Token == "" {
		t.Fatalf("no token in response: %s", rec.Body)
	}
}

func TestProtectedRoutesRequireToken(t *testing.T) {
	h, _, _ := newTestServer(t)

	// No token -> 401.
	if rec := do(t, h, http.MethodGet, "/api/users", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated list code = %d, want 401", rec.Code)
	}

	// Login then list -> 200.
	login := do(t, h, http.MethodPost, "/api/login", "", `{"username":"root","password":"pw"}`)
	var resp loginResponse
	_ = json.Unmarshal(login.Body.Bytes(), &resp)

	rec := do(t, h, http.MethodGet, "/api/users", resp.Token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated list code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var out struct {
		Total int `json:"total"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out.Total != 1 {
		t.Errorf("total = %d, want 1", out.Total)
	}
}

func TestSubscribe(t *testing.T) {
	h, _, _ := newTestServer(t)

	// Unknown token must 404 (no token-existence oracle), and stay public.
	if rec := do(t, h, http.MethodGet, "/sub/nope", "", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown token code = %d, want 404", rec.Code)
	}

	// Clash UA -> YAML body + usage header.
	req := httptest.NewRequest(http.MethodGet, "/sub/tok123", nil)
	req.Header.Set("User-Agent", "clash-verge/1.0")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("subscribe code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "yaml") {
		t.Errorf("clash UA content-type = %q, want yaml", ct)
	}
	if ui := rec.Header().Get("Subscription-Userinfo"); !strings.Contains(ui, "download=42") || !strings.Contains(ui, "total=1073741824") {
		t.Errorf("usage header wrong: %q", ui)
	}
	if !strings.Contains(rec.Body.String(), "proxies") {
		t.Errorf("clash body missing proxies:\n%s", rec.Body)
	}

	// ?format override beats UA detection.
	rec2 := do(t, h, http.MethodGet, "/sub/tok123?format=base64", "", "")
	if ct := rec2.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("base64 override content-type = %q", ct)
	}
}

func TestGetUserUsage(t *testing.T) {
	h, _, _ := newTestServer(t)
	login := do(t, h, http.MethodPost, "/api/login", "", `{"username":"root","password":"pw"}`)
	var resp loginResponse
	_ = json.Unmarshal(login.Body.Bytes(), &resp)

	rec := do(t, h, http.MethodGet, "/api/users/"+uuid.New().String()+"/usage", resp.Token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("usage code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var out struct {
		Points []struct {
			Up   int64 `json:"up"`
			Down int64 `json:"down"`
		} `json:"points"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || len(out.Points) != 2 {
		t.Fatalf("expected 2 usage points, got %s", rec.Body)
	}
}

func TestUpdateUser(t *testing.T) {
	h, _, _ := newTestServer(t)
	login := do(t, h, http.MethodPost, "/api/login", "", `{"username":"root","password":"pw"}`)
	var resp loginResponse
	_ = json.Unmarshal(login.Body.Bytes(), &resp)

	id := uuid.New().String()
	rec := do(t, h, http.MethodPut, "/api/users/"+id, resp.Token, `{"data_limit":999,"note":"edited","status":"active"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "edited") {
		t.Errorf("response missing updated note: %s", rec.Body)
	}

	// Without a token the route is rejected.
	if r := do(t, h, http.MethodPut, "/api/users/"+id, "", `{}`); r.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated update code = %d, want 401", r.Code)
	}
}

type fakeDeviceLimiter struct{ allow bool }

func (f fakeDeviceLimiter) Allow(context.Context, string, string, int, time.Duration) (bool, error) {
	return f.allow, nil
}

// subRouter builds a router serving the subscription for one user, with optional
// device limiting.
func subRouter(t *testing.T, user *domain.User, devices DeviceLimiter) http.Handler {
	t.Helper()
	nodeID := uuid.New()
	userRepo := &fakeUserRepo{
		subUser:  user,
		inbounds: []domain.Inbound{{ID: uuid.New(), NodeID: nodeID, Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true}},
	}
	nodeRepo := &fakeNodeRepo{node: &domain.Node{ID: nodeID, Name: "de1", Address: "9.9.9.9:50051"}}
	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	authSvc := service.NewAuthService(&fakeAdminRepo{}, iss)
	return NewRouter(Deps{
		Handlers:  &Handlers{Sub: service.NewSubscriptionService(userRepo, nodeRepo, nil), Devices: devices},
		Issuer:    iss,
		PanelAuth: &auth.PanelAuth{JWT: iss},
		Auth:      authSvc,
	})
}

func subReq(t *testing.T, h http.Handler, deviceHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/sub/tok123", nil)
	if deviceHeader != "" {
		req.Header.Set("X-Device-Id", deviceHeader)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSubscribeHWIDAllowlist(t *testing.T) {
	user := &domain.User{ID: uuid.New(), Username: "a", SubToken: "tok123",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}, AllowedHWIDs: []string{"hwid-1"}}
	h := subRouter(t, user, nil)

	if rec := subReq(t, h, "hwid-2"); rec.Code != http.StatusForbidden {
		t.Errorf("unauthorized device code = %d, want 403", rec.Code)
	}
	if rec := subReq(t, h, "hwid-1"); rec.Code != http.StatusOK {
		t.Errorf("authorized device code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
}

func TestSubscribeDeviceLimit(t *testing.T) {
	user := &domain.User{ID: uuid.New(), Username: "a", SubToken: "tok123",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}, DeviceLimit: 2}

	// Tracker denies → 403.
	if rec := subReq(t, subRouter(t, user, fakeDeviceLimiter{allow: false}), "dev-x"); rec.Code != http.StatusForbidden {
		t.Errorf("over-limit code = %d, want 403", rec.Code)
	}
	// Tracker allows → 200.
	if rec := subReq(t, subRouter(t, user, fakeDeviceLimiter{allow: true}), "dev-x"); rec.Code != http.StatusOK {
		t.Errorf("within-limit code = %d, want 200", rec.Code)
	}
}

func TestAdminEndpoints(t *testing.T) {
	h, iss, admin := newTestServer(t)
	sudoTok, _ := iss.Issue(admin.ID, true, nil)

	// Sudo admin can list and create admins.
	if rec := do(t, h, http.MethodGet, "/api/admins", sudoTok, ""); rec.Code != http.StatusOK {
		t.Fatalf("list admins code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, h, http.MethodPost, "/api/admins", sudoTok, `{"username":"ops","password":"pw","sudo":true}`); rec.Code != http.StatusCreated {
		t.Fatalf("create admin code = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, h, http.MethodPost, "/api/admins", sudoTok, `{"username":"reseller","password":"pw"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("non-sudo without role code = %d, want 400 (%s)", rec.Code, rec.Body)
	}

	// A non-sudo admin with no role is denied (admin:manage required).
	plebTok, _ := iss.Issue(uuid.New(), false, nil)
	if rec := do(t, h, http.MethodGet, "/api/admins", plebTok, ""); rec.Code != http.StatusForbidden {
		t.Errorf("non-sudo list admins code = %d, want 403", rec.Code)
	}

	// An admin cannot delete itself.
	if rec := do(t, h, http.MethodDelete, "/api/admins/"+admin.ID.String(), sudoTok, ""); rec.Code != http.StatusBadRequest {
		t.Errorf("self-delete code = %d, want 400", rec.Code)
	}
}

func TestTOTPSelfEnrollmentEndpoints(t *testing.T) {
	h, iss, admin := newTestServer(t)
	tok, _ := iss.Issue(admin.ID, true, nil)

	// Setup returns a secret + url.
	rec := do(t, h, http.MethodPost, "/api/account/2fa/setup", tok, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("setup code = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var setup struct {
		Secret string `json:"secret"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &setup); err != nil || setup.Secret == "" {
		t.Fatalf("no secret in setup response: %s", rec.Body)
	}

	// A bad code is rejected.
	if r := do(t, h, http.MethodPost, "/api/account/2fa/confirm", tok, `{"code":"000000"}`); r.Code != http.StatusBadRequest {
		t.Errorf("bad confirm code = %d, want 400", r.Code)
	}

	// A real code activates 2FA.
	code, _ := totp.GenerateCode(setup.Secret, time.Now())
	if r := do(t, h, http.MethodPost, "/api/account/2fa/confirm", tok, `{"code":"`+code+`"}`); r.Code != http.StatusOK {
		t.Fatalf("confirm code = %d, want 200 (%s)", r.Code, r.Body)
	}

	// 2FA endpoints require authentication.
	if r := do(t, h, http.MethodPost, "/api/account/2fa/setup", "", ""); r.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated setup = %d, want 401", r.Code)
	}
}

func TestCreateUserRequiresWritePermissionButSudoPasses(t *testing.T) {
	h, _, _ := newTestServer(t)
	login := do(t, h, http.MethodPost, "/api/login", "", `{"username":"root","password":"pw"}`)
	var resp loginResponse
	_ = json.Unmarshal(login.Body.Bytes(), &resp)

	rec := do(t, h, http.MethodPost, "/api/users", resp.Token, `{"username":"alice","data_limit":1073741824}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user code = %d, want 201 (%s)", rec.Code, rec.Body)
	}
}
