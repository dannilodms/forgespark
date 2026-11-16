# Coding standards

How code in this repository is written. The reasons and the rejected alternatives are in ADRs [0007](docs/adr/0007-backend-organized-by-domain-area.md) (structure), [0008](docs/adr/0008-use-cases-served-by-several-inbound-adapters.md) (inbound adapters), [0009](docs/adr/0009-rest-api-conventions.md) (API conventions), [0010](docs/adr/0010-sequential-identifiers-exposed-through-sqids.md) (identifiers), [0011](docs/adr/0011-login-through-external-providers-with-a-server-side-session.md) (login), [0012](docs/adr/0012-opentelemetry-for-observability-and-a-separate-audit-log.md) (observability), [0013](docs/adr/0013-frontend-data-in-collections-and-cursor-queries.md) (frontend state and data), [0014](docs/adr/0014-frontend-organized-by-domain-area-and-tested-through-screens.md) (frontend structure and tests), [0015](docs/adr/0015-hand-written-sql-checked-by-sqlc-and-forward-only-migrations.md) (database access and migrations), [0016](docs/adr/0016-background-jobs-in-postgresql-run-by-a-worker-command.md) (background jobs) and [0017](docs/adr/0017-files-in-garage-reached-directly-by-the-browser.md) (files).

## Backend (`apps/api`)

### Layout

```
apps/api/internal/
├── <domain area>/
│   ├── domain/         # aggregates, value objects, domain errors
│   ├── application/    # use cases, and the interfaces they need
│   └── adapters/
│       ├── rest/       # inbound: Huma operations, including SSE
│       ├── mcp/        # inbound: MCP tools
│       ├── river/      # inbound: job workers; outbound: saving a job
│       ├── postgres/   # outbound
│       ├── s3/         # outbound
│       ├── smtp/       # outbound
│       └── resend/     # outbound: sending; inbound: bounce notices
├── kernel/             # value objects used by more than one domain area
└── platform/           # shared technical infrastructure
```

A **domain area** is one part of the business with its own aggregates. It is the deep module of the backend: its public surface is its use cases, and everything else is internal to it.

An **inbound adapter** calls use cases on behalf of an outside caller: REST for the React app and for the agents of ADR-0004, MCP for coding agents, a job worker for work that runs in the background. An **outbound adapter** implements an interface that `application` declares: Postgres, the S3 storage, an LLM provider. `adapters` has one package per technology, one level deep, and an area creates only the ones it uses.

### Precedence

When two rules collide, the higher one wins:

1. The dependency rule.
2. A small public surface for the domain area.
3. The modelling and control-flow rules.

### Dependency rule

Dependencies point inward: `adapters` → `application` → `domain`.

| Package | Imports |
| --- | --- |
| `domain` | the standard library, `kernel`, its own area's `domain` |
| `application` | its own `domain`, `kernel`, interfaces from `platform`, the `application` of other areas |
| `adapters` | its own `application` and `domain`, `kernel`, `platform`, third-party libraries; no other adapter package |
| `kernel` | the standard library |
| `platform` | third-party libraries; no domain area |

`application` reaches infrastructure only through an interface: one it declares itself, or one `platform` exports (the LLM provider interface of ADR-0004). The area that uses an interface declares it; an adapter implements it.

### Between domain areas

- One area uses another by calling its use cases, and holds the other area's aggregates by ID.
- A domain type moves to `kernel` when a second area needs it. `kernel` holds value objects only, with no rule that belongs to a single area.

### Use cases

- Every operation an inbound adapter exposes is one use case in `application`. A Huma handler or an MCP tool parses its input, calls one use case and maps the result.
- A use case that only passes the call through is correct. So is the mapping between API types and domain types. Both are deliberate, not a middle man or duplication.
- A use case's inputs and outputs carry no type from Huma, the MCP SDK or any other adapter library.
- **Writes.** One use case is one transaction, opened and closed through an interface. It loads the aggregate from its repository, calls the aggregate, and saves it whole. It may save several aggregates of its own area. An effect on another area runs after the commit and before the response, inside the request. So does an effect on an external service whose result the caller waits to see, such as a streamed LLM answer or an on-demand translation. Any other effect on an external service is a job, saved inside the transaction (see Jobs).
- **Reads.** A query use case calls a read-only interface that returns data already shaped for the caller. It builds no aggregate.

### Inbound adapters

- **REST is complete.** Every use case has a REST operation. Streaming to the browser is a Server-Sent Events operation in `adapters/rest`, added with the first use case that needs it.
- **MCP is a chosen subset.** A use case becomes an MCP tool only when an agent has a real use for it. MCP exposes tools only.
- **One tool, one use case.** A tool is designed for what the agent is trying to do, not copied from a REST operation. When it needs a composed result or a different shape, that is a new use case.
- **Authentication in the adapter, authorization in the use case.** The inbound adapter authenticates the caller (a session or an access token) and passes the actor to the use case: the user, the scopes, and the access token when there is one. The use case checks the scopes. The actor type lives in `kernel`.
- **MCP language.** Tool names and descriptions are in English. Configuration content is returned in the preferred language of the access token's user.
- **Credentials over MCP.** A tool that returns the value of a credential requires the scope for reading credential values, returns one credential per call, and its use case records the read. No listing returns a value.
- **Nothing ahead of need.** No package, mount point or abstraction is written for an inbound adapter that does not exist.

### API conventions

These apply to REST. Names, identifiers, lists and error codes apply to MCP tools as well.

- **Names.** `snake_case` in bodies, query parameters, error fields and tool arguments.
- **Identifiers.** An identifier leaves the backend as a type prefix plus the Sqids encoding of the integer, at least 8 characters: `prj_Xk9fQ2mB`. An inbound adapter rejects a wrong prefix, decodes on the way in and encodes on the way out. `platform` holds the encoder and the prefix and alphabet of each aggregate type; neither ever changes. The integer appears in no response.
- **Paths.** Resources are plural nouns. A state change is a `POST` ending in the verb of its use case: `POST /projects/{id}/cancel`. The one singular path is the health report (see Health).
- **Editing details.** A `PUT` with every editable field required. No `PATCH`. A narrower edit that an agent needs is its own use case.
- **Input.** Unknown body properties and unknown query parameters are rejected.
- **Status.** 201 with `Location` for a creation, 202 for asynchronous work, 400 for malformed input, 422 for failed validation or a broken business rule, 409 for a conflict.
- **Lists.** An envelope with an opaque cursor. `limit` has a default and a ceiling, sorting accepts only listed fields, filters are query parameters, the total is computed only when asked for. No field selection.
- **Single resource.** It is the root of the body. No `success` wrapper.
- **Allowed actions.** When a screen or a tool needs to know what can be done now, the query use case returns `allowed_actions`. The client never derives it from the status.
- **Values.** Timestamps in ISO 8601 UTC. Business dates without a time. Money as a decimal string beside its currency, never a float. Enums as lowercase strings.
- **Limits.** A size limit and a read timeout on request bodies. A timeout on every outbound call, and a retry only when the call is idempotent.
- **Sessions.** A request authenticated by the session cookie that changes state must come from the app's own origin.
- **Compression and caching.** The API compresses no response and sets no cache header. Caddy does both (see Delivery).
- **Credential values.** A response that carries the value of a credential repeats nothing the caller sent in the request, because responses are compressed (see Delivery).
- **Breaking changes.** Allowed when the web client and the MCP tools change in the same commit. The path carries no version.

### Telemetry

How the telemetry is collected, kept and shown is in Observability.

- **Nothing sensitive.** Logs, spans and metric labels never carry personal data, the value of a credential, an access token or a session cookie. A test proves it for logs.
- **Where.** Spans, metrics and logs are created in `adapters` and `platform` only. `application` and `domain` create none: a use case returns a typed error, and the inbound adapter records it.
- **Spans.** An inbound adapter opens the span: of a request, of an MCP tool call (in a middleware of the MCP adapter), of a job (see Jobs). Outbound adapters are instrumented: pgx (see Database) and the HTTP clients.
- **Actor.** A span and a log line carry the public identifiers of the user, the access token, the client and the project. A metric label carries none of them.
- **Logs.** JSON lines on the container output, each with the trace identifier. Nothing is sent as OTLP logs. The level comes from configuration and is `info` by default.
- **What is logged.** A 5xx or a panic with its raw error, a failed outbound call, the start and the stop of the process, the requests cut by a shutdown. A successful request and a 4xx write no line. No line carries a body, a query string or a header.
- **Metrics.** HTTP requests, the Go runtime, the database pool and the use of LLM providers. No business metric until a case needs one.
- **Traces.** Every trace is kept. Nothing is sampled.
- **LLM calls.** A call records the provider, the model, the input and output tokens, the duration and the outcome, on its span and as metrics, under the `gen_ai` names of the OpenTelemetry semantic conventions. It never records the prompt or the answer.
- **Trace identifier.** The `trace_id` in an error body is the trace identifier of the request. An incoming `traceparent` is honoured. An outbound call carries it only to the agents of ADR-0004, never to an LLM provider, a login provider or the e-mail service.
- **Export.** Traces and metrics leave over OTLP, to an address taken from configuration. With no address, nothing is exported and the telemetry collector is `not_configured`. `platform` wraps the metric exporter and keeps the time and the error of the last export (see Health).
- **Tests.** Tests run with no telemetry.
- **Audit log.** A sensitive operation (reading the value of a credential, a login, creating or revoking an access token) writes to the audit log from its use case. The audit log is append-only and is not telemetry.

### Health

The terms system health, part and health report are defined in `GLOSSARY.md`.

- **Liveness.** `GET /livez` answers 200 and touches no part.
- **Readiness.** `GET /readyz` answers 200 when PostgreSQL answers and the database schema is the version the binary expects, and 503 otherwise. It feeds the Compose health check.
- **Outside the contract.** Liveness and readiness live in `platform`, outside `/api` and outside the OpenAPI spec. They need no authentication and answer with the status code alone. Caddy forwards only `/api`, so they are reachable only inside the Compose network.
- **Parts.** PostgreSQL is the only essential part. The file storage, the LLM provider, the telemetry collector, the job worker and the e-mail server are optional: one of them down makes the system `degraded`, never unready. An optional part with no configuration is `not_configured` and does not degrade the system.
- **Health report.** `GET /api/health` is a query use case in the system health area, which has no aggregate. It follows the API conventions, except that its path is singular, which the plural-path lint allows for this path only. It always answers 200, with the system health in the body: `operational`, `degraded` or `down`.
- **Checks.** They run when the health report is asked for, in parallel, each under a timeout taken from configuration (2 seconds by default). Nothing is checked in the background and nothing is stored. The LLM provider is checked by a call that spends no tokens, such as listing its models. The telemetry collector is not probed: its state is the last export of metrics, and it is `down` when that export failed or is overdue. The job worker is checked in the database, through the leader record River keeps there. The e-mail server is checked by a call that sends nothing.
- **Content.** For each part: its name, whether it is essential or optional, its state (`operational`, `down` or `not_configured`), the response time, the time of the check and, when it is down, a stable error code. The raw error goes to the logs only. The health report also carries the version, the commit, the build date, the time the process started, and the schema version applied and the one expected.
- **Access.** Only a user who administers the system reads the health report, through the session or through an access token that carries the read scope of the system health area.
- **When.** Liveness and readiness are written with the first backend code. The health report waits for the modelling of users, which has to say who administers the system. Nothing provisional shows the state of the optional parts until then.
- **Not built.** An MCP tool for the health report, a check of the login providers, an operation per part, and history, which would be a periodic job (see Jobs). Alerts come from Grafana (see Observability).

### Shutdown

The API stops without cutting work short. It runs as one instance, so a deploy is a restart, and a few seconds without service are accepted.

- **Signal.** `SIGTERM` or `SIGINT` starts the shutdown. A second signal exits at once.
- **Ownership.** `platform` owns the process lifecycle and exports the shutdown notice. No domain area handles a signal.
- **Health.** Readiness fails from the signal on. Liveness keeps answering.
- **Requests.** New connections are refused. A request in flight, an MCP tool call included, has the drain deadline to finish: 20 seconds, taken from configuration. No timeout on a non-streaming outbound call exceeds the drain deadline.
- **Streams.** A stream closes when the shutdown starts: an LLM answer being streamed, a progress stream, the MCP listening stream. A streaming operation watches the shutdown notice. An SSE operation first sends a final error event with the code `service_restarting`, carrying the fields of the error format. The MCP listening stream closes with no event, because MCP defines its own reconnection.
- **Deadline reached.** The context of each remaining request is cancelled, which rolls back its transaction, and the API logs how many requests were cut.
- **Exit.** The API then closes its connections to PostgreSQL, flushes telemetry under a short deadline of its own and exits: with status zero when every request finished, non-zero when one was cut.
- **No goroutine outlives its request.** Work that must survive the request is a job (see Jobs). The shutdown of the API does not wait for it.
- **Test.** A test in `platform` holds a slow request and a stream open, sends the signal and checks that the request finishes, the stream receives its final event, a new connection is refused, and a reached deadline cancels what is left. It is written with the first backend code.

### Domain modelling

These rules apply inside `domain`. API types and database rows in `adapters` are plain structs.

- **Aggregate.** It guards its invariants. State changes through a method named after the business operation, and fields are unexported. Callers tell the aggregate what to do; they do not read its state to decide for it. Read accessors exist for adapters to persist and map.
- **Value object.** A primitive becomes a value object when it carries a rule (validation, format, invariant) or is an identifier. It is built through a constructor that returns an error, and it is immutable. A collection with a rule of its own gets its own type on the same test. Free text and plain counters stay primitives.
- **Repository.** One per aggregate, declared in `application`. It reserves the next identifier from a database sequence, and the use case creates the aggregate with it, so an aggregate never exists without an identity.
- **One call per collaborator.** An expression talks to the object it holds, and that object does the rest.

### Errors

A broken rule is a typed error with a stable code and its parameters, declared beside the rule in `domain` or `application`. Each inbound adapter translates it: REST into `application/problem+json` with `code`, `params` and `trace_id`, MCP into a tool error carrying the same stable code and English message.

Every REST failure uses that format, including unknown routes, wrong methods and panics, which need their own handlers because they bypass Huma's error hook. No response carries a stack trace, SQL or an internal name.

### Database

PostgreSQL 18 is the only data service: one database, one schema.

- **Access.** A `postgres` adapter runs SQL through pgx, version 5.13.0 or later. Its queries are SQL files inside the adapter package, and sqlc generates the Go functions and the row structs from them. The generated code is committed and never edited.
- **Parameters.** A value reaches SQL as a parameter. No value is concatenated into SQL text.
- **Lists.** A list is fixed SQL: an optional filter is `(parameter IS NULL OR column = parameter)`, and each allowed sort is its own query. A list that does not fit is written by hand on pgx, with a comment saying why.
- **Pool.** One pgxpool in `platform`, shared by every domain area and by the session store. Its maximum size, the statement timeout and the idle-in-transaction timeout come from configuration. `platform` stops the session store's cleanup routine during the shutdown.
- **Migrations.** Numbered SQL files in `apps/api/migrations`, with zero-padded numbers (`00001_create_projects.sql`), embedded in the binary and applied by goose with its lock turned on. A migration has no down part. A mistake is fixed by a new migration, and a migration that has run outside a development machine is never edited.
- **Running them.** The `migrate` command of the API binary applies the pending migrations, as its own step before the API starts. The API never changes the schema.
- **Roles.** The migration role owns the tables. The API role reads and writes rows; on the audit log table it can only insert and select. The migration that creates a table grants the API role its permissions in the same file. The monitoring role reads the statistics of the server (`pg_monitor`) and no table; only the telemetry collector uses it.
- **Tables.** Foreign keys cross domain areas. An identifier is a `bigint` from the sequence of its aggregate type. An instant is `timestamptz`, a business date is `date`, money is `numeric`, and a fixed set of values is `text` with a `CHECK` constraint listing the values the API uses.
- **Extensions.** `pg_stat_statements` only. Another extension arrives in the migration of the first feature that needs it. One that the official PostgreSQL image does not ship starts with a new ADR.
- **Traces.** pgx is traced with otelpgx. The SQL text is recorded; query parameters never are.

### Jobs

Work that nobody waits for on a screen runs as a job. Jobs are stored in PostgreSQL by River.

- **Saving.** A use case saves a job through an interface it declares, inside its own transaction. The API saves jobs and never runs one.
- **Running.** The `worker` command of the API binary runs the jobs, as its own service. A job worker lives in `adapters/river`: it reads the job's arguments and calls one use case.
- **Arguments.** A job's arguments are identifiers. They never carry personal data or the value of a credential, because the River panel and the telemetry show them.
- **Idempotent.** A job can run more than once. A second run changes nothing that the first one already did.
- **Attempts.** Five by default. A job kind that needs more attempts, or more time than River's default timeout, declares it. When the attempts run out, the domain record is marked as failed, and a person can ask for a new attempt.
- **Queues.** Light work and heavy work run in separate queues, each with a limit of concurrent jobs taken from configuration.
- **Periodic work.** A periodic job is declared in the worker and scheduled by River. It tolerates a missed run.
- **Answer.** An operation that starts a job answers 202 with the resource, whose status says the work is in progress. The screen reads the resource again while it is.
- **Kept.** River deletes finished jobs. What must be kept is a domain record, never a job.
- **Shutdown.** The worker handles the signal as the API does. It stops taking jobs, gives the running ones a deadline taken from configuration, then cancels them. A cancelled job returns an error, so River runs it again later.
- **Tables.** River's tables come from SQL exported by the River CLI, one migration per River schema version. An update of River that changes its schema brings the new migration in the same commit. The worker uses the API role, which gets row permissions on these tables.
- **Traces.** River is traced with its OpenTelemetry middleware.
- **Panel.** The River panel is served by the API, behind the session, only to a user who administers the system, and outside the OpenAPI spec. In production it waits for the modelling of users. Until then it runs in development only.

### Files

Files live in S3 storage, served by Garage.

- **Access.** An `s3` adapter uses the AWS SDK for Go v2 and only what the S3 API defines: no bucket notification, nothing that one server alone offers.
- **Bucket.** One private bucket. An object key is random and never contains the file's name or an identifier.
- **Upload.** The use case checks the permission and returns a signed URL that expires in minutes. The browser sends the bytes to the storage. A large file goes in parts, each with its own signed URL.
- **Confirmation.** The browser then tells the API that the upload ended. The use case checks the object in the storage, its size and its type, and only then is the file available. An object over the size limit is removed.
- **Never confirmed.** A periodic job removes the uploads that were never confirmed.
- **Download.** The use case checks the permission and returns a signed URL that expires in minutes. No file passes through the API.
- **Address.** The browser reaches the storage through a host name of its own in the `Caddyfile`. The bucket accepts cross-origin requests from the app's origin only.
- **Processing.** A job reads a file from the storage and writes its result there. Nothing stays on the local disk after the job ends.

### E-mail

The system sends e-mail and receives none.

- **Record.** Every e-mail is a row: the recipient, the subject, what it refers to and its state, which is `queued`, `sent`, `failed` or `bounced`. The use case creates the row and saves the sending job in the same transaction.
- **Sending.** The job sends through one interface with two adapters, chosen by configuration: `smtp` (go-mail), which development and tests point at Mailpit, and `resend` (the Resend API), used in production.
- **Once.** The `resend` adapter sends the identifier of the e-mail record as the idempotency key and as a tag, and stores the identifier that Resend returns.
- **Bounces.** Resend reports a bounce to an operation under `/api`, outside the OpenAPI spec. The operation verifies the signature on the raw body with a secret taken from configuration, ignores an event it has already seen, and calls the use case that marks the e-mail as `bounced`.
- **Text.** Each e-mail is a template in the backend, with an HTML part and a plain-text part. Its texts are in translation files for every supported language, and the build fails when a language is missing a key. The language is the recipient's preferred language, or the default language when it is not known.
- **Privacy.** An address is in the e-mail record. It never appears in logs, spans, metric labels or job arguments.
- **Not built.** Attachments, links for people outside the system, receiving e-mail, and events other than a bounce.

### Control flow

Handle the exceptional case first and return, so the main path runs unindented to the end of the function. A function holds one level of nesting; a second level becomes a new function. Tooling rejects `else`.

### Naming

A domain concept uses the exact term from `GLOSSARY.md`, spelled in full, in every name that carries it: type, function, package, table, error code. Go's conventional short names (`ctx`, `err`, one-letter receivers) stay.

### Tests

The default seam is the use case, exercised with in-memory implementations of its interfaces. Each real adapter (Postgres, the S3 storage, SMTP, LLM) has its own tests against the real service.

A `postgres` adapter test runs against the development PostgreSQL container. Each test gets a temporary database, cloned by pgtestdb from a template that the migrations built, so tests run in parallel and never touch the development database.

A use case that saves a job is tested with an in-memory implementation of the interface that saves it. A job worker in `adapters/river` is tested with River's test helpers, in a temporary database. The `smtp` adapter is tested against Mailpit, and the `resend` adapter against Resend's test addresses, which deliver nothing.

### Enforcement

A rule that tooling can check fails the build: imports between layers and between areas, `else`, nesting depth, a mapping between API type, domain type and database row that leaves a field out. The tool is chosen with the first backend code.

The generated OpenAPI spec and the code generated by sqlc are committed, and the build fails when either is stale. A lint on the spec (vacuum) fails the build on `snake_case` and plural-path violations. A breaking-change report (oasdiff) is shown and does not fail the build. Once a check exists, its rule here shrinks to a pointer at the check.

### Deferred

Domain events, CQRS, event sourcing and a second bounded context each start with a new ADR, written when a concrete case needs one. So do WebSocket, GraphQL and gRPC as inbound adapters, live notifications (which need domain events), OAuth for agents, and MCP resources and prompts.

Running more than one instance of the API, and Kubernetes, are intended and not built. They start with a new ADR. The known obstacles are listed in ADR-0016.

How the document processing pipeline keeps its state, and where files are converted, are decided when the processors are implemented. How the system writes to a project's Git repository starts with its own decision. The domain area that owns e-mails is named with the first e-mail, and the language of a client's contact with the modelling of contacts. Progress pushed to the browser waits for a case that a periodic read does not serve.

Nothing is cached yet. A read measured as slow first gets an in-memory cache inside the API. A cache service (Redis, Valkey) and a connection pooler (PgBouncer) each start with a new ADR.

A lint for migrations and the column type for case-insensitive text are decided with the first case that needs one. How many dumps taken before a migration are kept, and where, is decided with the backup routine.

`ETag` with `If-Match`, `Idempotency-Key` and rate limiting are added to the first case that needs one, without a new ADR. A version in the path waits for the first consumer that does not ship with the backend, and reopens ADR-0009.

## Frontend (`apps/web`, `apps/prototype`, `packages/ui`)

These rules apply to `apps/web`. `apps/prototype` follows only "Naming", "Components" and "Prototype".

### Layout

```
apps/web/src/
├── routes/              # TanStack Router files: path, search parameters, loader
├── <domain area>/
│   ├── screens/         # one screen per route
│   ├── components/      # pieces of this area's screens
│   ├── data/            # collections, queries and writes
│   ├── locales/         # en-US.json, pt-BR.json, es.json
│   └── index.ts         # what routes and other areas import
└── shared/
    ├── api/generated/   # orval output, never edited
    ├── errors/          # translation of API error codes
    ├── stores/          # Zustand stores used by more than one area
    └── locales/         # texts used by more than one area
```

A domain area has the same name as in the backend and in `GLOSSARY.md`. A route file declares the path, the search parameters and the loader, and opens a screen from an area.

### Dependency rule

Dependencies point one way: `routes` → a domain area → `shared` → `packages/ui`.

| Code | Imports |
| --- | --- |
| `routes` | the `index.ts` of areas, `shared` |
| a domain area | its own files, the `index.ts` of other areas, `shared`, `packages/ui` |
| `data` of an area | the same, plus `shared/api/generated` |
| `shared` | `packages/ui`, third-party libraries; no area |
| `packages/ui` | third-party libraries; no app |

Only `data` imports the generated client. Screens and components never call it or build a URL.

### State

Each kind of state has one home:

- **Data from the API** lives in the area's `data` module, never in a store or in component state.
- **Filters, sorting and the current page** are search parameters in the URL, typed by TanStack Router and validated with Zod.
- **Interface state shared between screens** (theme, style variation, open panels, drafts) is a Zustand store, small and about one subject.
- **Everything else** is React state in the component that uses it.

### Data module

- **Names.** It exposes the area's reads and writes under glossary names. A module that only forwards a call is correct.
- **Bounded sets.** A set that grows with the business (clients, projects, configuration content) is a TanStack DB collection, loaded whole and queried in the browser with live queries.
- **Unbounded sets.** A set that grows with activity (hours logged, comments, the audit log) is read through TanStack Query with the API's cursor.
- **Optimistic writes.** Only for a simple, reversible edit. A creation and an operation with a business rule wait for the server. A refused write rolls back and shows the error.
- **Allowed actions.** A screen shows an action only when `allowed_actions` from the API includes it. It never derives that from the status.
- **Documentation.** TanStack DB is recent. Read its current documentation before writing a collection.

### Contract

The API client, the Zod schemas and the simulated responses in `shared/api/generated` are generated by orval from the committed OpenAPI spec. They are committed and never edited by hand. A change to the API is made in the backend and regenerated.

### Forms and validation

Forms use TanStack Form with the generated Zod schemas. A form validates only what the contract declares: required, format, length. Business rules are not repeated in the frontend. A 422 from the API is shown on the field it names.

### Errors

An API error is translated by its `code` and `params` through `shared/errors`, in the user's language, and shows its `trace_id`. A failure while rendering falls to an error screen that offers a reload. The frontend sends no telemetry.

### Values

The frontend formats values with the browser's `Intl`, in the user's language, and computes none: totals arrive ready from the API. No date or decimal library is added until a screen proves it needs one.

### Components

- `packages/ui` holds the design system: shadcn/ui components on Base UI primitives, styled with Tailwind CSS. Every colour, radius, font, spacing step and shadow is a CSS custom property.
- A specialised library (charts, rich text, drag and drop, calendar) is added when the first screen needs it. The one shadcn/ui already wraps is the first candidate.
- React Compiler memoizes components. Do not write `useMemo`, `useCallback` or `memo`.

### Naming

A domain concept uses the exact term from `GLOSSARY.md` in component, route, folder and variable names.

### Tests

The default test renders a whole screen in Vitest's Browser Mode, with the API simulated by MSW from the generated responses, and acts as a user does: it clicks, types and reads what is on the screen. A few end-to-end tests in Playwright cover critical flows against the real backend.

### Prototype

`apps/prototype` uses React, TanStack Router, react-i18next and `packages/ui`, with fake data in plain modules and its own translation files. A Zustand store may hold the theme and the style variation. Nothing is imported from it into `apps/web` (ADR-0006).

### Enforcement

- `vp check` (format, lint, types) and the screen tests fail the build. `vp staged` runs format and lint on changed files at each commit.
- Knip fails the build on unused files, exports and dependencies.
- The build fails when an import breaks the dependency rule, when the generated client is stale, or when a language is missing a translation key (ADR-0005).
- The size of the JavaScript of `apps/web`, initial and per route, has a limit that fails the build. It is measured on the Brotli copies the build writes (see Delivery), which is what the browser downloads. The numbers are set with the first real screen.
- Load and response times of the main screens (LCP, CLS, INP) are reported against the "good" Core Web Vitals thresholds (2.5 s, 0.1, 200 ms) and do not fail the build.

The tools for the import rule, the size limit and the time report are chosen with the first frontend code. Once a check exists, its rule here shrinks to a pointer at the check.

For investigating performance in a real browser, agents have the Chrome DevTools MCP: see [`docs/agents/browser-tools.md`](docs/agents/browser-tools.md).

### Deferred

A client for Server-Sent Events is written with the first streaming operation. Telemetry from the browser waits for users other than the maintainer.

The system health screen waits for the health report (see Health in the backend rules) and is designed in the prototype first. It reads the health report when it opens and on a refresh button, with no automatic refresh and no warning on other screens.

## Delivery (`deploy`)

How responses travel from the server to browsers and agents. Compression and cache headers are set in one place, the `Caddyfile` in `deploy/`.

### Compression

- **In Caddy only.** The `Caddyfile` turns on `encode` with its defaults for the whole site: zstd or gzip, chosen from the request's `Accept-Encoding`, for a response of 512 bytes or more with a compressible content type. That covers the files of the SPA, `/api` and MCP. Request bodies are not compressed.
- **Defaults kept.** The `match` block of `encode` is not written, because a custom block replaces the whole default list of content types. Error bodies (`application/problem+json`) are outside that list and stay uncompressed; they are small.
- **Streams.** A Server-Sent Events response is compressed like any other, and Caddy sends each event as it is written. The first streaming operation comes with a test that goes through Caddy and checks that events arrive one by one.
- **Brotli copies.** The build of `apps/web` writes a Brotli copy (`.br`) beside each text file it produces: JavaScript, CSS, HTML, SVG and JSON. Caddy serves the copy through `file_server` with `precompressed br`. Caddy does not compress Brotli on the fly, so a client that does not accept Brotli gets the file through `encode`.
- **Plugin.** A Vite plugin writes the Brotli copies. It is chosen with the first frontend code; none has been tried with Vite+ (Rolldown).
- **Risk accepted.** The size of a compressed response can reveal a secret when the same body carries text chosen by an attacker (the BREACH attack). No response is excluded from compression. The rule on credential values in the API conventions keeps the two apart.

### Caching

- **Files with a hash.** A file whose name carries a content hash is served with `Cache-Control: public, max-age=31536000, immutable`.
- **Files without a hash.** `index.html`, and any other file of the SPA without a hash in its name, is served with `Cache-Control: no-cache`, so the browser revalidates it and a deploy shows on the next load.
- **API.** Every response under `/api` carries `Cache-Control: no-store`, set by Caddy.

## Observability (`deploy`)

How telemetry is collected, kept and shown. The configuration of the collector and the dashboards, alert rules and contact point of Grafana are files in `deploy/`. What the backend emits is in Telemetry, in the backend rules. The reasons are in ADR-0012.

### Collection

- **Two containers.** An OpenTelemetry Collector gathers the telemetry and forwards it to `grafana/otel-lgtm`, which stores and shows it. The API and the worker send to the collector, never to `otel-lgtm`.
- **Collector.** It has no public port. It mounts, read-only, the server's file system and Docker's log folder. The Docker socket is never mounted.
- **Sources.** The traces and metrics of the API, the worker and Caddy. The output of every container, read from Docker's log files. The metrics of the server (disk, memory, CPU), PostgreSQL, Caddy and the file storage.
- **PostgreSQL.** The collector connects with the monitoring role (see Database).
- **File storage.** Its metrics endpoint is read with a token.
- **Docker logs.** The log files of every container rotate by size, set in `compose.yaml`.

### Privacy

The rule that telemetry carries nothing sensitive covers every container.

- **Caddy.** Tracing is on for the app's site, so the trace starts in Caddy and its identifier is in the access log. The access log carries the caller's IP address cut down to its network, and no query string.
- **Spans.** A span carries the caller's IP address cut down to its network, and no query string. The collector removes what a source cannot leave out.
- **PostgreSQL.** It logs an error without its detail (`log_error_verbosity = terse`), because the detail can carry the values of a row.

### Storage

- **Retention.** Logs, traces and metrics are kept for 30 days. The metrics also have a size limit. Both are set in `compose.yaml`, because the image keeps logs forever by default.
- **Volume.** One volume holds the stored telemetry. It is not in the backup.

### Grafana

- **Address.** Grafana has a host name of its own in the `Caddyfile`. No other port of `otel-lgtm` is published, because none of them has a login.
- **Login.** Through GitHub, for the maintainer's account only, as administrator. Anonymous access and the password form are turned off; the image starts with both on.
- **Files.** Dashboards, alert rules and the contact point are loaded from `deploy/` when Grafana starts. A change to one of them is a commit. Nothing made on Grafana's screens is kept.
- **Alerts.** Grafana evaluates the rules and warns by e-mail, through an SMTP server taken from the deploy configuration. The first rules are five: the API or the worker stopped sending metrics, the rate of 5xx is high, the disk is over 80% full, PostgreSQL is down or near its connection limit, and an optional part is down. The thresholds are set with the first real traffic.

### Development

The two telemetry containers are in a Compose profile that is off by default. Without them the API and the worker run as usual, and their logs are on the container output.

### Deferred

A monitor outside the server, the only thing that can warn when the whole server or `otel-lgtm` is down, waits for users other than the maintainer.

Logs over OTLP wait for stable exporters in the Go SDK.

The cost of LLM calls per client or per project is decided with the first LLM feature. If it matters, it is business data in the database, with its own ADR.

## Versions (the whole repository)

Every tool, library and container image runs at a version written in a committed file. Each version has one home, and an update is a change to that file.

### Where each version lives

| What | Home |
| --- | --- |
| Go toolchain | the `go` and `toolchain` lines of `go.mod` |
| Go libraries | `go.mod` and `go.sum` |
| Go command-line tools (sqlc, vacuum, oasdiff, the River CLI) | `tool` lines of `go.mod`, run with `go tool` |
| Node.js | `.node-version` at the root |
| pnpm | `packageManager` in the root `package.json` |
| JavaScript libraries and tools, Vite+ included | the catalog in `pnpm-workspace.yaml`, and `pnpm-lock.yaml` |
| Container images (PostgreSQL, Garage, Mailpit, Caddy, OpenTelemetry Collector, Grafana otel-lgtm) | the image tag in `compose.yaml` and in each Dockerfile |
| MCP servers for agents | `.mcp.json` |
| Docker, Task and the global `vp` | the prerequisites table of `README.md`, as a minimum version |

### Rules

- **Exact.** A version is exact: no range (`^`, `~`), no `latest`, no floating tag such as `postgres:18`. Only the prerequisites table holds minimums.
- **Catalog.** The three JavaScript packages take their versions from the pnpm catalog (`catalog:`), so a library has one version in the whole workspace.
- **Lockfiles.** `go.sum` and `pnpm-lock.yaml` are committed. The build installs from them and fails when they would change.
- **Nothing global.** A tool the build uses runs from its home above (`go tool`, `vp`), never from a copy installed on the machine. Docker, Task and the global `vp` are the only exceptions.
- **One-off commands.** A command that downloads what it runs (`npx`, `pnpm dlx`, `go run` with a module path) carries an exact version.
- **Copied code.** The shadcn/ui components in `packages/ui` have no version. Updating one means running the shadcn CLI again for that component and reviewing the diff.
- **Prose.** A document names a version only when a rule depends on it, as a minimum (PostgreSQL 18, pgx 5.13.0 or later). The version in force is read from its home. A number that must appear in two files is listed in both as such, as in [`docs/agents/browser-tools.md`](docs/agents/browser-tools.md).

### Updating

- **Listing.** `task outdated` lists what has a newer release: Go modules, JavaScript packages and container images. It is written with the first code.
- **One at a time.** An update is its own commit: one dependency, or one family released together (TanStack, OpenTelemetry), with the build and the tests passing.
- **Major versions.** A major version is updated after reading its migration notes. When it changes what an ADR decided, that ADR is updated in the same commit.
- **Pre-releases.** A library adopted before its stable release (TanStack DB) moves to the stable release as soon as there is one.

### Enforcement

The check that rejects a range, `latest` or a floating tag is chosen with the first code. Once it exists, the rules here shrink to a pointer at the check.
