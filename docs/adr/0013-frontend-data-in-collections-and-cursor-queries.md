---
status: accepted
---

# Frontend data in collections and cursor queries, with a store for interface state only

The React app reads and writes only through the REST API. Agents write its code, and without a rule each session picks its own place for state: data from the API gets copied into a store, a filter lives in a component, and two screens disagree about the same record. Every kind of state therefore has one home.

- **A bounded set from the API is a TanStack DB collection.** It is loaded whole, and live queries filter, sort and join it in the browser. A set is bounded when it grows with the business: clients, projects, configuration content.
- **An unbounded set is read through TanStack Query with the API's cursor.** A set is unbounded when it grows with activity: hours logged, comments, the audit log. ADR-0009 stands: the API keeps its opaque cursor.
- **Optimistic writes are for simple, reversible edits only.** A creation waits for the server, because the identifier comes from the database (ADR-0010). An operation with a business rule waits too, because the new status and the allowed actions come from the server. When the server refuses, the screen rolls back and shows the error.
- **Filters, sorting and the current page live in the URL**, as search parameters typed by TanStack Router.
- **Zustand holds interface state only**: theme, open panels, drafts. No data from the API enters a store. Stores are small, one per subject.
- **The client is generated from the contract.** orval reads the committed OpenAPI spec and produces typed fetch functions, Zod schemas and simulated responses for tests. The output is committed, is never edited by hand, and the build fails when it is stale.
- **A form validates what the contract declares** (required, format, length), with the generated Zod schemas, through TanStack Form. Business rules exist only in the backend; a 422 is shown on the field it names.
- **The frontend formats values and computes none.** Totals arrive ready from a query use case. Formatting uses the browser's `Intl`.

## Considered Options

- **TanStack Query for everything.** Stable and the data library models know best. The maintainer chose TanStack DB for bounded sets; TanStack Query stays underneath, since TanStack DB runs on it.
- **TanStack DB for every read, loading on demand.** One path for agents, but on-demand loading pages by offset, which ADR-0009 rejected.
- **Router loaders alone.** One library fewer, with no mutations, no infinite lists and no fine invalidation.
- **No store library until a case needs one.** Avoids the risk of duplicated API data altogether. The maintainer chose to have Zustand from the start, so the rule above contains that risk instead.
- **Jotai.** More granular, less known to models, same risk.
- **Zod schemas written by hand per form.** More control over each form, but the rules exist in two places and drift.
- **React Hook Form.** The form library models know best, with weaker typing of field names. TanStack Form types field names and values end to end, the same criterion that chose TanStack Router in ADR-0002.
- **Other generators.** `@hey-api/openapi-ts` is before 1.0 and had gone about five months without a release. `openapi-fetch` is in maintenance mode and generates no schemas. Kubb is active and covers Server-Sent Events, but is newer and less known to models.
- **Totals computed in the browser.** Immediate feedback while typing, but it needs a decimal library and puts each calculation in two places.

## Consequences

- TanStack DB was a release candidate (1.0.0-rc) when this was decided, and models know it little. An agent reads its current documentation instead of relying on memory.
- There are two ways to read. The data module of each area hides which one a set uses, so moving a set that outgrew its collection to a cursor query changes that module only.
- Loading a whole set through a cursor API means following the cursor to the end.
- TanStack Form 2 was in alpha at the time, so a migration lies ahead.
- orval does not generate a client for `text/event-stream`, so the first streaming operation is written by hand.
- orval 8 was first run against the spec Huma generates with the setup screen, and works on it unchanged: the client, the Zod schemas and the simulated responses for MSW. What its configuration (`apps/web/orval.config.ts`) had to say: the fetch client calls a fetch of the app (`shared/api/request.ts`) with `includeHttpResponseReturnType: false`, so a function resolves with the body of a success and rejects with an `ApiError` that carries the error body; without it, an error body comes back as data. `generateEachHttpStatus` writes a simulated answer for every status the spec declares, the 429 of Caddy included. Its own `--outdated` flag is the check that the committed files are current. The files it writes are not formatted, so the formatter and the linter leave them out, and Knip ignores the functions no screen uses yet. With MSW 3, orval 8.48 still imports `http` and `HttpResponse` from the root of `msw`, which MSW 3 keeps exporting though its documentation moved them to `msw/http`.
- A form does not hand a generated schema to TanStack Form as it is: `checkTheContract` (`shared/errors/problems.ts`) runs it and names each broken field with the code the API answers with (`required`, `too_long`, `invalid`), so a refusal of the browser and one of the API are translated the same way.
- The generated Zod schemas work in TanStack Form 1.33, tried on the form of the setup, a form of steps with nested objects. What had to be adjusted: a form holds an empty text for an optional field the body leaves out, and an empty e-mail breaks `zod.email()`, so `checkTheContract` takes a function that turns the typed values into the body before it checks them; a step checks its part of the body with the nested schema the generated one holds (`ConcludeSetupBody.shape.consultancy`, `.omit()` for a property no step types), from the `onSubmit` validator of a `form.FormGroup`, which names each field relative to the group as the schema does; and a field forgets its problems when it unmounts, so a 422 answered on the review is put on the fields of the earlier steps with `form.setErrorMap`. The form itself has no validator: the validators of a form also run, filtered, when a group submits, which would send its `onSubmitAsync` from every step.
- The browser sends no telemetry. An API error shows the `trace_id` from its body.
- A specialised interface library (charts, rich text, drag and drop) is chosen when the first screen needs it, starting with the one shadcn/ui already wraps.
