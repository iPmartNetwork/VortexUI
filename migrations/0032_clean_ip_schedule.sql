-- +goose Up
CREATE TABLE IF NOT EXISTS clean_ip_schedule (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    interval_minutes INTEGER NOT NULL DEFAULT 360,
    port             INTEGER NOT NULL DEFAULT 443,
    ips              TEXT NOT NULL DEFAULT '',
    last_run_at      TIMESTAMPTZ,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT clean_ip_schedule_singleton CHECK (id = 1)
);
INSERT INTO clean_ip_schedule (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS clean_ip_schedule;
