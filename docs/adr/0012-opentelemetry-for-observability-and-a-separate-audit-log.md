---
status: accepted
---

# OpenTelemetry for observability, and a separate audit log

The API and the worker are instrumented with OpenTelemetry for traces and metrics, and send them over OTLP to an address taken from configuration. Their logs go to the container output. In the Docker Compose deployment two containers handle telemetry: an OpenTelemetry Collector, which gathers everything, and `grafana/otel-lgtm`, which stores and shows it with Loki, Tempo, Prometheus and Grafana.

- **One identifier per request.** The trace identifier of a request is the `trace_id` returned in an error body (ADR-0009), so a failure seen on a screen or by an agent can be found in the telemetry. The trace starts in Caddy, which writes the identifier in its access log and passes it to the API. An incoming `traceparent` header is honoured. The API passes the identifier on to the agents of ADR-0004 and never to a third party.
- **Everything goes through the collector.** It has no public port. It receives the traces and metrics of the API, the worker and Caddy, reads the output of every container from Docker's log files, and collects the metrics of the server (disk, memory, CPU), PostgreSQL, Caddy and the file storage. It forwards all of it to `otel-lgtm` and holds what arrives while `otel-lgtm` restarts. The Docker socket is never mounted.
- **Logs go to the container output.** The API and the worker write JSON lines with the trace identifier, and the collector reads them like the output of any other container. A line written while the collector was down is still in Docker's files when it comes back.
- **The API logs what is not routine.** A 5xx or a panic with its raw error, a failed outbound call, the start and the stop of the process, the requests cut by a shutdown. A successful request and a 4xx write no line: the trace and Caddy's access log already record them.
- **Telemetry is created in `adapters` and `platform` only.** A use case and the domain create no span and write no log, so the dependency rule of ADR-0007 has no exception. An inbound adapter opens the span and records the error that the use case returned.
- **Who acted is an opaque identifier.** Spans and logs carry the public identifiers of the user, the access token, the client and the project. Metric labels carry none of them.
- **Nothing sensitive in telemetry.** Logs, spans and metric labels never carry personal data, the value of a credential, an access token or a session cookie. A test proves it for the logs of the API. The rule covers the other containers too: the IP address of a caller is cut down to its network, the query string is removed, and PostgreSQL logs an error without the detail that carries row values.
- **An LLM call records its use, never its content.** The provider, the model, the input and output tokens, the duration and the outcome. The prompt and the answer are free text and can hold personal data or the value of a credential.
- **Grafana faces the internet, for the maintainer alone.** It has a host name of its own in Caddy and a login through GitHub that accepts one account. The image starts with anonymous access as administrator and with a default password; both are turned off. The other ports of `otel-lgtm` have no login and are never published.
- **Dashboards and alerts are files.** Dashboards, alert rules and the contact point live in `deploy/`, and Grafana loads them when it starts. Grafana evaluates the rules over the telemetry it already has and warns by e-mail. The first rules are five: the API or the worker stopped sending metrics, the rate of 5xx is high, the disk is over 80% full, PostgreSQL is down or near its connection limit, and an optional part is down.
- **Thirty days.** Logs, traces and metrics are kept for 30 days, and the metrics also have a size limit. The image keeps logs forever and has no limit on disk by default.
- **Not in production yet.** The two telemetry containers run in development only, in a Compose profile that is off by default. Production and staging do not declare them until someone other than the maintainer uses the system (ADR-0019). While the maintainer is the only user, their cost buys little: `otel-lgtm` asks for 2 GB of memory on the one server that runs everything else, it releases almost every week, and Grafana would face the internet. A failure affects nobody else and can be read in the output of the containers. The API and the worker are instrumented all the same: with no address to send to, they export nothing. Until then production has the output of its containers, kept in Docker's files, and the monitor below.
- **A monitor outside the server.** A hosted monitor calls `/up`, a path that Caddy forwards to the readiness of the API and that answers with the status code alone. Each backup signals the same monitor when it ends (ADR-0020). The monitor warns the maintainer by e-mail when the system does not answer or when a backup does not report.
- **Health.** Liveness touches no part. Readiness fails when PostgreSQL is down or when the database schema is not the version the binary expects, and feeds the Compose health check. The file storage, the LLM provider, the telemetry collector, the job worker and the e-mail server are optional parts: one of them down leaves the system degraded and still serving. Their state is shown by the health report, an authenticated API operation built once users are modelled. The state of the telemetry collector is the last export of metrics, which the API records. The rules are in Health in [`CODING_STANDARDS.md`](../../CODING_STANDARDS.md).

The **audit log** is a different thing and lives in the database. It is the permanent, append-only record of sensitive operations: reading the value of a credential, logging in, and creating or revoking an access token. It is not telemetry, which may be lost, and it is not the history of work that screens show.

## Considered Options

- **The minimum: health endpoints, structured logs to the container output and a trace identifier.** Enough to find a failing request on a single binary, with nothing more to operate. Not chosen: the maintainer wants traces and metrics from the start.
- **A hosted service.** Nothing to operate, but telemetry leaves the server.
- **A metrics endpoint only.** Useful only with something collecting and showing it, which is most of the stack anyway.
- **One container: collection inside `otel-lgtm`.** The first version of this decision. Its bundled collector can read the server's files, but it is the container whose Grafana faces the internet.
- **The API sending straight to `otel-lgtm`.** One hop less, but its telemetry is lost while `otel-lgtm` restarts, and there are two ways in to maintain.
- **Logs sent over OTLP.** The logs API and SDK for Go are stable (v1.47.0), but the OTLP log exporters and the bridge from `log/slog` are not. Logs sent this way are also lost while the collector is down, and sent both ways they arrive twice.
- **A span for each use case.** It shows one use case calling another, at the cost of a telemetry interface in `platform` and the same lines in every use case, or of an exception to the dependency rule.
- **A line in the API log for every request.** It survives the loss of the traces, but it is the third record of the same request.
- **No identifier of who acted.** Nothing personal by any reading of the rule, but no way to find what one user or one access token did.
- **The full IP address, or none.** The first is personal data. The second leaves an abuse of the login with no trace of where it came from.
- **PostgreSQL logs kept out of the telemetry.** They would keep the full detail of an error, on the server only.
- **The prompt and the answer recorded in development.** It helps to debug a prompt, and it has to be proven never to turn on in production.
- **Grafana behind an SSH tunnel or a private network.** Nothing new faces the internet, but Grafana is out of reach without access to the server.
- **Grafana's own password, or a login through Google.** The first is a password form on the internet. Google restricts a login by Workspace domain only, so it cannot accept one personal account.
- **Grafana behind the session of the app.** One login, but it needs the modelling of users and new code.
- **Dashboards and alerts made on Grafana's screens.** Faster to try, but they live only in the container's volume and have no history.
- **Alerts by Telegram or a webhook.** No mail server to configure. The maintainer chose e-mail.
- **The telemetry containers in production from the first deploy.** Traces and metrics of every failure from the start, for 2 GB of memory, two more containers to keep updated and a Grafana on the internet, on a server that only the maintainer uses.
- **No monitor outside the server.** The first version of this decision left it for later, counting on the alerts of Grafana. With no telemetry in production, nothing would warn of anything.
- **The monitor reading the home page, or a signal sent by the worker.** The first shows only that Caddy answers. The second needs no public path, at the cost of one more job and one more outbound call.
- **Probing the collector, or leaving it out of the health report.** A probe says the collector is up, not that telemetry arrives.
- **The API role for the collector in PostgreSQL.** No new role, but the collector would hold a credential that reads and writes business data.
- **Auditing every write.** A complete trail, but it overlaps the history of work and grows fast.
- **The file storage in the readiness check.** The first version of this decision. Dropped: with a single instance no other replica takes the traffic, so a failed readiness makes every screen unavailable when only files are affected.
- **A health report open to anyone, or to any logged-in user.** Available sooner, but it tells which parts the system is made of and which are down. It is restricted to the users who administer the system, so it waits for the modelling of users.
- **Health checked in the background, with history and alerts.** Possible as a periodic job (ADR-0016). Not built: Grafana already evaluates alert rules over the telemetry it receives, with no code in the API.

## Consequences

- Grafana describes `grafana/otel-lgtm` as intended for development, demo and testing, and points production use to its cloud service. It is accepted here because losing telemetry loses no business data, and because the collector stands between the system and it: moving to another backend or a hosted service changes the collector's configuration and no code.
- The deployment has two more containers to run and keep updated. `otel-lgtm` asks for 2 GB of memory and releases almost every week (0.41.0 when this was decided).
- Everything this decision says about the collector, Grafana, the alerts and the retention describes the deployment once it has the telemetry containers. Until then a failure in production is investigated in the output of the containers, on the server, and there are no traces or metrics of it.
- The readiness of the API is reachable from the internet through `/up`. It tells whether the system is ready and nothing else.
- The hosted monitor is one more account. Which one is chosen with the deploy.
- Once Grafana runs in production, nobody is warned when `otel-lgtm` is down, because its alerts come from inside it. The monitor outside the server covers the server as a whole.
- The alerts reach nobody until Grafana has an SMTP server, taken from the deploy configuration. Whether the SMTP service of Resend serves it was not tried.
- The collector mounts, read-only, the server's file system and Docker's log folder.
- With the collector down, the logs wait in Docker's files, which rotate by size. A long outage loses the oldest lines.
- The telemetry volume is not in the backup. Dashboards and alerts are files in the repository, so losing the volume loses only the stored telemetry.
- PostgreSQL has a third role, for monitoring only (ADR-0015). Its log also loses the hint and the context of an error, which go with the detail.
- The cost of LLM calls per client or per project cannot come from telemetry, which may be lost and carries neither in a metric label. Whether it is business data is decided with the first LLM feature.
- The OpenTelemetry semantic conventions for LLM calls and for MCP are still in development. Their names can change, and the dashboards change with them.
- The MCP Go SDK has no OpenTelemetry support (v1.8.0). The MCP adapter opens the span of a tool call in a middleware.
- The Go SDK reports a failed export only, and has no record of the last one. The API wraps the metric exporter to keep the time and the error of each export. Traces leave only when there is something to send, so the periodic export of metrics is the heartbeat.
- The legal side of personal data protection (legal basis, retention periods, a path for deletion requests) is outside this repository until someone from outside the consultancy logs in to the system.
- The modelling of users has to say who administers the system. Until it does, the state of an optional part is seen only in the errors of the operations that use it and in telemetry.
- With PostgreSQL down, neither a session nor an access token can be validated, so the health report cannot be read. Readiness and the logs on the server are what is left.
- These were read in documentation and not tried: the folder from which Grafana loads alert rules inside the image, whether Tempo applies its retention when run this way, the exact mounts the collector needs, a GitHub login that refuses every account but one, the metrics endpoint of Garage, and the removal of the IP address and the query string from spans. Whether every LLM provider behind the interface of ADR-0004 offers a call that spends no tokens, such as listing models, was not tried either.
- Logs over OTLP are looked at again when the exporters for Go are stable.
