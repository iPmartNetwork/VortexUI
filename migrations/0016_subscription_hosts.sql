-- +goose Up
CREATE TABLE IF NOT EXISTS sub_hosts (
    id             UUID PRIMARY KEY,
    inbound_id     UUID NOT NULL,
    remark         TEXT NOT NULL DEFAULT '',
    address        TEXT NOT NULL DEFAULT '',
    port           INTEGER,                       -- NULL = inherit inbound port
    sni            TEXT NOT NULL DEFAULT '',
    host_header    TEXT NOT NULL DEFAULT '',
    path           TEXT NOT NULL DEFAULT '',
    alpn           TEXT NOT NULL DEFAULT '',
    fingerprint    TEXT NOT NULL DEFAULT '',
    security       TEXT NOT NULL DEFAULT 'inbound_default',
    allow_insecure BOOLEAN NOT NULL DEFAULT FALSE,
    mux_enable     BOOLEAN NOT NULL DEFAULT FALSE,
    fragment       TEXT NOT NULL DEFAULT '',
    priority       INTEGER NOT NULL DEFAULT 0,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sub_hosts_inbound ON sub_hosts (inbound_id, priority);

-- +goose Down
DROP TABLE IF EXISTS sub_hosts;
