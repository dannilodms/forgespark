## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature-slug>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`), recorded as a `Status:` line in each issue file. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

### Browser tools

Chrome DevTools MCP, for investigating performance, console and network in a real Chrome. See `docs/agents/browser-tools.md`.

## Coding standards

Read `CODING_STANDARDS.md` before writing or reviewing code under `apps/`, `packages/` or `deploy/`. Backend: layout by domain area, dependency rule, use cases, domain modelling, database access and migrations, background jobs, files, e-mail, naming, test seam. Frontend: layout by domain area, where each kind of state lives, the data module, forms, tests. Delivery: response compression and cache headers, both set in the `Caddyfile`. Observability: how telemetry is collected, kept and shown, and what must never appear in it; read it before changing the collector, Grafana or what a container logs. Versions: where the version of each tool, library and container image is recorded, and how it is updated; read it before adding or updating a dependency.
