-- WriteTrafficPoints bulk-inserts samples via pgx CopyFrom for high throughput
-- under heavy reporting load.
-- name: WriteTrafficPoints :copyfrom
INSERT INTO traffic_points (time, user_id, node_id, up, down)
VALUES ($1, $2, $3, $4, $5);

-- UsageSeries buckets a user's traffic over a time range. date_bin (Postgres 14+,
-- standard — not a TimescaleDB extension function) groups rows into fixed-width
-- intervals so this works even without the extension installed.
-- name: UsageSeries :many
SELECT
    date_bin(sqlc.arg(bucket)::interval, time, TIMESTAMPTZ '2000-01-01')::timestamptz AS bucket,
    sum(up)::bigint   AS up,
    sum(down)::bigint AS down
FROM traffic_points
WHERE user_id = sqlc.arg(user_id)
  AND time >= sqlc.arg(from_ts)
  AND time <  sqlc.arg(to_ts)
GROUP BY bucket
ORDER BY bucket;
