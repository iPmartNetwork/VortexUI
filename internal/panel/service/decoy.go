package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DecoyService manages decoy/fallback website configurations.
type DecoyService struct {
	repo port.DecoySiteRepository
	now  func() time.Time
}

// NewDecoyService wires the decoy service.
func NewDecoyService(repo port.DecoySiteRepository) *DecoyService {
	return &DecoyService{repo: repo, now: time.Now}
}

// CreateDecoyInput describes a new decoy site configuration.
type CreateDecoyInput struct {
	NodeID     *uuid.UUID
	Mode       domain.DecoyMode
	TargetURL  string
	StaticHTML string
}

// CreateDecoy persists a new decoy site configuration.
func (s *DecoyService) CreateDecoy(ctx context.Context, in CreateDecoyInput) (*domain.DecoySite, error) {
	if in.Mode == "" {
		return nil, errors.New("mode is required (proxy or static)")
	}
	if in.Mode == domain.DecoyProxy && in.TargetURL == "" {
		return nil, errors.New("target_url is required for proxy mode")
	}
	if in.Mode == domain.DecoyStatic && in.StaticHTML == "" {
		return nil, errors.New("static_html is required for static mode")
	}
	d := &domain.DecoySite{
		ID:         uuid.New(),
		NodeID:     in.NodeID,
		Mode:       in.Mode,
		TargetURL:  in.TargetURL,
		StaticHTML: in.StaticHTML,
		Enabled:    true,
		CreatedAt:  s.now(),
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// ListDecoys returns all decoy configurations.
func (s *DecoyService) ListDecoys(ctx context.Context) ([]*domain.DecoySite, error) {
	return s.repo.List(ctx)
}

// UpdateDecoy updates an existing decoy configuration.
func (s *DecoyService) UpdateDecoy(ctx context.Context, id uuid.UUID, mode domain.DecoyMode, targetURL, staticHTML string, enabled bool) (*domain.DecoySite, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d.Mode = mode
	d.TargetURL = targetURL
	d.StaticHTML = staticHTML
	d.Enabled = enabled
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// DeleteDecoy removes a decoy configuration.
func (s *DecoyService) DeleteDecoy(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
