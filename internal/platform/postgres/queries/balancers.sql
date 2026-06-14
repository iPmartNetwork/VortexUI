-- name: CreateBalancer :exec
INSERT INTO balancers (
    id, node_id, tag, selectors, strategy, observe, probe_url, probe_interval, enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetBalancerByID :one
SELECT * FROM balancers WHERE id = $1;

-- name: UpdateBalancer :exec
UPDATE balancers SET
    tag = $2, selectors = $3, strategy = $4, observe = $5,
    probe_url = $6, probe_interval = $7, enabled = $8
WHERE id = $1;

-- name: DeleteBalancer :exec
DELETE FROM balancers WHERE id = $1;

-- name: ListBalancersByNode :many
SELECT * FROM balancers WHERE node_id = $1 ORDER BY tag;
