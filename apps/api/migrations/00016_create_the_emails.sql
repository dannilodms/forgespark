-- +goose Up

-- The record of every e-mail the system sends (E-mail in CODING_STANDARDS.md).
-- It is created queued, in the transaction that saves the job that delivers
-- it, and the job marks it sent or, when the attempts run out, failed. The
-- address of the recipient is here and nowhere else: the job carries the
-- identifier of the record only. What the e-mail says beyond its subject is
-- written when it is delivered and is not kept.
CREATE SEQUENCE emails_id_seq AS bigint;

CREATE TABLE emails (
    id bigint PRIMARY KEY,
    recipient text NOT NULL,
    subject text NOT NULL,
    template text NOT NULL
        CONSTRAINT emails_template_check CHECK (template IN ('login_code')),
    -- The language the e-mail is written in. No CHECK, as for the preferred
    -- language of a user: adding a language never changes the schema
    -- (ADR-0005).
    language text NOT NULL,
    -- What the e-mail refers to: a record of another area, by its kind and its
    -- identifier. The kinds are of different tables, so no foreign key.
    reference_kind text NOT NULL
        CONSTRAINT emails_reference_kind_check CHECK (reference_kind IN ('login_code_request')),
    reference_id bigint NOT NULL,
    status text NOT NULL
        CONSTRAINT emails_status_check CHECK (status IN ('queued', 'sent', 'failed', 'bounced')),
    queued_at timestamptz NOT NULL,
    -- The identifier the sender gave the message: the Message-ID of SMTP.
    message_id text,
    sent_at timestamptz,
    failed_at timestamptz,
    -- Why the e-mail failed, as a stable code.
    failure_code text
        CONSTRAINT emails_failure_code_check CHECK (failure_code IN ('content_unavailable', 'sending_failed', 'delivery_failed'))
);

GRANT SELECT, INSERT, UPDATE ON emails TO api;
GRANT USAGE ON SEQUENCE emails_id_seq TO api;
