-- name: CreateUser :exec
INSERT INTO users (
    id, username, status, note, data_limit, used_traffic, expire_at,
    on_hold_expire, reset_strategy, last_reset, device_limit, allowed_hwids,
    vmess_uuid, vless_uuid, trojan_pass, ss_password, ss_method, sub_token,
    admin_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17, $18, $19, $20, $21
);

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserBySubToken :one
SELECT * FROM users WHERE sub_token = $1;

-- name: UpdateUser :exec
UPDATE users SET
    username = $2, status = $3, note = $4, data_limit = $5, used_traffic = $6,
    expire_at = $7, on_hold_expire = $8, reset_strategy = $9, last_reset = $10,
    device_limit = $11, allowed_hwids = $12, trojan_pass = $13,
    ss_password = $14, ss_method = $15, updated_at = now()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- AddUsedTraffic atomically folds a traffic delta into the aggregate counter in
-- one statement, so concurrent reports from multiple nodes never race.
-- name: AddUsedTraffic :exec
UPDATE users SET used_traffic = used_traffic + $2, updated_at = now()
WHERE id = $1;

-- AddUsedTrafficBatch folds many per-user deltas in ONE statement (parallel
-- id/delta arrays), so the stats flush is a single round-trip regardless of how
-- many users had traffic in the window.
-- name: AddUsedTrafficBatch :exec
UPDATE users u
SET used_traffic = u.used_traffic + d.delta, updated_at = now()
FROM (
    SELECT unnest(@ids::uuid[]) AS id, unnest(@deltas::bigint[]) AS delta
) d
WHERE u.id = d.id;

-- name: ListUsers :many
SELECT * FROM users
WHERE (sqlc.arg(search)::text = '' OR username ILIKE '%' || sqlc.arg(search) || '%')
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status))
  AND (sqlc.narg(admin_id)::uuid IS NULL OR admin_id = sqlc.narg(admin_id))
ORDER BY created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountUsers :one
SELECT count(*) FROM users
WHERE (sqlc.arg(search)::text = '' OR username ILIKE '%' || sqlc.arg(search) || '%')
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status))
  AND (sqlc.narg(admin_id)::uuid IS NULL OR admin_id = sqlc.narg(admin_id));

-- UsersToLimit returns active users who have crossed their data cap or expiry,
-- so the enforcement loop can disable+de-provision exactly those (not a full
-- table scan). now() is evaluated server-side to avoid clock skew.
-- name: UsersToLimit :many
SELECT * FROM users
WHERE status = 'active'
  AND (
    (data_limit > 0 AND used_traffic >= data_limit)
    OR (expire_at IS NOT NULL AND expire_at <= now())
  );

-- UsersToReset returns users whose periodic traffic reset is due, computing the
-- per-strategy window in SQL so the loop only touches users that need it.
-- name: UsersToReset :many
SELECT * FROM users
WHERE reset_strategy <> 'no_reset'
  AND (
    last_reset IS NULL
    OR (reset_strategy = 'daily'   AND last_reset <= now() - interval '1 day')
    OR (reset_strategy = 'weekly'  AND last_reset <= now() - interval '7 days')
    OR (reset_strategy = 'monthly' AND last_reset <= now() - interval '1 month')
  );

-- name: ClearUserInbounds :exec
DELETE FROM user_inbounds WHERE user_id = $1;

-- name: AddUserInbound :exec
INSERT INTO user_inbounds (user_id, inbound_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveUserInbound :exec
DELETE FROM user_inbounds WHERE user_id = $1 AND inbound_id = $2;

-- name: InboundsForUser :many
SELECT i.* FROM inbounds i
JOIN user_inbounds ui ON ui.inbound_id = i.id
WHERE ui.user_id = $1;

-- UsersByNode returns every (inbound tag, user) pair on a node's enabled
-- inbounds, the read model used to assemble a node's full desired core config.
-- name: UsersByNode :many
SELECT i.tag, sqlc.embed(u)
FROM inbounds i
JOIN user_inbounds ui ON ui.inbound_id = i.id
JOIN users u ON u.id = ui.user_id
WHERE i.node_id = $1 AND i.enabled = TRUE;

-- UserStats aggregates user counts and total used traffic per status, powering
-- the dashboard overview in a single round-trip.
-- name: UserStats :many
SELECT status, COUNT(*)::bigint AS count, COALESCE(SUM(used_traffic), 0)::bigint AS used_traffic
FROM users
GROUP BY status;

-- name: UsersExpiringSoon :many
SELECT * FROM users
WHERE status = 'active' AND expire_at IS NOT NULL AND expire_at <= $1 AND expire_at > now();

-- name: AdminUserStats :one
SELECT
    count(*)::bigint AS user_count,
    COALESCE(SUM(used_traffic), 0)::bigint AS traffic_used,
    COALESCE(SUM(data_limit), 0)::bigint AS traffic_allocated
FROM users
WHERE admin_id = $1;

-- name: AdminUserStatsByStatus :many
SELECT status, count(*)::bigint AS count
FROM users
WHERE admin_id = $1
GROUP BY status;

-- name: AdminTopUsersByTraffic :many
SELECT id, username, used_traffic, data_limit, status
FROM users
WHERE admin_id = $1
ORDER BY used_traffic DESC
LIMIT sqlc.arg(lim);

-- name: CountAdminUsersExpiringSoon :one
SELECT count(*)::bigint FROM users
WHERE admin_id = $1 AND status = 'active'
  AND expire_at IS NOT NULL AND expire_at <= now() + interval '7 days' AND expire_at > now();

-- name: CountAdminUsersCreatedSince :one
SELECT count(*)::bigint FROM users
WHERE admin_id = $1 AND created_at >= $2;

-- name: CountIPLimitEventsForAdminSince :one
SELECT count(*)::bigint
FROM ip_limit_events e
INNER JOIN users u ON u.id = e.user_id
WHERE u.admin_id = $1 AND e.created_at >= $2;
