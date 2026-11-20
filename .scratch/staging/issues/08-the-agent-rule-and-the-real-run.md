# 08: The agent's rule and the real run

**What to build:** An agent on this machine gives every pull request the choice of staging. In the same question that asks to open a pull request, it offers to open it with staging, without staging, or not at all. With staging, it turns staging on from the main checkout before pushing, and opens the pull request with the label only if that worked. Then it follows the publish and gives the links. It never turns staging off by itself, and asks whether to once the maintainer says the pull request was merged or closed. The effort ends with the whole path tried for real. See the spec, sections The agent and Records.

**Blocked by:** 05, 06, 07.

**Status:** ready-for-agent

- [ ] A global rule of Claude on this machine, written for this repository, says what the description above says. It recommends "without staging" for a pull request that changes only documents.
- [ ] The rule says that a failure to turn staging on is reported and stops the agent before the push.
- [ ] The `README.md` says how to turn staging on and off, what the machine needs (`gh` logged in, `HETZNER_TOKEN` in `.env`) and how to read the password of Mailpit.
- [ ] `CODING_STANDARDS.md`, `docs/providers.md` and ADR-0019 match what was built. `docs/providers.md` marks `sslip.io` as provisional and points to the ticket of the domain.
- [ ] The permissions of the `gh` token have one home: a row "Token of `gh` on the maintainer's machine" in the table of the GitHub section of `docs/providers.md`. The row says the token is fine-grained, with access to `dannilodms/software-factory-manager` alone and no account permission. It gives each repository permission below with its level, and the expiration date, "none" until the maintainer makes the token again. It never gives the value. Both READMEs point to that row instead of listing permissions: the paragraph of `task staging:on` in Deploy, which lists three since be4939d, and, if it fits, the row of the GitHub CLI in the prerequisites. The list, as the maintainer chose it on 2026-11-19 (relayed by another session of this repository):
  - Metadata: read (required).
  - Contents: read. Compares branches when a pull request is opened, and reads releases. Not write: agents never merge, and tags and releases come from the workflow.
  - Pull requests: read and write. Open and edit pull requests, put the label on and take it off, create the label of the repository and comment on a pull request. Those endpoints accept Issues or Pull requests write (`X-Accepted-GitHub-Permissions: issues=write; pull_requests=write`).
  - Actions: read and write. Runs, logs, rerun, cancel and `gh workflow run`; read also covers reading an environment.
  - Secrets: read and write. The repository secret `RESEND_TEST_API_KEY`.
  - Environments: read and write. The secrets and variables of `staging`, and later of `production`.
  - Administration: read and write. Creating the environments `staging` and `production`. The maintainer accepted the risk that it also allows deleting the repository and changing its rulesets.
  - Left out: Issues (Pull requests covers it), Workflows (git goes over SSH, and a token counts only for git over HTTPS), Variables (those of an environment come with Environments), Webhooks, and Packages (a fine-grained token cannot reach the GHCR; the pipeline uses `GITHUB_TOKEN`).
- [ ] The real run, done with the maintainer and recorded here:
  1. the rule's question is asked;
  2. staging is turned on;
  3. this effort's pull request is opened with the label and published;
  4. a login code is read in Mailpit and the shell opens;
  5. staging is turned off;
  6. nothing is left at the provider, and only the password of Mailpit is left in GitHub.
