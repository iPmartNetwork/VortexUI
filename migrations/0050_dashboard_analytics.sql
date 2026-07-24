-- +goose Up

-- ISP quality metrics: stores hourly quality scores per ISP for heatmap visualization.
CREATE TABLE isp_quality_metrics (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isp_name    TEXT NOT NULL,
    hour        INT NOT NULL CHECK (hour >= 0 AND hour <= 23),
    day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    score       FLOAT NOT NULL DEFAULT 0,
    sample_date DATE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_isp_quality_metrics_isp_date ON isp_quality_metrics (isp_name, sample_date);

-- Subscription analytics: tracks subscription fetch events for format/ISP/time analysis.
CREATE TABLE subscription_analytics (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    format_type TEXT NOT NULL,
    isp_name    TEXT NOT NULL DEFAULT '',
    client_app  TEXT NOT NULL DEFAULT '',
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Convert to TimescaleDB hypertable for efficient time-series queries.
SELECT create_hypertable('subscription_analytics', 'fetched_at', if_not_exists => TRUE);

-- Revenue entries: tracks income and expenses per admin for financial reporting.
CREATE TABLE revenue_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID NOT NULL,
    type        TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    amount      BIGINT NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_revenue_entries_admin_created ON revenue_entries (admin_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS revenue_entries;
DROP TABLE IF EXISTS subscription_analytics;
DROP TABLE IF EXISTS isp_quality_metrics;
