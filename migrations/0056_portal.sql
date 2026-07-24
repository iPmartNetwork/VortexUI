-- +goose Up

-- Portal push notification subscriptions (Web Push API).
CREATE TABLE portal_push_subscriptions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint    TEXT NOT NULL,
    p256dh      TEXT NOT NULL,
    auth        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, endpoint)
);

CREATE INDEX idx_push_subs_user ON portal_push_subscriptions (user_id);

-- Connection guides: per-app setup instructions.
CREATE TABLE connection_guides (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_name    TEXT NOT NULL,
    platform    TEXT NOT NULL DEFAULT '',
    icon_url    TEXT NOT NULL DEFAULT '',
    content     TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_connection_guides_enabled ON connection_guides (enabled, sort_order);

-- +goose Down
DROP TABLE IF EXISTS connection_guides;
DROP TABLE IF EXISTS portal_push_subscriptions;
