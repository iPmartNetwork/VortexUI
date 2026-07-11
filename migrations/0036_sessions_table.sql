-- +goose Up
CREATE TABLE IF NOT EXISTS admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL,
    token_hash VARCHAR(256) NOT NULL UNIQUE,
    ip_address INET NOT NULL,
    user_agent TEXT,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE CASCADE
);

CREATE INDEX idx_admin_sessions_admin_active ON admin_sessions(admin_id, revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_admin_sessions_expires ON admin_sessions(expires_at);
CREATE INDEX idx_admin_sessions_token ON admin_sessions(token_hash);

-- +goose Down
DROP TABLE IF EXISTS admin_sessions;
