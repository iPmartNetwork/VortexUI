package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

type FederationService struct {
	repo port.FederationRepository
	now  func() time.Time
}

func NewFederationService(repo port.FederationRepository) *FederationService {
	return &FederationService{repo: repo, now: time.Now}
}

func (s *FederationService) GetConfig(ctx context.Context) (*domain.FederationConfig, error) {
	c, err := s.repo.GetConfig(ctx)
	if err != nil {
		def := domain.DefaultFederationConfig()
		return &def, nil
	}
	return c, nil
}

func (s *FederationService) UpdateConfig(ctx context.Context, c *domain.FederationConfig) error {
	return s.repo.SaveConfig(ctx, c)
}

func (s *FederationService) AddPeer(ctx context.Context, name, endpoint, apiKey string, syncUsers, syncNodes bool) (*domain.FederationPeer, error) {
	if name == "" || endpoint == "" {
		return nil, errors.New("name and endpoint are required")
	}
	p := &domain.FederationPeer{
		ID:        uuid.New(),
		Name:      name,
		Endpoint:  endpoint,
		APIKey:    apiKey,
		Status:    domain.PeerDisconnected,
		SyncUsers: syncUsers,
		SyncNodes: syncNodes,
		CreatedAt: s.now(),
	}
	if err := s.repo.CreatePeer(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *FederationService) ListPeers(ctx context.Context) ([]*domain.FederationPeer, error) {
	return s.repo.ListPeers(ctx)
}

func (s *FederationService) DeletePeer(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePeer(ctx, id)
}

func (s *FederationService) ListSyncEvents(ctx context.Context, limit int) ([]*domain.FederationSyncEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListSyncEvents(ctx, limit)
}
