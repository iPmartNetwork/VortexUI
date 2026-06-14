-- +goose Up
-- +goose StatementBegin

-- Roles: named permission bundles for non-sudo admins (RBAC).
CREATE TABLE roles (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '[]'
);

-- Admins: panel operators (distinct from service users).
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

-- Nodes: remote proxy servers managed over gRPC.
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

-- Inbounds: listening endpoints on a node.
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

-- Users: the central, core-agnostic identity. Shared credentials are reused
-- across every inbound this user is bound to.
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

CREATE INDEX idx_users_status ON users(status);

-- user_inbounds: the many-to-many join that makes the model user-centric.
CREATE TABLE user_inbounds (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inbound_id UUID NOT NULL REFERENCES inbounds(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, inbound_id)
);

-- traffic_points: per-user time series. Converted to a TimescaleDB hypertable
-- below (in a statement sqlc ignores) for efficient time-bucketed reads.
CREATE TABLE traffic_points (
    time    TIMESTAMPTZ NOT NULL,
    user_id UUID NOT NULL,
    node_id UUID NOT NULL,
    up      BIGINT NOT NULL,
    down    BIGINT NOT NULL
);

CREATE INDEX idx_traffic_user_time ON traffic_points(user_id, time DESC);

-- +goose StatementEnd

-- TimescaleDB-specific; isolated so sqlc's static analyzer skips it. Falls back
-- to a plain table + index if the extension is unavailable.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'timescaledb') THEN
        CREATE EXTENSION IF NOT EXISTS timescaledb;
        PERFORM create_hypertable('traffic_points', 'time', if_not_exists => TRUE);
    END IF;
END$$;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS traffic_points;
DROP TABLE IF EXISTS user_inbounds;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS inbounds;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS roles;
