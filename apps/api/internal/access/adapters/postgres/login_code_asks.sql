-- name: HoldLoginCodeAsks :exec
-- Locks the asks for one address until the transaction ends, so two asks for
-- it at the same moment are counted one after the other. It is an advisory
-- lock, keyed by the address: an address asked for the first time has no row
-- to lock yet. Its key has two numbers, apart from goose's lock, which takes
-- one.
SELECT pg_advisory_xact_lock(@lock_class::integer, @lock_key::integer);

-- name: CountLoginCodeAsksSince :one
SELECT count(*)::integer
FROM login_code_asks
WHERE address = @address AND asked_at > @since;

-- name: AddLoginCodeAsk :exec
INSERT INTO login_code_asks (address, asked_at)
VALUES (@address, @asked_at);

-- name: RemoveLoginCodeAsksBefore :exec
DELETE FROM login_code_asks
WHERE asked_at < @before;
