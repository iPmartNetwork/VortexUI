-- name: CreateOutbound :exec
INSERT INTO outbounds (
    id, node_id, tag, protocol, address, port, uuid, password, username,
    method, flow, network, security, sni, path, host, raw, enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
);

-- name: GetOutboundByID :one
SELECT * FROM outbounds WHERE id = $1;

-- name: UpdateOutbound :exec
UPDATE outbounds SET
    tag = $2, protocol = $3, address = $4, port = $5, uuid = $6, password = $7,
    username = $8, method = $9, flow = $10, network = $11, security = $12,
    sni = $13, path = $14, host = $15, raw = $16, enabled = $17
WHERE id = $1;

-- name: DeleteOutbound :exec
DELETE FROM outbounds WHERE id = $1;

-- name: ListOutboundsByNode :many
SELECT * FROM outbounds WHERE node_id = $1 ORDER BY tag;
