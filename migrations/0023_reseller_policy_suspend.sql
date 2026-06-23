-- Reseller abuse policies (max per-user limits) and auto-suspend.

ALTER TABLE admins
    ADD COLUMN IF NOT EXISTS policy_max_data_limit BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS policy_max_expire_days INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS policy_allow_bulk_delete BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS policy_allow_bulk_create BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS suspended BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS suspend_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS auto_suspend_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS ip_violation_suspend_threshold INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS suspend_grace_minutes INT NOT NULL DEFAULT 60,
    ADD COLUMN IF NOT EXISTS quota_breached_at TIMESTAMPTZ;
