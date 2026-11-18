# 04 - Issue a setup code

Status: resolved
Blocked by: 01

Read `../spec.md`, ADR-0024 and, in `CODING_STANDARDS.md`, Setup, Use cases, Domain modelling, Database and Tests. The terms setup code and audit log are in `GLOSSARY.md`.

## What to build

Whoever installs the system runs `issue-setup-code` in the API container and gets a setup code on the screen. This is the first use case of the repository, so it also sets how a use case, its repository and its tests are written.

- The `consultancy` area, with the setup code as its first aggregate, and the use case that issues one. It ends every earlier code, stores the hash of the new one and returns the code, its expiry and what it opens. No REST operation exists for it (ADR-0008).
- The code as "The setup code" in `../spec.md` describes it: twenty characters of Crockford's Base32 in four groups of five, from the cryptographic random source, lasting one hour, stored only as the SHA-256 hash of its normalised form.
- What the code opens is the first setup for now. Ticket 12 makes it depend on whether the setup is concluded.
- The `issue-setup-code` command of the binary. It prints the code, when it stops working, what it opens and the address of the setup, taken from configuration, to its own output. The code reaches no log and no span.
- A Task target that runs the command in development.
- The `audit` area, with the use case that records an entry and its table, on which the API role can only insert and select. Issuing a setup code records an entry whose actor is whoever installs the system, with no identifier. The entry is written in the transaction of the operation it records: add that sentence to `CODING_STANDARDS.md`.
- The transaction opened and closed through an interface, and the repository that locks the row it loads, as Use cases and Domain modelling describe.
- sqlc for the queries of both areas, with its generated code committed and a check that fails when it is stale, in `task check` and in the pipeline.
- A use case takes the current time from a clock it receives when it is built, and the tests assert with testify. These are the other two conventions of the first backend code: record both in `CODING_STANDARDS.md`.
- `postgres` adapter tests on a temporary database cloned by pgtestdb from a template the migrations built.
- `README.md` and `README.pt-BR.md` say how to issue a setup code.

## Acceptance criteria

- [x] The command prints a code in four groups of five characters, with none of the letters I, L, O and U, together with its expiry, what it opens and the address of the setup.
- [x] Two codes issued in a row differ, and the first one is ended.
- [x] The table of setup codes holds a hash and never the code.
- [x] The code appears in no log line. The test of sensitive values of ticket 03 guards it.
- [x] Each code issued leaves one entry in the audit log, and a failure to write the entry leaves no code stored.
- [x] The API role cannot update or delete a row of the audit log. A test proves it against the database.
- [x] The command fails with a clear message, and a status other than zero, when the database is unreachable or a migration is pending.
- [x] The use case is tested with in-memory implementations of its interfaces and a fixed clock. The adapters are tested against PostgreSQL, each test on its own temporary database.
- [x] pgtestdb works with the version of goose in use. If it does not, ADR-0015 is updated before the work goes on.
- [x] Stale sqlc code fails the check.
- [x] `CODING_STANDARDS.md` records the clock, the assertion library and the sentence on the audit log and the transaction.

## Comments

### What was built

- **The use case.** `IssueSetupCode` in `apps/api/internal/consultancy/application`. In one transaction it ends every setup code not ended, stores the hash of a new one and records the operation in the audit log. It returns the code, its expiry and what it opens, which is the first setup for now.
- **The setup code.** `SetupCode` in `consultancy/domain` is the record: the hash, when it was issued, when it expires and when it was ended. `PlainSetupCode` is the code itself: 20 characters of Crockford's Base32, each from five bits of the random source, shown in four groups of five. It prints as `[setup code]` under every formatting verb, in JSON and in a log, and only `Reveal` gives its characters.
- **The audit area.** `RecordEntry` in `audit/application`, with the entry in `audit/domain`: when, which operation, the kind of actor with its identifier, and what was acted on. The domain holds one table with the actor and the subject each operation has, and refuses an entry that does not fit it.
- **Two tables.** `audit_log`, on which the API role has `SELECT` and `INSERT` and nothing else, and `setup_codes`, with a sequence for its identifiers and an index that lets at most one code be not ended.
- **The `postgres` adapters** of both areas, on SQL that sqlc generates from the `.sql` file beside each. sqlc 1.31.1 runs with `go tool` from `apps/api/sqlc.mod`. `task api:generate` writes the code and `task api:generate:check` fails when it is stale, in `task check` and in the backend job.
- **The transaction.** Each area declares `Transactions` in its `application`. `database.Transactions` in `platform` opens a pgx transaction and carries it in the context, and `database.Within` gives an adapter what to run its SQL on. The audit entry is written through it, in the transaction of the operation.
- **The repository locks what it loads**: `LoadNotEnded` is `SELECT ... FOR UPDATE`.
- **The clock.** A use case declares `Clock` and receives one. `clock.System` in `platform` is the one of the machine.
- **The in-memory implementations**, in `application/applicationtest` of each area: the repository, a clock that stands still, transactions that undo a failed operation, and the audit log, which can be told to fail.
- **`databasetest.Migrated`**: a database for each test, cloned by pgtestdb from a template the migrations built.
- **The command.** `issue-setup-code` prints the code, when it stops working, what it opens and the address of the setup, which is the new variable `APP_ORIGIN` followed by `/setup`. `task issue-setup-code` runs it in development.
- **The documents.** `CODING_STANDARDS.md`: the shape of a use case, the clock, the sentence on the audit log and the transaction, how the transaction reaches the adapters, how the code of sqlc is generated and checked, the in-memory implementations and testify. Both READMEs: how to issue a setup code, the variable, the three targets and testify. ADR-0015: pgtestdb with goose, and the module file of sqlc.

### Decided while building

- **The random source is a dependency of the use case**, an `io.Reader`, like the clock. The command passes the cryptographic one. A test passes one that counts, so the code and its hash are literals in the test and not values computed the way the code computes them.
- **sqlc has a module file of its own**, as the linter of ticket 02 has. As a `tool` line of `go.mod` it would bring its whole tree into the module of the API. The row of the Versions table changed, for every Go command-line tool.
- **The template of pgtestdb is built by `database.Migrate`**, the function the `migrate` command calls, through a small migrator of its own. pgtestdb then knows nothing of goose.
- **The transaction travels in the context.** The audit entry is written by another area, in the same transaction, and a `Run` inside a `Run` joins the outer one.
- **The audit use case takes the time from its caller.** An entry carries the moment of the operation it records, which the use case of the operation already read from its clock.
- **Another area names the words of the audit log through `audit/application`** (`vocabulary.go`), because it may not import `audit/domain`.
- **`audit_log.id` is an identity column.** An entry is not an aggregate, and nothing refers to one.
- **Each `CHECK` of `audit_log` lists the values written today.** Tickets 06 and 12 add theirs in a migration.
- **At most one setup code is not ended, by an index.** Two commands at the same moment cannot both leave a code that works: the second fails.
- **The command waits for PostgreSQL as long as a statement may run** (`DATABASE_STATEMENT_TIMEOUT`), not the two seconds of a health check. A refused connection fails at once.
- **A pending migration says what to do**: the line ends with "Apply the migrations with the migrate command first". `database.SchemaVersionError` tells that case from the others.
- **The command is traced** as ticket 03 set for a command: one span, with its queries under it.
- **The time is printed in UTC**, with the zone written.
- **The in-memory `Transactions` and `Clock` live in the consultancy area.** They were written before the import check of ticket 02 existed, which would also allow one copy in `platform`.

### What was checked

- **The command, run for real** with `task issue-setup-code`, on a database of its own in the development PostgreSQL. Before the migrations: status 1 and `api: the database schema is at version 0 and this binary expects version 3. Apply the migrations with the migrate command first`. After `task migrate`, twice: two different codes of four groups of five, each with `It opens: the first setup...`, `It stops working: 2026-11-17 15:11:48 UTC` and `Address: http://localhost:5173/setup`, and nothing on the error output.
- **What it stored**, read with `psql`: two rows in `setup_codes`, the first ended at the moment the second was issued. The `code_hash` of the second equals `sha256sum` of its 20 characters, computed in the shell. Two rows in `audit_log`: `setup_code_issued`, `installer`, no actor identifier, `setup_code` and the identifier of each code, at the time each was issued.
- **PostgreSQL out of reach**, with `DATABASE_URL` on port 1: status 1 and `api: PostgreSQL does not answer: ...`, at once, with no password in the line.
- **The use case**, with the in-memory implementations and the clock stopped: the code and its expiry, the hash as a literal, the second code ending the first, one entry for each code, and an audit log that fails leaving no code stored and the earlier one working.
- **The adapters**, each test on its own database: a setup code loaded as it was saved, the lock (a second transaction that will not wait gets `55P03`), the refusal of a second code not ended (`23505`), an entry appended, and an entry appended in a transaction that fails not being kept.
- **That the tests bite.** With the audit adapter running on the pool, the test of the failed transaction failed. With `FOR UPDATE` removed, the lock test failed. With the code logged at `debug`, the test of sensitive values failed with `the logs carry a sensitive value`, naming `the setup code`. Each change was undone.
- **The audit log is append-only for the API role**: `UPDATE`, `DELETE` and `TRUNCATE` each answer `42501`, and the row is still there.
- **A failure to write the entry, through the command**: with `INSERT` on `audit_log` revoked from the API role, status 1, no code printed, no new row in `setup_codes`, and the earlier code not ended.
- **The code reaches nothing but the output of the command.** A test reads every row of every table as text and finds the code in none, in any case, with or without hyphens. The test of sensitive values of ticket 03 has a scenario that issues a code with the logs at `debug` and the spans collected, and finds it in neither.
- **pgtestdb with goose 3.30.1.** pgtestdb 0.1.1 with its own goose migrator, 0.1.1, built the template and cloned it: the three tests of the helper passed. The helper was then moved to `database.Migrate`.
- **Stale sqlc code.** A query added to a `.sql` file and not generated, and a generated file edited by hand, each made `task api:generate:check` fail with the difference printed. Generated again, it passes.
- **On Linux**, in a `golang:1.27.4` container against the same PostgreSQL, with the merged tree: `go tool -modfile=sqlc.mod sqlc diff` built sqlc and exited 0, and every Go test passed, the two of ticket 03 for Unix included.
- **The guards of ticket 02**, after the merge: the formatter changed nothing in what this ticket wrote before it, golangci-lint reports 0 issues, the guard for imports, `else` and nesting reports none, and the check of exact versions passes with `sqlc.mod`.
- **`task check`** passes on the branch with `feat/foundation-and-setup` merged in.

### What was not checked

- **The pipeline was not run.** Nothing was pushed. The new step of `ci.yml` calls `task api:generate:check`, which ran in a Linux container on this machine. actionlint and zizmor were not run on the changed file.
- **The command in a container.** No container runs the API yet, so `docker compose exec` was not tried.
- **The development database.** `task migrate` and `task issue-setup-code` were not run against `software_factory`, to leave the database every worktree shares at the version the others expect.
- **Two commands at the same moment.** The index that refuses a second code not ended was tested through the repository, one transaction after the other, and not with two processes racing.
- **The address.** Nothing serves `/setup` yet: the screen is ticket 08.

### Found, and left for a decision

- **The suite can fail whole on a loaded machine.** One run of `go test ./...` on the merged tree failed 24 tests, most at exactly 30 seconds, with `CREATE DATABASE ...: timeout: context deadline exceeded` from `databasetest.New`. In the same run `tools/guard`, which uses no database, took 132 seconds instead of 30, so the machine was starved by something else, which had ended by the time I looked. The suite then passed three times: `go test ./...` by itself, inside `task check`, and on Linux. This ticket adds about 25 tests that each create a database, which makes it more likely. Raising the 30 seconds, or running the packages one at a time, changes how long every run takes, so I changed neither.
- **The first run of sqlc on a Windows machine takes about half an hour.** One of its dependencies, a ClickHouse parser, has 127,449 files, and Go extracted them in 24 minutes, looking idle. The build took 4.5 more. A run that is interrupted starts the extraction over. On Linux the whole step took 5 minutes. Both READMEs warn of it.
- **pgtestdb leaves databases behind** in the development PostgreSQL: the one of each test that failed, and a template for each state the migrations had. ADR-0015 records it. Nothing removes them.
- **The cache of the pipeline** is keyed by `apps/api/go.sum` alone, and sqlc and the linter are built from other files.
- **Migrating the development database** moves every worktree that is behind to "not ready", because readiness asks for the exact version. It will happen with the first ticket that needs the shared database migrated.
- **`installer`**, the kind of actor of whoever installs the system, is a word the glossary does not have.
- **A test of ticket 03 pinned the number of migrations** (`this binary expects version 1`) and failed once this ticket was merged with it. It accepts any version now. A test that names a migration or a version breaks with the next migration.
