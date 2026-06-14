// Package service holds the panel's application logic, sitting between the HTTP
// handlers and the repository/hub ports. Handlers stay thin; persistence and
// transport stay dumb; the rules live here and are unit-tested with fakes.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ErrInvalidCredentials is returned for any failed login. It is deliberately
// generic — never revealing whether the username, password, or 2FA code was the
// problem — to avoid handing attackers an oracle.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthService authenticates admins and issues tokens.
type AuthService struct {
	admins port.AdminRepository
	issuer *auth.Issuer
	now    func() time.Time
}

// NewAuthService wires the auth service.
func NewAuthService(admins port.AdminRepository, issuer *auth.Issuer) *AuthService {
	return &AuthService{admins: admins, issuer: issuer, now: time.Now}
}

// LoginInput carries credentials. TOTPCode is only consulted when the admin has
// 2FA enabled.
type LoginInput struct {
	Username string
	Password string
	TOTPCode string
}

// Login validates credentials and returns a signed JWT. Every failure path
// returns ErrInvalidCredentials so timing/branching does not distinguish them.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (string, error) {
	admin, err := s.admins.GetByUsername(ctx, in.Username)
	if err != nil {
		// Spend a hash anyway to keep the timing of "no such user" and "wrong
		// password" indistinguishable.
		_ = auth.CheckPassword("$2a$12$........................................................", in.Password)
		return "", ErrInvalidCredentials
	}
	if !auth.CheckPassword(admin.PasswordHash, in.Password) {
		return "", ErrInvalidCredentials
	}
	if admin.TOTPEnabled && !auth.VerifyTOTP(admin.TOTPSecret, in.TOTPCode) {
		return "", ErrInvalidCredentials
	}

	token, err := s.issuer.Issue(admin.ID, admin.Sudo, admin.RoleID)
	if err != nil {
		return "", err
	}

	// Record last login best-effort; a write failure must not block sign-in.
	admin.LastLogin = ptrTime(s.now())
	_ = s.admins.Update(ctx, admin)
	return token, nil
}

// Authorize loads the role behind a token's claims so a handler can check a
// permission. Sudo admins need no role lookup.
func (s *AuthService) Authorize(ctx context.Context, c *auth.Claims, p domain.Permission) (bool, error) {
	if c.Sudo {
		return true, nil
	}
	if c.RoleID == nil {
		return false, nil
	}
	role, err := s.admins.GetRole(ctx, *c.RoleID)
	if err != nil {
		return false, err
	}
	admin := &domain.Admin{Sudo: false}
	return admin.Has(p, role), nil
}

func ptrTime(t time.Time) *time.Time { return &t }
