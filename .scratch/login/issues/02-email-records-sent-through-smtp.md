# 02: E-mail records sent through SMTP

**What to build:** The system can send an e-mail and keep a record of it. A use case queues an e-mail, saving its record and its sending job in the caller's transaction; the job delivers it through the SMTP adapter, pointed at Mailpit in development and tests; the record ends `sent` or, when the attempts run out, `failed`. The first template is the login code, in the three supported languages, with an HTML part and a plain-text part. Rules: E-mail in `CODING_STANDARDS.md`. The format of the templates and of their translation files is chosen here and written into E-mail.

**Blocked by:** 01 (River and the worker command).

**Status:** resolved

- [x] An e-mail record has recipient, subject, what it refers to (kind and identifier), status (`queued`, `sent`, `failed`, `bounced`), the identifier the sender returned, when it was sent or failed, and a stable error code.
- [x] The use case that queues an e-mail saves the record and its job in one transaction; a use case test with in-memory implementations proves both are saved or neither.
- [x] The job calls the use case that delivers: it renders the template in the recipient's preferred language, sends through the adapter chosen by configuration, and marks the record; a second run on a record already `sent` sends nothing.
- [x] The `smtp` adapter (go-mail) sends to Mailpit; its test reads the message back through Mailpit's API and finds both parts and the right recipient.
- [x] Templates have an HTML part and a plain-text part, with texts in translation files for every supported language; a test fails when a language lacks a key.
- [x] The login code template says the code, how long it works, and that whoever did not ask for it should ignore it. The code itself is passed in by the caller; this ticket does not compute it.
- [x] No e-mail address appears in logs, spans, metric labels or job arguments; the test of the logs covers an address.
- [x] Mailpit is a service of the development Compose file, started by `task services`, with its ports in `.env.example`; the sender (`smtp` or `resend`), the SMTP address and the sender address and name are configuration with fictitious defaults.
- [x] The e-mail server is not yet a part of the health report (next effort).

## Comments

### What was built

- The domain area `email` (`apps/api/internal/email`):
  - `domain`: the e-mail record (`Email`, created by `Queue`, with `MarkSent` and `MarkFailed` asking a transition table of ADR-0022), its `Recipient`, its `Reference` (kind and identifier; the only kind is `login_code_request`), its statuses, the failure codes `content_unavailable`, `sending_failed` and `delivery_failed`, and the templates with their texts: `templates/login_code.html`, `templates/login_code.txt` and `locales/{en-US,pt-BR,es}.json`, embedded in the binary.
  - `application`: `QueueEmail` (record and job in the caller's transaction), `DeliverEmail` (run by the job), the interfaces `EmailRepository`, `DeliveryJobs`, `Sender` and `ContentSource`, the vocabulary another area imports, and the in-memory implementations in `applicationtest`.
  - `adapters/postgres`: `EmailRepository` on the table `emails`, migration `00016_create_the_emails.sql` with its grants and CHECK constraints, sqlc entry.
  - `adapters/river`: the job kind `deliver_email` (arguments `{"email_id": N}`, light queue), `DeliveryJobs` on `jobs.Saver`, and `DeliveryWorker`, which calls `DeliverEmail` and tells it whether the attempt is the last.
  - `adapters/smtp`: the go-mail sender, with a span `SMTP send` (server address and port only), and `smtptest.RefusingServer`, an SMTP server for tests that refuses every recipient and repeats the address in its answer.
- The worker registers the job in `mountJobs` (one `jobs.Add` line), with the SMTP sender built from configuration. `jobs.Runner` no longer stands by with no job kind, and its test is gone, as ticket 01 asked.
- Configuration: `EMAIL_SENDER` (`smtp`), `EMAIL_SMTP_ADDRESS` (`host:port`), `EMAIL_FROM_ADDRESS`, `EMAIL_FROM_NAME`, all required by `worker` and by no other command, with fictitious values in `.env.example`. New value types `EmailSender`, `ServerAddress`, `EmailAddress`.
- Mailpit: `axllent/mailpit:v1.34.0` in `deploy/compose.base.yaml`, the `mailpit` service of `compose.yaml` on `MAILPIT_SMTP_PORT` (1025) and `MAILPIT_PORT` (8025). go-mail v0.8.2 in `go.mod`.
- Records: E-mail and Tests in `CODING_STANDARDS.md` (the format of the templates and translation files; the line of Deferred that left it open now points there), `README.md` and `README.pt-BR.md` (variables, services, the worker), the comment of `kernel/language.go` on what adding a language takes.

### Decided while building

- **The content is given at delivery, by the area that asked.** The code of a login is in no table and no job argument, so the template cannot take it when the e-mail is queued. `DeliverEmail` asks the `ContentSource` wired for the template (`ContentSources` in `email/application`) for the content of the record the e-mail refers to. Ticket 03 implements the source of `login_code` in `access/application` (computing the code from the request) and adds it in `mountJobs`, where `contents` is empty today. The content of the login code is `LoginCodeContent{Code, Lifetime}`. Recorded in E-mail, Content.
- **The subject takes no value of the content.** It is written in the record when the e-mail is queued, from the template and the language, so it never carries the code.
- **The record keeps its template and its language** beside what the spec lists: the delivery writes the e-mail from them. The language is what the caller passes (the recipient's preferred language), with no CHECK, as `users.preferred_language`.
- **Sending inside the transaction that locks the record**, so two runs of one job never both send; a commit that fails after the server took the e-mail sends it again on the next attempt (the Resend idempotency key of ticket 11 makes that one e-mail). The last failed attempt marks the record `failed` in a second transaction and still returns the error, so River discards the job; a run cut by the shutdown (context cancelled) marks nothing, since River gives the attempt back.
- **Failure codes** by what failed: `content_unavailable` (no source, or the source failed), `sending_failed` (the sender failed), `delivery_failed` (anything else).
- **The format of the templates and translation files** (E-mail in `CODING_STANDARDS.md`): `html/template` and `text/template` files per template, texts in `locales/<language>.json` with one object per template, each text itself a `text/template` over the content (`{{.Minutes}}`), `missingkey=error` everywhere. The test of the keys compares each language with `en-US` both ways and fails naming the missing text.
- **The `smtp` adapter never repeats the server's answer.** go-mail's `SendError` text carries the server's answer and lists the affected recipients; both name the address. The adapter replaces it with the step, the SMTP code, the enhanced code and whether it is temporary.
- **One connection per e-mail**, TLS opportunistic (Mailpit offers no STARTTLS), timeout 15 s, Message-ID by go-mail, stored without angle brackets as the identifier the sender returned.
- **`EMAIL_SENDER` accepts `smtp` only.** A value `resend` would name an adapter that does not exist; ticket 11 adds it, with the switch in `emailSender` (`cmd/api/worker.go`) and with `EMAIL_SMTP_ADDRESS` becoming required only for `smtp` (today it is in `requiredBy` for `worker`, which is the same while `smtp` is the only value).
- **A `Recipient` of the `email` area**, not the `Email` value object of `consultancy`: the record needs one address and nothing else (no lowercase rule, no limit of a field), and moving `consultancy`'s type to `kernel` touches that area beyond this ticket. See "Found, and left for a decision".
- **No function of allowed actions** (ADR-0022) yet: no query use case reads e-mails. It comes with the first screen or tool that lists them.
- **The `river` adapter's tests use the in-memory repository**: the guard forbids one adapter to import another, test files included, so River's helper runs the job worker in a temporary database while the record lives in memory.

### What was checked

- Use cases with in-memory implementations and a fixed clock: queued with its job, nothing saved when the job or the record fails or the caller's transaction fails after, recipient refused without repeating it; delivered in the recipient's language and marked sent; a second run sends nothing; a failed attempt leaves it queued for the next; the last failed attempt marks it failed with its code; failed is not sent again; a content that cannot be given; a template with no source; a run cut by the shutdown; a job of a missing e-mail.
- Domain: the subject and both parts in each language with the code, the lifetime and the line to ignore it; the HTML escapes the content; transitions and their typed error; recipients refused; the test of the keys, run once with a Spanish text removed to see it fail naming `ignore`.
- `postgres` adapter in temporary databases: round trips of queued, sent and each failure code, the lock, the rollback, the CHECK constraints (and `bounced` accepted).
- `river` adapter with `rivertest`: the job saved in the transaction, in the light queue, with 5 attempts and `{"email_id": 37}` as its only argument; the worker delivers; a failed attempt before the last leaves the record queued; the last marks it failed and River discards the job.
- `smtp` adapter against the Mailpit of development (started with `docker compose --env-file .env.example up --detach --wait mailpit` from the worktree): both parts, the recipient, the sender's name and address, the subject, read back through `GET /api/v1/search?query=message-id:"..."` and `GET /api/v1/message/{ID}`; a refusing server gives the step and codes and never the address, in the error or the span; an unreachable server; an address go-mail refuses.
- `cmd/api`: the worker starts with the sender's configuration and stops with status 0; refuses to start without the database or the sender; the first sensitive-values scenario runs the worker with the sender; a new scenario queues an e-mail inside the span of a request and runs its job on the worker's wiring against the refusing server, then searches the logs (River's `Job errored; retrying` line included), the spans (`river.work/deliver_email`, `SMTP send`, `INSERT INTO emails`) and the rows of `river_job` for the address and the code; a helper test finds an address in an SMTP answer and in go-mail's list of recipients.
- Configuration: each new variable named when missing (worker) and when invalid, never its value; no other command needs them.
- `go build ./...`, `go vet`, `go mod tidy -diff`, `task api:lint`, `task api:generate:check`, `task api:spec:check`, `task api:spec:lint`, `task api:lists:check`, `task versions:check` and `go test -count=1 ./...` of `apps/api` pass. In the whole run two tests of packages this ticket does not touch failed at a timeout under load, `TestTheSeedFailsWhileAMigrationIsPending` (the 30 seconds of creating a database) and `TestTheLintPassesASpecThatKeepsEveryConvention` (vacuum's lookup timeout of 500 ms); each passed when run again alone, as Tests in `CODING_STANDARDS.md` says to do.
- A mutation that made the `smtp` adapter return go-mail's error as it is made the new sensitive-values scenario fail on the logs, the spans and the jobs River keeps; the real adapter passes it.

### What was not checked

- The worker delivering an e-mail through the command itself: no area gives the content of a template yet, so the sensitive-values scenario builds the job worker with the worker's own wiring (`deliverEmail`, `emailSender`) and a source of login codes of the test. Ticket 03, which brings the source, can run it through the command.
- `task services` was not run from the worktree (the shared stack); Mailpit alone was started with `docker compose up --wait mailpit` and reported healthy.
- `task dev` and the frontend were not run; nothing of `apps/web` changed.

### Found, and left for a decision

- **The e-mail address value object.** `consultancy/domain.Email` (lowercase, form, length) is the address of a user; the `email` area has its own minimal `Recipient`. Ticket 03 needs the same rule as `consultancy` to read the e-mail typed at the login, so the move of `Email` to `kernel` (as Between domain areas in `CODING_STANDARDS.md` says for a type a second area needs) is likely decided there.
- **The identifier the sender returned for SMTP** is the Message-ID the adapter set, not an identifier the server returned (Mailpit answers `250 2.0.0 Ok: queued as <id>`; go-mail keeps it as `Msg.ServerResponse`). The Message-ID is what Mailpit and any mail client show and search by; switching is one line if the maintainer prefers the server's queue identifier.

### Settled at the end of the effort

- **The e-mail address value object**: moved to `kernel` by ticket 03; in the review-fix pass its one rule became `kernel.IsEmailAddress`, which the recipient of an e-mail (`email/domain.NewRecipient`) and the addresses of the configuration (`config.EmailAddress`, `config.Mailbox`) are checked by. Recorded in Between domain areas of `CODING_STANDARDS.md`.
- **The identifier of an e-mail sent through SMTP** is the Message-ID the adapter sets: accepted by the maintainer as it is. Recorded in Record of E-mail in `CODING_STANDARDS.md`.
