-- +goose Up
-- Reverse-join index: UsersByNode and failover migration join user_inbounds on
-- inbound_id, which the (user_id, inbound_id) primary key does not serve.
CREATE INDEX idx_user_inbounds_inbound ON user_inbounds (inbound_id);

-- Partial index for the hourly reset sweep, which scans only users on a schedule.
CREATE INDEX idx_users_reset_strategy ON users (reset_strategy)
    WHERE reset_strategy <> 'no_reset';

-- +goose Down
DROP INDEX IF EXISTS idx_users_reset_strategy;
DROP INDEX IF EXISTS idx_user_inbounds_inbound;
