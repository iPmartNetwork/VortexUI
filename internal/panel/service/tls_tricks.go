package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// TLSTricksService manages advanced TLS trick profiles.
type TLSTricksService struct {
	repo port.TLSTricksRepository
	now  func() time.Time
}

func NewTLSTricksService(repo port.TLSTricksRepository) *TLSTricksService {
	return &TLSTricksService{repo: repo, now: time.Now}
}

func (s *TLSTricksService) CreateProfile(ctx context.Context, p *domain.TLSTrickProfile) (*domain.TLSTrickProfile, error) {
	if p.Name == "" {
		return nil, errors.New("name is required")
	}
	p.ID = uuid.New()
	p.CreatedAt = s.now()
	p.Enabled = true
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *TLSTricksService) CreateFromPreset(ctx context.Context, isp domain.ISPPreset) (*domain.TLSTrickProfile, error) {
	p := domain.ISPPresetDefaults(isp)
	p.ID = uuid.New()
	p.CreatedAt = s.now()
	if err := s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *TLSTricksService) ListProfiles(ctx context.Context) ([]*domain.TLSTrickProfile, error) {
	return s.repo.List(ctx)
}

func (s *TLSTricksService) UpdateProfile(ctx context.Context, p *domain.TLSTrickProfile) (*domain.TLSTrickProfile, error) {
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *TLSTricksService) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *TLSTricksService) GetAvailablePresets() []domain.ISPPreset {
	return []domain.ISPPreset{
		domain.ISPHamrahAval, domain.ISPIrancell, domain.ISPMokhaberat,
		domain.ISPShatel, domain.ISPAsiatech, domain.ISPCustom,
	}
}
