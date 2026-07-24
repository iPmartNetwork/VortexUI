-- +goose Up

-- Config versioning for inbound configuration change tracking and rollback.
CREATE TABLE config_versions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inbound_id  UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    version     INT NOT NULL DEFAULT 1,
    config_data JSONB NOT NULL DEFAULT '{}',
    comment     TEXT NOT NULL DEFAULT '',
    admin_id    UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (inbound_id, version)
);

CREATE INDEX idx_config_versions_inbound ON config_versions (inbound_id, version DESC);
CREATE INDEX idx_config_versions_created ON config_versions (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS config_versions;
