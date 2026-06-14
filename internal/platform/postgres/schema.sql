-- Schema mirror for sqlc's static type inference only. The runtime source of
-- truth is migrations/ (which also adds TimescaleDB hypertables). Keep the table
-- shapes here in sync with migrations/0001_init.sql.

CREATE TABLE roles (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE admins (
    id            UUID PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    sudo          BOOLEAN NOT NULL DEFAULT FALSE,
    role_id       UUID REFERENCES roles(id) ON DELETE SET NULL,
    totp_secret   TEXT NOT NULL DEFAULT '',
    totp_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    user_quota    INTEGER NOT NULL DEFAULT 0,
    traffic_quota BIGINT  NOT NULL DEFAULT 0,
    last_login    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE nodes (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    address       TEXT NOT NULL,
    core          TEXT NOT NULL DEFAULT 'xray',
    status        TEXT NOT NULL DEFAULT 'disconnected',
    usage_ratio   DOUBLE PRECISION NOT NULL DEFAULT 1,
    last_seen     TIMESTAMPTZ,
    cpu_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_percent   DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    core_running  BOOLEAN NOT NULL DEFAULT FALSE,
    connections   INTEGER NOT NULL DEFAULT 0,
    core_version  TEXT NOT NULL DEFAULT '',
    agent_version TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inbounds (
    id                 UUID PRIMARY KEY,
    node_id            UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag                TEXT NOT NULL,
    protocol           TEXT NOT NULL,
    listen             TEXT NOT NULL DEFAULT '',
    port               INTEGER NOT NULL,
    network            TEXT NOT NULL DEFAULT 'tcp',
    security           TEXT NOT NULL DEFAULT 'none',
    sni                JSONB NOT NULL DEFAULT '[]',
    path               TEXT NOT NULL DEFAULT '',
    host               JSONB NOT NULL DEFAULT '[]',
    flow               TEXT NOT NULL DEFAULT '',
    evasion_profile_id UUID,
    raw                JSONB NOT NULL DEFAULT '{}',
    enabled            BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (node_id, tag)
);

CREATE TABLE users (
    id             UUID PRIMARY KEY,
    username       TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'active',
    note           TEXT NOT NULL DEFAULT '',
    data_limit     BIGINT NOT NULL DEFAULT 0,
    used_traffic   BIGINT NOT NULL DEFAULT 0,
    expire_at      TIMESTAMPTZ,
    on_hold_expire BIGINT,
    reset_strategy TEXT NOT NULL DEFAULT 'no_reset',
    last_reset     TIMESTAMPTZ,
    device_limit   INTEGER NOT NULL DEFAULT 0,
    allowed_hwids  JSONB NOT NULL DEFAULT '[]',
    vmess_uuid     UUID NOT NULL,
    vless_uuid     UUID NOT NULL,
    trojan_pass    TEXT NOT NULL DEFAULT '',
    ss_password    TEXT NOT NULL DEFAULT '',
    ss_method      TEXT NOT NULL DEFAULT 'aes-128-gcm',
    sub_token      TEXT NOT NULL UNIQUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_inbounds (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inbound_id UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, inbound_id)
);

CREATE INDEX idx_user_inbounds_inbound ON user_inbounds (inbound_id);

CREATE TABLE traffic_points (
    time    TIMESTAMPTZ NOT NULL,
    user_id UUID NOT NULL,
    node_id UUID NOT NULL,
    up      BIGINT NOT NULL,
    down    BIGINT NOT NULL
);

CREATE TABLE audit_log (
    id       UUID PRIMARY KEY,
    time     TIMESTAMPTZ NOT NULL DEFAULT now(),
    admin_id UUID,
    method   TEXT NOT NULL,
    path     TEXT NOT NULL,
    status   INTEGER NOT NULL,
    ip       TEXT NOT NULL DEFAULT ''
);


CREATE TABLE outbounds (
    id        UUID PRIMARY KEY,
    node_id   UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag       TEXT NOT NULL,
    protocol  TEXT NOT NULL,
    address   TEXT NOT NULL DEFAULT '',
    port      INTEGER NOT NULL DEFAULT 0,
    uuid      TEXT NOT NULL DEFAULT '',
    password  TEXT NOT NULL DEFAULT '',
    username  TEXT NOT NULL DEFAULT '',
    method    TEXT NOT NULL DEFAULT '',
    flow      TEXT NOT NULL DEFAULT '',
    network   TEXT NOT NULL DEFAULT '',
    security  TEXT NOT NULL DEFAULT '',
    sni       TEXT NOT NULL DEFAULT '',
    path      TEXT NOT NULL DEFAULT '',
    host      TEXT NOT NULL DEFAULT '',
    raw       JSONB NOT NULL DEFAULT '{}',
    enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (node_id, tag)
);

CREATE TABLE routing_rules (
    id           UUID PRIMARY KEY,
    node_id      UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    priority     INTEGER NOT NULL DEFAULT 0,
    name         TEXT NOT NULL DEFAULT '',
    inbound_tags JSONB NOT NULL DEFAULT '[]',
    domains      JSONB NOT NULL DEFAULT '[]',
    ip           JSONB NOT NULL DEFAULT '[]',
    port         TEXT NOT NULL DEFAULT '',
    protocols    JSONB NOT NULL DEFAULT '[]',
    network      TEXT NOT NULL DEFAULT '',
    outbound_tag TEXT NOT NULL DEFAULT '',
    balancer_tag TEXT NOT NULL DEFAULT '',
    enabled      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_routing_rules_node ON routing_rules (node_id, priority);

CREATE TABLE balancers (
    id             UUID PRIMARY KEY,
    node_id        UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    tag            TEXT NOT NULL,
    selectors      JSONB NOT NULL DEFAULT '[]',
    strategy       TEXT NOT NULL DEFAULT 'random',
    observe        BOOLEAN NOT NULL DEFAULT FALSE,
    probe_url      TEXT NOT NULL DEFAULT '',
    probe_interval TEXT NOT NULL DEFAULT '',
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (node_id, tag)
);
