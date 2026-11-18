-- +goose Up

-- The audit log is the permanent record of sensitive operations (ADR-0012).
-- An entry says when, which operation, who acted and on what, by identifier
-- only: it never carries a name or an e-mail.
--
-- Each CHECK lists the values the API writes today. A new sensitive
-- operation comes with a migration that adds its values.
CREATE TABLE audit_log (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    occurred_at timestamptz NOT NULL,
    operation text NOT NULL
        CONSTRAINT audit_log_operation_check CHECK (operation IN ('setup_code_issued')),
    actor_kind text NOT NULL
        CONSTRAINT audit_log_actor_kind_check CHECK (actor_kind IN ('installer')),
    -- Empty for an actor that has no identifier, such as whoever installs
    -- the system.
    actor_id bigint,
    subject_kind text NOT NULL
        CONSTRAINT audit_log_subject_kind_check CHECK (subject_kind IN ('setup_code')),
    subject_id bigint NOT NULL
);

-- Append-only: the API role adds entries and reads them. It gets no right to
-- change or remove one, so the database itself keeps the record.
GRANT SELECT, INSERT ON audit_log TO api;
