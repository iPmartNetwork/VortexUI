-- name: CreateAdmin :exec
INSERT INTO admins (
    id, username, password_hash, sudo, role_id, totp_secret, totp_enabled,
    user_quota, traffic_quota, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetAdminByUsername :one
SELECT * FROM admins WHERE username = $1;

-- name: GetAdminByID :one
SELECT * FROM admins WHERE id = $1;

-- name: UpdateAdmin :exec
UPDATE admins SET
    password_hash = $2, sudo = $3, role_id = $4, totp_secret = $5,
    totp_enabled = $6, user_quota = $7, traffic_quota = $8, last_login = $9
WHERE id = $1;

-- name: ListAdmins :many
SELECT * FROM admins ORDER BY created_at;

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
