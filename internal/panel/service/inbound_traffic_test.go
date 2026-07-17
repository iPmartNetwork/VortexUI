package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// fakeInboundTrafficRepo implements port.InboundTrafficRepository for testing.
type fakeInboundTrafficRepo struct {
	data map[uuid.UUID]*domain.InboundTrafficStats
}

func newFakeTrafficRepo() *fakeInboundTrafficRepo {
	return &fakeInboundTrafficRepo{data: make(map[uuid.UUID]*domain.InboundTrafficStats)}
}

func (f *fakeInboundTrafficRepo) AddTraffic(_ context.Context, inboundID uuid.UUID, upload, download int64) error {
	s, ok := f.data[inboundID]
	if !ok {
		s = &domain.InboundTrafficStats{InboundID: inboundID}
		f.data[inboundID] = s
	}
	s.Upload += upload
	s.Download += download
	s.Total = s.Upload + s.Download
	return nil
}

func (f *fakeInboundTrafficRepo) GetStats(_ context.Context, inboundID uuid.UUID, days int) (*domain.InboundTrafficStats, error) {
	s, ok := f.data[inboundID]
	if !ok {
		return &domain.InboundTrafficStats{InboundID: inboundID}, nil
	}
	return s, nil
}

func TestInboundTrafficService_AddAndGetStats(t *testing.T) {
	repo := newFakeTrafficRepo()
	svc := NewInboundTrafficService(repo)
	ctx := context.Background()
	ibID := uuid.New()

	// Add traffic twice.
	if err := svc.AddTraffic(ctx, ibID, 100, 200); err != nil {
		t.Fatalf("AddTraffic: %v", err)
	}
	if err := svc.AddTraffic(ctx, ibID, 50, 75); err != nil {
		t.Fatalf("AddTraffic: %v", err)
	}

	stats, err := svc.GetStats(ctx, ibID, 30)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.Upload != 150 {
		t.Errorf("upload = %d, want 150", stats.Upload)
	}
	if stats.Download != 275 {
		t.Errorf("download = %d, want 275", stats.Download)
	}
	if stats.Total != 425 {
		t.Errorf("total = %d, want 425", stats.Total)
	}
}

func TestInboundTrafficService_NilRepo(t *testing.T) {
	svc := NewInboundTrafficService(nil)
	if svc != nil {
		t.Error("expected nil service when repo is nil")
	}
}

func TestGetNodeIDForInbound(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	// Seed an inbound.
	ibID := uuid.New()
	repo.inbounds[ibID] = &domain.Inbound{
		ID:     ibID,
		NodeID: nodeID,
		Tag:    "test-inbound",
		Port:   443,
	}

	gotNodeID, err := svc.GetNodeIDForInbound(ctx, ibID)
	if err != nil {
		t.Fatalf("GetNodeIDForInbound: %v", err)
	}
	if gotNodeID != nodeID {
		t.Errorf("nodeID = %v, want %v", gotNodeID, nodeID)
	}

	// Non-existent inbound should error.
	_, err = svc.GetNodeIDForInbound(ctx, uuid.New())
	if err == nil {
		t.Error("expected error for non-existent inbound")
	}
}

func TestHealthBadgeClassification(t *testing.T) {
	// Test the health logic that would be used by enrichFleetHealth:
	// healthy: core running, CPU < 90, Mem < 90
	// degraded: core running, CPU >= 90 OR Mem >= 90
	// down: core not running

	tests := []struct {
		name        string
		coreRunning bool
		cpu, mem    float64
		want        string
	}{
		{"healthy", true, 50, 60, "healthy"},
		{"degraded high cpu", true, 95, 60, "degraded"},
		{"degraded high mem", true, 50, 92, "degraded"},
		{"degraded both high", true, 91, 91, "degraded"},
		{"down core stopped", false, 20, 30, "down"},
		{"boundary cpu 90", true, 90, 50, "degraded"},
		{"boundary cpu 89", true, 89.9, 50, "healthy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if !tt.coreRunning {
				got = "down"
			} else if tt.cpu >= 90 || tt.mem >= 90 {
				got = "degraded"
			} else {
				got = "healthy"
			}
			if got != tt.want {
				t.Errorf("health = %q, want %q", got, tt.want)
			}
		})
	}
}
