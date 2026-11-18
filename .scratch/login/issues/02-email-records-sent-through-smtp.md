# 02: E-mail records sent through SMTP

**What to build:** The system can send an e-mail and keep a record of it. A use case queues an e-mail, saving its record and its sending job in the caller's transaction; the job delivers it through the SMTP adapter, pointed at Mailpit in development and tests; the record ends `sent` or, when the attempts run out, `failed`. The first template is the login code, in the three supported languages, with an HTML part and a plain-text part. Rules: E-mail in `CODING_STANDARDS.md`. The format of the templates and of their translation files is chosen here and written into E-mail.

**Blocked by:** 01 (River and the worker command).

**Status:** ready-for-agent

- [ ] An e-mail record has recipient, subject, what it refers to (kind and identifier), status (`queued`, `sent`, `failed`, `bounced`), the identifier the sender returned, when it was sent or failed, and a stable error code.
- [ ] The use case that queues an e-mail saves the record and its job in one transaction; a use case test with in-memory implementations proves both are saved or neither.
- [ ] The job calls the use case that delivers: it renders the template in the recipient's preferred language, sends through the adapter chosen by configuration, and marks the record; a second run on a record already `sent` sends nothing.
- [ ] The `smtp` adapter (go-mail) sends to Mailpit; its test reads the message back through Mailpit's API and finds both parts and the right recipient.
- [ ] Templates have an HTML part and a plain-text part, with texts in translation files for every supported language; a test fails when a language lacks a key.
- [ ] The login code template says the code, how long it works, and that whoever did not ask for it should ignore it. The code itself is passed in by the caller; this ticket does not compute it.
- [ ] No e-mail address appears in logs, spans, metric labels or job arguments; the test of the logs covers an address.
- [ ] Mailpit is a service of the development Compose file, started by `task services`, with its ports in `.env.example`; the sender (`smtp` or `resend`), the SMTP address and the sender address and name are configuration with fictitious defaults.
- [ ] The e-mail server is not yet a part of the health report (next effort).
