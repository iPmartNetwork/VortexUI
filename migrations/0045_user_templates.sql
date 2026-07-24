-- +goose Up
-- User Templates: reusable configuration templates for bulk user provisioning.

CREATE TABLE IF NOT EXISTS user_templates (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              TEXT NOT NULL UNIQUE,
    data_limit        BIGINT NOT NULL DEFAULT 0,
    expire_duration   BIGINT,              -- seconds; NULL = never expires
    device_limit      INT NOT NULL DEFAULT 0,
    reset_strategy    TEXT NOT NULL DEFAULT 'no_reset',
    note              TEXT NOT NULL DEFAULT '',
    protocol_settings JSONB NOT NULL DEFAULT '{}',
    groups            TEXT[] NOT NULL DEFAULT '{}',
    allowed_admins    UUID[] DEFAULT NULL,  -- NULL = all admins can use
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_templates_name ON user_templates(name);

-- +goose Down
DROP TABLE IF EXISTS user_templates;
