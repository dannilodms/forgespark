-- name: NextSetupCodeID :one
SELECT nextval('setup_codes_id_seq')::bigint;

-- name: LoadSetupCodesNotEnded :many
-- Loading locks each row until the transaction ends, so two writes to the
-- same setup code run one after the other.
SELECT id, code_hash, issued_at, expires_at, ended_at
FROM setup_codes
WHERE ended_at IS NULL
ORDER BY id
FOR UPDATE;

-- name: FindSetupCodeByHash :one
-- It locks nothing: it serves a use case that checks a setup code and saves
-- nothing. A hash belongs to one setup code at most.
SELECT id, code_hash, issued_at, expires_at, ended_at
FROM setup_codes
WHERE code_hash = @code_hash;

-- name: LoadSetupCodeByHash :one
-- Loading locks the row until the transaction ends: two operations that use
-- the same code run one after the other, and the second sees whether the
-- first ended it.
SELECT id, code_hash, issued_at, expires_at, ended_at
FROM setup_codes
WHERE code_hash = @code_hash
FOR UPDATE;

-- name: SaveSetupCode :exec
INSERT INTO setup_codes (id, code_hash, issued_at, expires_at, ended_at)
VALUES (@id, @code_hash, @issued_at, @expires_at, @ended_at)
ON CONFLICT (id) DO UPDATE
SET code_hash = excluded.code_hash,
    issued_at = excluded.issued_at,
    expires_at = excluded.expires_at,
    ended_at = excluded.ended_at;
