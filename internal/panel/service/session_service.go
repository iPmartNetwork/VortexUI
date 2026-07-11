package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SessionService manages admin sessions
type SessionService struct {
	repo              port.SessionRepository
	log               *slog.Logger
	sessionDuration   time.Duration
	cleanupInterval   time.Duration
}

// NewSessionService creates a new session service
func NewSessionService(repo port.SessionRepository, log *slog.Logger) *SessionService {
	if log == nil {
		log = slog.Default()
	}
	return &SessionService{
		repo:            repo,
		log:             log,
		sessionDuration: 24 * time.Hour,
		cleanupInterval: 1 * time.Hour,
	}
}

// SetSessionDuration sets the duration of a session
func (s *SessionService) SetSessionDuration(duration time.Duration) {
	s.sessionDuration = duration
}

// CreateSession creates a new session and returns the token
func (s *SessionService) CreateSession(ctx context.Context, adminID uuid.UUID, ipAddress, userAgent string) (string, error) {
	// Generate random token
	token, err := generateToken(32)
	if err != nil {
		s.log.Error("failed to generate session token", "err", err)
		return "", err
	}

	// Hash token
	tokenHash := hashToken(token)

	// Create session
	session := &domain.AdminSession{
		AdminID:      adminID,
		TokenHash:    tokenHash,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(s.sessionDuration),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		s.log.Error("failed to create session", "err", err)
		return "", err
	}

	s.log.Info("session created", "admin_id", adminID, "ip", ipAddress)

	// Return token in format: sess_<base64-encoded-token>
	return "sess_" + token, nil
}

// ValidateSession validates a session token
func (s *SessionService) ValidateSession(ctx context.Context, token string) (*domain.AdminSession, error) {
	if len(token) < 6 || token[:5] != "sess_" {
		return nil, domain.ErrUnauthorized
	}

	// Extract raw token
	rawToken := token[5:]

	// Hash token
	tokenHash := hashToken(rawToken)

	// Get session
	session, err := s.repo.GetSession(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if !session.IsActive() {
		return nil, domain.ErrSessionExpired
	}

	// Update last activity asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.UpdateSession(ctx, session); err != nil {
			s.log.Debug("failed to update session activity", "err", err)
		}
	}()

	return session, nil
}

// RevokeSession revokes a session
func (s *SessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		s.log.Error("failed to revoke session", "err", err)
		return err
	}
	s.log.Info("session revoked", "session_id", sessionID)
	return nil
}

// RevokeAdminSessions revokes all sessions for an admin
func (s *SessionService) RevokeAdminSessions(ctx context.Context, adminID uuid.UUID) error {
	if err := s.repo.RevokeAdminSessions(ctx, adminID); err != nil {
		s.log.Error("failed to revoke admin sessions", "err", err)
		return err
	}
	s.log.Info("all admin sessions revoked", "admin_id", adminID)
	return nil
}

// ListAdminSessions lists all active sessions for an admin
func (s *SessionService) ListAdminSessions(ctx context.Context, adminID uuid.UUID) ([]*domain.AdminSession, error) {
	return s.repo.ListAdminSessions(ctx, adminID, true)
}

// CleanupExpiredSessions cleans up expired sessions (typically called by cron)
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	deleted, err := s.repo.DeleteExpiredSessions(ctx)
	if err != nil {
		s.log.Error("failed to clean up expired sessions", "err", err)
		return err
	}
	s.log.Info("cleaned up expired sessions", "deleted", deleted)
	return nil
}

// StartCleanupWorker starts a background worker for cleaning expired sessions
func (s *SessionService) StartCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := s.CleanupExpiredSessions(cleanCtx); err != nil {
				s.log.Error("cleanup worker error", "err", err)
			}
			cancel()
		}
	}
}

// generateToken generates a cryptographically secure random token
func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
