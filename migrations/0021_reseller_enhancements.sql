-- +goose Up
-- Reseller enhancements: plan/node allowlists, traffic quota mode, quota alerts.

ALTER TABLE admins
    ADD COLUMN IF NOT EXISTS traffic_quota_mode TEXT NOT NULL DEFAULT 'allocated';

CREATE TABLE IF NOT EXISTS admin_plans (
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    plan_id  UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    PRIMARY KEY (admin_id, plan_id)
);

CREATE INDEX IF NOT EXISTS idx_admin_plans_plan ON admin_plans (plan_id);

CREATE TABLE IF NOT EXISTS admin_nodes (
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    node_id  UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    PRIMARY KEY (admin_id, node_id)
);

CREATE INDEX IF NOT EXISTS idx_admin_nodes_node ON admin_nodes (node_id);

CREATE TABLE IF NOT EXISTS admin_quota_notify_config (
    id                 INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled            BOOLEAN NOT NULL DEFAULT FALSE,
    threshold_pct      INT[] NOT NULL DEFAULT '{80,90,100}',
    notify_telegram    BOOLEAN NOT NULL DEFAULT TRUE,
    webhook_url        TEXT NOT NULL DEFAULT '',
    cooldown_minutes   INT NOT NULL DEFAULT 1440
);

INSERT INTO admin_quota_notify_config (id) VALUES (1) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS admin_quota_notify_events (
    id          UUID PRIMARY KEY,
    admin_id    UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    threshold   INT NOT NULL,
    metric      TEXT NOT NULL,
    usage_pct   INT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_admin_quota_notify_admin
    ON admin_quota_notify_events (admin_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_admin_quota_notify_admin;
DROP TABLE IF EXISTS admin_quota_notify_events;
DROP TABLE IF EXISTS admin_quota_notify_config;
DROP INDEX IF EXISTS idx_admin_nodes_node;
DROP TABLE IF EXISTS admin_nodes;
DROP INDEX IF EXISTS idx_admin_plans_plan;
DROP TABLE IF EXISTS admin_plans;
ALTER TABLE admins DROP COLUMN IF EXISTS traffic_quota_mode;
