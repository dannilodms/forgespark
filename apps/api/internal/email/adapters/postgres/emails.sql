-- name: NextEmailID :one
SELECT nextval('emails_id_seq')::bigint;

-- name: LoadEmail :one
-- Locks the e-mail until the transaction ends, so two runs of its job never
-- both send it.
SELECT id, recipient, subject, template, language, reference_kind, reference_id, status,
       queued_at, message_id, sent_at, failed_at, failure_code, bounced_at
FROM emails
WHERE id = @id
FOR UPDATE;

-- name: SaveEmail :exec
INSERT INTO emails (
    id, recipient, subject, template, language, reference_kind, reference_id, status,
    queued_at, message_id, sent_at, failed_at, failure_code, bounced_at
)
VALUES (
    @id, @recipient, @subject, @template, @language, @reference_kind, @reference_id, @status,
    @queued_at, @message_id, @sent_at, @failed_at, @failure_code, @bounced_at
)
ON CONFLICT (id) DO UPDATE
SET recipient = excluded.recipient,
    subject = excluded.subject,
    template = excluded.template,
    language = excluded.language,
    reference_kind = excluded.reference_kind,
    reference_id = excluded.reference_id,
    status = excluded.status,
    queued_at = excluded.queued_at,
    message_id = excluded.message_id,
    sent_at = excluded.sent_at,
    failed_at = excluded.failed_at,
    failure_code = excluded.failure_code,
    bounced_at = excluded.bounced_at;
