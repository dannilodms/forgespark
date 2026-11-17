-- +goose Up

-- A setup code opens the setup (ADR-0024). Only the SHA-256 hash of the code
-- is here, never the code. Whether a code still works follows from the three
-- times and is not stored.
CREATE SEQUENCE setup_codes_id_seq AS bigint;

CREATE TABLE setup_codes (
    id bigint PRIMARY KEY,
    code_hash bytea NOT NULL
        CONSTRAINT setup_codes_code_hash_key UNIQUE
        CONSTRAINT setup_codes_code_hash_check CHECK (octet_length(code_hash) = 32),
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    ended_at timestamptz
);

-- Issuing a setup code ends every earlier one, so at most one is not ended.
-- Two codes issued at the same moment cannot both stay.
CREATE UNIQUE INDEX setup_codes_one_not_ended ON setup_codes ((true)) WHERE ended_at IS NULL;

GRANT SELECT, INSERT, UPDATE ON setup_codes TO api;
GRANT USAGE ON SEQUENCE setup_codes_id_seq TO api;
