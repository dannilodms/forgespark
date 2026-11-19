-- +goose Up

-- When the sender reported that a sent e-mail bounced: the server of the
-- recipient refused it for good (E-mail in CODING_STANDARDS.md). The record
-- keeps when it was sent and the identifier the sender gave it: the
-- Message-ID of SMTP, or the identifier Resend answered with, which its
-- notice of the bounce names.
ALTER TABLE emails ADD COLUMN bounced_at timestamptz;

-- The bounce notices taken, by the identifier their sender gives each one:
-- Resend's webhooks, signed with Svix, the `svix-id` header. A notice the
-- sender delivers again keeps its identifier, so it changes nothing. A notice
-- carries the address of the recipient, which is not kept here.
CREATE TABLE bounce_notices (
    id text PRIMARY KEY,
    received_at timestamptz NOT NULL
);

GRANT SELECT, INSERT ON bounce_notices TO api;
