-- name: InsertAPIToken :exec
INSERT INTO api_tokens (id, name, token_hash, admin_id)
VALUES ($1, $2, $3, $4);

-- ResolveAPIToken returns the token row joined with its admin's auth attributes,
-- so the middleware can build the same claims a JWT would carry.
-- name: ResolveAPIToken :one
SELECT t.id, t.admin_id, a.sudo, a.role_id
FROM api_tokens t
JOIN admins a ON a.id = t.admin_id
WHERE t.token_hash = $1;

-- name: TouchAPIToken :exec
UPDATE api_tokens SET last_used_at = now() WHERE id = $1;

-- name: ListAPITokens :many
SELECT id, name, admin_id, created_at, last_used_at
FROM api_tokens
ORDER BY created_at DESC;

-- name: DeleteAPIToken :exec
DELETE FROM api_tokens WHERE id = $1;
