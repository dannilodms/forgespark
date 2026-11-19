# 09: The seed

**What to build:** A developer runs `task seed` on an empty development database and the system opens on the login screen with a consultancy and an administrator already registered, the administrator being the person Dex knows. Running it again changes nothing. Rules: Seed in `CODING_STANDARDS.md`, ADR-0019, ADR-0021 (nothing real).

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] The seed is a `main` package of its own in the Go module, built into its own image, declared by the development Compose file only; `task seed` runs it on the development database.
- [x] It creates its data through the use cases of the setup: it issues a setup code and concludes the setup with a fictitious consultancy and its administrator; it runs no SQL of its own.
- [x] The administrator's e-mail is `<local part>+admin@<domain>` of a mailbox taken from configuration, whose default is fictitious; the developer's address `<local part>+developer@<domain>` is documented for Dex (07) and created by the seed in the next effort.
- [x] Two runs on an empty database produce the same data with the same identifiers; a run on a database whose setup is concluded does nothing and exits successfully. A Go test proves both.
- [x] Nothing real: no real person, company, address or number.
- [x] `README.md` Local development names `task seed` as the way to skip the setup, and `.env.example` lists the mailbox; Commands in `CODING_STANDARDS.md` lists the `seed` command.

## Comments

### What was built

- `apps/api/cmd/seed`: a second `main` package of the module. `run` takes the arguments, the environment and the outputs and returns the exit status, as the commands of the API binary do. It loads the configuration for the `seed` command, checks the schema, reads the state of the setup through `TellSetupState` and, when the setup is not concluded, issues a setup code with `IssueSetupCode` and concludes the setup with `ConcludeSetup`, with the fictitious consultancy and administrator in `data.go`. When the setup is concluded it prints that it changes nothing and exits 0. It runs no SQL of its own; the schema check is the one of `platform/database` that `issue-setup-code` uses.
- `apps/api/internal/platform/config`: the `Seed` command, which requires `DATABASE_URL` and `SEED_MAILBOX`; the `Mailbox` type, read as one e-mail address, with `Subaddress(tag)` giving `local+tag@domain`.
- `apps/api/cmd/seed/Dockerfile`, with the Go module as its build context: a static build on `golang:1.27.4-alpine`, copied into `gcr.io/distroless/static-debian12:nonroot`, pinned by its digest. `apps/api/.dockerignore` keeps the local binaries out of the context.
- `compose.yaml`: the `seed` service, with `build` and `image`, in the Compose profile `seed`; it extends a `seed` entry of `deploy/compose.base.yaml` that holds the logging shared by every environment. `task seed` in `Taskfile.yml` runs `go run ./cmd/seed`.
- `.env.example`: `SEED_MAILBOX=seed@fabrica.example`. `README.md` and `README.pt-BR.md`: the variable, `task seed` in Local development and in the commands table, the image in Build. `CODING_STANDARDS.md`: the seed in Commands, the Seed section in Environments rewritten with what was decided, the joint in Deferred marked proven. `GLOSSARY.md`: the term seed.
- The shared note for ticket 07 is in `codebase-map.md` of the notes directory, under "Seed (ticket 09)": the variable, its default and the two derived addresses.

### Decided while building

- **`task seed` runs the Go package on the machine**, with `go run ./cmd/seed`, as `task migrate` does, and not the image. Development in `CODING_STANDARDS.md` says the tools run on the machine and Compose runs the services they talk to; the seed is a tool, and the development connection of `.env.example` (`localhost`) serves it as it is. The image is still declared by `compose.yaml`, so development builds and proves what staging will run; it lives in a Compose profile of its own because it runs once and exits, and `task services` (`up --wait`) refuses a service that exited unless another waits for it. Recorded in Seed in `CODING_STANDARDS.md`.
- **The Dockerfile sits beside the command**, at `apps/api/cmd/seed/Dockerfile`, inside its build context (the Go module), rather than under `deploy/` as Caddy's: Compose resolves `dockerfile` against the context, and a Dockerfile outside it would be reached by `../../`. The image of the API, when the deploy builds it, can follow (`apps/api/cmd/api/Dockerfile`).
- **The mailbox is required, with the fictitious value in `.env.example`**, and not a default in the code: a staging publish that forgot the variable stops before seeding with an address nobody owns, which is how every required variable behaves (Secrets in `CODING_STANDARDS.md`). "Default" in the Seed rule is the value of `.env.example`, which is what development runs with.
- **The identifiers are deterministic through the use cases as they are**: each identifier is the next value of a sequence, and the seed takes the same steps in the same order, so two empty databases get consultancy 1, roles 1 and 2, user 1, setup code 1. No use case was extended. What differs between two runs is the moment: the times recorded and the hash of the setup code, which stays random (`crypto/rand`) so that a run cut short between issuing and concluding leaves no code anyone could guess waiting for an hour. The test compares every row of the seeded tables without those two things.
- **The data**: consultancy "Fábrica Exemplo", legal name "Fábrica Exemplo Software Ltda.", Brazil, `America/Sao_Paulo`, BRL, e-mail the mailbox itself, website `https://fabrica.example`, no logo, no telephone, no address; administrator "Ana Exemplo", `seed+admin@fabrica.example`, Brazil, preferred language `pt-BR`, no photo, no telephone. The tax ID `00.000.000/0000-00` and the national ID `000.000.000-00` are placeholders in the form of Brazil that no check digit holds for. The preferred language is Brazilian Portuguese so that the shell shows the preferred language of a user at work where the browser asks for another (story 52).
- **The output names no address.** In a container the output of the command is a log, and the mailbox of staging is the maintainer's own. The seed prints the names of the consultancy and of the administrator and says the administrator logs in with the address of `SEED_MAILBOX` tagged `+admin`.
- **The storage of the files area** is a stand-in, `noFileStorage`, whose every method answers an error: `ConcludeSetup` is built with `ConfirmUpload`, which the seed never reaches because it chooses no image, and the seed requires nothing of the file storage in its configuration.

### What was checked

- `go test -count=1 ./...` from `apps/api`, against the development PostgreSQL and Garage, the new tests of `cmd/seed` and `platform/config` included: one run registers the consultancy, the administrator with the administrator role and ends the setup code; two runs on two empty databases leave the same rows with the same identifiers; a second run on a seeded database changes no row and prints that it changes nothing; the output carries no `@`; a pending migration, a missing mailbox and a PostgreSQL that does not answer fail with status 1 and no password in the output; an argument fails with status 2 and the usage.
- In the full run, nine tests of `platform/database` failed at the 30-second limit to create a database (`context deadline exceeded`), while the image of the seed was building beside it: the case Tests in `CODING_STANDARDS.md` names. Run again alone, `./internal/platform/database/...` passed. Every other package passed in the full run.
- `go build ./...`, `task api:lint` (golangci-lint and the guard), `task versions:check`, `task api:generate:check`, `task api:spec:check`, `task api:lists:check`, `task api:spec:lint`.
- The image built from `apps/api/cmd/seed/Dockerfile` with `docker build`, and run with `docker run` on the network of the Compose project against a database created for the check and dropped afterwards, migrated from the machine: it registered the data and exited 0, and a second run exited 0 changing nothing.

### What was not checked

- `task e2e`, `task dev` and the frontend checks: the ticket touches no screen and no contract, and the ports belong to the main checkout. `docker compose --profile seed run seed` against the development database was not run from the worktree, to leave the shared development database as the maintainer left it; the image was exercised on a database of its own instead.
- The seed on a database where a setup code is waiting and the setup is not concluded: `IssueSetupCode` ends every earlier code, so the seed concludes the setup and leaves no code waiting, which the use case tests of the setup already prove; no test of the seed repeats it.

### Found, and left for a decision

- Nothing builds the seed image in the pipeline until the deploy effort writes the publish: a Dockerfile that stops building would show only on the maintainer's machine. A `docker build` step in the backend job of `ci.yml` would catch it before then; it was not added, because the pipeline is read before being changed and the deploy effort owns the images.
- The developer, `seed+developer@fabrica.example`, is listed in Dex by ticket 07 and created by the seed in the next effort, when registering a user exists, as the spec says. Until then logging in as the developer shows the notice of an e-mail with no active user.

### Settled at the end of the effort

- **Nothing built the seed image in the pipeline**: decided by the maintainer and built in the review-fix pass. `task seed:image` builds the `seed` service of `compose.yaml`, publishing nothing, and the backend job of `ci.yml` runs it; `task workflows:check` passes. Recorded in Checks on a pull request of `CODING_STANDARDS.md`.
- **The developer of Dex is created by the seed in the next effort**: nothing to settle; the spec of the next effort holds it.
