package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// InboundTrafficService is a thin wrapper around the inbound traffic repository.
type InboundTrafficService struct {
	repo port.InboundTrafficRepository
}

// NewInboundTrafficService creates a new InboundTrafficService. Returns nil if
// repo is nil (feature disabled).
func NewInboundTrafficService(repo port.InboundTrafficRepository) *InboundTrafficService {
	if repo == nil {
		return nil
	}
	return &InboundTrafficService{repo: repo}
}

// GetStats returns accumulated totals and daily breakdown for an inbound.
func (s *InboundTrafficService) GetStats(ctx context.Context, inboundID uuid.UUID, days int) (*domain.InboundTrafficStats, error) {
	return s.repo.GetStats(ctx, inboundID, days)
}

// AddTraffic records incremental traffic for an inbound.
func (s *InboundTrafficService) AddTraffic(ctx context.Context, inboundID uuid.UUID, upload, download int64) error {
	return s.repo.AddTraffic(ctx, inboundID, upload, download)
}
