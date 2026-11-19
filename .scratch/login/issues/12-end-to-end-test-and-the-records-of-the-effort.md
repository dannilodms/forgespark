# 12: End-to-end test and the records of the effort

**What to build:** The end-to-end test proves the whole way in against the real system, and the records of the repository say what is now built and what was proven. The test continues the script of the setup on the same database: the setup registers the administrator with the e-mail Dex knows; then a login through Dex, the shell, the logout, a login with a code read from Mailpit, the shell again; and, last, a check that spends the login limit and reads the 429. Rules: Tests in `CODING_STANDARDS.md`; the issue tracker rules for closing a spec in `docs/agents/issue-tracker.md`.

**Blocked by:** 06 (Signed read URLs and the shell), 08 (The provider on the login screen), 09 (The seed), 10 (Periodic cleanup of the storage), 11 (Resend adapter and the bounce webhook).

**Status:** resolved

- [x] `task e2e` starts Dex, Mailpit and the worker beside what it starts today, on the end-to-end database.
- [x] The setup script registers the administrator with the e-mail Dex knows; a test logs in through Dex, sees the shell with the name and the logo of the consultancy and the user's own menu, logs out and lands on the login with its notice.
- [x] A test asks for a login code, reads it from Mailpit's API, types it and sees the shell.
- [x] A test logs in through Dex as the person who is no user and reads the notice with their e-mail.
- [x] A last test spends the `login` limit and reads the 429 in the error format.
- [x] `CODING_STANDARDS.md`: Deferred no longer calls the login and the shell the next effort; each joint of this effort is marked proven or corrected, with its ADR changed where it did not hold; Rate limit has the numbers of the login code; E-mail has the format of the templates.
- [x] `README.md`: the status line, Local development with the login, Dex and Mailpit, and the diagram of the parts still match what runs.
- [x] `.env.example` lists every new variable with a value that works in development, and `docs/machine-tools.yaml` was updated by 01 if River's tooling is installed on the machine.
- [ ] Every item under "Found, and left for a decision" in the tickets of this effort was settled with the maintainer, and the spec is set to `resolved`.

## Comments

### What was built

- `apps/web/e2e/login.spec.ts`, a second file of the end-to-end test, which continues on the database the setup left. In order: a login through Dex as the administrator, from the button "Log in with Local login provider" and the login page of Dex, landing at `/` in the shell, with the menu of Ada Example, the consultancy's name, the sentence of the work surface, and the photo in the button of the menu and the logo beside the name read from the storage (the object of each URL the API answers, with its bytes read by the browser); the menu with her name and e-mail, and "Log out" to `/login?notice=logged_out` with "You logged out". Then a login with a code: the e-mail typed, the step of the code, the code read from Mailpit's API, typed, and the shell again. Then Dex as `seed+developer@fabrica.example`, back at `/login?notice=no_active_user` with "No active user has this e-mail", the e-mail and "Ask an administrator for access.". Last, outside the serial group, ten asks for a code for an e-mail of no user (202 each) until Caddy answers 429 in the error format with `Retry-After` and `Cache-Control: no-store`, a check of a code meeting the same 429, and the login screen showing "Too many attempts from this address" with the trace identifier.
- `apps/web/e2e/setup.spec.ts`: the first administrator is `Seed+Admin@Fabrica.example`, kept as `seed+admin@fabrica.example`, the administrator Dex knows; the second setup code types it in capitals to meet the refusal on the e-mail, as before.
- `apps/web/e2e/system.ts`: the bucket of the test, the address of the storage the browser reads it at, the display name of the local login provider, the two people of Dex (`administrator`, `developer`), and `mailboxOf(address)`, whose `nextLoginCode()` waits for the first e-mail to the address that Mailpit did not hold before, checks its subject and reads the six digits of its plain-text part.
- `apps/web/playwright.config.ts`: two projects, `setup` then `login`, which depends on it; a third server, the `worker` command of the binary, ready once its log says "the worker started"; the API and the worker on the database and the bucket of the test.
- `Taskfile.yml`, `e2e`: the bucket `software-factory-e2e-files`, created in the Garage of development on the first run (`garage bucket info`, else `bucket create`), the key of the API allowed on it, and the cross-origin rule set by the `garage-cors` step run on it; `E2E_FILE_STORAGE_BUCKET` handed to Playwright. Dex and Mailpit come from `task services`, which the target already depended on.
- Records: `CODING_STANDARDS.md` (Deferred: the login and the shell are built and the next effort follows; the joint of the login through Mailpit and Dex proven, the one of the verified e-mail corrected; Files: a worker cleans the bucket it is given, the end-to-end test has its own; Login: the end-to-end test drives Dex; Tests of the frontend: what the end-to-end test starts, its projects, its bucket, its order and its two limits; Delivery: the `Caddyfile` sets both limits; Rate limit: the count of each zone; Checks: Dex among the containers of the adapter tests). ADR-0011 (the check of how a provider states a verified e-mail moves to the configuration of each provider). `README.md` and `README.pt-BR.md` (the status line; the roles; the nine ports of development; the login and the shell after the setup; the login code on the screen; `task e2e`; the bounces of Resend in the diagram of the parts).

### Decided while building

- **A bucket of its own for the end-to-end test.** The worker cleans up the storage when it starts, and on the database of the test it treats every object older than a day that this database does not name as left behind: on the bucket of development it would remove the images of the development database. Files in `CODING_STANDARDS.md` already said no test runs the cleanup on the bucket of development. The bucket is created by the `e2e` target with Garage's own commands, inside its container, rather than by a new step of `compose.yaml`, so no service of the shared stack changes; it is kept between runs, and the worker of each run removes what earlier runs left there more than a day before. Recorded in Files and in Tests of the frontend rules.
- **Two files, in two projects of Playwright.** `login.spec.ts` continues on the database the setup leaves, so it runs after `setup.spec.ts`, as a project that depends on the setup's, rather than in one long file or on the order Playwright happens to read files in. Each file ends with the test that spends its zone of the limit: the setup's 429 does not touch the zone `login`.
- **The e-mail of the code is told by what Mailpit held before.** The Mailpit of development keeps every e-mail, the ones of earlier runs and of the work in development to the same address among them, so the test notes the messages to the address before it asks and takes the first new one. A run with the worker pointed at an SMTP address that refuses failed with "No new e-mail reached ..." and did not pick up an older code.
- **The images are told by their object.** The photo and the logo are decorative (`alt=""`), and the API signs a new URL at each read, so the test reads who is logged in and the consultancy with the browser's cookie and looks for an image of the same object, from the bucket of the test, whose bytes the browser read.
- **The first administrator is typed in capitals in part**, so the test also shows the e-mail kept in lowercase, which is the form Dex gives and the login matches.
- **The worker is ready on its log line**, since it listens on nothing (Playwright's `wait` on standard error).
- **No new variable.** `E2E_FILE_STORAGE_BUCKET`, like `E2E_DATABASE_URL`, is set by the target for Playwright, and the API and the worker receive it as `FILE_STORAGE_BUCKET`. `.env.example` already lists every variable `config.Load` reads (33) and every one Compose and the tests read; nothing was added to it.

### What was checked

- `task e2e` from this worktree, four times: three with every test passing (the bucket created on the first run and kept after; the last run, on the final files, 7 passed in 3.5 minutes), and once with the worker given `EMAIL_SMTP_ADDRESS=localhost:1`, where the login with a code failed on "No new e-mail reached seed+admin@fabrica.example in 30 seconds", the test after it in the serial group did not run, and the worker's lines of the failed sending carried no address.
- `task web:check` (generate check, `vp check` with the e2e files formatted and linted, imports, knip, missing keys, 155 screen tests), with `OXLINT_TSGOLINT_PATH` set to the worktree's `tsgolint.exe`.
- `go build ./...`, `go vet ./cmd/...`, `go test -count=1 -p 2 ./...` of `apps/api`, `task api:lint`, `task api:generate:check`, `task api:spec:check`, `task api:spec:lint`, `task api:lists:check`, `task versions:check`. No Go code changed.
- `.env.example` against the `env:` tags of `config.go`: every variable is there.
- `docs/machine-tools.yaml`: River's CLI runs from `apps/api/river.mod` (ticket 01), nothing was installed on the machine, so the file needs no change. `docs/providers.md` needs none either.

### What was not checked

- The last criterion: settling the items under "Found, and left for a decision" of the tickets with the maintainer, and setting the spec to `resolved`, belong to the orchestrator at the end of the effort.
- The end-to-end job of the pipeline: it runs on GitHub, on Docker Engine, once the branch is pushed. The commands of Garage and the step `garage-cors` run there in a stack started fresh, which creates the bucket on every run.
- `task check` and `task check:quick` as wholes, which start the shared stack and run `task api:test`; each part ran on its own. `task workflows:check`: no workflow changed.
- The shell at a narrow width and in the dark theme, and the preferred language other than the browser's, in the end-to-end test: the screen tests of ticket 06 cover them, and the end-to-end administrator prefers `en-US`, the language of its browser.

### Found, and left for a decision

- **The tests of Go still share the bucket of development**, as ticket 10 found: a worker of development removes, a day later, what only a temporary database names. The end-to-end test no longer does; a bucket or a prefix for the tests would close the rest.
- **The people of Dex are written in four places**: the configuration of Dex in `compose.yaml`, `access/adapters/oidc/oidctest`, `apps/web/e2e/system.ts` and the README. A change of the default of `SEED_MAILBOX` changes all four together; the configuration of Dex does not follow `SEED_MAILBOX` itself.
- `docs/machine-tools.yaml` and `docs/providers.md` need no change.
