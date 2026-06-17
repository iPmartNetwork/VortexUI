-- +goose Up

-- SNI Domains
CREATE TABLE IF NOT EXISTS sni_domains (
    id          UUID PRIMARY KEY,
    inbound_id  UUID NOT NULL,
    domain      TEXT NOT NULL,
    auto_cert   BOOLEAN NOT NULL DEFAULT TRUE,
    cert_status TEXT NOT NULL DEFAULT 'pending',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sni_domains_inbound ON sni_domains (inbound_id);

-- SSL Certificates
CREATE TABLE IF NOT EXISTS ssl_certificates (
    id         UUID PRIMARY KEY,
    domain     TEXT NOT NULL,
    wildcard   BOOLEAN NOT NULL DEFAULT FALSE,
    issuer     TEXT NOT NULL DEFAULT 'letsencrypt',
    status     TEXT NOT NULL DEFAULT 'pending',
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- SNI Routes
CREATE TABLE IF NOT EXISTS sni_routes (
    id          UUID PRIMARY KEY,
    inbound_id  UUID NOT NULL,
    sni         TEXT NOT NULL,
    action      TEXT NOT NULL DEFAULT 'proxy',
    target_tag  TEXT NOT NULL DEFAULT '',
    priority    INTEGER NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sni_routes_inbound ON sni_routes (inbound_id);

-- TLS Trick Profiles
CREATE TABLE IF NOT EXISTS tls_trick_profiles (
    id                 UUID PRIMARY KEY,
    name               TEXT NOT NULL,
    description        TEXT NOT NULL DEFAULT '',
    fragment_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
    fragment_length    TEXT NOT NULL DEFAULT '',
    fragment_interval  TEXT NOT NULL DEFAULT '',
    fingerprint        TEXT NOT NULL DEFAULT 'chrome',
    mux_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    mux_protocol       TEXT NOT NULL DEFAULT '',
    enabled            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Fingerprint Policy (singleton)
CREATE TABLE IF NOT EXISTS fingerprint_policy (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    default_action  TEXT NOT NULL DEFAULT 'allow',
    log_unknown     BOOLEAN NOT NULL DEFAULT TRUE
);

-- Fingerprint Rules
CREATE TABLE IF NOT EXISTS fingerprint_rules (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    fingerprint TEXT NOT NULL DEFAULT '',
    ja3_hash    TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'allow',
    priority    INTEGER NOT NULL DEFAULT 0,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Fingerprint Events
CREATE TABLE IF NOT EXISTS fingerprint_events (
    id          UUID PRIMARY KEY,
    client_ip   TEXT NOT NULL,
    fingerprint TEXT NOT NULL DEFAULT '',
    user_agent  TEXT NOT NULL DEFAULT '',
    action      TEXT NOT NULL DEFAULT 'allow',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_fp_events_time ON fingerprint_events (created_at DESC);

-- Federation Config (singleton)
CREATE TABLE IF NOT EXISTS federation_config (
    id             INTEGER PRIMARY KEY DEFAULT 1,
    enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    cluster_name   TEXT NOT NULL DEFAULT '',
    sso_enabled    BOOLEAN NOT NULL DEFAULT FALSE,
    sync_interval  INTEGER NOT NULL DEFAULT 60
);

-- Federation Peers
CREATE TABLE IF NOT EXISTS federation_peers (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    endpoint    TEXT NOT NULL,
    api_key     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'disconnected',
    sync_users  BOOLEAN NOT NULL DEFAULT TRUE,
    sync_nodes  BOOLEAN NOT NULL DEFAULT TRUE,
    last_sync   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Federation Sync Events
CREATE TABLE IF NOT EXISTS federation_sync_events (
    id          UUID PRIMARY KEY,
    peer_name   TEXT NOT NULL,
    direction   TEXT NOT NULL DEFAULT 'push',
    entity_type TEXT NOT NULL DEFAULT '',
    count       INTEGER NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'success',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_fed_sync_time ON federation_sync_events (created_at DESC);

-- Deep Link Config (singleton)
CREATE TABLE IF NOT EXISTS deeplink_config (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    base_url        TEXT NOT NULL DEFAULT '',
    app_scheme      TEXT NOT NULL DEFAULT 'vortex',
    include_name    BOOLEAN NOT NULL DEFAULT TRUE,
    qr_logo_url     TEXT NOT NULL DEFAULT ''
);

-- Quota Notification Config (singleton)
CREATE TABLE IF NOT EXISTS quota_notify_config (
    id              INTEGER PRIMARY KEY DEFAULT 1,
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    threshold_pct   INTEGER NOT NULL DEFAULT 80,
    notify_telegram BOOLEAN NOT NULL DEFAULT TRUE,
    notify_email    BOOLEAN NOT NULL DEFAULT FALSE,
    message_template TEXT NOT NULL DEFAULT ''
);

-- Quota Notification Events
CREATE TABLE IF NOT EXISTS quota_notify_events (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    username   TEXT NOT NULL DEFAULT '',
    threshold  INTEGER NOT NULL DEFAULT 0,
    usage_pct  INTEGER NOT NULL DEFAULT 0,
    notified   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_quota_notify_time ON quota_notify_events (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS quota_notify_events;
DROP TABLE IF EXISTS quota_notify_config;
DROP TABLE IF EXISTS deeplink_config;
DROP TABLE IF EXISTS federation_sync_events;
DROP TABLE IF EXISTS federation_peers;
DROP TABLE IF EXISTS federation_config;
DROP TABLE IF EXISTS fingerprint_events;
DROP TABLE IF EXISTS fingerprint_rules;
DROP TABLE IF EXISTS fingerprint_policy;
DROP TABLE IF EXISTS tls_trick_profiles;
DROP TABLE IF EXISTS sni_routes;
DROP TABLE IF EXISTS ssl_certificates;
DROP TABLE IF EXISTS sni_domains;
