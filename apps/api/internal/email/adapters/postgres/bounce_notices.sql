-- name: KeepBounceNotice :execrows
-- Keeps the notice, or nothing when it was kept before. A second delivery of
-- the notice in a transaction of its own waits for the first one to end.
INSERT INTO bounce_notices (id, received_at)
VALUES (@id, @received_at)
ON CONFLICT (id) DO NOTHING;
