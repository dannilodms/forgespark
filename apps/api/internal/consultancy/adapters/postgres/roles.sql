-- name: NextRoleID :one
SELECT nextval('roles_id_seq')::bigint;

-- name: FindRoleByKey :one
-- The key is unique (roles_key_key), so one row or none.
SELECT id, key FROM roles WHERE key = @key;

-- name: SaveRole :exec
INSERT INTO roles (id, key)
VALUES (@id, @key)
ON CONFLICT (id) DO UPDATE
SET key = excluded.key;
