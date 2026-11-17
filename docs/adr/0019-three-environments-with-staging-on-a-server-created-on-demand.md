---
status: accepted
---

# Three environments, with staging on a server created on demand

The system runs in three environments, each with its own Compose file: development on the maintainer's machine, staging on a server that exists only while a branch is being looked at, and production on one server that is always on. Both servers are Hetzner Cloud virtual machines prepared by the same cloud-init file, and GitHub Actions publishes to them over SSH. The lowest price decided the provider and the region.

- **Development runs the tools on the machine.** Go, Vite+ and Task run on Windows, so an agent runs a test without going through Docker. The development Compose file declares every service the code talks to.
- **Staging shows one branch.** A pull request that receives the `staging` label gets a new server with that branch on it, and every later push to the pull request is published again. There is one staging environment: labelling another pull request replaces what was there.
- **Staging data is the seed.** Every publish to staging recreates the database and runs the seed. Nothing from production is ever copied there.
- **Staging does not outlive its use.** The server is destroyed when the label is removed or the pull request closes. A scheduled sweep destroys any staging server older than 24 hours, so a forgotten one stops costing money.
- **Staging is public, with the real login.** A person logs in through Google, Microsoft or GitHub, or with a login code. The e-mail goes to Mailpit, which has a host name of its own behind a password asked by the browser. The seed creates its users with addresses derived from one mailbox taken from configuration, so the maintainer can log in as any of them.
- **Production receives `main`.** A merge into `main` is published to production once the pipeline passes. Nothing else publishes there.
- **One cloud-init file prepares every server**, applied through the provider's command-line tool: Docker, the deploy user, the firewall and automatic security updates. Each staging server is a rehearsal of rebuilding production from nothing.
- **Staging has a fixed address.** A reserved IP address moves from one staging server to the next, and its DNS records are set once, by hand.
- **Publishing is one workflow.** GitHub Actions builds the images, tags them with the commit, pushes them to the GitHub container registry, and logs in to the server over SSH to pull them and restart the Compose services. The Compose files stay the description of the deploy.
- **Secrets live in GitHub Environments**, `staging` and `production`. A value can be written and never read back. The secrets of `production` can be used only by a run of `main`. The publish writes the variables file on the server, after checking that every name in `.env.example` has a value.
- **The seed is its own command, in its own image.** It creates data by calling use cases, so seeded data obeys the domain rules. The production Compose file does not declare it, the same way it does not declare the local login provider (ADR-0011).
- **A stand-in answers for the LLM provider outside production.** It is a small server written in this repository that answers in the Chat Completions format (ADR-0004) with fixed answers, streamed ones included, and lists models for the health check. Development, the pipeline and staging use it. The production Compose file does not declare it.

What each environment runs:

| Service | Development | Staging | Production |
| --- | --- | --- | --- |
| API, worker, `migrate` | run by Task, on the machine | the image of the branch | the image of `main` |
| Caddy | container | container | container |
| PostgreSQL | container | container, recreated on every publish | container, backed up (ADR-0020) |
| File storage | Garage | Garage | Backblaze B2 (ADR-0020) |
| E-mail | Mailpit | Mailpit, behind a password | Resend |
| Login | Dex, or a login code read from Mailpit | the three providers, or a login code read from Mailpit | the three providers, or a login code sent by Resend |
| LLM provider | the stand-in | the stand-in | the real provider |
| Telemetry | a Compose profile, off by default | none | none yet (ADR-0012) |
| Seed | yes | on every publish | never |

## Considered Options

- **Staging as a gate before production.** `main` would go to staging by itself and to production by a manual promotion. The maintainer wants a merge into `main` to reach production.
- **One environment for each pull request.** More isolation, with a wildcard DNS record, a cleanup routine and more memory to pay for.
- **Staging on the maintainer's machine only.** Nothing remote besides production, and nothing to show from another device.
- **A second server that is always on.** Simpler than creating one each time, for a monthly fee paid while nothing is being looked at.
- **Staging on the production server.** No cost at all, but a branch that nobody has reviewed would run beside the data and the secrets of the clients.
- **Staging closed to the internet, on a private network.** Nothing to protect, but it needs one more account and it shuts out anyone the maintainer wants to show a branch to.
- **Staging behind a browser password on the whole site.** The password travels in the same header as the access token of an agent, so the API and MCP would stop working for agents.
- **The local login provider in staging.** A simple way in for automated tests, but anyone who found the address could log in as any seeded user.
- **Resend in staging.** The login code would reach a real mailbox, with a second key to keep and real e-mail leaving a test environment.
- **A real LLM provider outside production, with a spending limit on each key.** LLM features could be tried before a merge. The maintainer chose to spend nothing outside production.
- **Scheduled tests of the LLM adapter against a real provider.** The first version of this decision. A change at the provider would show before production met it, for a second key and a little spent every day.
- **A small local model as the stand-in.** Real answers at no cost per call, but it takes gigabytes of memory, its answers change from one run to the next, and it does not fit the staging server.
- **A server in São Paulo or in the United States.** Lower latency for people in Brazil, at a higher price.
- **Kamal.** Deploys with no interruption through a proxy of its own, which competes with Caddy, and the Compose files would no longer describe production.
- **Coolify or Dokploy.** A panel on the server that pulls and publishes. It is one more system to operate, and part of the configuration would live in the panel.
- **The server pulling new images by itself.** No SSH key in GitHub, but nothing orders the dump, the migration and the restart.
- **SOPS and age in the repository.** A change of secret is a commit, but the private key on the maintainer's machine opens the production secrets to any agent running there.
- **A variables file placed on the server by hand.** Nothing in GitHub but the SSH key. Every new variable becomes a manual step that, when forgotten, stops the API at the start.
- **A hosted secret manager.** A panel and a history of changes, with one more provider.
- **OpenTofu or Terraform.** Servers, firewall and DNS declared in code, with a state file to keep and one more tool in the pipeline.
- **Production set up by hand.** Rebuilding it would depend on a written procedure being up to date.
- **A Dev Container, or every service of development in Compose.** The same Linux as the pipeline. The first means setting up the agent's own configuration again inside the container. The second makes an agent run every test through `docker compose exec`.
- **The seed as a command of the API binary, as SQL files, or as an anonymised copy of production.** The first puts the command in production. The second goes around the domain rules. The third needs an anonymisation routine and carries client data out of production.
- **DNS updated by the pipeline.** No reserved address to pay for, but the DNS would have to be at a provider with an API, with one more secret.
- **SSH through a private network, or a key that can only start the publish script.** Either would limit what a leaked key can do. The first is one more account, the second a more complex script.

## Consequences

- The servers are in Europe. Every request from Brazil crosses the ocean, which was accepted for the price.
- The plan chosen (CX23) had limited stock when this was decided, and its page showed it as unavailable. Hetzner raised its prices twice in 2026 and limits the resources of a new account in its first month. The price of a reserved IP address was not confirmed.
- Port 22 is open on both servers, with key login only, no password and no root. Whoever holds the deploy key controls the server.
- A secret cannot be read back from GitHub. The two that cannot be made again, the key that encrypts credential values (ADR-0018) and the passphrase of the backups (ADR-0020), are also kept in the maintainer's password manager.
- A new configuration variable needs its value set in both environments before the merge. When one is missing, the publish stops before it restarts anything.
- Let's Encrypt refuses more than five identical certificates in a week, and each staging server would ask for the same ones. Caddy's certificate data is kept in a private bucket between one staging server and the next. This was not tried.
- Staging needs its own application registered with each of the three login providers.
- An automated test cannot get past a real login provider, so nothing automated logs in to staging. The end-to-end tests run in the pipeline, against the development Compose file.
- An LLM feature meets a real provider only in production. Nothing automated calls one anywhere else, so the tests of the LLM adapter run against the stand-in, and a change at the provider shows first in production. The one exception is the maintainer pointing the development configuration at a real provider, by hand, to work on a prompt.
- The three Compose files extend one base file, so the version of each container image keeps a single home. Tried with development, which extends the base for PostgreSQL, Garage, the step that sets the rule of the bucket, and Caddy. It holds, with two things the base cannot hold: `extends` copies a service and nothing else, so the volumes and the configs a service names are declared in the file that extends it; and an image built in this repository, Caddy's, has its versions in its Dockerfile, and the file that builds it names the result with `build` and `image` together.
- Going back to an earlier version is a manual workflow that publishes an earlier commit of `main`. Going back in the data is a restore (ADR-0020).
- Earlier documents call this environment staging. The label on a pull request and the GitHub Environment use the same name.
