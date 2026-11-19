-- +goose Up

-- River's main migration 003, exported by River's CLI and kept as exported
-- (see 00009). It creates no table, so there is nothing to grant.

-- River main migration 003 [up]
ALTER TABLE river_job ALTER COLUMN tags SET DEFAULT '{}';
UPDATE river_job SET tags = '{}' WHERE tags IS NULL;
ALTER TABLE river_job ALTER COLUMN tags SET NOT NULL;
