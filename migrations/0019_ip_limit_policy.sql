-- +goose Up
CREATE TABLE IF NOT EXISTS ip_limit_policy (
    id             INTEGER PRIMARY KEY DEFAULT 1,
    enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    action         TEXT NOT NULL DEFAULT 'warn',
    alert_cooldown INTEGER NOT NULL DEFAULT 900,
    restore_after  INTEGER NOT NULL DEFAULT 900
);
CREATE TABLE IF NOT EXISTS ip_limit_events (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    username   TEXT NOT NULL DEFAULT '',
    online_ips INTEGER NOT NULL DEFAULT 0,
    limit_val  INTEGER NOT NULL DEFAULT 0,
    action     TEXT NOT NULL DEFAULT 'warn',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ip_limit_events_time ON ip_limit_events (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS ip_limit_events;
DROP TABLE IF EXISTS ip_limit_policy;
