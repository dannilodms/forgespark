-- name: NextRoleID :one
SELECT nextval('roles_id_seq')::bigint;

-- name: SaveRole :exec
INSERT INTO roles (id, key)
VALUES (@id, @key)
ON CONFLICT (id) DO UPDATE
SET key = excluded.key;
