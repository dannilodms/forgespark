# 05 - REST contract, and opening the setup with a setup code

Status: ready-for-agent
Blocked by: 03, 04

Read `../spec.md`, ADR-0003, ADR-0009 and, in `CODING_STANDARDS.md`, Inbound adapters, API conventions, Errors and Enforcement in the backend rules.

## What to build

The first REST operation, and with it everything every later operation relies on. A caller sends a setup code and learns whether it opens the setup.

- The use case that opens the setup: it checks a setup code and answers what it opens and when it expires. It saves nothing, and a wrong code changes nothing. Reading a typed code ignores case, hyphens and spaces, and reads O as 0 and I or L as 1.
- `POST /api/setup/open`, with the setup code in the body. A code that is wrong, expired or ended gets one answer: 422 with the code `setup_code_invalid`.
- The REST adapter on Huma, under `/api`, with its documentation page turned off.
- The error format on every failure: `application/problem+json` with `code`, `params` and the `trace_id` of the request. That includes a failed validation, an unknown route, a wrong method and a panic, which need handlers of their own. No response carries a stack trace, SQL or an internal name.
- Strict input: an unknown property of the body and an unknown query parameter are rejected. A size limit and a read timeout on bodies.
- The OpenAPI spec generated from Go and committed, with a check that fails when it is stale.
- The lint of the spec with vacuum: `snake_case`, and plural paths, with `/api/setup` and `/api/health` listed as the singular ones. The breaking-change report with oasdiff, shown and not failing.
- The mapping check chosen in ticket 02, turned on.
- Task targets for generating the spec, and all the checks in `task check` and in the pipeline.

The check of the origin of a request arrives with the session, in the second effort.

## Acceptance criteria

- [x] A valid code answers 200 with what it opens, `first_setup`, and when it expires.
- [x] A wrong code, an expired one and an ended one get the same 422, and the stored code is unchanged afterwards.
- [x] A code typed in lower case, without hyphens, with spaces, or with O, I and L in place of 0 and 1 is accepted.
- [x] A validation error, an unknown route, a wrong method and a panic each answer in the error format, with a stable code and a `trace_id`. Tests through HTTP cover the four.
- [x] A body with an unknown property is rejected, and a body over the size limit is rejected.
- [x] The setup code appears in no path, query parameter or header of the spec, and in no log line or span.
- [x] The committed spec matches the code, and a stale spec fails the check.
- [x] The lint passes on the spec Huma generates, the `$schema` property it adds included. What had to be configured for it is noted in ADR-0009 if it differs from what the ADR expected.
- [x] A path with a singular noun that is not on the list fails the lint.
- [x] The breaking-change report runs and does not fail the build.

## Comments

### What was built

- **The use case.** `OpenSetup` in `apps/api/internal/consultancy/application`. It reads the code a person typed, finds the setup code by its hash, asks it whether it works at the time of the clock, and answers what it opens, the first setup for now, and when it expires. It saves nothing and opens no transaction.
- **Reading a typed code.** `ReadPlainSetupCode` in `consultancy/domain`: case, hyphens and every kind of space do not count, and O is read as 0, I and L as 1. A text that is not 20 characters of the alphabet after that is answered as a wrong code is.
- **Whether a code works.** `(*SetupCode).Works(at)`: not ended, and before its expiry. At the hour exactly it no longer works.
- **The broken rule.** `SetupCodeInvalidError` in `consultancy/domain`, with the code `setup_code_invalid`: one error for a code that is wrong, expired or ended.
- **The repository** gained `FindByHash`, which finds a setup code, ended or not, and locks nothing. Its SQL is in `setup_codes.sql`, and sqlc generated the function.
- **The REST foundation**, `internal/platform/restapi`: Huma 2.40.0 on the router of the standard library, with the documentation page, the copy of the spec and the schemas turned off. `restapi.Register` adds an operation with the broken rules it answers with, sets the limits of its body and writes the codes of each status into the spec.
- **The error format.** Every failure answers `application/problem+json` with `title`, `status`, `detail`, `code`, `params` and `trace_id`. A failed validation adds `errors`, one entry for each field, with its `location`, a `code` (`required` or `invalid`), `params` and a `message`. A broken rule answers with its own code under the status the operation lists it at. Any other failure takes its code from the name of its status, and a request that does not follow the contract is `validation_failed`. A 5xx carries the code and the `trace_id` alone, and its raw error goes to the one line `telemetry.Requests` writes.
- **Unknown routes, wrong methods and panics.** One handler on `/api` and `/api/` answers 404, or 405 with `Allow` when the path is a route under another method. A middleware of the adapter answers a panic with a 500 and hands its value and its stack to the logs.
- **Strict input and limits.** An unknown property of the body and an unknown query parameter answer 422. A body over one megabyte answers 413, and one that takes more than five seconds to arrive answers 408.
- **`POST /api/setup/open`**, in `internal/consultancy/adapters/rest`, registered in `mountREST` in `cmd/api/rest.go`, where each use case is wired. 200 with `opens` and `expires_at`, 422 with `setup_code_invalid`.
- **The spec.** `apps/api/openapi.yaml`, OpenAPI 3.1.0, generated from what `mountREST` registers. `task api:spec` writes it and `task api:spec:check` fails when the committed file differs, naming the first line that does.
- **The lint.** `task api:spec:lint` runs vacuum 0.41.0 with the rules in `apps/api/tools/speclint`: `snake_case` in the properties of every schema and in query and path parameters, and plural paths, by a function written there, with `/api/setup` and `/api/health` listed as the singular ones.
- **The breaking-change report.** `task api:spec:breaking` runs oasdiff 1.38.0 against `origin/main` and never fails.
- **The mapping check.** `exhaustruct_v5` is on in `apps/api/.golangci.yml`, for the types of adapter packages, with a fixture and an assertion in the test of the linter.
- **Task and the pipeline.** The four targets of the spec are in `task check` and in the backend job of `ci.yml`. The report runs on a pull request, after a step that fetches the branch it merges into.
- **The test of sensitive values** gained the scenario of a request that carries a setup code.
- **The documents.** `CODING_STANDARDS.md`: how an operation is registered, the error format, the limits, the `$schema` links, the repository method that locks nothing, the tests of the REST adapter, the mapping check and the checks of the spec. Both READMEs: the four targets and where the contract is. ADR-0009: what the lint needed, the `$schema` links and the three failures that bypass Huma.

### Decided while building

- **Huma's `$schema` links are turned off.** Huma adds to every body a `$schema` property and to every response a `Link` header. The value is built from the `Host` or the `X-Forwarded-Host` the caller sent, so every response would repeat something the caller chose, which Credential values in API conventions forbids for the responses that will carry a credential. A body now carries only what its Go type declares. The lint still accepts `$schema`, and a test proves it on a spec generated with Huma's defaults, so turning the links back on is one line. See "Left for a decision".
- **A failure that is not a broken rule takes its code from the name of its status**: `not_found`, `method_not_allowed`, `request_entity_too_large`, `internal_server_error`. That gives `too_many_requests` for the 429 ticket 07 writes in Caddy. The one exception is `validation_failed`, because a 422 is also the status of a broken rule.
- **An operation lists the broken rules it answers with**, by an example of each error under its status. One it does not list answers 500. So the spec names every code an operation answers with, and a new code cannot reach a client unnoticed.
- **A broken rule is any error with `Code() string`.** `restapi` asks for the method and nothing else, so `domain` and `application` import nothing to declare one. `Params() map[string]any` is optional.
- **A field of `errors` has a code.** `required` and `invalid` for what the contract refuses, and room for the code of a rule of the domain about a field, which ticket 06 brings.
- **A missing required property is named as its field.** Huma places it at the object that lacks it. `restapi` reads the name from Huma's message, with the format Huma exports, and answers `body.code`.
- **The value that was refused is never repeated.** Huma puts it in the error of a field, and puts the whole body in the error of a body that is not JSON. Both are dropped, so a setup code is in no answer.
- **`trace_id` is always there**, and `params` is always an object.
- **The spec is written by the test that compares it**, run with a flag. The operations are registered in `mountREST`, in package `main`, which only a test of that package can call. A fourth command of the binary would also do it, and the spec of the effort names three.
- **The spec is YAML**, which reads better in a review.
- **`info.version` is `unversioned`.** ADR-0009 gives the contract no version, and how a build reads `VERSION` is not decided yet.
- **`FindByHash` locks nothing.** Opening the setup saves nothing, so a lock would only make a caller wait. It is a method of the repository and returns the aggregate, so that `Works` stays the one place that says whether a code works. Domain modelling gained the sentence.
- **Plural means "ends in `s`"** in the lint. The last segment of a path whose only method is `POST` is taken for its verb.
- **The lint runs only the rules of this repository.** vacuum's recommended set judges the documentation of a public API.
- **The limits of a body are constants**, one megabyte and five seconds, in `cmd/api/rest.go`. No rule asks for a variable.
- **Outside `/api` an unknown path keeps the plain 404 of the router.** Caddy forwards only `/api`, and liveness and readiness answer with a status alone.
- **`opens` lists `first_setup` only.** Ticket 12 adds `new_administrator` with the code that returns it.

### What was checked

- **The binary, run as a process**, on a database of its own in the development PostgreSQL: `migrate`, then `issue-setup-code`, then the API on port 8085, with `curl`. The code as printed and the same code in lower case with no hyphen: 200, `{"opens":"first_setup","expires_at":"2026-11-17T16:42:10.985323Z"}`, the time the command had printed. A wrong code: 422, `application/problem+json`, `"code":"setup_code_invalid"`, `"params":{}` and a `trace_id` of 32 characters. An unknown property: 422, `validation_failed`, `"location":"body.step"`. No code: `"location":"body.code"`, `"code":"required"`. A body that is not JSON: 400, `bad_request`. `/api/setup/nothing`: 404, `not_found`. `GET /api/setup/open`: 405, `method_not_allowed`, `Allow: POST`. `/docs`: 404. The code again after the wrong ones: 200. The logs of the API held the start line and nothing else, and the code in none of its forms.
- **The use case**, with the in-memory implementations and the clock stopped: a valid code at 59 minutes, a wrong one, one at the hour exactly, one ended by a new code, with what is stored compared before and after. The code `01234-56789-ABCDE-FGHJK` typed in eight forms, with `O`, `I` and `L` for its `0` and `1`. Five texts that are not a code.
- **Through HTTP, on the binary** (`cmd/api/open_setup_test.go`), each test on its own database: the 200 with the `expires_at` the database holds, the same 422 for a code that is wrong, that is not a code, that expired and that was ended, with every row of every table the same before and after. Three wrong codes and then the right one, which still works. Six typed forms. An unknown property, an unknown query parameter, no code, an empty code and a code four times too long, each a 422 that names its field and does not repeat the code. A body over one megabyte: 413. An unknown route and a wrong method.
- **Through HTTP, on the foundation** (`platform/restapi`), on an operation of the test: a validation error, an unknown route, a wrong method and a panic, each as the exact JSON it answers, with the `trace_id` of the `traceparent` the test sent. A declared broken rule with its status, code and parameters. An error of the infrastructure and an undeclared broken rule, each a 500 that says nothing of it and one log line that does. A caller that stalls on its body: 408. Another media type: 415. A body at the limit and one over it. No documentation page, no copy of the spec, no schema, no `$schema` and no `Link` header.
- **The three failures bypass Huma's hook.** The tests of the unknown route, the wrong method and the panic failed before the handlers were written: the router answered `404 page not found` and `Method Not Allowed` as plain text, and the panic a 500 with no body.
- **The spec.** `task api:spec` wrote it. With `maxLength: 64` changed to `32` in the file, `task api:spec:check` failed with `The first difference is at line 30`, the line of the file and the line from the code. Generated again, it passes. A test generates the spec six times and gets the same text.
- **The setup code in the spec.** A test walks every path under `/api/setup`: none has a variable, and no operation takes a parameter.
- **The lint on the spec Huma generates.** `task api:spec:lint` passes on `apps/api/openapi.yaml`. A test generates a spec with Huma's defaults, checks that it holds `$schema`, and the lint passes on it.
- **The lint fails on what it should.** On the fixture `unconventional.yaml`: `/api/project` (a singular noun not on the list), `/api/projects/{projectId}/detail` under `GET` (a singular last segment that is no verb), `legalName`, `PostalCode` inside a nested object, the query parameter `sortBy` and the path parameter `projectId`. Six errors, status 1. On `conventional.yaml`, with `/api/setup`, `/api/health`, `POST /api/setup/open` and `$schema`: it passes.
- **That the tests of the lint bite.** With `/api/health` taken off the list and `$schema` taken out of the pattern, the two tests that expect a pass failed, naming each. Undone.
- **The breaking-change report.** `task api:spec:breaking SPEC_BASE=HEAD` on the committed spec: `No changes detected`, status 0. With a required request property added and a response property removed in the working tree: `2 changes: 2 error`, each named, status 0. Against `origin/main`, which has no spec yet: the error of oasdiff is shown and the target still ends with status 0.
- **The mapping check.** With `exhaustruct_v5` on, golangci-lint reports 0 issues on the code. On the fixture: `mapping.go:21:9: rest.Body is missing field LegalName (exhaustruct_v5)`, and nothing for the literal that fills every field, the zero value returned beside an error, or the type of the domain.
- **The setup code in the logs and the spans.** `TestTheLogsAndTheSpansOfOpeningTheSetupNeverCarryTheCode` sends the code in four forms, a wrong code, the code in four requests the API refuses, one of them with the code in the address, and the code in a request that fails with a 5xx, with the logs at `debug` and the spans collected. Neither holds any form, and no answer repeats it. With the handler changed to wrap the error with the code, the test failed with `the logs carry a sensitive value`, naming `the setup code`. Undone.
- **That the other tests bite.** With `RejectUnknownQueryParameters` off, the test of the unknown query parameter failed. With the hook of Huma left on, the test of the schema link failed. With `FOR UPDATE` added to the query, `TestFindingASetupCodeByItsHashLocksNothing` failed with `55P03`. Each change was undone.
- **On Linux**, in a `golang:1.27.4` container against the same PostgreSQL: the check of the spec, vacuum with the function (`Quality Score: 100/100`), oasdiff on two files (one breaking change shown, status 0), golangci-lint (0 issues), the guard, and every Go test, the tests of the lint included.
- **The workflow file.** actionlint 1.7.12 and zizmor 1.32.0 (offline), each run from its container image on the changed `ci.yml`: no finding.
- **The versions.** `task versions:check` passes with `vacuum.mod` and `oasdiff.mod`. Huma 2.40.0, vacuum 0.41.0 and oasdiff 1.38.0 were the newest releases `go list -m -versions` showed.
- **`task check`** passes on the branch. `feat/foundation-and-setup` had nothing new to merge.

### What was not checked

- **The pipeline was not run.** Nothing was pushed. The new steps call Task targets that ran on this machine and in a Linux container.
- **oasdiff reading a Git revision on Linux, and the step that fetches the base branch.** On this machine `task api:spec:breaking` read `HEAD:apps/api/openapi.yaml`. In the Linux container the checkout is a Git worktree whose `.git` names a Windows path, so only the comparison of two files ran there. The `git fetch` of the base branch in `ci.yml` was not run.
- **The read timeout and the panic through the binary.** Both are tested on the foundation, behind the same `telemetry.Requests` the command mounts. The binary has no operation that panics, and the test of the binary does not hold a body for five seconds.
- **The development database.** `task migrate` was not run against `software_factory`, to leave the database every worktree shares where the others expect it. `task api` against it answers "not ready" until someone does.
- **Behind Caddy.** Nothing forwards `/api` yet: ticket 07.
- **O, I and L through HTTP with a code that has a 0 or a 1.** The command draws the code, so the test of the binary replaces them only when the code has one. The use case test covers it with a known code.

### Found, and left for a decision

- **`$schema`.** The acceptance criterion reads "the lint passes on the spec Huma generates, the `$schema` property it adds included". The lint does pass on such a spec, and a test keeps proving it, but the spec of this API has no `$schema`: the links are turned off, for the reason under "Decided while building". If the links are wanted, delete `config.CreateHooks = nil` in `restapi.New`, set `SchemasPath` to `/api/schemas` and run `task api:spec`. Every body then gains the property, and so does every type ticket 08 generates.
- **The `trace_id` of the 429.** Caddy answers it before the API sees the request, so there is no trace. The schema requires the property and accepts an empty string. Ticket 07 decides what Caddy writes there.
- **A rule of the domain about a field** (ticket 06) has the shape of its answer ready, an entry in `errors` with the code of the rule, and nothing that builds it yet.
- **Outside `/api`** an unknown path answers the plain 404 of the router, not the error format.
- **The cache of the pipeline** now also names `vacuum.sum` and `oasdiff.sum`. `sqlc.sum` is still out, as ticket 04 left it.
- **The first build of vacuum on Linux took eight minutes** in the container, most of it compiling. On Windows the two tools together took about ten minutes.
- **The breaking-change report prints an error until `main` holds the spec**, and the target ignores every error of oasdiff, a real malfunction included.
