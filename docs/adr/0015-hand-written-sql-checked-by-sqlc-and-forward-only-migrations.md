---
status: accepted
---

# Hand-written SQL checked by sqlc, and forward-only migrations

The backend reaches PostgreSQL through pgx. SQL is written by hand in files, and sqlc generates the typed Go functions and the row structs from those queries and from the schema in the migration files. The schema changes only through numbered SQL migrations applied by goose, and a migration has no way back.

- **SQL is checked when the code is generated.** A wrong column or type fails the generation, before any test runs. Agents write the code, so the earlier a mistake surfaces, the cheaper it is.
- **A list is fixed SQL.** An optional filter is written as `(parameter IS NULL OR column = parameter)`, and each allowed sort is its own query. A list that does not fit is written by hand on pgx, as a documented exception.
- **Migrations go forward only.** A mistake is fixed by a new migration, and in development the database is recreated. A down migration is code that almost never runs and is never tested, and one that drops a column loses the data anyway.
- **Migrations run as their own step.** The API binary has a `migrate` command. In Docker Compose a one-shot service runs it, and the API starts only when it succeeds. When a migration is pending, the deploy takes a `pg_dump` first.
- **Three database roles.** The migration role owns the tables. The API role reads and writes rows and cannot change the schema. On the audit log table (ADR-0012) it can only insert and select, so the database itself keeps that table append-only. The monitoring role reads the statistics of the server and no table, and only the telemetry collector uses it (ADR-0012).
- **One database, one schema.** Foreign keys cross domain areas. The areas are one bounded context (ADR-0007), and the database guards the integrity between them.
- **PostgreSQL is the only database.** Files live in the file storage (ADR-0017), and nothing else keeps data. The API runs as one instance, so the pool inside it (pgxpool) is enough, and nothing needs a cache yet. A connection pooler and a cache service each start with a new ADR.
- **PostgreSQL 18, from the official Debian image.** `pg_stat_statements` is the only extension from the start. The superuser creates it when the data directory is created, in the same script that creates the roles, because no migration can (see Consequences). Another one arrives with the first feature that needs it, and one that the official image does not ship (pgvector, pg_cron, pgaudit) starts with a new ADR, because it changes the image. The image is built in this repository, to add pgBackRest for the backups (ADR-0020).
- **Adapter tests use the development container.** Each test gets a temporary database cloned from a template that the same migrations built (pgtestdb). The development database is never touched.

## Considered Options

- **pgx alone, with SQL in Go strings.** Nothing is generated, but a wrong column or type shows up only in the adapter's test.
- **An ORM (GORM, Bun, ent).** Less SQL to write, but the real SQL is hidden, and the ORM's model competes with the aggregates that a repository saves whole (ADR-0007).
- **A query builder for lists (Squirrel, Jet, Bob).** Covers any dynamic query, with one more tool for agents to know and SQL that is no longer checked at generation.
- **Down migrations**, plain or exercised by a test that applies, reverts and applies again. The tested form gives a real guarantee, at a cost paid on every schema change.
- **Migrations applied when the API starts.** Simpler, but the API role would need to change the schema, and a failing migration becomes an API that restarts in a loop.
- **golang-migrate.** The best known, but it works with up and down file pairs and leaves the database marked dirty when a migration fails halfway.
- **tern.** By the author of pgx, with automatic locking, and less known to models.
- **Atlas.** It derives the migration from the desired schema. Its free edition had no migration lint and no support for extensions, views, functions or sequences when this was decided, and sequences are the base of the identifiers (ADR-0010).
- **PgBouncer.** It solves many processes opening too many connections. Here there is one process, and transaction pooling breaks `LISTEN`, session-level advisory locks and session `SET`.
- **Redis or Valkey, or an in-memory cache from the start.** A cache brings invalidation before a slow read exists to justify it. The cached automatic translation of ADR-0005 is permanent data and goes in an ordinary table.
- **A PostgreSQL schema per domain area, or no foreign keys between areas.** The boundary would show in the database, at the cost of more names and permissions, or of identifiers that can point at nothing.
- **PostgreSQL 19.** It was a release candidate when this was decided.
- **A separate test container, a fixed test database, or tests inside the development database.** The first runs a second PostgreSQL, the second runs tests one at a time, and the third can leave test rows among the development data.

## Consequences

- Returning to the previous version of the API after a migration that drops or rewrites data means restoring the dump taken before it. The last three dumps are kept on the server (ADR-0020).
- sqlc builds no dynamic SQL. A list with many sort options multiplies its queries, and an `IS NULL OR` filter can get a worse plan than a tailored query. Both are accepted at this size.
- sqlc releases slowly: v1.31.1, from April 2026, was the latest when this was decided.
- pgx is 5.13.0 or later. Version 5.12.0 fixed two high-severity security issues.
- goose does not lock against a concurrent run unless told to. The lock is turned on, although only one step applies migrations.
- Adapter tests need the development PostgreSQL container running and a role that may create databases there. pgtestdb had its last release in 2024. Tried with goose 3.30.1, pgtestdb 0.1.1 and its own goose migrator build the template and clone it. The test helper does not use that migrator: it builds the template with the function the `migrate` command calls, so pgtestdb knows nothing of goose and a new version of goose cannot break the two apart.
- pgtestdb leaves behind the database of a test that failed, to be looked at, and a template for each state the migrations had. Both stay in the development PostgreSQL until someone drops them.
- sqlc runs with `go tool` from a module file of its own, `apps/api/sqlc.mod`, so the dependencies of a code generator never meet the ones of the API. One of them, a ClickHouse parser, has more than a hundred thousand files: the first run on a Windows machine spends about half an hour extracting it, and looks idle meanwhile.
- `pg_stat_statements` must be listed in `shared_preload_libraries`, which is server configuration in the Compose file.
- The migration role cannot create `pg_stat_statements`. The first migration was meant to create it. Tried on PostgreSQL 18.7 as the role that owns the database, `CREATE EXTENSION pg_stat_statements` answers `permission denied to create extension`, with the hint that only a superuser may: PostgreSQL does not mark the extension as trusted. An extension it does mark as trusted, such as `pg_trgm`, was created by the same role. So the superuser creates `pg_stat_statements` once, beside the roles, and the migration role stays without that power. Marking the extension as trusted in an image of our own was set aside: it changes what PostgreSQL ships.
- The script that creates the roles and the extension runs only when the data directory is created. A server that already has data gets a new role or extension by hand, as the superuser.
- The official image for PostgreSQL 18 keeps its volume at `/var/lib/postgresql`, not at `/var/lib/postgresql/data` as earlier versions did.
- Moving to PostgreSQL 19 is a planned major upgrade.
- Moving a domain area to a database of its own would mean dropping the foreign keys that cross into it.
