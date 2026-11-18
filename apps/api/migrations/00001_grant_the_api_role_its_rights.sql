-- +goose Up

-- The migration role owns the schema and every table in it. The API role
-- uses the schema and creates nothing there: it reads and writes rows. The
-- migration that creates a table grants the API role its rights on that
-- table, in the same file.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO api;

-- Readiness compares the version recorded here with the one the binary
-- expects.
GRANT SELECT ON goose_db_version TO api;
