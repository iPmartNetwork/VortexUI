-- +goose Up
-- Auto-Protocol Switching & Dynamic Port Hopping: protocol groups, ISP profiles,
-- switch events, and hop_interval for inbounds.

CREATE TABLE IF NOT EXISTS protocol_groups (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id        UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    inbound_ids    JSONB NOT NULL DEFAULT '[]',
    probe_url      TEXT NOT NULL DEFAULT 'https://www.gstatic.com/generate_204',
    probe_interval INTEGER NOT NULL DEFAULT 60,
    probe_timeout  INTEGER NOT NULL DEFAULT 5,
    max_retries    INTEGER NOT NULL DEFAULT 3,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_protocol_groups_node_id ON protocol_groups(node_id);

CREATE TABLE IF NOT EXISTS isp_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id            UUID NOT NULL REFERENCES protocol_groups(id) ON DELETE CASCADE,
    isp_identifier      TEXT NOT NULL,
    country_code        CHAR(2) NOT NULL DEFAULT '',
    preferred_protocols JSONB NOT NULL DEFAULT '[]',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(group_id, isp_identifier)
);

CREATE INDEX IF NOT EXISTS idx_isp_profiles_group_id ON isp_profiles(group_id);
CREATE INDEX IF NOT EXISTS idx_isp_profiles_isp ON isp_profiles(isp_identifier);

CREATE TABLE IF NOT EXISTS switch_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    node_id         UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    group_id        UUID REFERENCES protocol_groups(id) ON DELETE SET NULL,
    source_protocol TEXT NOT NULL,
    target_protocol TEXT NOT NULL,
    isp             TEXT NOT NULL DEFAULT '',
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_switch_events_user_id ON switch_events(user_id);
CREATE INDEX IF NOT EXISTS idx_switch_events_node_id ON switch_events(node_id);
CREATE INDEX IF NOT EXISTS idx_switch_events_timestamp ON switch_events(timestamp DESC);

ALTER TABLE inbounds
    ADD COLUMN IF NOT EXISTS hop_interval INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE inbounds DROP COLUMN IF EXISTS hop_interval;
DROP TABLE IF EXISTS switch_events;
DROP TABLE IF EXISTS isp_profiles;
DROP TABLE IF EXISTS protocol_groups;
