# 01: River and the worker command

**What to build:** The system can run work that nobody waits for on a screen. A `worker` command of the API binary runs jobs stored in PostgreSQL by River, and `task dev` starts it beside the API and the app. Nothing visible changes for a user yet; this is the ground the e-mails and the cleanup stand on. Rules: Jobs in `CODING_STANDARDS.md`, ADR-0016. Prove the joints the standards list as not tried: sqlc reading River's exported SQL, goose applying it, the permissions the API role needs, and whether River's OpenTelemetry middleware links the trace that saved a job to the one that ran it. Record each outcome in the ticket and correct the ADR or the standards where a joint does not hold.

**Blocked by:** None (can start immediately).

**Status:** ready-for-agent

- [ ] River's tables arrive as numbered goose migrations exported by River's tooling, one per River schema version, with the grants the API role needs in the same migration; readiness still checks one schema version.
- [ ] Where River's tooling runs from is recorded: `docs/machine-tools.yaml` if installed on the machine, otherwise as a Go tool of the module listed in Versions.
- [ ] The `worker` command runs with a light queue and a heavy queue, each with a limit of concurrent jobs from configuration, and a shutdown deadline from configuration; it requires the database and nothing of the API's listening address.
- [ ] On a signal the worker takes no new job, gives the running ones the deadline, then cancels them; a cancelled job runs again later.
- [ ] A job worker in the `river` adapter reads identifiers from the job's arguments and calls one use case; a test with River's helpers in a temporary database proves it.
- [ ] Jobs are traced with River's OpenTelemetry middleware; whether the saving trace links to the running trace is checked and recorded.
- [ ] The River panel runs in development only, as the standards allow.
- [ ] `task dev` starts the worker; the new variables are in `.env.example` with values that work in development, and the configuration names a missing or invalid one at the start.
- [ ] Commands in `CODING_STANDARDS.md` lists the `worker` command.
