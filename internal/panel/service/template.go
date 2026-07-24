package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// TemplateService manages user provisioning templates and bulk user creation.
type TemplateService struct {
	repo  port.UserTemplateRepository
	users port.UserRepository
	nodes NodeOps
	now   func() time.Time
}

// NewTemplateService wires the template service.
func NewTemplateService(repo port.UserTemplateRepository, users port.UserRepository, nodes NodeOps) *TemplateService {
	return &TemplateService{repo: repo, users: users, nodes: nodes, now: time.Now}
}

// Create persists a new user template after validating name uniqueness.
func (s *TemplateService) Create(ctx context.Context, t *domain.UserTemplate) error {
	if t.Name == "" {
		return errors.New("template name is required")
	}
	existing, err := s.repo.GetByName(ctx, t.Name)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("check name uniqueness: %w", err)
	}
	if existing != nil {
		return errors.New("template name already exists")
	}
	now := s.now()
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.CreatedAt = now
	t.UpdatedAt = now
	return s.repo.Create(ctx, t)
}

// Update persists changes to an existing template.
func (s *TemplateService) Update(ctx context.Context, t *domain.UserTemplate) error {
	if t.Name == "" {
		return errors.New("template name is required")
	}
	// Check name uniqueness against other templates.
	existing, err := s.repo.GetByName(ctx, t.Name)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("check name uniqueness: %w", err)
	}
	if existing != nil && existing.ID != t.ID {
		return errors.New("template name already exists")
	}
	t.UpdatedAt = s.now()
	return s.repo.Update(ctx, t)
}

// Delete removes a template by ID.
func (s *TemplateService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// Get retrieves a template by ID.
func (s *TemplateService) Get(ctx context.Context, id uuid.UUID) (*domain.UserTemplate, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns all templates.
func (s *TemplateService) List(ctx context.Context) ([]*domain.UserTemplate, error) {
	return s.repo.List(ctx)
}

// ListForAdmin returns templates accessible by a specific admin (filtered by
// AllowedAdmins).
func (s *TemplateService) ListForAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.UserTemplate, error) {
	return s.repo.ListForAdmin(ctx, adminID)
}

// BulkCreate generates `count` users from a template, applying all template
// fields and generating unique credentials for each. Returns the created users.
func (s *TemplateService) BulkCreate(ctx context.Context, templateID uuid.UUID, count int, adminID uuid.UUID) ([]*domain.User, error) {
	if count < 1 || count > 1000 {
		return nil, errors.New("count must be between 1 and 1000")
	}

	tmpl, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	// Enforce AllowedAdmins restriction.
	if tmpl.AllowedAdmins != nil {
		if !containsUUID(tmpl.AllowedAdmins, adminID) {
			return nil, errors.New("admin is not allowed to use this template")
		}
	}

	now := s.now()
	users := make([]*domain.User, 0, count)

	for i := 0; i < count; i++ {
		creds, err := newCredentials()
		if err != nil {
			return nil, fmt.Errorf("generate credentials: %w", err)
		}
		tok, err := randToken()
		if err != nil {
			return nil, fmt.Errorf("generate sub token: %w", err)
		}
		suffix, err := randSuffix(6)
		if err != nil {
			return nil, fmt.Errorf("generate username suffix: %w", err)
		}
		username := tmpl.Name + "_" + suffix

		// Determine status and expire_at based on template ExpireDuration.
		var expireAt *time.Time
		status := domain.UserStatusActive

		if tmpl.ExpireDuration == nil {
			// No expire duration: check if on_hold_expire behavior applies.
			// on_hold status means timer starts on first connection.
			status = domain.UserStatusOnHold
		} else {
			// ExpireDuration is set: compute expire_at from now.
			exp := now.Add(time.Duration(*tmpl.ExpireDuration) * time.Second)
			expireAt = &exp
		}

		aid := adminID
		u := &domain.User{
			ID:            uuid.New(),
			Username:      username,
			Status:        status,
			Note:          tmpl.Note,
			AdminID:       &aid,
			DataLimit:     tmpl.DataLimit,
			ExpireAt:      expireAt,
			DeviceLimit:   tmpl.DeviceLimit,
			ResetStrategy: tmpl.ResetStrategy,
			Proxies:       creds,
			SubToken:      tok,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.users.Create(ctx, u); err != nil {
			return users, fmt.Errorf("persist user %d: %w", i+1, err)
		}
		users = append(users, u)
	}

	return users, nil
}

// CloneUser creates a new user by copying all fields from the source user
// except: ID (new), Username (generated), SubToken (generated), UsedTraffic (0),
// CreatedAt/UpdatedAt (now). New credentials are generated.
func (s *TemplateService) CloneUser(ctx context.Context, sourceID uuid.UUID, adminID uuid.UUID) (*domain.User, error) {
	src, err := s.users.GetByID(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("get source user: %w", err)
	}

	creds, err := newCredentials()
	if err != nil {
		return nil, fmt.Errorf("generate credentials: %w", err)
	}
	tok, err := randToken()
	if err != nil {
		return nil, fmt.Errorf("generate sub token: %w", err)
	}
	suffix, err := randSuffix(6)
	if err != nil {
		return nil, fmt.Errorf("generate username suffix: %w", err)
	}
	username := src.Username + "_clone_" + suffix

	now := s.now()
	aid := adminID
	clone := &domain.User{
		ID:             uuid.New(),
		Username:       username,
		Status:         src.Status,
		Note:           src.Note,
		AdminID:        &aid,
		DataLimit:      src.DataLimit,
		UsedTraffic:    0,
		ExpireAt:       src.ExpireAt,
		OnHoldExpire:   src.OnHoldExpire,
		ResetStrategy:  src.ResetStrategy,
		DeviceLimit:    src.DeviceLimit,
		AllowedHWIDs:   src.AllowedHWIDs,
		TelegramChatID: src.TelegramChatID,
		Proxies:        creds,
		SubToken:       tok,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.users.Create(ctx, clone); err != nil {
		return nil, fmt.Errorf("persist cloned user: %w", err)
	}

	return clone, nil
}

// containsUUID checks if a UUID slice contains a specific UUID.
func containsUUID(list []uuid.UUID, id uuid.UUID) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

// randSuffix generates a random hex string of the given character length.
func randSuffix(chars int) (string, error) {
	b := make([]byte, (chars+1)/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:chars], nil
}

// isNotFound checks if an error represents a "not found" condition.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, domain.ErrNotFound)
}
