-- +goose Up
-- +goose StatementBegin

-- Outbounds: per-node egress handlers (freedom/blackhole/dns + proxy protocols).
-- Replaces the previously-hardcoded direct/blocked pair with operator-managed
-- egress, and provides the members balancers select among.
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

-- Routing rules: per-node traffic steering, evaluated in ascending priority.
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

-- Balancers: distribute traffic across outbounds; leastPing/leastLoad drive an
-- observatory that probes members.
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

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS balancers;
DROP TABLE IF EXISTS routing_rules;
DROP TABLE IF EXISTS outbounds;
