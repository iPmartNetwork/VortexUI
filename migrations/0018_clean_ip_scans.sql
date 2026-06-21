-- +goose Up
CREATE TABLE IF NOT EXISTS clean_ip_scans (
    id         UUID PRIMARY KEY,
    ip         TEXT NOT NULL,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    loss_pct   INTEGER NOT NULL DEFAULT 0,
    score      INTEGER NOT NULL DEFAULT 0,
    reachable  BOOLEAN NOT NULL DEFAULT FALSE,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_clean_ip_score ON clean_ip_scans (score DESC);

-- +goose Down
DROP TABLE IF EXISTS clean_ip_scans;
