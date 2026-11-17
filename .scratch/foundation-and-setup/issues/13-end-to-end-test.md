# 13 - End-to-end test of the setup, through Caddy

Status: ready-for-agent
Blocked by: 12

Read `../spec.md` and, in `CODING_STANDARDS.md`, Tests in the frontend rules and Checks on a pull request.

## What to build

One test drives the real system the way whoever installs it does, and the pipeline runs it on every pull request. This closes the effort.

- Playwright, at an exact version, with a Task target that starts the development Compose file, applies the migrations to an empty database, starts the API and the app, and runs the test against the address of Caddy.
- The test issues a setup code with the command, opens the app, types the code, fills in the consultancy and the administrator with a logo and a photo, concludes, and lands on the login screen with its notice.
- It then issues a second setup code and registers a new administrator.
- A separate check sends requests past the limit to an operation of the setup and reads the 429 in the error format of the API.
- A job in `ci.yml`, added to the `needs` of `Checks`. A pull request that changes only documents skips it. On a failure it keeps the trace of the test as an artifact.
- The sample data of the test is fictitious: no real company, person or address (ADR-0021).
- Closing the effort: the list "Joints this effort proves" in `../spec.md` is checked one by one, and an ADR whose joint needed something it did not expect is updated. `README.md` and `README.pt-BR.md` describe the whole path of development, from the prerequisites to a concluded setup. `docs/machine-tools.yaml` records any tool installed on the maintainer's machine along the way.

## Acceptance criteria

- [ ] One Task target runs the end-to-end test from an empty database, on the machine and in the pipeline.
- [ ] The test completes the first setup with both images and reaches the login screen.
- [ ] The test registers a new administrator with a second setup code.
- [ ] The check of the limit reads a 429 with the code `too_many_requests`.
- [ ] Run twice in a row, the test passes both times.
- [ ] The job runs on a pull request, the `Checks` job needs it, and a failure leaves a trace to read.
- [ ] Every joint in "Joints this effort proves" is marked as proven in `../spec.md`, or its ADR says what changed.
- [ ] The READMEs take a new person from a clean machine to a concluded setup.
- [ ] `task check` and the pipeline pass on `main` with everything of this effort.
