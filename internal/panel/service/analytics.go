package service

import (
	"context"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// AnalyticsService provides advanced analytics data.
type AnalyticsService struct {
	repo port.AnalyticsRepository
}

// NewAnalyticsService wires the analytics service.
func NewAnalyticsService(repo port.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// GetOverview returns a full analytics overview for the given time range.
func (s *AnalyticsService) GetOverview(ctx context.Context, from, to time.Time) (*domain.AnalyticsOverview, error) {
	q := port.SeriesQuery{
		FromUnix: from.Unix(),
		ToUnix:   to.Unix(),
		Bucket:   "1h",
	}

	// Graceful: if any query fails, return empty data for that section
	// rather than failing the entire endpoint.
	geo, _ := s.repo.GeoBreakdown(ctx, q)
	top, _ := s.repo.TopUsers(ctx, 20)
	peaks, _ := s.repo.PeakHours(ctx, q)
	totalUp, totalDown, _ := s.repo.TotalTraffic(ctx, q)

	return &domain.AnalyticsOverview{
		GeoBreakdown: geo,
		TopUsers:     top,
		PeakHours:    peaks,
		TotalUp:      totalUp,
		TotalDown:    totalDown,
	}, nil
}
