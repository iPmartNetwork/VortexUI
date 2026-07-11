package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

// TokenResolver maps a hashed API token to admin auth context.
type TokenResolver interface {
	Resolve(ctx context.Context, hash string) (adminID uuid.UUID, sudo bool, roleID *uuid.UUID, tokenID uuid.UUID, err error)
	Touch(ctx context.Context, tokenID uuid.UUID) error
}

// PanelAuth verifies JWTs and vtx_ API tokens.
type PanelAuth struct {
	JWT    *Issuer
	Tokens TokenResolver
}

func hashAPIToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// Verify returns claims for a bearer token (JWT or vtx_ API token).
func (a *PanelAuth) Verify(ctx context.Context, token string) (*Claims, error) {
	if strings.HasPrefix(token, "vtx_") && a.Tokens != nil {
		adminID, sudo, roleID, tokenID, err := a.Tokens.Resolve(ctx, hashAPIToken(token))
		if err != nil {
			return nil, ErrInvalidToken
		}
		if err := a.Tokens.Touch(ctx, tokenID); err != nil {
			// Log but don't fail auth - token is still valid, just update failed
			slog.Default().Debug("failed to update token timestamp", "err", err)
		}
		return &Claims{AdminID: adminID, Sudo: sudo, RoleID: roleID}, nil
	}
	return a.JWT.Verify(token)
}
