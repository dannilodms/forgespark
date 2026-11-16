---
status: accepted
---

# OpenTelemetry for observability, and a separate audit log

The backend is instrumented with OpenTelemetry for traces, metrics and logs, and exports them over OTLP to an address taken from configuration. In the Docker Compose deployment that address is one extra container, `grafana/otel-lgtm`, which bundles the collector, Loki, Tempo, Prometheus and Grafana.

- **One identifier per request.** The trace identifier of a request is the `trace_id` returned in an error body (ADR-0009), so a failure seen on a screen or by an agent can be found in the telemetry. An incoming `traceparent` header is honoured.
- **Health.** Liveness touches no part. Readiness fails when PostgreSQL is down or when the database schema is not the version the binary expects, and feeds the Compose health check. The file storage, the LLM provider and the telemetry collector are optional parts: one of them down leaves the system degraded and still serving. Their state is shown by the health report, an authenticated API operation built once users are modelled. The rules are in Health in [`CODING_STANDARDS.md`](../../CODING_STANDARDS.md).
- **Nothing sensitive in telemetry.** Logs, spans and metric labels never carry personal data, the value of a credential, an access token or a session cookie. A test proves it for logs.

The **audit log** is a different thing and lives in the database. It is the permanent, append-only record of sensitive operations: reading the value of a credential, logging in, and creating or revoking an access token. It is not telemetry, which may be lost, and it is not the history of work that screens show.

## Considered Options

- **The minimum: health endpoints, structured logs to the container output and a trace identifier.** Enough to find a failing request on a single binary, with nothing more to operate. Not chosen: the maintainer wants traces and metrics from the start.
- **A hosted service.** Nothing to operate, but telemetry leaves the server.
- **A metrics endpoint only.** Useful only with something collecting and showing it, which is most of the stack anyway.
- **Auditing every write.** A complete trail, but it overlaps the history of work and grows fast.
- **The file storage in the readiness check.** The first version of this decision. Dropped: with a single instance no other replica takes the traffic, so a failed readiness makes every screen unavailable when only files are affected.
- **A health report open to anyone, or to any logged-in user.** Available sooner, but it tells which parts the system is made of and which are down. It is restricted to the users who administer the system, so it waits for the modelling of users.
- **Health checked in the background, with history and alerts.** It would tell the maintainer when a part goes down, but it needs work that runs outside a request, which has no mechanism yet.

## Consequences

- Grafana describes `grafana/otel-lgtm` as intended for development, demo and testing, and points production use to its cloud service. It is accepted here because losing telemetry loses no business data, and because the destination is one address in configuration: moving to another collector or a hosted service changes no code.
- The deployment has one more container to run and keep updated.
- The legal side of personal data protection (legal basis, retention periods, a path for deletion requests) is outside this repository until someone from outside the consultancy logs in to the system.
- The modelling of users has to say who administers the system. Until it does, the state of an optional part is seen only in the errors of the operations that use it and in telemetry.
- With PostgreSQL down, neither a session nor an access token can be validated, so the health report cannot be read. Readiness and the logs on the server are what is left.
- Two checks of the health report were not tried: whether every LLM provider behind the interface of ADR-0004 offers a call that spends no tokens, such as listing models, and how the Go OpenTelemetry SDK reports the result of an export.
- How mature the logs signal is in the Go OpenTelemetry SDK was not checked. If it is not stable, logs go to the container output with the trace identifier until it is.
