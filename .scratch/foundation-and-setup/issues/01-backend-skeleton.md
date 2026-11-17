# 01 - Backend skeleton: the API starts, migrates and answers liveness and readiness

Status: ready-for-agent
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

- [ ] On a clean machine with the prerequisites, the Task targets start PostgreSQL, apply the migrations and start the API, with no step done by hand.
- [ ] Liveness answers 200 and touches no part.
- [ ] Readiness answers 200 with the migrations applied, and 503 with PostgreSQL stopped or with a migration still pending.
- [ ] The API refuses to start with a required variable missing or invalid. The line names the variable and never shows a value. A test covers it.
- [ ] `migrate` run twice changes nothing the second time.
- [ ] The first migration creates `pg_stat_statements` without a superuser. If it cannot, ADR-0015 is updated with what was found before the work goes on.
- [ ] The API role cannot create or alter a table.
- [ ] Every image tag is exact and lives in the base Compose file only.
- [ ] The backend job runs on a pull request and the `Checks` job needs it.
- [ ] `CODING_STANDARDS.md` records where dependencies are wired.
