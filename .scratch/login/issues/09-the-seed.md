# 09: The seed

**What to build:** A developer runs `task seed` on an empty development database and the system opens on the login screen with a consultancy and an administrator already registered, the administrator being the person Dex knows. Running it again changes nothing. Rules: Seed in `CODING_STANDARDS.md`, ADR-0019, ADR-0021 (nothing real).

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] The seed is a `main` package of its own in the Go module, built into its own image, declared by the development Compose file only; `task seed` runs it on the development database.
- [ ] It creates its data through the use cases of the setup: it issues a setup code and concludes the setup with a fictitious consultancy and its administrator; it runs no SQL of its own.
- [ ] The administrator's e-mail is `<local part>+admin@<domain>` of a mailbox taken from configuration, whose default is fictitious; the developer's address `<local part>+developer@<domain>` is documented for Dex (07) and created by the seed in the next effort.
- [ ] Two runs on an empty database produce the same data with the same identifiers; a run on a database whose setup is concluded does nothing and exits successfully. A Go test proves both.
- [ ] Nothing real: no real person, company, address or number.
- [ ] `README.md` Local development names `task seed` as the way to skip the setup, and `.env.example` lists the mailbox; Commands in `CODING_STANDARDS.md` lists the `seed` command.
