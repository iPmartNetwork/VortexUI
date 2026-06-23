-- name: InsertAudit :exec
INSERT INTO audit_log (id, admin_id, impersonator_id, method, path, status, ip)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- ListAudit returns recent entries newest-first, joining the admin's username
-- (left join so deleted admins still show their past actions).
-- name: ListAudit :many
SELECT a.id, a.time, a.admin_id, a.impersonator_id, a.method, a.path, a.status, a.ip,
       COALESCE(ad.username, '') AS username
FROM audit_log a
LEFT JOIN admins ad ON ad.id = a.admin_id
ORDER BY a.time DESC
LIMIT $1 OFFSET $2;

-- ListAuditForAdmin returns audit rows for one reseller (actions they performed).
-- name: ListAuditForAdmin :many
SELECT a.id, a.time, a.admin_id, a.impersonator_id, a.method, a.path, a.status, a.ip,
       COALESCE(ad.username, '') AS username
FROM audit_log a
LEFT JOIN admins ad ON ad.id = a.admin_id
WHERE a.admin_id = $1
ORDER BY a.time DESC
LIMIT $2 OFFSET $3;
