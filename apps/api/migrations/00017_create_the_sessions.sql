-- +goose Up

-- The sessions of the browsers (ADR-0011), kept by scs through pgxstore, which
-- reads and writes this table by its name and its columns. A session exists
-- before the login too: it holds what the login needs between two requests.
-- The token is the SHA-256 of the cookie, never the cookie, so a dump of the
-- database opens no session. The data is what scs encodes. A session whose
-- expiry passed is not read, and the cleanup of the store removes it.
CREATE TABLE sessions (
    token text PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamptz NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

GRANT SELECT, INSERT, UPDATE, DELETE ON sessions TO api;
