-- name: NextFileID :one
SELECT nextval('files_id_seq')::bigint;

-- name: LoadFile :one
-- Locks the file until the transaction ends, so two confirmations of the
-- same file run one after the other.
SELECT id, object_key, purpose, media_type, size, started_at, confirmed_at
FROM files
WHERE id = @id
FOR UPDATE;

-- name: SaveFile :exec
INSERT INTO files (id, object_key, purpose, media_type, size, started_at, confirmed_at)
VALUES (@id, @object_key, @purpose, @media_type, @size, @started_at, @confirmed_at)
ON CONFLICT (id) DO UPDATE
SET object_key = excluded.object_key,
    purpose = excluded.purpose,
    media_type = excluded.media_type,
    size = excluded.size,
    started_at = excluded.started_at,
    confirmed_at = excluded.confirmed_at;
