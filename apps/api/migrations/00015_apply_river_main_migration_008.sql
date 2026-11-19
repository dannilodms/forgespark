-- +goose Up

-- River's main migration 008, exported by River's CLI and kept as exported
-- (see 00009). It changes nothing on PostgreSQL, and is kept so that each
-- version of River's main line has its migration here: the version River's
-- library expects is read off the last of these files.

-- River main migration 008 [up]
-- No-op. Postgres sequences already prevent automatically generated job IDs
-- from being reused.
SELECT 1;
