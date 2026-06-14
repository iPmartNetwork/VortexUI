-- +goose Up
-- Audit log of mutating admin actions. admin_id has no FK so history survives
-- the admin being deleted.
CREATE TABLE audit_log (
    id       UUID PRIMARY KEY,
    time     TIMESTAMPTZ NOT NULL DEFAULT now(),
    admin_id UUID,
    method   TEXT NOT NULL,
    path     TEXT NOT NULL,
    status   INTEGER NOT NULL,
    ip       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_audit_time ON audit_log (time DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_log;
