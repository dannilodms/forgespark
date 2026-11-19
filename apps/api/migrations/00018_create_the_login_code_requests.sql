-- +goose Up

-- The requests of login codes (Login in CODING_STANDARDS.md, ADR-0011). The
-- code is in no column: it is computed from the identifier and the random
-- value of its request, with a key that is not in the database. Whether a code
-- still works follows from the times and the attempts, and is not stored.
CREATE SEQUENCE login_code_requests_id_seq AS bigint;

CREATE TABLE login_code_requests (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users,
    random_value bytea NOT NULL
        CONSTRAINT login_code_requests_random_value_check CHECK (octet_length(random_value) = 32),
    asked_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    -- The wrong codes typed for the request. The third ends it.
    attempts integer NOT NULL
        CONSTRAINT login_code_requests_attempts_check CHECK (attempts BETWEEN 0 AND 3),
    ended_at timestamptz
);

-- The limits count the requests of a user asked for in the last minute and
-- the last hour.
CREATE INDEX login_code_requests_user_id_asked_at ON login_code_requests (user_id, asked_at);

-- A new request ends the one before it, so a user has one request that was
-- not ended at most.
CREATE UNIQUE INDEX login_code_requests_one_not_ended ON login_code_requests (user_id) WHERE ended_at IS NULL;

GRANT SELECT, INSERT, UPDATE ON login_code_requests TO api;
GRANT USAGE ON SEQUENCE login_code_requests_id_seq TO api;
