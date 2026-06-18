-- +goose Up

-- The v1.2.0 feature set added a number of tables to the schema mirror
-- (internal/platform/postgres/schema.sql) but only a subset of them were
-- carried into the runtime migrations (0010 covered SNI/TLS-tricks/fingerprint/
-- federation/deeplink/quota-notify). This migration creates the remaining
-- v1.2.0 tables so the portal-ticket, reality-scan, quota-policy, relay-chain,
-- decoy-site, analytics-geo, node-migration, probing, family, referral and
-- DoH features have their backing tables at runtime. Definitions mirror
-- schema.sql verbatim.

-- Self-service portal tickets
CREATE TABLE IF NOT EXISTS tickets (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject    TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'open',
    priority   TEXT NOT NULL DEFAULT 'medium',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tickets_user ON tickets (user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets (status);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id         UUID PRIMARY KEY,
    ticket_id  UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    sender     TEXT NOT NULL,
    sender_id  UUID NOT NULL,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket ON ticket_messages (ticket_id);

-- Reality Scanner results cache
CREATE TABLE IF NOT EXISTS reality_scans (
    id         UUID PRIMARY KEY,
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    sni        TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    score      INTEGER NOT NULL DEFAULT 0,
    valid      BOOLEAN NOT NULL DEFAULT TRUE,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reality_scans_node ON reality_scans (node_id, score DESC);

-- Smart Quota tiers
CREATE TABLE IF NOT EXISTS quota_policies (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    tiers           JSONB NOT NULL DEFAULT '[]',
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- CDN/Relay chains
CREATE TABLE IF NOT EXISTS relay_chains (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    hops       JSONB NOT NULL DEFAULT '[]',
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_relay_chains_node ON relay_chains (node_id);

-- Decoy website configuration
CREATE TABLE IF NOT EXISTS decoy_sites (
    id          UUID PRIMARY KEY,
    node_id     UUID REFERENCES nodes(id) ON DELETE CASCADE,
    mode        TEXT NOT NULL DEFAULT 'proxy',
    target_url  TEXT NOT NULL DEFAULT '',
    static_html TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Analytics geo breakdown cache
CREATE TABLE IF NOT EXISTS traffic_geo (
    time       TIMESTAMPTZ NOT NULL,
    node_id    UUID NOT NULL,
    country    TEXT NOT NULL DEFAULT '',
    connections INTEGER NOT NULL DEFAULT 0,
    bytes_up   BIGINT NOT NULL DEFAULT 0,
    bytes_down BIGINT NOT NULL DEFAULT 0
);

-- Node auto-migration
CREATE TABLE IF NOT EXISTS migration_events (
    id           UUID PRIMARY KEY,
    user_id      UUID,
    username     TEXT NOT NULL DEFAULT '',
    from_node_id UUID NOT NULL,
    to_node_id   UUID NOT NULL,
    reason       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending',
    error        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS migration_policy (
    id                    INTEGER PRIMARY KEY DEFAULT 1,
    enabled               BOOLEAN NOT NULL DEFAULT FALSE,
    health_check_interval INTEGER NOT NULL DEFAULT 30,
    unhealthy_threshold   INTEGER NOT NULL DEFAULT 3,
    cpu_threshold         DOUBLE PRECISION NOT NULL DEFAULT 90,
    mem_threshold         DOUBLE PRECISION NOT NULL DEFAULT 90,
    packet_loss_max       DOUBLE PRECISION NOT NULL DEFAULT 10,
    migrate_back          BOOLEAN NOT NULL DEFAULT TRUE
);

-- Active probing protection
CREATE TABLE IF NOT EXISTS probing_policy (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    action           TEXT NOT NULL DEFAULT 'block',
    block_duration   INTEGER NOT NULL DEFAULT 3600,
    max_probe_per_min INTEGER NOT NULL DEFAULT 5,
    whitelisted_ips  JSONB NOT NULL DEFAULT '[]',
    honeypot_html    TEXT NOT NULL DEFAULT '',
    notify_telegram  BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS probe_events (
    id          UUID PRIMARY KEY,
    source_ip   TEXT NOT NULL,
    port        INTEGER NOT NULL DEFAULT 0,
    method      TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'log',
    node_id     UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_probe_events_ip ON probe_events (source_ip);

CREATE TABLE IF NOT EXISTS blocked_ips (
    ip         TEXT PRIMARY KEY,
    reason     TEXT NOT NULL DEFAULT '',
    blocked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- Family/Group subscriptions
CREATE TABLE IF NOT EXISTS family_groups (
    id           UUID PRIMARY KEY,
    name         TEXT NOT NULL,
    owner_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data_limit   BIGINT NOT NULL DEFAULT 0,
    used_traffic BIGINT NOT NULL DEFAULT 0,
    max_members  INTEGER NOT NULL DEFAULT 5,
    member_quota BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_family_groups_owner ON family_groups (owner_id);

CREATE TABLE IF NOT EXISTS family_members (
    id           UUID PRIMARY KEY,
    group_id     UUID NOT NULL REFERENCES family_groups(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_traffic BIGINT NOT NULL DEFAULT 0,
    label        TEXT NOT NULL DEFAULT '',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (group_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_family_members_group ON family_members (group_id);

-- Invite/Referral system
CREATE TABLE IF NOT EXISTS referral_config (
    id            INTEGER PRIMARY KEY DEFAULT 1,
    enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    reward_type   TEXT NOT NULL DEFAULT 'data',
    reward_amount BIGINT NOT NULL DEFAULT 1073741824,
    max_referrals INTEGER NOT NULL DEFAULT 0,
    require_paid  BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS referral_codes (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT NOT NULL UNIQUE,
    uses       INTEGER NOT NULL DEFAULT 0,
    max_uses   INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_referral_codes_user ON referral_codes (user_id);

CREATE TABLE IF NOT EXISTS referral_events (
    id             UUID PRIMARY KEY,
    referrer_id    UUID NOT NULL,
    referred_id    UUID NOT NULL,
    code_used      TEXT NOT NULL,
    reward_type    TEXT NOT NULL DEFAULT 'data',
    reward_amount  BIGINT NOT NULL DEFAULT 0,
    reward_applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- DNS-over-HTTPS
CREATE TABLE IF NOT EXISTS doh_config (
    id               INTEGER PRIMARY KEY DEFAULT 1,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    listen_addr      TEXT NOT NULL DEFAULT ':8053',
    upstream_dns     JSONB NOT NULL DEFAULT '["1.1.1.1","8.8.8.8"]',
    block_ads        BOOLEAN NOT NULL DEFAULT FALSE,
    block_malware    BOOLEAN NOT NULL DEFAULT TRUE,
    custom_blocklist JSONB NOT NULL DEFAULT '[]',
    log_queries      BOOLEAN NOT NULL DEFAULT FALSE,
    cache_ttl        INTEGER NOT NULL DEFAULT 300
);

CREATE TABLE IF NOT EXISTS doh_query_logs (
    domain     TEXT NOT NULL,
    type       TEXT NOT NULL DEFAULT 'A',
    client_ip  TEXT NOT NULL DEFAULT '',
    blocked    BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_doh_query_logs_time ON doh_query_logs (timestamp DESC);

-- +goose Down
DROP TABLE IF EXISTS doh_query_logs;
DROP TABLE IF EXISTS doh_config;
DROP TABLE IF EXISTS referral_events;
DROP TABLE IF EXISTS referral_codes;
DROP TABLE IF EXISTS referral_config;
DROP TABLE IF EXISTS family_members;
DROP TABLE IF EXISTS family_groups;
DROP TABLE IF EXISTS blocked_ips;
DROP TABLE IF EXISTS probe_events;
DROP TABLE IF EXISTS probing_policy;
DROP TABLE IF EXISTS migration_policy;
DROP TABLE IF EXISTS migration_events;
DROP TABLE IF EXISTS traffic_geo;
DROP TABLE IF EXISTS decoy_sites;
DROP TABLE IF EXISTS relay_chains;
DROP TABLE IF EXISTS quota_policies;
DROP TABLE IF EXISTS reality_scans;
DROP TABLE IF EXISTS ticket_messages;
DROP TABLE IF EXISTS tickets;
