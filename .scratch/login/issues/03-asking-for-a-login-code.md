# 03: Asking for a login code

**What to build:** A person sends their e-mail to the API and, when it belongs to an active user, a six-digit login code reaches their mailbox within seconds, in their preferred language. The answer is the same for every e-mail, and nothing of the code is stored anywhere. The browser that asked is remembered in a session that exists before any login. Rules: Login and Rate limit in `CODING_STANDARDS.md`, ADR-0011. Verifiable with `curl` and Mailpit.

**Blocked by:** 02 (E-mail records sent through SMTP).

**Status:** ready-for-agent

- [ ] The session exists before the login: `alexedwards/scs/v2` with `pgxstore`, its table from a migration, an opaque `HttpOnly` `SameSite=Lax` cookie (`Secure` off in development), idle life 7 days, absolute life 30 days.
- [ ] A `POST` that takes the e-mail answers 202 with no body for every e-mail. For an active user it ends their open requests, saves a login code request (identifier, user, random value, asked at, expires at, attempts, ended at) and queues the e-mail with its job in one transaction, and stores the request's identifier in the session. For an unknown or deactivated e-mail nothing is saved.
- [ ] The code is HMAC-SHA256 of the request's identifier and random value under a key from configuration, reduced to six digits; the delivery computes it to write the e-mail. It is in no table, job argument, log or span, and the test of the logs proves it.
- [ ] One request a minute and five an hour for each user, counted on their requests; past them the operation answers 422 with a rule of its own.
- [ ] A new request ends the one before it.
- [ ] Use case tests cover a known, an unknown and a deactivated e-mail with the same answer, the two limits with a fixed clock, and the ending of the earlier request.
- [ ] The `postgres` adapter tests cover the session store and the requests with their counts.
- [ ] Caddy has a zone `login` of 10 requests a minute for each IP address over the two operations of the login code, asking and checking (the path of checking is reserved here and built in 04), separate from the zone of the setup; `caddy_test.go` reads the 429 in the error format; the numbers go to Rate limit in `CODING_STANDARDS.md`.
- [ ] In development, asking with `curl` through Caddy puts an e-mail with the code in Mailpit.
