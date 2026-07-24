-- +goose Up

-- Client templates: match proxy client apps via regex and apply custom config.
CREATE TABLE client_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    client_pattern  TEXT NOT NULL,
    routing_rules   JSONB NOT NULL DEFAULT '[]',
    dns_settings    JSONB NOT NULL DEFAULT '{}',
    custom_outbounds JSONB NOT NULL DEFAULT '[]',
    priority        INT NOT NULL DEFAULT 0,
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Subscription approval queue: admin-reviewed subscription requests.
CREATE TABLE subscription_approval_queue (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_data JSONB NOT NULL DEFAULT '{}',
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_id     UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at  TIMESTAMPTZ
);

CREATE INDEX idx_approval_queue_pending ON subscription_approval_queue (status) WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS subscription_approval_queue;
DROP TABLE IF EXISTS client_templates;
