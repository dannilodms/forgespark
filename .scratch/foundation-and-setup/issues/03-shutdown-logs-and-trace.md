# 03 - Shutdown, logs and trace

Status: ready-for-agent
Blocked by: 01

Read `../spec.md` and, in `CODING_STANDARDS.md`, Shutdown and Telemetry. The reasons are in ADR-0012.

## What to build

The API stops without cutting work short, and what it writes about itself is safe to keep.

- The shutdown that Shutdown describes: the signal, readiness failing from then on while liveness keeps answering, new connections refused, requests in flight given the drain deadline from configuration, the remaining ones cancelled at the deadline and counted in a log line, the pool closed, and the exit status. `platform` owns the lifecycle and exports the shutdown notice.
- Logs as JSON lines on the output, each with the trace identifier, at a level taken from configuration. What is logged is what Telemetry lists: a 5xx or a panic with its raw error, a failed outbound call, the start and the stop of the process, the requests cut by a shutdown. No line carries a body, a query string or a header.
- A span for each request, opened by the inbound side, and pgx traced with otelpgx: the SQL text is recorded and the parameters never are. An incoming `traceparent` is honoured.
- The exporter of traces wired to an address taken from configuration. With no address nothing is exported. Tests run with no telemetry.
- A test that proves the logs carry nothing sensitive. It is written so that later tickets add their own values to it: a setup code, an e-mail, a name.

Metrics and the wrapper around the metric exporter come with the health report, in the third effort. The streaming part of the shutdown test waits for the first stream.

## Acceptance criteria

- [x] With a slow request held open, the signal lets it finish, refuses a new connection, and the process exits with status zero.
- [x] With the drain deadline reached, the remaining request is cancelled, a log line says how many were cut, and the process exits with a status other than zero.
- [x] Readiness fails from the signal on, and liveness keeps answering.
- [x] A second signal exits at once.
- [x] Every log line is one JSON object with the trace identifier of its request.
- [x] A successful request and a 4xx write no line.
- [x] A query made during a request shows as a span under the span of the request, with its SQL text and no parameter.
- [x] With no export address the process starts and exports nothing.
- [x] The test of sensitive values fails when a value it guards is written to the logs.

## Comments

### What was built

- **The process lifecycle.** `apps/api/internal/platform/lifecycle`. `Signals` listens for `SIGTERM` and `SIGINT`. `ShutdownNotice` turns them into the shutdown notice, a context that ends at the first signal, and exits with status 1 at the second. `Serve` answers requests until the notice, then drains.
- **The drain.** From the notice on the listener is closed, so new connections are refused, and the connections already open are still answered. The requests in flight have `SHUTDOWN_DRAIN_DEADLINE` to finish (20 seconds by default). At the deadline their contexts are cancelled, they get two seconds to notice, and one line says how many were cut: `requests cut by the shutdown`, with `count`.
- **The command that serves.** `cmd/api/serve.go` watches the notice. Readiness fails from the signal on, without asking PostgreSQL. When `Serve` returns, the pool is closed, the last spans are sent under a deadline of five seconds, and the command exits: with status 0 when every request finished, with status 1 when one was cut.
- **Logs.** `telemetry.NewLogger` writes one JSON object for each line, on the error output, from `LOG_LEVEL` up (`info` by default). A line written with the context of a request carries its `trace_id`, at the root of the object. The API writes `the API started` with its address, `the API stopped`, the line of a 5xx or of a panic, the requests cut, and a failed export of traces.
- **The telemetry of a request.** `telemetry.Requests` wraps the handler of the API. It opens the span of each request, under the trace of an incoming `traceparent`, and writes the one line a request may leave: `request failed` for a 5xx, with the method, the route, the status and the raw error, and `request panicked` for a panic, with the stack. Whatever answers a 5xx hands its raw error to `telemetry.RecordFailure`; readiness does. A panic that reaches it answers 500 with no body.
- **Traces.** `telemetry.StartTracing` builds the tracer of a command. With `OTEL_EXPORTER_OTLP_ENDPOINT` set, spans leave over OTLP/HTTP, to `/v1/traces` under that base URL. With it empty nothing is exported, and a span still has its trace identifier. Nothing is sampled, whatever the caller decided.
- **pgx.** `database.NewPool` takes a tracer provider and traces the pool with otelpgx 0.13.0. A query is a span under the span of its request, with the SQL text in `db.query.text`. The option that records the parameters stays off.
- **Configuration.** Three variables, none required: `SHUTDOWN_DRAIN_DEADLINE`, `LOG_LEVEL` and `OTEL_EXPORTER_OTLP_ENDPOINT`. They are in `.env.example` and in the table of both READMEs.
- **The test of sensitive values.** `cmd/api/sensitive_test.go`. It runs `migrate` and the API with PostgreSQL unreachable, with a marked value as the password of each database role and in the query string, a cookie, an authorization header and the body of a request that fails. It then searches the logs and the exported spans for each value. The comment at the top says how a later ticket adds a setup code, an e-mail or a name.
- **For the tests.** `telemetrytest.Collector` stands in for the telemetry collector and keeps the spans a command exported. In `cmd/api`, `launchAPI` runs the command and gives its logs, its signal and its exit status, and a proxy in front of PostgreSQL holds an answer back, which is how a test keeps a request in flight.
- **Documents.** `CODING_STANDARDS.md`: the two new packages in Layout, and what this code settled in Configuration, Telemetry and Shutdown.

### Decided while building

- **The notice is the context a command receives.** It ends when the shutdown starts. `migrate` and the commands to come pass it on, so a signal stops them. The command that serves only watches it: the contexts of its requests are cut loose from it, and only the drain deadline cancels them.
- **A connection already open is still answered while the drain lasts.** Shutdown asks for two things at once: new connections refused, and liveness answering. `http.Server.Shutdown` closes every idle connection the moment it is called and lets no connection take a second request, so after it nothing is answered. The drain closes the listener only, waits until no request is in flight, and calls `Shutdown` then. A monitor or a proxy that holds a connection gets 200 from liveness and 503 from readiness; one that opens a connection is refused.
- **The second signal exits from inside `platform`**, with `os.Exit(1)`, and runs nothing more: no flush, no closing of the pool.
- **One line for a 5xx, written in one place.** A handler does not log. It records its raw error, and `telemetry.Requests` writes the line when the request ends, with the route and the status it saw. A 5xx that recorded nothing still gets its line.
- **A span and a line name the route, never the path.** The route is the pattern the mux matched, `/things/{id}`. A method the standard does not know is written as `_OTHER`.
- **The span of a request is written by hand.** `otelhttp` puts the path, the query string, the address of the caller and the user agent on the span, which Privacy forbids.
- **OTLP over HTTP**, which needs no gRPC client. The variable has the name the OpenTelemetry specification gives it and the same meaning: a base URL, with `/v1/traces` appended. `WithEndpointURL` stopped appending it in v1.45.0, so `StartTracing` does.
- **A failed export is a log line.** The exporter is wrapped, so the failure is written by the logger of the command as `export of traces failed`. The global error handler of OpenTelemetry would print plain text.
- **The logs go to the error output.** The standard output stays for what a command prints, such as the setup code of ticket 04.
- **An invalid configuration is still one plain line**, `api: configuration: ...`, as ticket 01 wrote it. It is written before a logger can exist, because the level of the logger is part of the configuration.
- **Exit status 1 for a cut request.** The command logs its failure as `the API stopped`, at the level of an error, and returns `errLogged`, so `run` prints nothing after it.
- **Signals in the tests.** Go cannot send `SIGTERM` or `SIGINT` to one process on Windows. The tests hand the signal to the code: a channel for `lifecycle`, the context for `run`. They run on every system. Two more tests run on Unix only: the test binary starts again as the API, and the test sends it the real signals.

### What was checked

- **The shutdown, through the command run whole.** With a migrated database behind the stalling proxy and readiness held in flight: before the signal `/readyz` answered 200. After it a new connection was refused, `/livez` answered 200 and `/readyz` 503 on the connection opened before, and the command had not exited. Once PostgreSQL answered, the slow request received its 200 and the command returned 0.
- **The drain deadline.** With `SHUTDOWN_DRAIN_DEADLINE=200ms` and the same request held, the command returned 1, the request did not receive a 200, and the logs had `requests cut by the shutdown` with `count` 1 and `the API stopped` with the error. In `lifecycle`, two requests held gave `count` 2 and neither received a 200.
- **Real signals, on Linux.** In a `golang:1.27.4` container, against the same PostgreSQL: `SIGINT` stopped the process with status 0 and the line `the API stopped`. With a request in flight, `SIGTERM` refused new connections and left the process running, and a second `SIGTERM` ended it within five seconds with a status other than zero and no `the API stopped`.
- **The whole suite on Linux**, 76 tests, with the race detector, in the same container. On Windows the 74 tests that are not for Unix pass.
- **Logs.** Every test of the command reads the logs as JSON and fails on a line that is not an object. `/livez` (200), an unknown path (404) and `POST /livez` (405) wrote nothing. `/readyz` on a database with no migration, sent with a `traceparent`, wrote one line with the `trace_id` of that header and the error `the database schema is at version 0 and this binary expects version 1`, and without the query string that was sent.
- **By hand, on Windows.** The built binary on port 8093 at level `debug` wrote one line, `the API started`, as JSON on the error output, and nothing for `/livez`, `/readyz`, an unknown path and a wrong method.
- **Spans.** With the collector of `telemetrytest` as the export address, the API exported the span `GET /readyz` in the trace of the incoming `traceparent`, and under it the span of `SELECT max(version_id) FROM goose_db_version`. In `database`, a query with the parameter `ana@example.com` gave a span with the SQL text, and no span carried the parameter or the result. Turning on the option of otelpgx that records parameters made that test fail.
- **No export address.** The API started, answered and stopped with status 0, and wrote only its start and its stop. The tracer reports that it does not export.
- **The test of sensitive values.** Adding the query string to the line of a failed request made it fail, naming `a value in a query string`. A second test writes a guarded value through the logger in five ways and checks that each is found.
- **The guards of ticket 02**, read from its worktree before they were merged. golangci-lint with gofumpt reported nothing, the Unix tests included. The guard for `else`, nesting and imports found three test helpers with an `if` inside a `for`, which were rewritten, and then reported nothing. The check of exact versions passed with the new modules.
- **`task check`** passes.

### What was not checked

- **The pipeline was not run.** Nothing was pushed. The two tests for Unix ran in a container on this machine only.
- **A real collector.** The export was received by a stand-in that speaks OTLP/HTTP. The OpenTelemetry Collector and `grafana/otel-lgtm` are not in the Compose file yet.
- **A signal from a console on Windows.** Ctrl+C in `task api` was not tried. `kill` from Git Bash ends the process without a signal.
- **A container being stopped.** No container runs the API yet, so `docker stop` was not tried.
- **The stream.** The part of the shutdown test that holds a stream waits for the first stream, as the ticket says.

### Found, and left for a decision

- **Liveness and readiness open a span each.** The ticket asks for a span for each request, and they are requests. With a probe every few seconds and no sampling, they will be most of the traces once a collector runs. Leaving them out is a filter in the collector, or two lines in `serve.go`.
- **Closing the pool waits for PostgreSQL.** After a query is cancelled, pgx asks PostgreSQL to cancel it and waits up to 15 seconds. With PostgreSQL hung, the exit takes the drain deadline plus that. The stop grace period of the API container has to be longer than both, and Docker gives 10 seconds by default.
- **The OpenTelemetry SDK reads variables on its own.** Configuration said no other code reads the environment. The SDK reads the `OTEL_` variables of its specification: the headers and the timeout of an export, the resource attributes. `CODING_STANDARDS.md` now records it as the one exception.
- **`migrate` writes plain lines**, on both outputs. They are what a person reads when running it, and what the container of the migration step will write.
