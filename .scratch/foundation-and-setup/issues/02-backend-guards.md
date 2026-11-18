# 02 - Backend guards: the build fails on a broken rule

Status: resolved
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

- [x] A Go file that is not formatted fails `task check`.
- [x] An import from `domain` to `application`, from `application` to an adapter, from one adapter package to another, and from one area into the `domain` of another, each fail the check. A test or a fixture proves each case.
- [x] An `else` and a second level of nesting each fail the check.
- [x] A version written as a range, as `latest` or as a floating tag fails the check, in each kind of file listed above.
- [x] The existing code of the repository passes every check.
- [ ] The pipeline fails the same way the machine does.
- [x] `CODING_STANDARDS.md` names the linter, the formatter and each check.

## Comments

### What was built

- **The linter and the formatter.** golangci-lint 2.15.0, configured in `apps/api/.golangci.yml`: its standard set (errcheck, govet, ineffassign, staticcheck, unused), nolintlint, and gofumpt 0.12.0 as its formatter. It runs with `go tool` from a module file of its own, `apps/api/golangci-lint.mod`.
- **The guard.** `apps/api/tools/guard`, a command of the API module with three checks, one package each, and their tests beside them. It prints one line for each finding, `file:line: what to do`, and ends with status 1.
  - `imports`: the table of the Dependency rule, on every import under `internal`, test files included. It reads the place of a package from its path, so it holds no list of areas. A package under `internal` that is outside the Layout is a finding too.
  - `controlflow`: no `else`, an `else if` included, and no block (`if`, `for`, `switch`, `select`) inside another block of the same function. A function literal starts again from zero.
  - `versions`: every pinned version is exact. It recognises a file by its name, wherever it is: Go module files, Compose files, Dockerfiles, the YAML files under `.github`, `pnpm-workspace.yaml`, `package.json` and `.mcp.json`.
- **Task targets.** `api:lint` (the linter, then `guard code`), `api:format` and `versions:check`. `check` runs `versions:check` first and `api:lint` before the build of the API.
- **The pipeline.** The `backend` job of `ci.yml` runs `task versions:check` and `task api:lint` before it starts the services. The cache of setup-go also follows `golangci-lint.sum`.
- **Documents.** `CODING_STANDARDS.md`: Enforcement in the backend rules names the linter, the formatter and each check; Control flow and the Exact rule of Versions shrank to a pointer; Enforcement in Versions says what the check reads; the Versions table has a row for the linter; Layout names `apps/api/tools`. Both READMEs gained the three commands and the two tools.

### Decided while building

- **golangci-lint has its own module file.** The ticket asks for `go tool`. The documentation of golangci-lint advises against it and, for those who use it, asks for a module file apart, so that the linter is built with the dependencies it was tested with. In `go.mod` it would have added about 210 indirect lines and could have raised the versions the API is built with. `go tool -modfile=golangci-lint.mod` keeps the exact version and `go tool`, and leaves `go.mod` with one new line. The Versions table says where it lives.
- **gofumpt runs inside golangci-lint.** One tool and one version to pin. A file that is not formatted is a finding of the lint.
- **Nesting is checked by the guard, not by the linter.** revive has a rule for it, `max-control-nesting`. With a limit of 1 it accepted an `if` inside a `for ... range`, a loop inside a `range` loop and an `if` inside a case of a `select`, because it counts neither. That is the most common second level in Go, so the rule is in `controlflow`, beside `else`, which no linter has.
- **The rule is read strictly**, as ticket 01 read it (`565328f`): an `if err != nil` inside a `for` is a second level, and so is an `if` inside a `case`.
- **The checks are written on the standard library**, with `go/parser`. The only library added is `go.yaml.in/yaml/v4` 4.0.0, which reads the YAML files and, with them, the JSON ones, so a finding has a line number.
- **Where the table is silent**, the import check allows: a third-party library in a test file of any layer; a package of this module outside `internal`, where a library is allowed; `kernel` importing `kernel`, and `platform` importing `kernel`; a package below `adapters/<technology>` as part of that adapter. It refuses an adapter importing the `application` of another area, and a third-party library in `application`, `domain` and `kernel` outside their tests.
- **Generated code** is a file with the comment Go defines, `// Code generated ... DO NOT EDIT.`. The linter, the formatter and the control-flow check skip it; the import check does not. No path is listed anywhere.
- **errcheck exemptions.** The error of a `Close` and of a print is not checked (the `std-error-handling` preset), nor the one of `(pgx.Tx).Rollback`, which is how pgx ends a transaction that did not commit. With them the code of ticket 01 passes unchanged.
- **`//nolint` must name its rule and say why** (nolintlint). The guard has no way to silence a finding.
- **Images need three numbers in the tag.** `caddy:2.10` floats over its patches. The images whose full version has two numbers are listed in the check: `postgres` today. A digest always passes. A service with `build:` is not judged.
- **More than the ticket lists, because it was one line each:** a runner label ending in `-latest` is rejected, and so is a `with` input named `version` or `*-version` that is not exact.
- **The mapping check** is the `exhaustruct_v5` linter of golangci-lint, limited to the types of adapter packages. It is written, commented out, in `.golangci.yml`, for ticket 05 to turn on. `exhaustruct` without the suffix is deprecated since golangci-lint 2.13.
- **The test of the linter is left out of a short run.** It runs the real linter on fixtures and `go tool` is slow to start. `task check` and the pipeline run it; `task api:test -- -short ./...` does not.
- **The table of the Dependency rule stays in `CODING_STANDARDS.md`**, with a pointer at the check below it. Enforcement asks a checked rule to shrink to a pointer; the table is the only place the rule can be read without reading Go.

### What was checked

- **`task check` passes** on the branch, with the code of ticket 01 unchanged: the linter reported `0 issues`, the guard and the version check printed nothing, and the Go tests passed. It passed again after the integration branch was merged in with ticket 03, whose code needed no change either.
- **A Go file that is not formatted.** With `apps/api/cmd/api/zz_unformatted.go`, which gofmt accepts and gofumpt does not, `task check` failed at `api:lint`: `cmd\api\zz_unformatted.go:4:1: File is not properly formatted (gofumpt)`. A test runs the linter on a fixture with the same fault.
- **Imports.** `imports_test.go` has one case for each forbidden import, seventeen in all, the four of the criterion among them, and one tree with every allowed import in areas that do not exist yet. By hand, with five packages under `apps/api/internal/zzarea` and `zzother`, `task api:lint` failed with one line for each: domain to application, domain to the domain of another area, application to an adapter, an adapter to another adapter.
- **`else` and nesting.** `controlflow_test.go` covers an `else`, an `else if`, ten shapes of a second level and four shapes that hold one. The same run by hand reported the `else` and the `if` inside the `for ... range`.
- **Versions.** `versions_test.go` holds, for each of twenty-one kinds of version, the values that pass and the ones that fail: a range, `latest` and a floating tag as that file writes them. By hand, with one version broken in each kind of file (`deploy/compose.base.yaml`, a Dockerfile, `ci.yml`, `golangci-lint.mod`, `pnpm-workspace.yaml`, `package.json`, `.mcp.json`), `task versions:check` failed with nine findings and named each line.
- **The mapping check.** On a scratch fixture, with the configuration written in `.golangci.yml`: a struct literal of an adapter type that left a field out failed, also behind a pointer; a domain type was not judged; the zero value returned beside an error passed.
- **The worktree** was clean after each of those runs.

### What was not checked

- **The pipeline.** Nothing is pushed, so the `backend` job did not run. It calls the same Task targets as the machine. The first run builds golangci-lint from its sources, which took about five minutes here.
- **actionlint and zizmor on the changed `ci.yml`.** Neither is installed, and the command that downloads actionlint was refused in this session. The file parses as YAML, which the version check does on every run.
- **The checks on Linux.** They are Go with no path of Windows in them, and ran on Windows only.

### Checked in the pipeline

In run [44225480372](https://github.com/dannilodms/software-factory-manager/actions/runs/44225480372) of pull request #5, on `0c751a9`, the `backend` job ran `task versions:check` and `task api:lint` (golangci-lint and the guard) on the Linux runner and passed, as on the machine. No broken rule was pushed on purpose, so a failure of the pipeline matching one of the machine is still not seen, and the box stays open. The run before it, [44222903777](https://github.com/dannilodms/software-factory-manager/actions/runs/44222903777), failed in the pipeline on a race the machine had not shown (`TestConcludingRequestsAtTheSameMomentSaveOneConsultancy`), fixed in `0c751a9`.
