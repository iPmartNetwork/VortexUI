-- name: CreateInbound :exec
INSERT INTO inbounds (
    id, node_id, tag, protocol, listen, port, network, security,
    sni, path, host, flow, evasion_profile_id, raw, enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
);

-- name: GetInboundByID :one
SELECT * FROM inbounds WHERE id = $1;

-- name: UpdateInbound :exec
UPDATE inbounds SET
    tag = $2, protocol = $3, listen = $4, port = $5, network = $6,
    security = $7, sni = $8, path = $9, host = $10, flow = $11,
    evasion_profile_id = $12, raw = $13, enabled = $14
WHERE id = $1;

-- name: DeleteInbound :exec
DELETE FROM inbounds WHERE id = $1;

-- name: ListInboundsByNode :many
SELECT * FROM inbounds WHERE node_id = $1 ORDER BY tag;
