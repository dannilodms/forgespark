# 01 - Backend skeleton: the API starts, migrates and answers liveness and readiness

Status: resolved
Blocked by: None (can start immediately)

Read `../spec.md` and the backend rules of `CODING_STANDARDS.md` first: Layout, Dependency rule, Configuration, Health, Database, and Compose files and Development under Environments. The reasons are in ADR-0001, ADR-0007, ADR-0015 and ADR-0019.

## What to build

The first Go code of the repository. One command starts the services of development, one applies the migrations, one starts the API, and the API tells whether it is alive and ready.

- The Go module in `apps/api`, laid out as Layout describes, with `platform` and nothing ahead of need. It builds one binary with two commands for now: the one that serves the API and `migrate`. Commands are parsed with the standard library.
- Dependencies wired by hand in the `main` of each command. This is one of the three conventions of the first backend code: record it in `CODING_STANDARDS.md`.
- Configuration from environment variables, read once into one typed struct, failing at the start with a line that names the variable and never its value. A committed example file lists every variable.
- One pgx pool in `platform`, with its maximum size and timeouts from configuration.
- Migrations as numbered SQL files, embedded in the binary and applied by goose with its lock on. The first migration gives the API role its rights on the schema and creates `pg_stat_statements`, as the migration role and not as a superuser.
- Two database roles: the migration role, which owns the tables, and the API role. The development PostgreSQL creates both.
- Liveness and readiness outside `/api` and outside any contract. Readiness answers 200 when PostgreSQL answers and the schema is the version the binary expects, and 503 otherwise.
- The development Compose file at the root, extending a base file in `deploy/` that holds the tag of every image. It runs PostgreSQL 18 with `pg_stat_statements` preloaded and a health check. Log files rotate by size.
- Task targets that start the services, apply the migrations, start the API and run the Go tests, and `task check` running the Go tests beside what it runs today.
- `VERSION` at the root, holding 0.0.0.
- A job for the backend in `ci.yml`, added to the `needs` of `Checks`: it builds the binary and runs the Go tests against a PostgreSQL container, through Task targets. A pull request that changes only documents skips it.
- `README.md` and `README.pt-BR.md`: the prerequisites and how to start the system in development, as far as this ticket goes.

## Acceptance criteria

- [x] On a clean machine with the prerequisites, the Task targets start PostgreSQL, apply the migrations and start the API, with no step done by hand.
- [x] Liveness answers 200 and touches no part.
- [x] Readiness answers 200 with the migrations applied, and 503 with PostgreSQL stopped or with a migration still pending.
- [x] The API refuses to start with a required variable missing or invalid. The line names the variable and never shows a value. A test covers it.
- [x] `migrate` run twice changes nothing the second time.
- [ ] The first migration creates `pg_stat_statements` without a superuser. If it cannot, ADR-0015 is updated with what was found before the work goes on.
- [x] The API role cannot create or alter a table.
- [x] Every image tag is exact and lives in the base Compose file only.
- [ ] The backend job runs on a pull request and the `Checks` job needs it.
- [x] `CODING_STANDARDS.md` records where dependencies are wired.

## Comments

### What was built

- **The Go module.** `apps/api`, module `github.com/dannilodms/software-factory-manager/apps/api`, on Go 1.27.4. `platform` is three packages: `config`, `database` and `health`. Nothing else exists yet.
- **One binary, two commands.** `apps/api/cmd/api`. With no command it serves the API, and `migrate` applies the pending migrations. Each command is a file that wires its own dependencies by hand. A command takes its arguments, the environment and its outputs and returns the exit status, so a test runs it whole.
- **Configuration.** `config.Load` reads the environment once into one struct, for one command. The API requires `API_LISTEN_ADDRESS` and `DATABASE_URL`, and `migrate` requires `MIGRATION_DATABASE_URL`. A missing or invalid variable gives one line, such as `api: configuration: DATABASE_URL is required`. `.env.example` lists every variable.
- **The pool.** One pgxpool, built in `database.NewPool`. Its maximum size, the statement timeout and the idle-in-transaction timeout come from configuration. It opens no connection when it is built, so the API starts while PostgreSQL is down.
- **Migrations.** Numbered SQL files in `apps/api/migrations`, embedded in the binary and applied by goose 3.30.1 with its session lock. The first one, `00001_grant_the_api_role_its_rights.sql`, lets the API role use the schema and read the version goose records, and takes the right to create in the schema from everyone but the owner.
- **Two roles.** `migration` owns the database and `api` reads and writes rows. The PostgreSQL container creates both on its first start, with `pg_stat_statements`, from a script written in `compose.yaml`.
- **Liveness and readiness.** `GET /livez` and `GET /readyz`, outside `/api`, with no body. Readiness passes when PostgreSQL answers and the newest version goose recorded equals the newest migration the binary carries. It runs under `HEALTH_CHECK_TIMEOUT`.
- **Compose.** `compose.yaml` at the root, with the fixed name `software-factory-manager`, extends `deploy/compose.base.yaml`, which holds `postgres:18.7-trixie`, the preload of `pg_stat_statements`, the health check and the log rotation (10 MB, 5 files).
- **Task targets.** `services`, `services:stop`, `migrate`, `api`, `api:build`, `api:test` and `web:check`. `check` runs the checks of the app, the build of the API and the Go tests.
- **`VERSION`**, holding 0.0.0.
- **The pipeline.** `ci.yml` gained two jobs. `changes` tells whether a pull request changes something that is not a document. `backend` installs Go and Task, starts the services, builds the binary and runs the Go tests, all through Task targets. `Checks` needs both.
- **Documents.** `CODING_STANDARDS.md` gained Commands, with the wiring by hand, and what this code settled in Configuration, Health, Database, Compose files and Development. `README.md` and `README.pt-BR.md` list the variables, the steps to start the system and the commands. ADR-0015 records the finding below.

### The joint that did not hold

The migration role cannot create `pg_stat_statements`. On PostgreSQL 18.7, as the role that owns the database, `CREATE EXTENSION pg_stat_statements` answers `permission denied to create extension`, with the hint `Must be superuser to create this extension`. PostgreSQL does not mark the extension as trusted. The same role created `pg_trgm`, which is trusted.

The superuser now creates the extension once, in the script that creates the two roles, when the data directory is created. The first migration only grants the rights of the API role. ADR-0015 and Database in `CODING_STANDARDS.md` say so. Marking the extension as trusted in an image of our own would let a migration create it, and was set aside because it changes what PostgreSQL ships.

### Decided while building

- **Required by command.** One struct holds every variable, and each command requires only what it uses. The API process never receives the connection of the migration role.
- **A connection URL prints as a placeholder.** `config.PostgresURL` gives the URL only through `ConnectionString`, so a configuration written to a log shows no password.
- **Development needs no `.env`.** Task loads `.env` first and `.env.example` for what it lacks, and passes the variables to Compose. A variable of the shell wins over both, which is how a check picks another port: `API_LISTEN_ADDRESS=127.0.0.1:8081 task api`. Compose run without Task and without a `.env` stops and names the missing variable.
- **No file of the checkout is mounted.** Tried with two directories and one project name: a service with a bind mount is created again each time `up` runs from the other directory, which would cut the connections of every other worktree. A config with its content in the Compose file is not. So the script of the roles is written in `compose.yaml`. `extends` copies a service and not the top-level volumes and configs, so each environment file declares them.
- **The health check goes over TCP.** While the first-start scripts run, PostgreSQL listens on its socket only. A check over the socket would pass before the roles exist.
- **`log_error_verbosity=terse`** is set in the base file now, because Observability already asks for it and it sits beside the preload.
- **Readiness asks for the exact version.** A database ahead of the binary is not ready either.
- **A temporary database for each test.** `databasetest.New` creates an empty database owned by the migration role and removes it when the test ends, as the superuser, through `TEST_DATABASE_ADMIN_URL`. Ticket 04 brings pgtestdb and the template the migrations build.
- **testify** is used by the tests, as the spec decided. Recording it as a convention is left to ticket 04, which the ticket says.
- **The API stops when its context ends**, letting requests finish, with no deadline. That is what lets a test run the command. The shutdown of ticket 03 replaces it.
- **`task api:test` starts the services first and never reads the test cache**, because a test that reaches PostgreSQL can fail with no file changed.
- **Documents in the pipeline** are `*.md`, `docs/`, `.scratch/` and `LICENSE`. Anything else, `VERSION` included, runs the builds and the tests.

### What was checked

- **From nothing.** With the volume of the project removed and no `.env`, `task services`, `task migrate` and `task api` started PostgreSQL, applied the migration and served the API, with nothing done by hand.
- **Liveness and readiness, by hand**, on `http://127.0.0.1:8080`. Before `task migrate`: `/livez` 200 and `/readyz` 503. After it: `/readyz` 200. With the container stopped: `/livez` 200 and `/readyz` 503. Started again: `/readyz` 200. Every answer had an empty body. `/api/livez` answered 404.
- **`task migrate` twice.** The first run printed `applied 00001_grant_the_api_role_its_rights.sql`, the second `no migration to apply`.
- **The API role, in the development database.** `CREATE TABLE` answered `permission denied for schema public`, and `ALTER TABLE goose_db_version` answered `must be owner of table`. `pg_stat_statements` was installed.
- **The lines of a refused start**, from the built binary: a missing variable, an invalid one, and both at once. A wrong password in `MIGRATION_DATABASE_URL` made `migrate` fail with status 1 and no password in its line.
- **The tests**, 38 in 5 packages, against the PostgreSQL container: the configuration, liveness and readiness, the pool settings, the migrations, the rights of the API role, and the two commands run whole. Removing the settings of the pool, and giving the API role every right on the table of versions, each made a test fail.
- **On Linux.** The binary was built and the same tests passed in a `golang:1.27.4` container against the same PostgreSQL.
- **Another checkout.** `docker compose up --detach --wait` from a copy of the two Compose files in another directory left the container as it was.
- **The container.** The image is `postgres:18.7-trixie`, its log configuration is `json-file` with `max-size` 10m and `max-file` 5, and `shared_preload_libraries` and `log_error_verbosity` read as set.
- **The workflow.** actionlint 1.7.12 and zizmor 1.32.0, in offline mode, reported nothing. The script that tells documents from the rest was run on seven lists of files.
- **`task check`** passes.

### What was not checked

- **The pipeline was not run.** Nothing was pushed. The box of the backend job stays open until a pull request shows the job running and `Checks` waiting for it. The `changes` job calls the GitHub API, which was not exercised, and zizmor ran without its online audits.
- **A machine that is really clean.** The check ran on the maintainer's machine, where the image was already pulled, from an empty volume and with no `.env`.
- **`start_interval`** in the health check needs Docker Engine 25 or later. The runner was not checked for it.
- **Staging and production.** No Compose file of theirs exists yet, so nothing shows how they will declare the script of the roles.

### Found, and left for a decision

- **The API role can still create a temporary table.** PostgreSQL gives that right on a database to everyone. It lives in the session and changes no schema. Taking it away is one statement in the script of the roles.
- **A file a container needs cannot be mounted from the checkout** while worktrees share containers. Ticket 07 brings a `Caddyfile` and meets this.
- **A test that is killed leaves its temporary database behind.** Its name starts with `test_`.
