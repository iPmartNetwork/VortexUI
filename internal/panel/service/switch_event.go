package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SwitchEventService records and queries protocol switch events.
type SwitchEventService struct {
	repo port.SwitchEventRepository
	pub  events.Publisher
}

func NewSwitchEventService(repo port.SwitchEventRepository) *SwitchEventService {
	return &SwitchEventService{repo: repo, pub: events.Nop{}}
}

func (s *SwitchEventService) SetPublisher(p events.Publisher) {
	if p != nil {
		s.pub = p
	}
}

// Record validates and persists a switch event, then publishes a notification.
func (s *SwitchEventService) Record(ctx context.Context, e *domain.SwitchEvent) error {
	if e.UserID == (uuid.UUID{}) {
		return errors.New("user_id is required")
	}
	if e.NodeID == (uuid.UUID{}) {
		return errors.New("node_id is required")
	}
	if e.SourceProtocol == "" || e.TargetProtocol == "" {
		return errors.New("source_protocol and target_protocol are required")
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	if err := s.repo.Record(ctx, e); err != nil {
		return err
	}
	// Publish event for notification dispatch.
	s.pub.Publish(events.Event{
		Type:    events.ProtocolSwitch,
		UserID:  e.UserID.String(),
		NodeID:  e.NodeID.String(),
		Message: "Protocol switch: " + e.SourceProtocol + " \u2192 " + e.TargetProtocol,
		Data: map[string]any{
			"source_protocol": e.SourceProtocol,
			"target_protocol": e.TargetProtocol,
			"isp":             e.ISP,
		},
	})
	return nil
}

// Summary returns aggregated switch event data for a time window.
func (s *SwitchEventService) Summary(ctx context.Context, filter domain.SwitchEventFilter) (*domain.SwitchSummary, error) {
	if filter.FromTime.IsZero() {
		filter.FromTime = time.Now().Add(-24 * time.Hour)
	}
	if filter.ToTime.IsZero() {
		filter.ToTime = time.Now()
	}
	return s.repo.Summary(ctx, filter)
}
