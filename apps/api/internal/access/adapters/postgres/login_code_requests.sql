-- name: NextLoginCodeRequestID :one
SELECT nextval('login_code_requests_id_seq')::bigint;

-- name: LoadNotEndedLoginCodeRequestsOf :many
-- Locks each request until the transaction ends, so two requests of one user
-- at the same moment run one after the other.
SELECT id, user_id, random_value, asked_at, expires_at, attempts, ended_at
FROM login_code_requests
WHERE user_id = @user_id AND ended_at IS NULL
ORDER BY id
FOR UPDATE;

-- name: FindLoginCodeRequest :one
SELECT id, user_id, random_value, asked_at, expires_at, attempts, ended_at
FROM login_code_requests
WHERE id = @id;

-- name: LoadLoginCodeRequest :one
-- Locks the request until the transaction ends, so two checks of its code
-- count their attempts one after the other.
SELECT id, user_id, random_value, asked_at, expires_at, attempts, ended_at
FROM login_code_requests
WHERE id = @id
FOR UPDATE;

-- name: SaveLoginCodeRequest :exec
INSERT INTO login_code_requests (id, user_id, random_value, asked_at, expires_at, attempts, ended_at)
VALUES (@id, @user_id, @random_value, @asked_at, @expires_at, @attempts, @ended_at)
ON CONFLICT (id) DO UPDATE
SET user_id = excluded.user_id,
    random_value = excluded.random_value,
    asked_at = excluded.asked_at,
    expires_at = excluded.expires_at,
    attempts = excluded.attempts,
    ended_at = excluded.ended_at;

-- name: RemoveLoginCodeRequestsEndedOrExpiredBefore :exec
DELETE FROM login_code_requests
WHERE expires_at < @before OR ended_at < @before;
