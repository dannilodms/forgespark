# 07 - Development behind Caddy, with the limit on the setup code

Status: ready-for-agent
Blocked by: 05

Read `../spec.md` and, in `CODING_STANDARDS.md`, Delivery, Health and Compose files. The reasons are in ADR-0009, ADR-0011 and ADR-0024.

## What to build

In development the browser reaches the system through Caddy, at one origin, and the operations that take a setup code are limited for each IP address.

- A Caddy image built in this repository with a rate-limit plugin, both at exact versions. Choosing the plugin is part of this ticket.
- The `Caddyfile` of development in `deploy/`, and Caddy in the development Compose file. One site on `localhost` forwards `/api` to the API on the machine and the rest to the Vite server on the machine.
- A limit, for each IP address, on the `POST` operations of the setup, starting at ten requests a minute. Over it, Caddy answers 429 in the error format of the API, with the code `too_many_requests`.
- `Cache-Control: no-store` on every response under `/api`.
- `/up`, answered by Caddy with the status code of readiness and no body. Liveness and readiness stay unreachable from outside.
- The Task target that starts the app in development starts the API and Vite together, and the address to open is the one of Caddy.
- `CODING_STANDARDS.md` records the plugin and the number of the limit, and the two places that said they were still to be chosen are updated. `README.md` and `README.pt-BR.md` give the address to open.

Compression, the security headers, the Brotli copies and the cache headers of the built app wait for the deploy.

## Acceptance criteria

- [x] With the services and the app started, the address of Caddy shows the app, and a request under `/api` reaches the API.
- [x] Requests past the limit to an operation of the setup answer 429 as `application/problem+json` with the code `too_many_requests`, and requests within it are untouched.
- [x] `GET /api/setup` and every path outside the setup have no limit.
- [x] A response under `/api` carries `Cache-Control: no-store`.
- [x] `/up` answers 200 with no body when the API is ready and a failure status when it is not.
- [x] Liveness and readiness cannot be reached through Caddy.
- [ ] Caddy reaches the API and Vite on the machine both under Docker Desktop and on the runner of the pipeline.
- [x] The Caddy image and its plugin have exact versions, and the check of ticket 02 covers its Dockerfile.
- [x] `CODING_STANDARDS.md` names the plugin and the limit.

## Comments

### What was built

- **The Caddy image.** `deploy/caddy/Dockerfile` builds Caddy 2.11.9 with xcaddy, in `caddy:2.11.9-builder-alpine`, with the rate-limit plugin `github.com/mholt/caddy-ratelimit` at v0.1.0, and copies the binary into `caddy:2.11.9-alpine`.
- **The `Caddyfile` of development**, `deploy/Caddyfile`. One site, `http://localhost`, with plain HTTP. `/api` and `/api/*` go to the API, `/up` is rewritten to `/readyz` and goes to the API too, and everything else goes to Vite. Every response under `/api` carries `Cache-Control: no-store`, set as the response leaves, so it replaces one the API might send. `tracing` is on for the site.
- **The limit.** One zone of the plugin, for every `POST` under `/api/setup/`, keyed by `{remote_host}`: 10 a minute for each IP address, counted together. Past it the plugin returns a 429 error, and `handle_errors 429` writes the body of the error format with `Content-Type: application/problem+json` and `Cache-Control: no-store`. The plugin adds `Retry-After`.
- **Caddy in Compose.** The service `caddy` of `compose.yaml` builds the image as `software-factory-manager-caddy:development`, publishes it at `127.0.0.1:${CADDY_PORT}` (80), mounts the `Caddyfile` read-only, maps `host.docker.internal` to `host-gateway`, and sets `OTEL_TRACES_EXPORTER=none`. `deploy/compose.base.yaml` gives it the rotating logs. The upstreams are `{$API_UPSTREAM}` and `{$WEB_UPSTREAM}` in the `Caddyfile`, defaulting to ports 8080 and 5173 of the machine.
- **Task.** `task services` builds the image (`docker compose up --detach --wait --build`). `task dev` starts the API and Vite together; `task web` starts Vite alone, on `$WEB_LISTEN_HOST`, port 5173, `--strictPort`; `task api` is unchanged. `task dev` prints the address to open.
- **Variables.** `.env.example` gains `CADDY_PORT=80` and `WEB_LISTEN_HOST=127.0.0.1`, and `APP_ORIGIN` is now `http://localhost`, so `issue-setup-code` prints `http://localhost/setup`.
- **The version check** (`apps/api/tools/guard/versions/xcaddy.go`) reads `RUN xcaddy build` in a Dockerfile, across the lines a backslash continues: the Caddy version when one is named, and each `--with module@version`. A plugin with no version, `latest`, a branch or a partial version is a finding.
- **The tests** (`apps/api/cmd/api/caddy_test.go`). Each starts a container of the image of `compose.yaml`, with the `Caddyfile` of development mounted, in front of the binary of the API and of a stand-in for Vite, both listening on the host of `API_LISTEN_ADDRESS`. The 429 is checked against the schema `Problem` read from the committed `apps/api/openapi.yaml`, with Huma's validator.
- **The pipeline.** `ci.yml` sets `API_LISTEN_ADDRESS: 0.0.0.0:8080` and `WEB_LISTEN_HOST: 0.0.0.0` for every job, so that Caddy reaches what listens on the runner. The backend job already runs `task services` and `task api:test`, which now build the image and run the tests of Caddy.
- **The documents.** `CODING_STANDARDS.md`: the plugin, the limit and its 429 in Rate limit; Deferred no longer leaves the plugin to be chosen; One origin and the mount of Caddy in Development; tracing in Caddy in the Development of Observability; the tests of Caddy in Tests; `xcaddy build` in the Enforcement of Versions. ADR-0011 names the plugin in its consequence. Both READMEs give the address to open, the steps and the commands.

### Decided while building

- **The plugin: `mholt/caddy-ratelimit`.** It is written by the author of Caddy, registers its directive in the Caddyfile with an order, keeps a sliding window per key in memory, and returns the 429 as an error that `handle_errors` writes, so the body is the error format with no code of ours. Its last release is v0.1.0 (August 2024), which builds with Caddy 2.11.9; the main branch has gone on since (key prefixes for IPv4 and IPv6, metrics) with no new release, and the brief asks for the latest stable one. No other plugin was tried.
- **One limit for the four `POST` of the setup, not one each.** Every one of them checks a setup code, so one budget is what limits the guessing. Ten a minute leaves room for a person: open, two uploads, conclude.
- **The `trace_id` of the 429 is the trace Caddy starts.** Caddy's `tracing` gives each request a trace and puts it in `{http.vars.trace_id}`, 32 hexadecimal characters like every other error body, and passes it to the API in `traceparent`, so the trace of a request that reaches the API is the same one. This is what Observability says of Caddy (the trace starts there), so tracing is on for the site already. Without an exporter Caddy sends to `https://localhost:4318` and logs each failure: `OTEL_TRACES_EXPORTER=none` stops that and keeps the identifiers. If tracing were removed, the placeholder would be empty, which the schema still accepts.
- **The `Caddyfile` is mounted, not copied into the image.** An edit shows after `docker compose restart caddy`, and the image stays the same in every environment. Compose creates the container again when `up` runs from another checkout, which costs Caddy nothing. The note in `compose.yaml` and Development in `CODING_STANDARDS.md` said no service mounts a file of the checkout; they now say PostgreSQL mounts none, for its connections.
- **The image is built by the development Compose file**, which names it. The base keeps the logs. Staging and production will name the image the pipeline publishes. A built service is not judged by the version check, so the name needs no version. The first version put `build` in the base, and the check then judged the name in `compose.yaml` as a floating tag.
- **`task services` builds every time** (`--build`), so the image is never older than its Dockerfile, with `BUILDX_NO_DEFAULT_ATTESTATIONS=1`. BuildKit's default provenance record carries the time of the build: each build was a new image, and Compose created Caddy again on every `task services` and every `task api:test`. Without it the same Dockerfile gives the same image and the container stays. Compose 2.38.2 does not accept `provenance` in the `build` of a service.
- **Plain HTTP on `localhost`.** A bare host name makes Caddy issue a certificate from its own authority and redirect to HTTPS, which a browser would warn about. Browsers treat `http://localhost` as a secure context.
- **Caddy reaches the machine as `host.docker.internal`, mapped to `host-gateway`.** Docker Desktop answers that name by itself; Docker Engine on Linux needs the mapping. Docker Desktop forwards to `127.0.0.1` of the machine but not to `::1`, so Vite listens on `127.0.0.1` (`--host`) rather than `localhost`. Docker Engine on Linux reaches the machine at the address of its bridge, which a server on `127.0.0.1` does not hear, hence `WEB_LISTEN_HOST` and the host of `API_LISTEN_ADDRESS` set to `0.0.0.0` there.
- **The tests of Caddy listen where `API_LISTEN_ADDRESS` says**, rather than choosing by the operating system: the same variable that makes the API reachable makes the tests pass, and the pipeline sets it once.
- **A port taken fails.** Vite runs with `--strictPort`, so a second `task dev` stops instead of moving to 5174, where Caddy does not look. The API already fails on a taken port.
- **Caddy publishes on `127.0.0.1` only**, as PostgreSQL does.
- **The upstreams are variables of the `Caddyfile` with defaults**, so Compose sets nothing and the tests point Caddy at their own servers.
- **`/up` answers 502 when the API does not run.** Caddy's own error, with no body. `lb_try_duration`, which the README gives the deploy, is not in development.

### What was checked

- **The tests of Caddy**, each on its own container and database: the app and a 404 of the API through one origin, with `no-store` on the 404; ten `POST /api/setup/open` each answered by the API with its exact 422 and `no-store`, then the eleventh a 429 with `application/problem+json`, `no-store`, `Retry-After`, a `trace_id` of 32 hexadecimal characters and the exact body, which validates against `Problem` in the committed spec; from an address already over the limit, fifteen rounds of `GET /api/setup`, `POST /api/sessions` and `GET /setup`, none a 429; `/livez`, `/readyz`, `/api/../readyz`, `/api/%2e%2e/livez` and `/api/..%2freadyz` all answered by the stand-in for Vite; `/up` 200 with an empty body on a ready API, 503 with an empty body when PostgreSQL does not answer, and a 5xx with an empty body when the API has stopped.
- **That the tests bite.** Before the `Caddyfile` existed every test failed. With only the routing written, the tests of `no-store`, of the limit and of `/up` failed and the one of liveness and readiness passed. The check against the spec, given a body with an extra property, a status written as text and no `trace_id`, reported each, and passed the body Caddy writes.
- **`task dev` through the shared Caddy**, against a database of my own (`fs07_manual`, migrated, then dropped), with `curl` on `http://localhost`: `/` and `/setup` answered Vite's `index.html`; `/api/setup/unknown` the API's 404 in the error format with `Cache-Control: no-store`; `/up` 200 with `Content-Length: 0`, and `HEAD /up` 200; `/livez` and `/readyz` the `index.html` of the app; twelve `POST /api/setup/open`, ten 422 and then 429 with `Retry-After: 57` and a `trace_id`; `GET /api/setup` after that, not limited. With `task dev` stopped, `/up` answered 502 with `Content-Length: 0`. Vite's HMR socket opened through Caddy (`ws://localhost/`, its token from `/@vite/client`) and received `{"type":"connected"}`.
- **`issue-setup-code`** on that database printed `Address: http://localhost/setup`, and that address answered the `index.html` of the app through Caddy.
- **The trace.** Through a container of the image in front of a server that echoes headers: with no `traceparent`, the upstream received one Caddy made; with one, it received the same trace with Caddy's span. A `Cache-Control: max-age=60` from the upstream left Caddy as `no-store`.
- **Docker Engine on Linux.** In a `docker:28.3.2-dind` container standing for the runner, with the Go tests in a `golang:1.27.4` container on its network: the tests of Caddy passed with `API_LISTEN_ADDRESS=0.0.0.0:8080`, and failed with `127.0.0.1:8080` (Caddy answered 502), which is why `ci.yml` sets it. `host.docker.internal` resolved to the gateway of the bridge there. The `caddy` service of `compose.yaml`, started with `docker compose up --wait` in that engine, reached two servers on ports 8080 and 5173 on every interface of its machine: the app, the API and `/up` 200.
- **The version check.** Its tests on the plugin (no version, `latest`, `master`, `v0.1`; a release and a pseudo-version pass) and on the Caddy version (`latest`, `master`, `v2.11`; `v2.11.9` passes). On the repository it passes; on a copy of the Dockerfile with the plugin unpinned it reports `deploy/caddy/Dockerfile:9: the plugin github.com/mholt/caddy-ratelimit names no version`.
- **The versions.** Caddy 2.11.9 is the newest release `go list -m -versions github.com/caddyserver/caddy/v2` shows, and both images exist (`docker manifest inspect`). v0.1.0 is the only release of the plugin.
- **The workflow file.** actionlint 1.7.12 and zizmor 1.32.0 (offline), from their images: no finding.
- **The shared PostgreSQL container** was not created again by `task services`; Caddy was added beside it. Two `task services` in a row now leave Caddy `Running`; before the attestations were turned off, each one printed `Recreate`.
- **`task check`** passes on the branch. Two runs before that failed in the Go tests for reasons outside this ticket: in the first, the shared PostgreSQL was in crash recovery after another process of its container ended with an error; in the second, under load, about thirty tests timed out creating their database, and their two packages then passed run alone. `feat/foundation-and-setup` had nothing new to merge.

### What was not checked

- **The runner of the pipeline.** Nothing was pushed, so the backend job never ran with Caddy: neither the build of the image there (about three and a half minutes cold here, nothing cached) nor the tests of Caddy on it. Docker Engine on Linux was checked in a Docker-in-Docker container instead (above), which is why the criterion of the runner is left open.
- **Vite on Linux.** On Docker Engine a stand-in listened on port 5173; `vp dev --host 0.0.0.0` was not run there. On this machine Vite ran behind Caddy with `--host 127.0.0.1`.
- **A browser.** The app was read through Caddy with `curl` and its HMR socket opened with Node; no browser opened `http://localhost`, because the browser tools did not connect in this session.
- **The limit for two IP addresses.** Every request of a test, and every request from the machine in development, reaches Caddy from the same address (`X-Forwarded-For: 172.17.0.1`). That the key is the address is read from the configuration and the plugin's code, not seen with two callers.
- **The development database** was not migrated: it is at version 1, so `/up` through the shared Caddy answers 503 while the API of development runs against it.

### Found, and left for a decision

- **The limit is one bucket for the whole machine in development**, for the reason above. The end-to-end test (ticket 13) runs the whole setup twice and then checks the limit, all from one address: close to ten `POST` in a minute. It has to check the limit last, or wait for the window.
- **The operations of the setup do not list the 429 in the spec.** The description at the top of the spec names `too_many_requests`, and the client of ticket 08 has to handle it as a `Problem` it was not told about. Listing it under each `POST` of the setup is a change to `restapi` and to the spec.
- **A 502 of Caddy has no body.** When the API does not run, a request under `/api` gets Caddy's own 502, not the error format. The screen sees a failure with no `code`. (Fixed after the review of the effort: under `/api`, Caddy now answers a 502, 503 or 504 in the error format, with `bad_gateway`, `service_unavailable` or `gateway_timeout` and its `trace_id`, and the app translates them. `/up` still answers with no body.)
- **The plugin logs the full IP address** of each refused request, at the info level, on the output of the container. Privacy in Observability wants it cut to its network, which the collector does once it exists. (Fixed after the review of the effort: the default log of the `Caddyfile` filters every line, the plugin's included, and keeps the network of the address only; `caddy_test.go` checks it.)
- **v0.1.0 is from August 2024.** The plugin's main branch has the prefixes of IPv4 and IPv6 keys (so an IPv6 caller is limited by its /64, not by one address) and metrics, with no release. Moving to a commit of the main branch is a pseudo-version, which the version check accepts.
- **Caddy's spans carry what `otelhttp` records**, the client address and the path among them. Nothing is exported now. When the telemetry containers export them, the collector has to cut them as Privacy says.
