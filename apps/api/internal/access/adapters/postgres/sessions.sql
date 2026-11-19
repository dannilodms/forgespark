-- name: RemoveEndedSessions :exec
-- The sessions table is the one pgxstore keeps for scs (platform/session):
-- a session whose expiry passed is never read again.
DELETE FROM sessions
WHERE expiry < @at;
