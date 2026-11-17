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
- [x] The test completes the first setup with both images and reaches the login screen.
- [x] The test registers a new administrator with a second setup code.
- [x] The check of the limit reads a 429 with the code `too_many_requests`.
- [x] Run twice in a row, the test passes both times.
- [ ] The job runs on a pull request, the `Checks` job needs it, and a failure leaves a trace to read.
- [x] Every joint in "Joints this effort proves" is marked as proven in `../spec.md`, or its ADR says what changed.
- [x] The READMEs take a new person from a clean machine to a concluded setup.
- [ ] `task check` and the pipeline pass on `main` with everything of this effort.

## Comments

### What was built

- **`task e2e`** (`Taskfile.yml`): starts the services and builds the API binary, together; restarts Caddy; drops and creates the database `software_factory_e2e`, owned by `migration`, through `docker compose exec postgres psql`; applies the migrations with the binary; and runs `vp exec playwright test` in `apps/web`. `task check` ends with it.
- **The configuration** (`apps/web/playwright.config.ts`): the tests in `e2e/`, one worker, no retry, a test timeout of five minutes (Vite bundles on its first load), the trace kept on a failure in `test-results/`, Chromium in `en-US` and the time zone of São Paulo, at `APP_ORIGIN`. Playwright starts the API binary, handing it the database of the test as `DATABASE_URL`, and Vite as `task web` does; each is ready once Caddy reaches it (`/up` for the API, `/` for Vite), and a server already there fails the run.
- **`apps/web/e2e/system.ts`**: where the binary is, the origin, the connection of the test, `issueSetupCode()` (runs `issue-setup-code` on that database and reads the code, what it opens and the address from what it printed), and `required()`, which stops a run started without `task e2e` with what to do.
- **The test** (`apps/web/e2e/setup.spec.ts`), with fictitious data (Example Consultancy, Ada Example, Bruno Example, addresses under `.example`, identifiers of zeros):
  - the first setup: the code of the command, typed at the address it printed; the consultancy with a logo (PNG) and the first administrator with a photo (JPEG), each waited for until sent; the review showing both images; concluding, which leads to `/login?notice=setup_concluded` with its notice;
  - the way back in: a second code, which the command says opens the registration of a new administrator; the step of the new administrator with no Back; the e-mail of the first administrator in other case refused on its field; another e-mail, registered, leading to `/login?notice=administrator_registered` with its notice;
  - last, the limit: wrong codes sent to `POST /api/setup/open` until the first 429, which carries `application/problem+json`, `Cache-Control: no-store`, a `Retry-After` and the body of the schema `Problem` with `too_many_requests` and a trace identifier of 32 hexadecimal characters; then a code that works, typed on the screen, meets the limit too, and the screen shows "Too many attempts from this address" with the identifier the 429 carried.
- **The images** (`apps/web/e2e/images`): a logo and a photo of 96 pixels, abstract shapes drawn in a Chromium canvas.
- **The job** `end-to-end` in `.github/workflows/ci.yml`: Go, Vite+, Task, `task install`, the libraries of Chromium, `task e2e`, and on a failure the artifact `end-to-end-traces` with `apps/web/test-results` (`actions/upload-artifact` v7.0.2, by its hash). It is skipped for a change of documents only, and `Checks` needs it.
- **Around it**: Vitest looks for screen tests in `src` only (`test.dir`), or it took the test of Playwright for its own; Knip reads `playwright.config.ts` and `e2e/**/*.spec.ts` as entries and `e2e/` as part of the project; `test-results/` is ignored by Git.
- **Closing the effort**: each joint of "Joints this effort proves" in `../spec.md` is marked, with the issue that tried it: ten proven and `pg_stat_statements` that did not hold, which ADR-0015 already said. ADR-0013 (each step puts its own problems on the form, because of the second mount of development), ADR-0014 (what the screen tests needed) and ADR-0019 (what the base Compose file cannot hold) gained what their joints needed. The sentence of the spec on `pg_stat_statements` points at the joint.
- **Documents**: `CODING_STANDARDS.md` (Tests of the frontend: where the end-to-end tests live and how `task e2e` runs them; Rate limit: development counts the whole machine); both READMEs (the status, where each prerequisite comes from, the ports, the steps from a clean machine to a concluded setup, `task e2e` and `task check` in the table of commands); `docs/machine-tools.yaml` (`checked`).

### Decided while building

- **A database of its own**, `software_factory_e2e`, dropped and created empty at the start of each run and kept after it, so that a failure can be read in it. The development database and its volume are never touched; `task services` recovers nothing because nothing is removed.
- **The limit.** Every request of the machine reaches Caddy from the gateway of Docker, so the browser and the test share one count of ten `POST` a minute with anything else on the machine. The flow spends seven. Two runs in a row would spend fourteen in less than a minute, and the check of the limit spends what is left. So `task e2e` restarts Caddy, whose count lives in its memory only, and the check of the limit is the last test of the file, with one worker. Set aside: running the check from a container on the network of Compose, which has an address of its own, because two runs in a row would still meet the limit; and waiting for the window to pass, a minute on every run.
- **The test runner is `playwright/test`**, a subpath of the `playwright` package already in the catalog for Vitest, at the same 1.65.1. `@playwright/test` only re-exports it. Knip's Playwright plugin turns on with `@playwright/test` only, and would load the configuration, which needs the variables of `task e2e`; Knip reads the files as entries instead.
- **The connections are written on the commands**, and the API gets `DATABASE_URL` from the configuration of Playwright: a variable of the shell wins over Task's `env:` (tried), so a `DATABASE_URL` exported in a shell would otherwise send the test to another database.
- **The API runs from the binary** that `task e2e` builds, which also issues the codes, so the command the test runs is the one a person runs on a server. Vite runs as `task web` starts it.
- **The e-mail in use is part of the way back in**, one `POST` more than the ticket asks: it is the fault ticket 12 found only in a real browser.
- **`task check` runs `task e2e`**, because it runs every check of a pull request. It now needs ports 8080 and 5173 free and restarts the shared Caddy, so it cannot run in two checkouts at once or beside `task dev`.
- **No `@types/node`.** The test uses `node:child_process` and `node:path`, whose types already reach the type check through the dependencies, as they do for `vite.config.ts`.

### What was checked

- **The whole test passed six times** through `task e2e` on this machine (Docker Desktop), each from an empty database. Twice in a row, starting at 20:37:05 and 20:37:34; three times in a row, the last step of `task check` ending at 20:54:55, then runs starting at 20:55:20 and 20:55:49. A warm run takes about 25 seconds; the first took 69, most of it Vite bundling.
- **Without the restart of Caddy**, the second run in a row failed at the step of the code: the code was answered 429. That failed run left `trace.zip` and `error-context.md` in `apps/web/test-results`.
- **What the run leaves in its database**: the consultancy with its logo, the time zone of São Paulo and BRL; two users with the administrator role, e-mails in lower case, the first with her photo; the three setup codes, the first two ended; the files confirmed, `image/png` of 2362 bytes and `image/jpeg` of 1792; the audit log `setup_code_issued`, `setup_concluded`, `setup_code_issued`, `administrator_registered`, `setup_code_issued`. The development database stayed at version 1.
- **The guards of the run**: with `DATABASE_URL` and `MIGRATION_DATABASE_URL` of the shell naming a database that does not exist, the test still ran on its own and passed; with a server already on port 5173, the run stopped with `http://localhost/ is already used`; run without `task e2e`, Playwright stops with `APP_ORIGIN is not set. Run the end-to-end test with task e2e, which sets it.`
- **The steps of the README**, on this machine, with a database of their own named by variables of the shell (`fs13_readme`, dropped after): `task install`, `task services`, `task migrate` applied seven migrations, `task dev` served `/up` with 200 and the app at `/`, `GET /api/setup` said no code was waiting, `task issue-setup-code` printed the code, that it opens the first setup and the address `http://localhost/setup`, after which a code was waiting and the address served the app. The rest of the path, in the browser, is what `task e2e` does.
- **The checks of the app**: `vp check`, Knip (an unused export added to `e2e/system.ts` failed it), and Vitest lists no test of `e2e/`.
- **The workflow**: actionlint 1.7.12 and zizmor 1.32.0 (offline), from their images, report nothing; the one finding zizmor suppresses is the same as before this change. `task versions:check` passes on it.
- **The joints**, each against its ADR and the record of the issue that tried it.
- **`docs/machine-tools.yaml`**: every version was read again from the tool and is the recorded one (Node.js 22.23.3 and pnpm 11.22.0 are the ones of NVM and Corepack, before the shims of Vite+ on the path). No earlier ticket installed a tool, so only `checked` moved.
- **`task check`** passes, in five and a half minutes, ending with the end-to-end test; 85 screen tests. `feat/foundation-and-setup` had not moved from `39ecab6` (ticket 12), where the branch started, so the merge was empty.

### What was not checked

- **The pipeline.** Nothing was pushed: the job on the Linux runner (Docker Engine, the API and Vite on `0.0.0.0`, Caddy reaching them through the host gateway), the artifact of a failure, and `Checks` waiting for the job. The same holds for the last criterion, `task check` and the pipeline on `main`.
- **A clean machine.** The steps of the README were run on this one, where Go, Vite+, Task, Docker and Chromium were already installed and the caches warm; its box is ticked on that run, as issue 01 ticked its own.
- **Another browser than Chromium**, and the screen in another language than English in this test (the screen tests cover the three).
