# 11: Resend adapter and the bounce webhook

**What to build:** Production can send its e-mail through Resend and learn of a bounce: the adapter sends with the record's identifier as idempotency key and stores the identifier Resend returns; an operation under `/api`, present only when its secret is configured, verifies Resend's signature on a bounce report, ignores a report already seen and marks the record `bounced`. The maintainer chose to build this now, before an account exists: the Resend adapter is covered by simulated responses and by a scheduled test that skips until a key is configured. Rules: E-mail and Pipeline in `CODING_STANDARDS.md`.

**Blocked by:** 02 (E-mail records sent through SMTP).

**Status:** ready-for-agent

- [ ] The `resend` adapter sends through Resend's API with the identifier of the record as the idempotency key and as a tag, and stores the identifier Resend returns on the record; chosen by configuration as the sender.
- [ ] Its tests against Resend's test addresses run only in a scheduled workflow, never on a pull request, and skip when no key is configured; unit tests cover the request and the response with simulated answers.
- [ ] The bounce webhook is registered only when its secret is configured; without it the path answers as unknown.
- [ ] It verifies the signature on the raw body, refuses a wrong one, ignores an event identifier already seen, and calls the use case that marks the record `bounced`; it carries no cookie and is outside the origin check and the OpenAPI spec.
- [ ] HTTP tests by the binary cover the wrong signature, the repeated event, and the absence without a secret.
- [ ] The ticket's closing comment names which paths were never exercised against Resend, and `docs/providers.md` is left as it is: account not created, sender address not chosen.
