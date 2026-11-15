# Coding standards

How code in this repository is written. The reasons and the rejected alternatives are in [ADR-0007](docs/adr/0007-backend-organized-by-domain-area.md) and [ADR-0008](docs/adr/0008-use-cases-served-by-several-inbound-adapters.md).

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
│       ├── postgres/   # outbound
│       └── minio/      # outbound
├── kernel/             # value objects used by more than one domain area
└── platform/           # shared technical infrastructure
```

A **domain area** is one part of the business with its own aggregates. It is the deep module of the backend: its public surface is its use cases, and everything else is internal to it.

An **inbound adapter** calls use cases on behalf of an outside caller: REST for the React app and for the agents of ADR-0004, MCP for coding agents. An **outbound adapter** implements an interface that `application` declares: Postgres, MinIO, an LLM provider. `adapters` has one package per technology, one level deep, and an area creates only the ones it uses.

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
- **Writes.** One use case is one transaction, opened and closed through an interface. It loads the aggregate from its repository, calls the aggregate, and saves it whole. It may save several aggregates of its own area. An effect on another area or on an external service runs after the commit.
- **Reads.** A query use case calls a read-only interface that returns data already shaped for the caller. It builds no aggregate.

### Inbound adapters

- **REST is complete.** Every use case has a REST operation. Streaming to the browser is a Server-Sent Events operation in `adapters/rest`, added with the first use case that needs it.
- **MCP is a chosen subset.** A use case becomes an MCP tool only when an agent has a real use for it. MCP exposes tools only.
- **One tool, one use case.** A tool is designed for what the agent is trying to do, not copied from a REST operation. When it needs a composed result or a different shape, that is a new use case.
- **Authentication in the adapter, authorization in the use case.** The inbound adapter authenticates the caller (a session or an access token) and passes the actor to the use case: the user, the scopes, and the access token when there is one. The use case checks the scopes. The actor type lives in `kernel`.
- **MCP language.** Tool names and descriptions are in English. Configuration content is returned in the preferred language of the access token's user.
- **Credentials over MCP.** A tool that returns the value of a credential requires the scope for reading credential values, returns one credential per call, and its use case records the read. No listing returns a value.
- **Nothing ahead of need.** No package, mount point or abstraction is written for an inbound adapter that does not exist.

### Domain modelling

These rules apply inside `domain`. API types and database rows in `adapters` are plain structs.

- **Aggregate.** It guards its invariants. State changes through a method named after the business operation, and fields are unexported. Callers tell the aggregate what to do; they do not read its state to decide for it. Read accessors exist for adapters to persist and map.
- **Value object.** A primitive becomes a value object when it carries a rule (validation, format, invariant) or is an identifier. It is built through a constructor that returns an error, and it is immutable. A collection with a rule of its own gets its own type on the same test. Free text and plain counters stay primitives.
- **Repository.** One per aggregate, declared in `application`.
- **One call per collaborator.** An expression talks to the object it holds, and that object does the rest.

### Errors

A broken rule is a typed error with a stable code and its parameters, declared beside the rule in `domain` or `application`. Each inbound adapter translates it: REST into the status and response format of ADR-0005, MCP into a tool error carrying the same stable code and English message.

### Control flow

Handle the exceptional case first and return, so the main path runs unindented to the end of the function. A function holds one level of nesting; a second level becomes a new function. Tooling rejects `else`.

### Naming

A domain concept uses the exact term from `GLOSSARY.md`, spelled in full, in every name that carries it: type, function, package, table, error code. Go's conventional short names (`ctx`, `err`, one-letter receivers) stay.

### Tests

The default seam is the use case, exercised with in-memory implementations of its interfaces. Each real adapter (Postgres, MinIO, LLM) has its own tests against the real service.

### Enforcement

A rule that tooling can check fails the build: imports between layers and between areas, `else`, nesting depth. The tool is chosen with the first backend code. Once a check exists, its rule here shrinks to a pointer at the check.

### Deferred

Domain events, CQRS, event sourcing and a second bounded context each start with a new ADR, written when a concrete case needs one. So do WebSocket, GraphQL and gRPC as inbound adapters, live notifications (which need domain events), OAuth, and MCP resources and prompts.

## Frontend (`apps/web`, `apps/prototype`, `packages/ui`)

A domain concept uses the exact term from `GLOSSARY.md` in component, route and variable names. The remaining frontend standards are not written yet.
