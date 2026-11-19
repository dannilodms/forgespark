-- +goose Up

-- The asks for a login code (Login in CODING_STANDARDS.md, ADR-0011): one row
-- for each ask the limits counted, for any typed address, an address of no
-- user included, so that the limits answer the same for every address. The
-- address is kept as its HMAC-SHA256 under a key derived from the key of the
-- login code, which is not in the database, and never as it was typed. A row
-- is removed once no limit counts it, an hour after the ask, by the daily
-- cleanup of the worker.
CREATE TABLE login_code_asks (
    address bytea NOT NULL
        CONSTRAINT login_code_asks_address_check CHECK (octet_length(address) = 32),
    asked_at timestamptz NOT NULL
);

-- The limits count the asks for an address in the last minute and the last
-- hour.
CREATE INDEX login_code_asks_address_asked_at ON login_code_asks (address, asked_at);

-- The limits no longer count the requests of a user.
DROP INDEX login_code_requests_user_id_asked_at;

GRANT SELECT, INSERT, DELETE ON login_code_asks TO api;
