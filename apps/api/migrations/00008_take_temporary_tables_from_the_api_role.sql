-- +goose Up

-- PostgreSQL gives every role the right to create temporary tables on a new
-- database. The API role reads and writes rows and creates nothing, so the
-- right is taken from everyone; the migration role keeps it as the owner of
-- the database. A right on the database itself lives on the database, not in
-- it: it is not copied with a template, which is why the tests give each copy
-- the rights of its template (databasetest), and the script of the roles,
-- which runs only when the data directory is created, is not the place for it.
-- The name of the database differs in each environment, so it is read here.
-- +goose StatementBegin
DO $$
BEGIN
    EXECUTE format('REVOKE TEMPORARY ON DATABASE %I FROM PUBLIC', current_database());
END
$$;
-- +goose StatementEnd
