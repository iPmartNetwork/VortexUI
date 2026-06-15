package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// APITokenRepo is the storage interface for API tokens.
type APITokenRepo interface {
	Insert(ctx context.Context, id uuid.UUID, name, hash string, adminID uuid.UUID) error
	List(ctx context.Context) ([]domain.APIToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// APITokenService manages personal access tokens for API automation.
type APITokenService struct {
	repo APITokenRepo
}

// NewAPITokenService wires the service.
func NewAPITokenService(repo APITokenRepo) *APITokenService {
	return &APITokenService{repo: repo}
}

// CreateResult is the one-time response when a token is created (raw secret shown once).
type CreateResult struct {
	Token domain.APIToken `json:"token"`
	Raw   string          `json:"raw"` // only shown once
}

// Create generates a new API token bound to the given admin. The raw token is
// returned once; only its SHA-256 hash is persisted.
func (s *APITokenService) Create(ctx context.Context, adminID uuid.UUID, name string) (*CreateResult, error) {
	if name == "" {
		return nil, fmt.Errorf("token name is required")
	}
	raw, err := generateRawToken()
	if err != nil {
		return nil, err
	}
	hash := hashToken(raw)
	id := uuid.New()
	if err := s.repo.Insert(ctx, id, name, hash, adminID); err != nil {
		return nil, err
	}
	return &CreateResult{
		Token: domain.APIToken{ID: id, Name: name, AdminID: adminID},
		Raw:   raw,
	}, nil
}

// List returns all tokens (without secrets).
func (s *APITokenService) List(ctx context.Context) ([]domain.APIToken, error) {
	return s.repo.List(ctx)
}

// Delete removes a token.
func (s *APITokenService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func generateRawToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "vtx_" + hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
