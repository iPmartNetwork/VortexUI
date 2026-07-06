-- +goose Up
ALTER TABLE tls_trick_profiles ADD COLUMN IF NOT EXISTS isp TEXT NOT NULL DEFAULT 'custom';

UPDATE tls_trick_profiles SET isp = 'hamrah_aval' WHERE isp = 'custom' AND (name ILIKE '%Hamrah Aval%' OR name ILIKE '%MCI%');
UPDATE tls_trick_profiles SET isp = 'irancell' WHERE isp = 'custom' AND (name ILIKE '%Irancell%' OR name ILIKE '%MTN%');
UPDATE tls_trick_profiles SET isp = 'mokhaberat' WHERE isp = 'custom' AND (name ILIKE '%Mokhaberat%' OR name ILIKE '%TCI%');
UPDATE tls_trick_profiles SET isp = 'shatel' WHERE isp = 'custom' AND name ILIKE '%Shatel%';
UPDATE tls_trick_profiles SET isp = 'asiatech' WHERE isp = 'custom' AND name ILIKE '%Asiatech%';

-- +goose Down
ALTER TABLE tls_trick_profiles DROP COLUMN IF EXISTS isp;
