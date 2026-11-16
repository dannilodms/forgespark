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

Read `CODING_STANDARDS.md` before writing or reviewing code under `apps/`, `packages/`, `deploy/` or `.github/`. Backend: layout by domain area, dependency rule, use cases, domain modelling, login, configuration, database access and migrations, background jobs, files, credentials, e-mail, naming, test seam. Frontend: layout by domain area, where each kind of state lives, the data module, forms, tests. Delivery: response compression, cache headers, security headers and the limit on the login code, all set in the `Caddyfile`. Observability: how telemetry is collected, kept and shown, and what must never appear in it; read it before changing the collector, Grafana or what a container logs. Environments: the list of providers in `docs/providers.md`, what development, staging and production each run, the seed, the servers, secrets and backup; read it before changing a Compose file or anything in `deploy/`. Pipeline: how a change reaches `main` and the servers, and the rules a public repository sets; read it before changing a workflow in `.github/`. Versions: where the version of each tool, library and container image is recorded, and how it is updated; read it before adding or updating a dependency.
