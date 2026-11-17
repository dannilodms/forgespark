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

- [ ] With a slow request held open, the signal lets it finish, refuses a new connection, and the process exits with status zero.
- [ ] With the drain deadline reached, the remaining request is cancelled, a log line says how many were cut, and the process exits with a status other than zero.
- [ ] Readiness fails from the signal on, and liveness keeps answering.
- [ ] A second signal exits at once.
- [ ] Every log line is one JSON object with the trace identifier of its request.
- [ ] A successful request and a 4xx write no line.
- [ ] A query made during a request shows as a span under the span of the request, with its SQL text and no parameter.
- [ ] With no export address the process starts and exports nothing.
- [ ] The test of sensitive values fails when a value it guards is written to the logs.
