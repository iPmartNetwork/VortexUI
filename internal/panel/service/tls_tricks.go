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
	repo     port.TLSTricksRepository
	inbounds inboundEvasionLinker
	now      func() time.Time
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

// inboundEvasionLinker assigns a TLS/DPI profile to live inbounds via
// evasion_profile_id so subscriptions can emit fragment/uTLS settings.
type inboundEvasionLinker interface {
	ApplyEvasionProfile(ctx context.Context, profileID uuid.UUID, inboundIDs []uuid.UUID) (int64, error)
}

// SetInboundLinker wires optional inbound assignment for ApplyToInbounds.
func (s *TLSTricksService) SetInboundLinker(l inboundEvasionLinker) { s.inbounds = l }

// ApplyToInbounds links the profile to the given inbound IDs (or every enabled
// inbound when inboundIDs is empty) and returns how many rows were updated.
func (s *TLSTricksService) ApplyToInbounds(ctx context.Context, profileID uuid.UUID, inboundIDs []uuid.UUID) (int64, error) {
	p, err := s.repo.GetByID(ctx, profileID)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, domain.ErrNotFound
	}
	if s.inbounds == nil {
		return 0, errors.New("inbound linker not configured")
	}
	return s.inbounds.ApplyEvasionProfile(ctx, profileID, inboundIDs)
}
