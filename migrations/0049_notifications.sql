-- +goose Up
CREATE TABLE notification_channels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,
    config      JSONB NOT NULL DEFAULT '{}',
    scope_type  TEXT NOT NULL DEFAULT 'global',
    scope_id    TEXT,
    events      TEXT[] NOT NULL DEFAULT '{}',
    template    TEXT,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE webhook_delivery_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id  UUID NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL,
    status_code INT,
    attempts    INT NOT NULL DEFAULT 0,
    next_retry  TIMESTAMPTZ,
    delivered   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_webhook_log_channel ON webhook_delivery_log(channel_id);
CREATE INDEX idx_webhook_log_pending ON webhook_delivery_log(delivered, next_retry)
    WHERE NOT delivered;

-- +goose Down
DROP TABLE IF EXISTS webhook_delivery_log;
DROP TABLE IF EXISTS notification_channels;
