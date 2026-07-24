-- +goose Up

-- Admin IP whitelist: restrict panel access to specific IPs/CIDRs.
CREATE TABLE admin_ip_whitelist (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID REFERENCES admins(id) ON DELETE CASCADE,
    cidr        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ip_whitelist_admin ON admin_ip_whitelist (admin_id);

-- Admin sessions: track active login sessions.
CREATE TABLE admin_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    ip_address  TEXT NOT NULL,
    user_agent  TEXT NOT NULL DEFAULT '',
    country     TEXT NOT NULL DEFAULT '',
    last_active TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked     BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_admin_sessions_admin ON admin_sessions (admin_id, revoked);

-- Login audit log: every authentication attempt.
CREATE TABLE login_audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID REFERENCES admins(id) ON DELETE SET NULL,
    username    TEXT NOT NULL,
    ip_address  TEXT NOT NULL,
    user_agent  TEXT NOT NULL DEFAULT '',
    country     TEXT NOT NULL DEFAULT '',
    success     BOOLEAN NOT NULL,
    failure_reason TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_login_audit_admin ON login_audit_log (admin_id, created_at DESC);
CREATE INDEX idx_login_audit_ip ON login_audit_log (ip_address, created_at DESC);

-- Security audit log: sensitive operation tracking.
CREATE TABLE security_audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id    UUID REFERENCES admins(id) ON DELETE SET NULL,
    operation   TEXT NOT NULL,
    resource    TEXT NOT NULL DEFAULT '',
    before_state JSONB,
    after_state  JSONB,
    ip_address  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_security_audit_admin ON security_audit_log (admin_id, created_at DESC);
CREATE INDEX idx_security_audit_operation ON security_audit_log (operation, created_at DESC);

-- IP ban list: temporary or permanent bans for suspicious IPs.
CREATE TABLE ip_ban_list (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address  TEXT NOT NULL,
    reason      TEXT NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (ip_address)
);

CREATE INDEX idx_ip_ban_active ON ip_ban_list (ip_address) WHERE expires_at IS NULL OR expires_at > now();

-- Add scopes to API tokens for fine-grained access control.
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS scopes TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE api_tokens DROP COLUMN IF EXISTS scopes;
DROP TABLE IF EXISTS ip_ban_list;
DROP TABLE IF EXISTS security_audit_log;
DROP TABLE IF EXISTS login_audit_log;
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS admin_ip_whitelist;
