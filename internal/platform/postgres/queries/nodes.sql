-- name: CreateNode :exec
INSERT INTO nodes (id, name, address, core, enabled_cores, status, usage_ratio, endpoint, region, country_code, location_auto, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: GetNodeByID :one
SELECT * FROM nodes WHERE id = $1;

-- name: UpdateNode :exec
UPDATE nodes SET
    name = $2, address = $3, core = $4, enabled_cores = $5, status = $6, usage_ratio = $7, endpoint = $8,
    region = $9, country_code = $10, location_auto = $11
WHERE id = $1;

-- name: DeleteNode :exec
DELETE FROM nodes WHERE id = $1;

-- name: ListNodes :many
SELECT * FROM nodes ORDER BY created_at;

-- UpdateNodeHealth persists the latest heartbeat snapshot from the hub.
-- name: UpdateNodeHealth :exec
UPDATE nodes SET
    cpu_percent = $2, mem_percent = $3, disk_percent = $4,
    core_running = $5, connections = $6, core_version = $7, agent_version = $8,
    ping_ms = $9, last_seen = now()
WHERE id = $1;
