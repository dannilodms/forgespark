---
status: accepted
---

# Background jobs in PostgreSQL, run by a worker command

Work that nobody waits for on a screen runs as a job. Jobs are stored in PostgreSQL through River, and the API binary has a `worker` command that runs them, as its own service in Docker Compose.

- **A job is saved with the business change.** The use case inserts the job inside its own transaction. A job never exists without the change that asked for it, and no change loses its job. A queue outside the database needs an outbox table and a relay to give the same guarantee.
- **What becomes a job.** Sending an e-mail, converting a file, a call to an LLM in the background, a write to a project's Git repository. What the user waits to see stays inside the request: a streamed LLM answer, an on-demand translation.
- **The worker is a command of the same binary**, beside `migrate`. Converting a video takes no CPU from requests, the API only saves jobs, and the API and the workers can later scale apart. Several workers can run at once: River elects a leader among them for scheduling and maintenance.
- **A job calls one use case.** A job worker is one more inbound adapter (ADR-0008): it reads the job's arguments and calls a use case.
- **At least once.** A job can run twice, so every job is idempotent.
- **Few attempts, then a visible failure.** Five attempts by default, about six minutes, and a job kind declares its own when it needs more. When they run out, the domain record is marked as failed and a person can ask for a new attempt.
- **River's tables are ordinary migrations.** River's CLI exports the SQL of each of its schema versions, and each becomes a numbered goose migration (ADR-0015). There is still one way to change the schema and one schema version for readiness to check.
- **Progress is read, not pushed.** The operation answers 202, and the screen reads the resource again while the work is in progress.
- **The River panel is served by the API**, behind the session, only to a user who administers the system. Job arguments carry identifiers only, because the panel shows them.

Running more than one instance of the API, and Kubernetes, are intended and not built. Nothing decided here prevents either.

## Considered Options

- **Redis or Valkey, with asynq.** A fast and well-known queue, in a second data service, and saving the change and the job are no longer one operation.
- **RabbitMQ.** A complete broker with more to operate, no scheduling of its own, and the same two steps.
- **Amazon SQS.** Nothing to operate, but it ties the system to AWS, needs an emulator in development, and has the same two steps.
- **A workflow engine (Temporal, Hatchet).** Built for long pipelines that wait for a person. It is a server of its own with a database of its own.
- **A queue written here**, on `SELECT ... FOR UPDATE SKIP LOCKED`. No new dependency, but retries, the rescue of a job whose process died, scheduling and leader election become our code.
- **gue or pgmq.** Smaller PostgreSQL queues. gue has no periodic jobs or panel. pgmq gives the queue primitives and no worker, and as an extension it is not in the official PostgreSQL image.
- **Jobs inside the API process.** One container less, but heavy work competes with requests, and the shutdown rule that no goroutine outlives its request would need an exception.
- **A separate Go binary for the worker.** Two builds, and two versions that can drift apart.
- **River's own migrator, called by `migrate`.** The documented path, with two schema versions to track.
- **River's default retries.** 25 attempts, the last about three weeks later. An e-mail could leave days after the fact, and a permanent failure would take long to show.
- **Progress pushed by Server-Sent Events.** The worker is another process, so it needs a channel between the two (`LISTEN`/`NOTIFY`). It waits for a case that needs it.
- **The River panel as its own container with a password, or in development only.** The first is one more password, outside the system's login. The second leaves production with telemetry and SQL alone.

## Consequences

- River is not yet at version 1.0. It releases almost every week and still breaks compatibility in minor versions (v0.51.2 when this was decided). An update is its own commit, and one that changes River's schema brings the new migration with it.
- Three things were not tried: whether sqlc reads River's exported SQL, whether goose needs annotations in it, and which permissions the API role needs on River's tables. River's reindexer needs the `MAINTAIN` privilege on its job table, or has to be turned off.
- Workflows, a concurrency limit across processes and a dead-letter queue are in River's paid edition. A limit here is per worker process, and chaining steps is our code.
- River deletes a finished job after a day and a discarded one after a week. What must be kept is a domain record.
- A periodic schedule lives in the memory of the leader. A run can be missed when the leader changes, so a periodic job tolerates it.
- A job cut by a shutdown runs again later, so a long job tolerates being restarted. Whether a cut uses up an attempt is settled with the first long job.
- A job whose worker died is picked up again only after River's rescue period, one hour by default.
- A River worker uses `LISTEN`/`NOTIFY` and session features. If a connection pooler ever arrives (ADR-0015), the workers connect to PostgreSQL directly.
- Whether River's OpenTelemetry middleware links the trace that saved a job to the trace that ran it was not checked.
- In production the River panel waits for the modelling of users, which has to say who administers the system. Until then it runs in development only.
- How the document processing pipeline keeps its state, and where files are converted, were left for the implementation of the processors.
- Known obstacles to a second API instance: the MCP Go SDK keeps sessions in the memory of the process (it has a stateless mode), an in-memory cache would differ between instances, and a deploy is a restart.
