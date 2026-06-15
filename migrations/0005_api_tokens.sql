-- +goose Up
-- Long-lived API tokens for automation. The raw token is shown once at creation;
-- only its SHA-256 hash is stored. A token authenticates as its owning admin, so
-- it inherits that admin's RBAC permissions.
CREATE TABLE api_tokens (
    id           UUID PRIMARY KEY,
    name         TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    admin_id     UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_api_tokens_admin ON api_tokens (admin_id);

-- +goose Down
DROP TABLE IF EXISTS api_tokens;
