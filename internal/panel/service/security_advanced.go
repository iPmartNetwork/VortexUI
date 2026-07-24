package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SecurityAdvancedService implements IP whitelisting, login audit, account lockout,
// session management, scoped tokens, auto-ban, and security audit logging.
type SecurityAdvancedService struct {
	whitelistRepo port.IPWhitelistRepository
	banRepo       port.IPBanRepository
	sessionRepo   port.AdminSessionRepository
	loginRepo     port.LoginAuditRepository
	auditRepo     port.SecurityAuditRepository

	// LockoutThreshold is the number of consecutive failures before lockout.
	LockoutThreshold int
	// LockoutDuration is how long an account stays locked.
	LockoutDuration time.Duration
	// LockoutWindow is the time window for counting failures.
	LockoutWindow time.Duration
	// AutoBanThreshold is the number of failed logins from one IP to trigger a ban.
	AutoBanThreshold int
	// AutoBanDuration is how long the IP stays banned.
	AutoBanDuration time.Duration
}

// NewSecurityAdvancedService wires the security service with default thresholds.
func NewSecurityAdvancedService(
	whitelistRepo port.IPWhitelistRepository,
	banRepo port.IPBanRepository,
	sessionRepo port.AdminSessionRepository,
	loginRepo port.LoginAuditRepository,
	auditRepo port.SecurityAuditRepository,
) *SecurityAdvancedService {
	return &SecurityAdvancedService{
		whitelistRepo:    whitelistRepo,
		banRepo:          banRepo,
		sessionRepo:      sessionRepo,
		loginRepo:        loginRepo,
		auditRepo:        auditRepo,
		LockoutThreshold: 5,
		LockoutDuration:  15 * time.Minute,
		LockoutWindow:    30 * time.Minute,
		AutoBanThreshold: 10,
		AutoBanDuration:  1 * time.Hour,
	}
}

// --- IP Whitelist ---

// CheckIPAllowed verifies if the given IP is allowed for the admin.
// If no whitelist entries exist, all IPs are allowed.
// Matches against CIDR ranges.
func (s *SecurityAdvancedService) CheckIPAllowed(ctx context.Context, adminID uuid.UUID, ip string) (bool, error) {
	entries, err := s.whitelistRepo.ListByAdmin(ctx, adminID)
	if err != nil {
		return false, err
	}
	// No whitelist means no restriction.
	if len(entries) == 0 {
		return true, nil
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, nil
	}

	for _, entry := range entries {
		_, network, err := net.ParseCIDR(entry.CIDR)
		if err != nil {
			// Try as single IP
			if net.ParseIP(entry.CIDR) != nil && entry.CIDR == ip {
				return true, nil
			}
			continue
		}
		if network.Contains(parsedIP) {
			return true, nil
		}
	}
	return false, nil
}

// AddWhitelist adds a new IP whitelist entry.
func (s *SecurityAdvancedService) AddWhitelist(ctx context.Context, adminID *uuid.UUID, cidr, description string) (*domain.AdminIPWhitelist, error) {
	entry := &domain.AdminIPWhitelist{
		ID:          uuid.New(),
		AdminID:     adminID,
		CIDR:        cidr,
		Description: description,
		CreatedAt:   time.Now(),
	}
	if err := s.whitelistRepo.Create(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// RemoveWhitelist removes a whitelist entry.
func (s *SecurityAdvancedService) RemoveWhitelist(ctx context.Context, id uuid.UUID) error {
	return s.whitelistRepo.Delete(ctx, id)
}

// ListWhitelist returns all whitelist entries.
func (s *SecurityAdvancedService) ListWhitelist(ctx context.Context) ([]*domain.AdminIPWhitelist, error) {
	return s.whitelistRepo.ListAll(ctx)
}

// --- Login Audit ---

// RecordLogin records a login attempt and checks for auto-ban conditions.
func (s *SecurityAdvancedService) RecordLogin(ctx context.Context, entry *domain.LoginAuditEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if err := s.loginRepo.Create(ctx, entry); err != nil {
		return err
	}

	// Check auto-ban on failure
	if !entry.Success {
		count, err := s.loginRepo.CountFailedSince(ctx, uuid.Nil, time.Now().Add(-s.LockoutWindow))
		if err == nil && count >= s.AutoBanThreshold {
			_ = s.AutoBan(ctx, entry.IPAddress, "too many failed login attempts")
		}
	}

	return nil
}

// ListLoginAudit returns recent login audit entries.
func (s *SecurityAdvancedService) ListLoginAudit(ctx context.Context, limit int) ([]*domain.LoginAuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.loginRepo.ListRecent(ctx, limit)
}

// --- Account Lockout ---

// CheckAccountLocked returns true if the admin account is currently locked.
func (s *SecurityAdvancedService) CheckAccountLocked(ctx context.Context, adminID uuid.UUID) (bool, error) {
	since := time.Now().Add(-s.LockoutWindow)
	count, err := s.loginRepo.CountFailedSince(ctx, adminID, since)
	if err != nil {
		return false, err
	}
	return count >= s.LockoutThreshold, nil
}

// --- Session Management ---

// CreateSession creates a new admin session.
func (s *SecurityAdvancedService) CreateSession(ctx context.Context, adminID uuid.UUID, ip, userAgent, country string) (*domain.AdminSession, error) {
	session := &domain.AdminSession{
		ID:        uuid.New(),
		AdminID:   adminID,
		IPAddress: ip,
		UserAgent: userAgent,
		Country:   country,
		CreatedAt: time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

// ListSessions returns active sessions for an admin.
func (s *SecurityAdvancedService) ListSessions(ctx context.Context, adminID uuid.UUID) ([]*domain.AdminSession, error) {
	return s.sessionRepo.ListByAdmin(ctx, adminID)
}

// RevokeSession revokes a single session.
func (s *SecurityAdvancedService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.Revoke(ctx, sessionID)
}

// RevokeAllSessions revokes all sessions for an admin.
func (s *SecurityAdvancedService) RevokeAllSessions(ctx context.Context, adminID uuid.UUID) error {
	return s.sessionRepo.RevokeAllForAdmin(ctx, adminID)
}

// TouchSession updates the last_active timestamp.
func (s *SecurityAdvancedService) TouchSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.UpdateLastActive(ctx, sessionID)
}

// --- Token Scopes ---

// CheckTokenScope verifies that a token has the required scope.
func (s *SecurityAdvancedService) CheckTokenScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == domain.ScopeAll || scope == required {
			return true
		}
	}
	return false
}

// --- Auto-Ban ---

// AutoBan adds a temporary IP ban.
func (s *SecurityAdvancedService) AutoBan(ctx context.Context, ip, reason string) error {
	expiresAt := time.Now().Add(s.AutoBanDuration)
	ban := &domain.IPBan{
		ID:        uuid.New(),
		IPAddress: ip,
		Reason:    reason,
		ExpiresAt: &expiresAt,
		CreatedAt: time.Now(),
	}
	return s.banRepo.Create(ctx, ban)
}

// CheckIPBanned returns whether the IP is currently banned.
func (s *SecurityAdvancedService) CheckIPBanned(ctx context.Context, ip string) (bool, string, error) {
	ban, err := s.banRepo.GetByIP(ctx, ip)
	if err != nil {
		// Not found means not banned.
		return false, "", nil
	}
	return true, ban.Reason, nil
}

// ListBans returns all active IP bans.
func (s *SecurityAdvancedService) ListBans(ctx context.Context) ([]*domain.IPBan, error) {
	return s.banRepo.ListActive(ctx)
}

// RemoveBan removes an IP ban.
func (s *SecurityAdvancedService) RemoveBan(ctx context.Context, id uuid.UUID) error {
	return s.banRepo.Delete(ctx, id)
}

// --- Security Audit ---

// RecordSecurityAudit logs a sensitive operation.
func (s *SecurityAdvancedService) RecordSecurityAudit(ctx context.Context, adminID *uuid.UUID, operation, resource, ip string, before, after map[string]any) error {
	entry := &domain.SecurityAuditEntry{
		ID:          uuid.New(),
		AdminID:     adminID,
		Operation:   operation,
		Resource:    resource,
		BeforeState: before,
		AfterState:  after,
		IPAddress:   ip,
		CreatedAt:   time.Now(),
	}
	return s.auditRepo.Create(ctx, entry)
}

// ListSecurityAudit returns recent security audit entries.
func (s *SecurityAdvancedService) ListSecurityAudit(ctx context.Context, limit int) ([]*domain.SecurityAuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.auditRepo.ListRecent(ctx, limit)
}

// --- Utility ---

// CheckIPAccess is the combined check: not banned + whitelisted (if whitelist exists).
func (s *SecurityAdvancedService) CheckIPAccess(ctx context.Context, adminID uuid.UUID, ip string) (bool, string, error) {
	// Check ban first
	banned, reason, err := s.CheckIPBanned(ctx, ip)
	if err != nil {
		return false, "", fmt.Errorf("check ban: %w", err)
	}
	if banned {
		return false, "IP banned: " + reason, nil
	}

	// Check whitelist
	allowed, err := s.CheckIPAllowed(ctx, adminID, ip)
	if err != nil {
		return false, "", fmt.Errorf("check whitelist: %w", err)
	}
	if !allowed {
		return false, "IP not in whitelist", nil
	}

	return true, "", nil
}
