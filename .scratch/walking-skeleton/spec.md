# Walking skeleton

Status: needs-triage

## Goal

The ADRs name many joints between tools that were read in documentation and never run. The walking skeleton is the first backend code: one thin slice that crosses every one of them, before any feature, so a joint that does not hold changes a decision and not a delivery.

It is written after the deploy tools are decided, because it includes the Compose files and the pipeline. This spec is a first list. The issues are written then.

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
- Login with a login code read from Mailpit, and through the local provider (ADR-0011).
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

## Conventions to fix

Each is recorded in `CODING_STANDARDS.md` when the skeleton settles it.

- Where a use case takes the current time from.
- Where dependencies are wired.
- The assertion library of the tests.
- Whether a record is deleted or archived.
- The tools behind Enforcement, backend and frontend, and the Go linter and formatter.

## Out of scope

- Real domain areas and their names.
- The streaming operation and its test through Caddy, which come with the first use case that streams.
- How each login provider states that an e-mail is verified, which needs the real providers (ADR-0011).
