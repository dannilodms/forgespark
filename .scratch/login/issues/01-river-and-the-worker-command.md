# 01: River and the worker command

**What to build:** The system can run work that nobody waits for on a screen. A `worker` command of the API binary runs jobs stored in PostgreSQL by River, and `task dev` starts it beside the API and the app. Nothing visible changes for a user yet; this is the ground the e-mails and the cleanup stand on. Rules: Jobs in `CODING_STANDARDS.md`, ADR-0016. Prove the joints the standards list as not tried: sqlc reading River's exported SQL, goose applying it, the permissions the API role needs, and whether River's OpenTelemetry middleware links the trace that saved a job to the one that ran it. Record each outcome in the ticket and correct the ADR or the standards where a joint does not hold.

**Blocked by:** None (can start immediately).

**Status:** resolved

- [x] River's tables arrive as numbered goose migrations exported by River's tooling, one per River schema version, with the grants the API role needs in the same migration; readiness still checks one schema version.
- [x] Where River's tooling runs from is recorded: `docs/machine-tools.yaml` if installed on the machine, otherwise as a Go tool of the module listed in Versions.
- [x] The `worker` command runs with a light queue and a heavy queue, each with a limit of concurrent jobs from configuration, and a shutdown deadline from configuration; it requires the database and nothing of the API's listening address.
- [x] On a signal the worker takes no new job, gives the running ones the deadline, then cancels them; a cancelled job runs again later.
- [x] A job worker in the `river` adapter reads identifiers from the job's arguments and calls one use case; a test with River's helpers in a temporary database proves it.
- [x] Jobs are traced with River's OpenTelemetry middleware; whether the saving trace links to the running trace is checked and recorded.
- [x] The River panel runs in development only, as the standards allow.
- [x] `task dev` starts the worker; the new variables are in `.env.example` with values that work in development, and the configuration names a missing or invalid one at the start.
- [x] Commands in `CODING_STANDARDS.md` lists the `worker` command.

## Comments

### What was built

- River v0.51.2 in `go.mod` (the library, the pgx driver, `rivertype`) with `otelriver` v0.12.0 of rivercontrib. River's CLI runs as a Go tool from `apps/api/river.mod` (`go tool -modfile=river.mod river migrate-get ...`); nothing was installed on the machine, so `docs/machine-tools.yaml` is unchanged.
- Migrations `00009` to `00015`: River's main line, versions 2 to 8, one goose file each, exported with `river migrate-get --line main --version N --up` and kept as exported, with two additions: goose `StatementBegin`/`StatementEnd` around each function body and `DO` block, and the grants of the API role after the exported SQL of each file that creates a table (`SELECT, INSERT, UPDATE, DELETE` on the table, `USAGE` on its sequence, `MAINTAIN` on `river_job`). Version 1, River's own migration table, is left out as River documents; its version 5 tolerates the absence, and version 8 is a no-op kept so the mapping of versions to files is complete. Readiness still compares one number, the newest goose file.
- `platform/jobs`: the two queues (`light`, `heavy`), `Kinds` where each area registers its job workers, `Saver`, which saves a job in the transaction the context carries and refuses a context with none, and `Runner`, which runs the queues until the shutdown notice and stops with River's soft stop: no new job, the running ones have the deadline, then their contexts are cancelled and River waits for them to return. Five attempts by default on both clients. `database.TransactionOf` exposes the transaction of a use case for it.
- The `worker` command (`cmd/api/worker.go`), `config.Worker` with `WORKER_LIGHT_QUEUE_MAX_JOBS`, `WORKER_HEAVY_QUEUE_MAX_JOBS` and `WORKER_SHUTDOWN_DEADLINE`, requiring `DATABASE_URL` and nothing of the listening address or the file storage. It checks the schema at its start and stops when a migration is pending or PostgreSQL does not answer. `mountJobs` wires the job kinds of the areas, none yet.
- `task worker`; `task dev` depends on it. The River panel: the `riverui` service of `compose.yaml` (image `ghcr.io/riverqueue/riverui:0.22.0`, in `deploy/compose.base.yaml`), on `RIVER_PANEL_PORT` (8081), reading the development database as the API role. The servers' Compose files do not declare it.
- `telemetrytest.Span` gained `Links`, so a test reads what a span points at.
- Records: ADR-0016 (consequences corrected), `CODING_STANDARDS.md` (Layout, Commands, Jobs, the proven joints in Deferred, the module files of Enforcement), `README.md` and `README.pt-BR.md` (variables, commands, local development, River's CLI).

### Decided while building

- **River's version 1 is left out.** It only creates `river_migration`, the table of River's own migrator, which goose replaces here. River documents the omission for other frameworks and its version 5 checks for the table before touching it. River UI 0.22.0 ran against the resulting schema (index, jobs, queues and features endpoints answered 200 on a throwaway database).
- **Goose annotations in the exported SQL.** Goose splits a file at semicolons and does not read dollar quotes, so the three dollar-quoted statements (the trigger function of version 2, the `DO` block of version 5 and the SQL function of version 6) are wrapped in `StatementBegin`/`StatementEnd`. Without them the migration fails with `unterminated dollar-quoted string`. Everything else is verbatim.
- **The reindexer stays on**, with `MAINTAIN` granted on `river_job`: as the API role, `REINDEX INDEX CONCURRENTLY river_job_kind` succeeds, and `ALTER TABLE`, `DROP TABLE`, `TRUNCATE` and `ALTER TYPE` are still refused.
- **The worker checks the schema at its start.** It has no readiness probe, so a pending migration stops it at once with the same line as `issue-setup-code`, instead of running on a schema without River's tables.
- **A worker with no job kind stands by.** River refuses to start with no registered worker. Until ticket 02 registers the first kind, `Runner.Run` logs `the worker has no job kind to run` and waits for the signal, so `task dev` has a worker to start. Ticket 02 removes that branch and its test.
- **The stop sequence** uses River's `SoftStopTimeout` (the shutdown deadline) through one `Stop` call, which escalates by itself to cancelling the contexts, then waits for the jobs to return under a grace of 15 seconds; `StopAndCancel` is not called by hand. River's context carries `ErrStop` as its cause, and a job that returns `ctx.Err()` is set back to `available` with its attempt given back.
- **Trace propagation** is turned on in the middleware, on the saving client of the API and on the running client of the worker, so the saving span sits inside the trace of the request and the run links to it. Work spans carry the job kind as a suffix (`river.work/<kind>`); the middleware records the kind, the queue, the attempt, the identifier of the job and the status, never the arguments.
- **No `adapters/river` package exists yet**: no area has a job in this ticket. The shape (a job kind whose arguments are identifiers, a worker whose `Work` calls one use case, the saving in the transaction) is written in the package doc of `platform/jobs` and proven by its tests, which stand in for the first adapter.

### What was checked

- Joint: goose applies River's exported SQL once annotated (`migrations` tests, `TestMigrateAppliesTheMigrationsAndASecondRunAppliesNothing`).
- Joint: sqlc reads the schema with River's SQL in it (`task api:generate:check` passes; `sqlc diff` parses every migration).
- Joint: the permissions of the API role (`TestTheAPIRoleReadsWritesAndReindexesTheTablesOfRiver`, `TestTheAPIRoleCannotChangeTheTablesOfRiver`, and River's own queries in the `platform/jobs` tests, all as the API role).
- Joint: the trace link (`TestTheRunnerRunsTheJobsOfBothQueuesAndTheRunLinksToTheTraceThatSavedTheJob`): the run is its own trace with one link to the saving span, whose parent is the span of the request.
- The shutdown (`TestTheShutdownTakesNoNewJobCutsTheRunningOneAtTheDeadlineAndBothRunAgainLater`): a job of a queue of one is cut after the deadline, the waiting one is not taken, both are left `available` at attempt 0, and a next runner completes both.
- A job worker reads the identifiers and calls its use case with `rivertest.NewWorker` in a temporary database; `Saver` saves in the transaction, nothing on a rollback, refuses a missing transaction.
- The `worker` command through the binary: starts on the database alone and stops with status 0 on the signal with JSON logs; refuses to start without `DATABASE_URL` or with an invalid limit; stops at once on a pending migration and on an unreachable PostgreSQL, without the password in its logs. The sensitive-values scenario runs the worker too.
- `go build`, `go vet`, `go mod tidy -diff`, `go test -count=1 ./...` of `apps/api`, `task api:lint`, `task api:generate:check`, `task api:spec:check`, `task api:spec:lint`, `task api:lists:check`, `task versions:check`, after `feat/login` with the seed (ticket 09) was brought in. In the whole run, the tests of `platform/database`, two of `cmd/api` and one of `cmd/seed` failed at the 30-second limit of creating a database under load; each passed when run again alone, as Tests in `CODING_STANDARDS.md` says to do.
- River UI 0.22.0 stays up on a database with no migrations (its API answers 500 until they are applied), so `task services` still runs before `task migrate`.
- River UI 0.22.0 against a throwaway migrated database, by hand (see above).

### What was not checked

- `task dev` itself was not run (ports of the main checkout); `task worker` is the same `go run` the command tests exercise.
- The `riverui` service was not started through Compose from this worktree; the image was run by hand against a migrated throwaway database with the same connection shape.
- Nothing of the frontend changed, so `task web:check` was not run.
- Metrics of River: the middleware emits them to the global meter provider, which is a no-op until the metric exporter of the standards exists.

### Found, and left for a decision

- **The worker on the servers.** ADR-0016 says the worker is its own Compose service; `deploy/compose.staging.yaml` and `deploy/compose.production.yaml` declare no API or worker service yet (the deploy effort builds the images). The `worker` command and `WORKER_*` variables are ready for it; the stop grace period of its container must exceed `WORKER_SHUTDOWN_DEADLINE` plus the 15 seconds of grace and the flush of telemetry, as the API's does.
- **River's version and the panel's.** River UI 0.22.0 is built with River v0.49.0 and works on the schema of v0.51.2 (version 8 of the main line). An update of River that changes its schema should be checked against the panel image, and the comment in `deploy/compose.base.yaml` says to update both together.
- **`service.name` of the worker's spans** is `api`, as for every command of the binary (`tracedDatabase`). Whether the worker should name itself apart (`worker`) in traces is for the health report and Grafana work; nothing here depends on it.
- **The sensitive-values test** now runs the worker against an unreachable PostgreSQL. A worker that reaches a database but fails inside River would log River's own lines; those carry job identifiers and kinds, never arguments, by what the middleware and River's executor record. A ticket that brings a job with arguments should extend the scenario with one that fails.
