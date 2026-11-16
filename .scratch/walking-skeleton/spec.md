# Walking skeleton

Status: needs-triage

## Goal

The ADRs name many joints between tools that were read in documentation and never run. The walking skeleton is the first backend code: one thin slice that crosses every one of them, before any feature, so a joint that does not hold changes a decision and not a delivery.

It includes the Compose files and the pipeline, which ADR-0019, ADR-0020 and ADR-0021 now decide. This spec is a first list. The issues are still to be written.

## The slice

One placeholder aggregate with one write and one list, reached through REST and through one MCP tool, with one job, one file upload and one e-mail, shown on one screen of `apps/web`, running in Compose and checked by the pipeline. It defines no real domain term.

## Joints to prove

Each one names where it was recorded as not tried.

### Backend

- The MCP handler on the same router as Huma, at `/api/mcp` (ADR-0008).
- The vacuum lint on the spec Huma generates: `snake_case`, and the custom rule for plural paths (ADR-0009). Huma adds a `$schema` property to response bodies by default, which the lint will meet.
- The error format on unknown routes, wrong methods and panics, which bypass Huma's error hook (ADR-0009).
- sqlc reading the SQL that River exports, goose applying it, and the permissions the API role needs on River's tables, the reindexer included (ADR-0016).
- Whether River's OpenTelemetry middleware links the trace that saved a job to the trace that ran it (ADR-0016).
- pgtestdb with the version of goose in use (ADR-0015).
- The migration role creating `pg_stat_statements` without being a superuser (ADR-0015).
- The AWS SDK for Go v2 against Garage: path-style addresses, checksum settings, and a URL signed for the public host name (ADR-0017).
- The AWS SDK for Go v2 against Backblaze B2: the default checksum and the setting that turns it off, a signed URL for each part of an upload in parts, the bucket's cross-origin rule, and a lifecycle rule that removes old versions after 30 days (ADR-0020). If one fails, the same tests run against Scaleway or OVHcloud.
- Login with a login code read from Mailpit, and through the local provider with Dex (ADR-0011).
- The LLM stand-in: a fixed answer, a streamed answer and the list of models, through the adapter of ADR-0004 (ADR-0019).
- The seed creating its data through use cases, twice on an empty database with the same result (ADR-0019).
- A call to the LLM provider that spends no tokens, for the health report (ADR-0012).

### Frontend

- orval on the spec Huma generates: the client, the Zod schemas and the simulated responses (ADR-0013).
- The generated Zod schemas in a TanStack Form (ADR-0013).
- A TanStack DB collection loaded by following the API's cursor to the end (ADR-0013).
- A screen test in Vitest's Browser Mode with MSW serving the generated responses (ADR-0014).
- A Vite plugin that writes Brotli copies under Vite+ (Delivery in `CODING_STANDARDS.md`).

### Delivery and observability

- The Content-Security-Policy with Base UI and the fonts of `packages/ui` (Delivery).
- The Caddy image built with a rate-limit plugin, and its 429 in the error format of the API (Delivery).
- The items ADR-0012 lists as read and not tried: the folder from which Grafana loads alert rules, the retention of Tempo, the mounts of the collector, a GitHub login that accepts one account, the metrics endpoint of Garage, and the removal of the IP address and the query string from spans.

### Environments and pipeline

- Three Compose files extending one base, with the tag of every image in the base only (Environments in `CODING_STANDARDS.md`).
- A server created from `deploy/cloud-init.yaml` by the provider's command-line tool, and whether the plan chosen can be ordered (ADR-0019).
- The staging lifecycle: the label creates the server, a push publishes again with a fresh database and the seed, removing the label destroys it, and the sweep destroys an old one (ADR-0019).
- The reserved IP address moving from one staging server to the next, and its price (ADR-0019).
- Caddy's certificate data saved to a private bucket and restored on the next staging server (ADR-0019).
- Mailpit behind a password on a host name of its own, and a login code read there (ADR-0019).
- The publish: the variables file written from the secrets of a GitHub Environment, and the stop when a name in `.env.example` has no value (ADR-0019).
- The PostgreSQL image with pgBackRest: archiving to Backblaze B2, a backup started by a timer, and the signal to the monitor (ADR-0020).
- The restore test on a server created for it, and what it checks (ADR-0020).
- `/up` through Caddy, answering with the status code of readiness and no body (Health in `CODING_STANDARDS.md`).
- A pull request from a fork: its workflow waits for approval, gets no secret and cannot be published to staging (ADR-0021).
- actionlint and zizmor on the workflows, and every action pinned to a commit hash (Pipeline in `CODING_STANDARDS.md`).
- `task outdated` finding a newer release for each kind of pinned version: a Go module, a JavaScript package, an image tag in a Compose file and in a Dockerfile, an action pinned to a commit hash, a tool a workflow installs, the system image of the servers and an MCP server (Versions in `CODING_STANDARDS.md`).
- The weekly workflow that keeps that list in one issue, and whether GitHub sends the maintainer an e-mail when a workflow opens the issue or comments on it (Pipeline).
- `task outdated:machine` on Windows, for tools installed from winget and from other sources (Versions).
- The first release, 0.1.0: the pull request that changes `VERSION` and `CHANGELOG.md`, the tag and the GitHub Release created after the publish, the number added to images already in the registry without a new build, the ruleset that refuses moving or deleting a tag, and going back to a release (ADR-0023).

## Conventions to fix

Each is recorded in `CODING_STANDARDS.md` when the skeleton settles it.

- Where a use case takes the current time from.
- Where dependencies are wired.
- The assertion library of the tests.
- Whether a record is deleted or archived.
- The tools behind Enforcement, backend and frontend, and the Go linter and formatter.
- The rate-limit plugin of Caddy and the numbers of the limit.
- The hosted monitor that calls `/up` and receives the signal of the backups.
- How the Go build and the web build read `VERSION`.
- The headings of a section of `CHANGELOG.md`.

## Out of scope

- Real domain areas and their names.
- The streaming operation and its test through Caddy, which come with the first use case that streams.
- How each login provider states that an e-mail is verified, which needs the real providers (ADR-0011).
