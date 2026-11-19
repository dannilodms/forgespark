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

-- name: RemoveAbandonedFiles :exec
-- A file a confirmation holds is locked: the removal waits for it, and then
-- finds it confirmed or not.
DELETE FROM files
WHERE confirmed_at IS NULL
  AND started_at < @started_before;

-- name: ConfirmedObjectKeyOf :one
-- The key the bytes of a confirmed file are kept under, for a read: it locks
-- nothing.
SELECT object_key
FROM files
WHERE id = @id
  AND confirmed_at IS NOT NULL;

-- name: ConfirmedObjectKeys :many
SELECT object_key
FROM files
WHERE confirmed_at IS NOT NULL
  AND object_key = ANY (@object_keys::text[])
ORDER BY object_key;
