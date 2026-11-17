# 02 - Backend guards: the build fails on a broken rule

Status: ready-for-agent
Blocked by: 01

Read `../spec.md` and, in `CODING_STANDARDS.md`, Dependency rule, Control flow, Enforcement in the backend rules, and Versions.

## What to build

The rules of the backend that tooling can check are checked, on the machine and in the pipeline, from the first lines of Go.

- golangci-lint as the linter and gofumpt as the formatter, at exact versions, run with `go tool`.
- The dependency rule as a check: `domain`, `application`, `adapters`, `kernel` and `platform` import only what the table in Dependency rule allows, one area reaches another only through its `application`, and no adapter package imports another.
- `else` rejected, and a second level of nesting rejected. Where the linter has no rule for one of them, a small analyzer in the repository checks it.
- A check that fails on a version that is not exact, across what exists today: the Go module and its tools, the image tags of the Compose files and of any Dockerfile, the actions and the tools of the workflows, the catalog of `apps/web` and `.mcp.json`. It rejects a range, `latest` and a floating tag.
- All of it runs in `task check` and in the backend job of the pipeline.
- `CODING_STANDARDS.md` records the tools chosen, and the rules that a check now covers shrink to a pointer at the check, as Enforcement asks.

The mapping check, the one that fails when a mapping between API type, domain type and database row leaves a field out, is chosen here and turned on with the first mapping, in ticket 05.

## Acceptance criteria

- [ ] A Go file that is not formatted fails `task check`.
- [ ] An import from `domain` to `application`, from `application` to an adapter, from one adapter package to another, and from one area into the `domain` of another, each fail the check. A test or a fixture proves each case.
- [ ] An `else` and a second level of nesting each fail the check.
- [ ] A version written as a range, as `latest` or as a floating tag fails the check, in each kind of file listed above.
- [ ] The existing code of the repository passes every check.
- [ ] The pipeline fails the same way the machine does.
- [ ] `CODING_STANDARDS.md` names the linter, the formatter and each check.
