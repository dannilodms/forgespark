# 12: End-to-end test and the records of the effort

**What to build:** The end-to-end test proves the whole way in against the real system, and the records of the repository say what is now built and what was proven. The test continues the script of the setup on the same database: the setup registers the administrator with the e-mail Dex knows; then a login through Dex, the shell, the logout, a login with a code read from Mailpit, the shell again; and, last, a check that spends the login limit and reads the 429. Rules: Tests in `CODING_STANDARDS.md`; the issue tracker rules for closing a spec in `docs/agents/issue-tracker.md`.

**Blocked by:** 06 (Signed read URLs and the shell), 08 (The provider on the login screen), 09 (The seed), 10 (Periodic cleanup of the storage), 11 (Resend adapter and the bounce webhook).

**Status:** ready-for-agent

- [ ] `task e2e` starts Dex, Mailpit and the worker beside what it starts today, on the end-to-end database.
- [ ] The setup script registers the administrator with the e-mail Dex knows; a test logs in through Dex, sees the shell with the name and the logo of the consultancy and the user's own menu, logs out and lands on the login with its notice.
- [ ] A test asks for a login code, reads it from Mailpit's API, types it and sees the shell.
- [ ] A test logs in through Dex as the person who is no user and reads the notice with their e-mail.
- [ ] A last test spends the `login` limit and reads the 429 in the error format.
- [ ] `CODING_STANDARDS.md`: Deferred no longer calls the login and the shell the next effort; each joint of this effort is marked proven or corrected, with its ADR changed where it did not hold; Rate limit has the numbers of the login code; E-mail has the format of the templates.
- [ ] `README.md`: the status line, Local development with the login, Dex and Mailpit, and the diagram of the parts still match what runs.
- [ ] `.env.example` lists every new variable with a value that works in development, and `docs/machine-tools.yaml` was updated by 01 if River's tooling is installed on the machine.
- [ ] Every item under "Found, and left for a decision" in the tickets of this effort was settled with the maintainer, and the spec is set to `resolved`.
