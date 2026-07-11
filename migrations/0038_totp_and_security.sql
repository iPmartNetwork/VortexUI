-- Migration: 0038_totp_and_security.sql
-- Purpose: Add TOTP, MFA, password policies, and IP access control tables

-- TOTP Secrets table
CREATE TABLE IF NOT EXISTS totp_secrets (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    secret VARCHAR(100) NOT NULL,
    qr_code TEXT,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(admin_id)
);

CREATE INDEX IF NOT EXISTS idx_totp_admin_id ON totp_secrets(admin_id);

-- MFA Configuration table
CREATE TABLE IF NOT EXISTS mfa_configs (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    totp_enabled BOOLEAN DEFAULT FALSE,
    email_enabled BOOLEAN DEFAULT FALSE,
    backup_codes TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(admin_id)
);

CREATE INDEX IF NOT EXISTS idx_mfa_configs_admin_id ON mfa_configs(admin_id);

-- Password Policy table
CREATE TABLE IF NOT EXISTS password_policies (
    id UUID PRIMARY KEY,
    min_length INT DEFAULT 12,
    require_uppercase BOOLEAN DEFAULT TRUE,
    require_lowercase BOOLEAN DEFAULT TRUE,
    require_numbers BOOLEAN DEFAULT TRUE,
    require_special_chars BOOLEAN DEFAULT TRUE,
    expiration_days INT DEFAULT 90,
    history_count INT DEFAULT 5,
    failed_attempts_limit INT DEFAULT 5,
    lockout_duration_mins INT DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Password History table (track old passwords)
CREATE TABLE IF NOT EXISTS password_history (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_history_admin_id ON password_history(admin_id, created_at DESC);

-- Admin Password Status table
CREATE TABLE IF NOT EXISTS admin_password_status (
    admin_id UUID PRIMARY KEY REFERENCES admin_users(id) ON DELETE CASCADE,
    last_changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '90 days',
    failed_attempts INT DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    must_change_password BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- IP Access Rules table (whitelist/blacklist)
CREATE TABLE IF NOT EXISTS ip_access_rules (
    id UUID PRIMARY KEY,
    admin_id UUID REFERENCES admin_users(id) ON DELETE CASCADE,
    rule_type VARCHAR(20) NOT NULL CHECK (rule_type IN ('whitelist', 'blacklist')),
    ip_address INET NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ip_rules_admin_id ON ip_access_rules(admin_id);
CREATE INDEX IF NOT EXISTS idx_ip_rules_type ON ip_access_rules(rule_type, active);
CREATE INDEX IF NOT EXISTS idx_ip_rules_ip ON ip_access_rules(ip_address) WHERE active = TRUE;

-- Global Security Config table
CREATE TABLE IF NOT EXISTS global_security_config (
    id UUID PRIMARY KEY,
    password_policy_id UUID REFERENCES password_policies(id),
    require_mfa_for_admins BOOLEAN DEFAULT FALSE,
    ip_whitelist_enabled BOOLEAN DEFAULT FALSE,
    ip_blacklist_enabled BOOLEAN DEFAULT TRUE,
    brute_force_lockout_mins INT DEFAULT 30,
    max_login_attempts INT DEFAULT 5,
    max_concurrent_sessions INT DEFAULT 3,
    session_timeout_minutes INT DEFAULT 30,
    auto_logout_idle_minutes INT DEFAULT 15,
    require_password_change_on TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Login Attempts tracking table (for rate limiting)
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    success BOOLEAN NOT NULL,
    reason VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_admin_id ON login_attempts(admin_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip ON login_attempts(ip_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_attempts_created_at ON login_attempts(created_at DESC);

-- Auto cleanup for old login attempts (keep 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_login_attempts()
RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts 
    WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;
