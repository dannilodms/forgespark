# 11: Resend adapter and the bounce webhook

**What to build:** Production can send its e-mail through Resend and learn of a bounce: the adapter sends with the record's identifier as idempotency key and stores the identifier Resend returns; an operation under `/api`, present only when its secret is configured, verifies Resend's signature on a bounce report, ignores a report already seen and marks the record `bounced`. The maintainer chose to build this now, before an account exists: the Resend adapter is covered by simulated responses and by a scheduled test that skips until a key is configured. Rules: E-mail and Pipeline in `CODING_STANDARDS.md`.

**Blocked by:** 02 (E-mail records sent through SMTP).

**Status:** resolved

- [x] The `resend` adapter sends through Resend's API with the identifier of the record as the idempotency key and as a tag, and stores the identifier Resend returns on the record; chosen by configuration as the sender.
- [x] Its tests against Resend's test addresses run only in a scheduled workflow, never on a pull request, and skip when no key is configured; unit tests cover the request and the response with simulated answers.
- [x] The bounce webhook is registered only when its secret is configured; without it the path answers as unknown.
- [x] It verifies the signature on the raw body, refuses a wrong one, ignores an event identifier already seen, and calls the use case that marks the record `bounced`; it carries no cookie and is outside the origin check and the OpenAPI spec.
- [x] HTTP tests by the binary cover the wrong signature, the repeated event, and the absence without a secret.
- [x] The ticket's closing comment names which paths were never exercised against Resend, and `docs/providers.md` is left as it is: account not created, sender address not chosen.

## Comments

### What was built

- **Sending through Resend** (`apps/api/internal/email/adapters/resend/sender.go`): `POST https://api.resend.com/emails` with `Authorization: Bearer <RESEND_API_KEY>`, `Idempotency-Key: email-<record id>`, the tag `email_id=<record id>`, both parts, and the sender as Go's `mail.Address` writes it (`"My Software Factory Manager" <no-reply@...>`, or an RFC 2047 encoded word for a name with accents). The identifier Resend answers is stored as the message identifier of the record. A failure reads `the API of Resend did not take the e-mail: answered <status> <name>`; the message of Resend's answer is never read. A span `Resend send` (server, method, status, Resend's identifier); no `traceparent` leaves. Timeout 15 s.
- **The worker chooses it** with `EMAIL_SENDER=resend` (`emailSender` in `cmd/api/worker.go`). `EMAIL_SENDER` takes `smtp` or `resend`; `EMAIL_SMTP_ADDRESS` is required only for `smtp`, `RESEND_API_KEY` only for `resend` (`requiredBySender` in `platform/config`).
- **The bounce webhook** (`bounce_webhook.go`): `POST /api/resend/bounces`, mounted by `mountBounceWebhook` (`cmd/api/bounce_webhook.go`) beside `mountREST`, only when `RESEND_WEBHOOK_SECRET` is set (new value type `config.WebhookSecret`, `whsec_` followed by base64, checked at the start). It verifies the signature of Svix on the raw body, reads the record from the tag and runs the new use case `MarkEmailBounced` (`email/application`), which keeps the `svix-id` in the new table `bounce_notices` in its transaction and marks the record `bounced`, with `bounced_at`.
- **Domain**: `MarkBouncedOperation` in the transition table (from `sent` only), `Email.MarkBounced(messageID, at)`, `BouncedAt()`, `BounceOfAnotherMessageError`.
- **Migration** `00023_record_the_bounces_of_emails.sql`: `emails.bounced_at` and `bounce_notices` (API role: SELECT and INSERT).
- **Handlers outside Huma** take the limits of a body and the error format from `restapi.ReadBody`, `restapi.WriteRefusal` and `restapi.WriteFailure` (`platform/restapi/own_handler.go`).
- **Scheduled tests**: `TestAgainstResend*` in `against_resend_test.go`, `task api:test:resend`, and the workflow `.github/workflows/resend.yml` (every day at 06:17 UTC, and by hand) with the secret `RESEND_TEST_API_KEY`.
- Records: E-mail, Tests and Scheduled workflows in `CODING_STANDARDS.md`; `.env.example` (a section "Resend, the e-mail of production", and `RESEND_TEST_API_KEY` under the Go tests); the variables and the new task in `README.md` and `README.pt-BR.md`. `docs/providers.md` is left as it is: account not created, sender address not chosen.

### Decided while building

- **A client of our own, no SDK.** One call on `net/http`. Resend's Go SDK would carry the messages of Resend's errors, which may quote the address, into the errors River logs and keeps; our client keeps the status and the name only. No dependency was added. Recorded in E-mail, Once.
- **The signature of Svix is checked by hand**, not with Svix's Go library, which brings the whole client of the Svix API: HMAC-SHA256 with the key after `whsec_`, over `svix-id.svix-timestamp.body`, against any `v1` signature of the header (there are several while a secret is replaced), compared in constant time, with a timestamp within 5 minutes of now either way, as Svix's libraries allow. Recorded in E-mail, Bounces.
- **The idempotency key is `email-<record id>`** and the tag is `email_id`, both within the characters Resend allows in a tag.
- **The record is found by its tag and confirmed by Resend's identifier.** A notice marks the record only when the identifier Resend gave the e-mail is the one the record stored, so a notice of another sending that carries the same tag value (a test run on the same account, another installation) changes nothing.
- **What a notice is answered.** 204 when it is taken, and for whatever it does not act on (another type of event, no tag of a record, an unknown record, another message, a record `failed` or already `bounced`), so Resend stops delivering it; 401 `unauthorized` for a signature that is wrong, missing or out of time; 400 for a signed body that is not JSON; 500 when the use case fails, which includes a record still `queued` (the notice came before the commit of the delivery): nothing is kept, and Resend delivers it again later.
- **A notice is kept by its `svix-id`**, which Svix repeats on every delivery of a notice, for `email.bounced` events only, in the transaction that marks the record, so a failure keeps nothing. Two deliveries at once: the second waits on the unique key and finds it kept (tested).
- **Tags are read as an object or as a list.** Resend's documentation shows the tags of an event as an object, and a sending takes a list. Both are read, since no event was ever seen for real.
- **`bounced_at` was added** beside `sent_at` and `failed_at`.
- **The scheduled tests read `RESEND_TEST_API_KEY`**, a variable of their own that no other workflow sets, not `RESEND_API_KEY`, so a production key in a `.env` never makes `task api:test` call Resend. They send from `onboarding@resend.dev`, the address Resend gives an account with no domain, and take record identifiers from the clock, so no two runs share an idempotency key within its 24 hours.
- **The webhook is mounted on the mux, not in Huma**, so the spec leaves it out and a Huma middleware (the origin check of ticket 04) never runs on it. A middleware that wraps the whole mux skips it by `emailresend.BounceWebhookPath`, or because it carries no session cookie.

### What was checked

- Use case, with in-memory implementations and a fixed clock: marked bounced with its time; a notice delivered again (same identifier, another e-mail) changes nothing; another message, an unknown record, a failed or already bounced record are taken and change nothing; a queued record fails and keeps nothing, and the same notice marks it once it is sent; a failed save keeps nothing; a notice with no identifier is refused. Domain: the transition from `sent` only, and the check of the message.
- `postgres` adapter in temporary databases: the round trip of a bounced e-mail; a notice kept once, undone with its transaction, and kept once when delivered twice at once; the API role cannot change or remove a notice. The migration tests pass with 00023.
- `resend` sender against a simulated Resend (`httptest`): the method, the path, the headers, the idempotency key, the tag and the body; the quoted and the encoded name of the sender; seven statuses and names of failure, never Resend's message or the address; an answer with no known name; an accepted answer with no identifier; an unreachable server; a context deadline; the span without the address or the key.
- The webhook on the adapter: a signed bounce; a repeated notice; another secret; a body changed after it was signed; another `svix-id` or timestamp than the ones signed; each header of the signature missing, and a timestamp that is not a number; 5 minutes and 1 second either way refused, 5 minutes taken; several signatures with one right, and a `v2` alone refused; events of other types; bounces with no tag of a record; tags as a list; a signed body that is not JSON; a queued record (500); `GET` answered 405 with `Allow: POST`. Two mutations (no check of the time, any version accepted) each made a test fail.
- `restapi`: a handler of its own refuses and fails in the error format with the trace identifier, logs the raw error of a 500 only, refuses a body over the limit, takes one at the limit, and answers 408 to a caller that stalls.
- Through the binary (`cmd/api/bounce_webhook_test.go`): without the secret the path is unknown (404 in the error format); a wrong signature answers 401 and changes nothing; a notice marks the record, and the same notice again, naming another record, changes nothing, with one row in `bounce_notices`; the generated spec has no path of the webhook. The worker starts with `resend` and no SMTP server, and refuses `resend` without its key.
- Configuration: `EMAIL_SENDER=resend`; the variable of each sender required only for it; an invalid webhook secret named without its value; neither secret printed; only the worker needs the key.
- `sensitive_test.go`: the first scenario now carries the key of Resend, the webhook secret, a worker started with each sender, and a signed bounce notice with a guarded address while PostgreSQL does not answer. A mutation that put the body of the notice in the raw error made it fail on the address.
- After bringing in ticket 10 (`e9e7ef0`), the conflicts in `serve.go`, `config.go`, `worker_test.go` and both READMEs were resolved, and the checks ran again on the merged tree.

### What was not checked

- **Never exercised against Resend**, since no account exists yet: every request and answer of the sender (the form of `POST /emails`, the idempotency key, the tag, the identifier answered, the body of an error with its `name`); how Resend reads a sender name with accents written as an RFC 2047 encoded word; and the whole webhook: the payload of `email.bounced` (`data.email_id`, `data.tags` as an object or a list), the headers and the signature of Svix as Resend sends them, the form of the signing secret Resend shows, and Resend's new deliveries of a notice answered 500. Once `RESEND_TEST_API_KEY` exists, the scheduled tests cover the sender, the tag, the idempotency key and the bounce of `bounced@resend.dev` as Resend reports it in its API; nothing automated covers the webhook against Resend, which needs a public address registered in Resend, that is, production.
- The scheduled workflow has not run: it runs on GitHub after the merge, and skips until the secret exists.
- `task services` and the end-to-end test were not run (the shared stack); nothing of `apps/web` changed.

### Found, and left for a decision

- **The secret `RESEND_TEST_API_KEY`** in GitHub, and whether the scheduled tests use the account of production or one of their own. With the same account, the webhook of production receives the bounce of `bounced@resend.dev` every day, which it answers 204 and keeps one row for (no record has that identifier of Resend).
- **`bounce_notices` grows** by one row per bounce and is never pruned. Svix stops delivering a notice within days, so a periodic job could remove the rows older than a month; not needed at this volume.
- **Migration numbers.** 00023 follows 00016 and 00020 here. A development database migrated from this branch before ticket 03 (00017-00019) lands would refuse the lower numbers, since goose does not apply out of order. Temporary test databases are unaffected.
- **One recipient per e-mail.** The adapter sends to the one address the record holds; a bounce of an e-mail with several recipients would mark the whole record, which no template does today.

### Settled at the end of the effort

- **The secret `RESEND_TEST_API_KEY` and the account the scheduled tests use**: deferred by the maintainer to the deploy effort. `docs/providers.md` now names what is fixed on our side: the webhook `POST /api/resend/bounces`, the event `email.bounced`, its signing secret, the separate test key and the repository secret. Recorded under Deferred in `CODING_STANDARDS.md`.
- **`bounce_notices` grew and was never pruned**: decided by the maintainer and built in the review-fix pass. The daily cleanup removes the notices received more than 30 days before (`CleanUpBounceNotices` of `email`; migration `00027` lets the API role delete them). Recorded in Jobs of `CODING_STANDARDS.md` and in ADR-0016.
- **Migration numbers**: no longer open; every migration of the effort is merged.
- **One recipient per e-mail**: accepted as it is; no template sends to several.
