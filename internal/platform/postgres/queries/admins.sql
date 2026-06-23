-- name: CreateAdmin :exec
INSERT INTO admins (
    id, username, password_hash, sudo, role_id, totp_secret, totp_enabled,
    user_quota, traffic_quota, traffic_quota_mode, parent_admin_id,
    wallet_traffic_bytes, wallet_user_credits, webhook_url, webhook_secret, webhook_enabled,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);

-- name: GetAdminByUsername :one
SELECT * FROM admins WHERE username = $1;

-- name: GetAdminByID :one
SELECT * FROM admins WHERE id = $1;

-- name: UpdateAdmin :exec
UPDATE admins SET
    password_hash = $2, sudo = $3, role_id = $4, totp_secret = $5,
    totp_enabled = $6, user_quota = $7, traffic_quota = $8, traffic_quota_mode = $9,
    parent_admin_id = $10, wallet_traffic_bytes = $11, wallet_user_credits = $12,
    webhook_url = $13, webhook_secret = $14, webhook_enabled = $15, last_login = $16,
    policy_max_data_limit = $17, policy_max_expire_days = $18,
    policy_allow_bulk_delete = $19, policy_allow_bulk_create = $20,
    auto_suspend_enabled = $21, ip_violation_suspend_threshold = $22, suspend_grace_minutes = $23
WHERE id = $1;

-- name: SuspendAdmin :exec
UPDATE admins SET suspended = TRUE, suspended_at = $2, suspend_reason = $3, quota_breached_at = NULL
WHERE id = $1;

-- name: UnsuspendAdmin :exec
UPDATE admins SET suspended = FALSE, suspended_at = NULL, suspend_reason = '', quota_breached_at = NULL
WHERE id = $1;

-- name: SetAdminQuotaBreachedAt :exec
UPDATE admins SET quota_breached_at = $2 WHERE id = $1;

-- name: ListResellerCandidates :many
SELECT * FROM admins WHERE sudo = FALSE AND suspended = FALSE ORDER BY created_at;

-- name: ListAdmins :many
SELECT * FROM admins ORDER BY created_at;

-- name: ListChildAdmins :many
SELECT * FROM admins WHERE parent_admin_id = $1 ORDER BY created_at;

-- name: AdjustAdminWallet :exec
UPDATE admins SET
    wallet_traffic_bytes = wallet_traffic_bytes + $2,
    wallet_user_credits = wallet_user_credits + $3
WHERE id = $1;

-- name: UpdateAdminWebhook :exec
UPDATE admins SET webhook_url = $2, webhook_secret = $3, webhook_enabled = $4 WHERE id = $1;

-- name: DeleteAdmin :exec
DELETE FROM admins WHERE id = $1;

-- CountSudoAdmins guards against locking everyone out by deleting/demoting the
-- last full-privilege admin.
-- name: CountSudoAdmins :one
SELECT count(*) FROM admins WHERE sudo = TRUE;

-- name: GetRole :one
SELECT * FROM roles WHERE id = $1;

-- name: CreateRole :exec
INSERT INTO roles (id, name, permissions) VALUES ($1, $2, $3);

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: UpdateRole :exec
UPDATE roles SET name = $2, permissions = $3 WHERE id = $1;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: ClearAdminInbounds :exec
DELETE FROM admin_inbounds WHERE admin_id = $1;

-- name: AddAdminInbound :exec
INSERT INTO admin_inbounds (admin_id, inbound_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListInboundIDsForAdmin :many
SELECT inbound_id FROM admin_inbounds WHERE admin_id = $1;

-- name: CountAdminInboundAccess :one
SELECT count(*)::bigint FROM admin_inbounds
WHERE admin_id = sqlc.arg(admin_id) AND inbound_id = ANY(sqlc.arg(inbound_ids)::uuid[]);

-- name: ClearAdminPlans :exec
DELETE FROM admin_plans WHERE admin_id = $1;

-- name: AddAdminPlan :exec
INSERT INTO admin_plans (admin_id, plan_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListPlanIDsForAdmin :many
SELECT plan_id FROM admin_plans WHERE admin_id = $1;

-- name: CountAdminPlanAccess :one
SELECT count(*)::bigint FROM admin_plans
WHERE admin_id = sqlc.arg(admin_id) AND plan_id = ANY(sqlc.arg(plan_ids)::uuid[]);

-- name: ClearAdminNodes :exec
DELETE FROM admin_nodes WHERE admin_id = $1;

-- name: AddAdminNode :exec
INSERT INTO admin_nodes (admin_id, node_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListNodeIDsForAdmin :many
SELECT node_id FROM admin_nodes WHERE admin_id = $1;

-- name: CountAdminNodeAccess :one
SELECT count(*)::bigint FROM admin_nodes
WHERE admin_id = sqlc.arg(admin_id) AND node_id = ANY(sqlc.arg(node_ids)::uuid[]);

-- name: GetAdminQuotaNotifyConfig :one
SELECT * FROM admin_quota_notify_config WHERE id = 1;

-- name: UpsertAdminQuotaNotifyConfig :exec
INSERT INTO admin_quota_notify_config (id, enabled, threshold_pct, notify_telegram, webhook_url, cooldown_minutes)
VALUES (1, $1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    enabled = EXCLUDED.enabled,
    threshold_pct = EXCLUDED.threshold_pct,
    notify_telegram = EXCLUDED.notify_telegram,
    webhook_url = EXCLUDED.webhook_url,
    cooldown_minutes = EXCLUDED.cooldown_minutes;

-- name: InsertAdminQuotaNotifyEvent :exec
INSERT INTO admin_quota_notify_events (id, admin_id, threshold, metric, usage_pct, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListAdminQuotaNotifyEvents :many
SELECT * FROM admin_quota_notify_events ORDER BY created_at DESC LIMIT $1;

-- name: LastAdminQuotaNotifyEvent :one
SELECT * FROM admin_quota_notify_events
WHERE admin_id = $1 AND metric = $2 AND threshold = $3
ORDER BY created_at DESC LIMIT 1;

-- name: InsertWalletLedger :exec
INSERT INTO admin_wallet_ledger (id, admin_id, delta_traffic, delta_users, reason, actor_admin_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListWalletLedger :many
SELECT * FROM admin_wallet_ledger WHERE admin_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: GetPortalBranding :one
SELECT * FROM portal_branding WHERE admin_id = $1;

-- name: GetPortalBrandingBySlug :one
SELECT * FROM portal_branding WHERE portal_slug = $1;

-- name: UpsertPortalBranding :exec
INSERT INTO portal_branding (admin_id, panel_title, logo_url, accent_color, footer_text, portal_slug, custom_domain)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (admin_id) DO UPDATE SET
    panel_title = EXCLUDED.panel_title,
    logo_url = EXCLUDED.logo_url,
    accent_color = EXCLUDED.accent_color,
    footer_text = EXCLUDED.footer_text,
    portal_slug = EXCLUDED.portal_slug,
    custom_domain = EXCLUDED.custom_domain;

-- name: ListAdminsWithWebhooks :many
SELECT id, username, webhook_url, webhook_secret, webhook_enabled FROM admins
WHERE webhook_enabled = TRUE AND webhook_url <> '';
