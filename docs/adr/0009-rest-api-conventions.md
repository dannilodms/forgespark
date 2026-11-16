---
status: accepted
---

# REST API conventions chosen for agents and a single deployment

The REST API is consumed by the React app, and the same use cases are reached by coding agents through MCP. Every consumer is built from this repository and shipped in the same deploy as the backend. The conventions below were picked from a general list of REST practices, keeping what helps agents and one consultancy and leaving out what protects public APIs with partners and several replicas.

- **Names.** `snake_case` in JSON bodies, query parameters, error fields and MCP tool arguments, so a term is spelled the same in the database, the API and a tool. Huma has no central switch for this (the name comes from each field's tag), so a lint on the generated spec enforces it.
- **Writes are named operations.** A state change is a `POST` whose path ends in the verb of the use case, such as `POST /projects/{id}/cancel`. Resources are plural nouns. Editing plain details is a `PUT` carrying the full set of editable fields. There is no `PATCH`.
- **Strict input.** An unknown property in a body or an unknown query parameter is rejected with 422.
- **Errors.** Every failure is RFC 9457 `application/problem+json`, extended with `code` (the stable code of ADR-0005), `params` and `trace_id`. `type` keeps its default. 400 is malformed input, 422 is failed validation or a broken business rule, 409 is a conflict. Unknown routes, wrong methods and panics do not pass through Huma's error hook, so they get their own handlers and a test that checks the format.
- **Collections.** A list is an envelope with an opaque cursor. `limit` has a default and a ceiling, sorting accepts only listed fields, filters are query parameters, and the total is computed only when asked for. Clients cannot pick fields. MCP tools that list follow the same rules.
- **Allowed actions.** When a screen or a tool needs to know what can be done to a resource now, the read returns `allowed_actions`, computed by the use case.
- **Values.** Timestamps are ISO 8601 in UTC, business dates carry no time, money is a decimal string beside its currency, and enums are lowercase strings. A single resource is the root of the body, with no `success` wrapper. Creation answers 201 with `Location`, asynchronous work answers 202.
- **Evolution.** No version in the path and no deprecation policy. A breaking change is allowed when every consumer changes in the same commit; the generated TypeScript client makes the frontend fail to compile otherwise. The generated spec is committed and a check fails when it is stale, so contract changes show up in review. A lint (vacuum) blocks convention violations, and a breaking-change report (oasdiff) is shown without blocking.
- **Limits.** Request bodies have a size limit and a read timeout. `X-Forwarded-*` headers from Caddy are honoured. Every outbound call has a timeout, and only idempotent calls are retried.

## Considered Options

- **`PATCH`, as JSON Merge Patch or JSON Patch.** Huma can generate it, but a generic patch hides which business operation is meant, and the aggregate rule in ADR-0007 is that state changes through a method named after the operation. It also needs a way to tell an absent field from a null one.
- **Actions as noun sub-resources** (`POST /projects/{id}/cancellations`). Closer to classic REST, but the verb form matches the use case and the MCP tool by name.
- **Ignoring unknown fields.** It keeps old clients working across versions, which does not apply here, and it lets an agent's misspelled field or filter pass in silence.
- **The error code inside `type`, as a documentation URL.** More faithful to the RFC, but it needs those pages to exist and the frontend would parse the code out of a URL.
- **Offset pagination.** Allows jumping to a page, but the envelope is contract and moving to cursors later breaks clients.
- **HAL links.** A hypermedia format does not fit a typed client generated from OpenAPI. Its one useful idea is kept as `allowed_actions`.
- **Versioned paths, blocking breaking-change checks and deprecation windows.** They protect consumers that ship separately, and none exist.
- **camelCase.** Idiomatic in the frontend, but the same term would be spelled two ways.

## Consequences

- **Deferred, with the risk accepted:** `ETag` with `If-Match`, so a `PUT` of details can overwrite a change made since the form was loaded; and `Idempotency-Key`, so an agent that repeats a call after a timeout can create a duplicate. Each is added to the first resource where the problem shows.
- Rate limiting exists only for the login code, which can be guessed (ADR-0011): a limit per e-mail address in the use case and a limit per IP address in Caddy. The rest of the API has none: access tokens are long and random.
- The first consumer that does not ship with the backend, such as an agent of ADR-0004, reopens versioning and makes the breaking-change report blocking.
- Not adopted because nothing here needs them: quotas and tiers, webhooks, a sandbox, generated SDKs, a BFF, CORS, an offline queue, 304 caching, circuit breakers and expand/contract migrations.
- The lint rules have not been tried on a spec generated by Huma. Plural path segments need a custom rule.
