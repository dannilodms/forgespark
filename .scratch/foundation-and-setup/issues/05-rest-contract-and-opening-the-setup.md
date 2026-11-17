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

- [ ] A valid code answers 200 with what it opens, `first_setup`, and when it expires.
- [ ] A wrong code, an expired one and an ended one get the same 422, and the stored code is unchanged afterwards.
- [ ] A code typed in lower case, without hyphens, with spaces, or with O, I and L in place of 0 and 1 is accepted.
- [ ] A validation error, an unknown route, a wrong method and a panic each answer in the error format, with a stable code and a `trace_id`. Tests through HTTP cover the four.
- [ ] A body with an unknown property is rejected, and a body over the size limit is rejected.
- [ ] The setup code appears in no path, query parameter or header of the spec, and in no log line or span.
- [ ] The committed spec matches the code, and a stale spec fails the check.
- [ ] The lint passes on the spec Huma generates, the `$schema` property it adds included. What had to be configured for it is noted in ADR-0009 if it differs from what the ADR expected.
- [ ] A path with a singular noun that is not on the list fails the lint.
- [ ] The breaking-change report runs and does not fail the build.
