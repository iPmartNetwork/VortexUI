package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// AdminQuotaNotifyRepo persists reseller quota alert config and events.
type AdminQuotaNotifyRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func (r *AdminQuotaNotifyRepo) GetConfig(ctx context.Context) (*domain.AdminQuotaNotifyConfig, error) {
	row, err := r.q.GetAdminQuotaNotifyConfig(ctx)
	if err != nil {
		return nil, mapErr(err)
	}
	return &domain.AdminQuotaNotifyConfig{
		Enabled:         row.Enabled,
		NotifyTelegram:  row.NotifyTelegram,
		WebhookURL:      row.WebhookUrl,
		NotifyAtPercent: int32SliceToInts(row.ThresholdPct),
		CooldownMinutes: int(row.CooldownMinutes),
	}, nil
}

func (r *AdminQuotaNotifyRepo) SaveConfig(ctx context.Context, cfg *domain.AdminQuotaNotifyConfig) error {
	pcts := make([]int32, len(cfg.NotifyAtPercent))
	for i, p := range cfg.NotifyAtPercent {
		pcts[i] = int32(p)
	}
	if len(pcts) == 0 {
		pcts = []int32{80, 90, 100}
	}
	return r.q.UpsertAdminQuotaNotifyConfig(ctx, db.UpsertAdminQuotaNotifyConfigParams{
		Enabled:          cfg.Enabled,
		ThresholdPct:     pcts,
		NotifyTelegram:   cfg.NotifyTelegram,
		WebhookUrl:       cfg.WebhookURL,
		CooldownMinutes:  int32(cfg.CooldownMinutes),
	})
}

func (r *AdminQuotaNotifyRepo) SaveEvent(ctx context.Context, e *domain.AdminQuotaNotifyEvent) error {
	return r.q.InsertAdminQuotaNotifyEvent(ctx, db.InsertAdminQuotaNotifyEventParams{
		ID:        e.ID,
		AdminID:   e.AdminID,
		Threshold: int32(e.Threshold),
		Metric:    e.Metric,
		UsagePct:  int32(e.UsagePct),
		CreatedAt: timeToTS(e.CreatedAt),
	})
}

func (r *AdminQuotaNotifyRepo) ListEvents(ctx context.Context, limit int) ([]*domain.AdminQuotaNotifyEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.q.ListAdminQuotaNotifyEvents(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AdminQuotaNotifyEvent, len(rows))
	for i, row := range rows {
		out[i] = &domain.AdminQuotaNotifyEvent{
			ID:        row.ID,
			AdminID:   row.AdminID,
			Threshold: int(row.Threshold),
			Metric:    row.Metric,
			UsagePct:  int(row.UsagePct),
			CreatedAt: row.CreatedAt.Time,
		}
	}
	return out, nil
}

func (r *AdminQuotaNotifyRepo) LastEvent(ctx context.Context, adminID uuid.UUID, metric string, threshold int) (*domain.AdminQuotaNotifyEvent, error) {
	row, err := r.q.LastAdminQuotaNotifyEvent(ctx, db.LastAdminQuotaNotifyEventParams{
		AdminID:   adminID,
		Metric:    metric,
		Threshold: int32(threshold),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &domain.AdminQuotaNotifyEvent{
		ID:        row.ID,
		AdminID:   row.AdminID,
		Threshold: int(row.Threshold),
		Metric:    row.Metric,
		UsagePct:  int(row.UsagePct),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func int32SliceToInts(in []int32) []int {
	if len(in) == 0 {
		return nil
	}
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}
