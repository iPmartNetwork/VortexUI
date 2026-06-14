-- name: CreateRoutingRule :exec
INSERT INTO routing_rules (
    id, node_id, priority, name, inbound_tags, domains, ip, port,
    protocols, network, outbound_tag, balancer_tag, enabled
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
);

-- name: GetRoutingRuleByID :one
SELECT * FROM routing_rules WHERE id = $1;

-- name: UpdateRoutingRule :exec
UPDATE routing_rules SET
    priority = $2, name = $3, inbound_tags = $4, domains = $5, ip = $6,
    port = $7, protocols = $8, network = $9, outbound_tag = $10,
    balancer_tag = $11, enabled = $12
WHERE id = $1;

-- name: DeleteRoutingRule :exec
DELETE FROM routing_rules WHERE id = $1;

-- name: ListRoutingRulesByNode :many
SELECT * FROM routing_rules WHERE node_id = $1 ORDER BY priority, id;
