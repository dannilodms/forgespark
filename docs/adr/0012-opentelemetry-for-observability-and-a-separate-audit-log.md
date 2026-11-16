---
status: accepted
---

# OpenTelemetry for observability, and a separate audit log

The backend is instrumented with OpenTelemetry for traces, metrics and logs, and exports them over OTLP to an address taken from configuration. In the Docker Compose deployment that address is one extra container, `grafana/otel-lgtm`, which bundles the collector, Loki, Tempo, Prometheus and Grafana.

- **One identifier per request.** The trace identifier of a request is the `trace_id` returned in an error body (ADR-0009), so a failure seen on a screen or by an agent can be found in the telemetry. An incoming `traceparent` header is honoured.
- **Health.** Liveness touches no dependency. Readiness checks PostgreSQL and MinIO and feeds the Compose health check.
- **Nothing sensitive in telemetry.** Logs, spans and metric labels never carry personal data, the value of a credential, an access token or a session cookie. A test proves it for logs.

The **audit log** is a different thing and lives in the database. It is the permanent, append-only record of sensitive operations: reading the value of a credential, logging in, and creating or revoking an access token. It is not telemetry, which may be lost, and it is not the history of work that screens show.

## Considered Options

- **The minimum: health endpoints, structured logs to the container output and a trace identifier.** Enough to find a failing request on a single binary, with nothing more to operate. Not chosen: the maintainer wants traces and metrics from the start.
- **A hosted service.** Nothing to operate, but telemetry leaves the server.
- **A metrics endpoint only.** Useful only with something collecting and showing it, which is most of the stack anyway.
- **Auditing every write.** A complete trail, but it overlaps the history of work and grows fast.

## Consequences

- Grafana describes `grafana/otel-lgtm` as intended for development, demo and testing, and points production use to its cloud service. It is accepted here because losing telemetry loses no business data, and because the destination is one address in configuration: moving to another collector or a hosted service changes no code.
- The deployment has one more container to run and keep updated.
- The legal side of personal data protection (legal basis, retention periods, a path for deletion requests) is outside this repository until someone from outside the consultancy logs in to the system.
- How mature the logs signal is in the Go OpenTelemetry SDK was not checked. If it is not stable, logs go to the container output with the trace identifier until it is.
