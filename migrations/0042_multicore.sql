-- Multi-core Phase 1: run xray and sing-box on the same node for A/B testing.
-- nodes.enabled_cores lists active engines; inbounds.core overrides the node default.

ALTER TABLE nodes
    ADD COLUMN IF NOT EXISTS enabled_cores JSONB NOT NULL DEFAULT '["xray"]';

UPDATE nodes
SET enabled_cores = jsonb_build_array(core)
WHERE enabled_cores IS NULL OR enabled_cores = '[]'::jsonb OR enabled_cores = 'null'::jsonb;

ALTER TABLE inbounds
    ADD COLUMN IF NOT EXISTS core TEXT NOT NULL DEFAULT '';
