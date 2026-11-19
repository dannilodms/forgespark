# 08: The agent's rule and the real run

**What to build:** An agent on this machine gives every pull request the choice of staging. In the same question that asks to open a pull request, it offers to open it with staging, without staging, or not at all. With staging, it turns staging on from the main checkout before pushing, and opens the pull request with the label only if that worked. Then it follows the publish and gives the links. It never turns staging off by itself, and asks whether to once the maintainer says the pull request was merged or closed. The effort ends with the whole path tried for real. See the spec, sections The agent and Records.

**Blocked by:** 05, 06, 07.

**Status:** ready-for-agent

- [ ] A global rule of Claude on this machine, written for this repository, says what the description above says. It recommends "without staging" for a pull request that changes only documents.
- [ ] The rule says that a failure to turn staging on is reported and stops the agent before the push.
- [ ] The `README.md` says how to turn staging on and off, what the machine needs (`gh` logged in, `HETZNER_TOKEN` in `.env`) and how to read the password of Mailpit.
- [ ] `CODING_STANDARDS.md`, `docs/providers.md` and ADR-0019 match what was built. `docs/providers.md` marks `sslip.io` as provisional and points to the ticket of the domain.
- [ ] The real run, done with the maintainer and recorded here:
  1. the rule's question is asked;
  2. staging is turned on;
  3. this effort's pull request is opened with the label and published;
  4. a login code is read in Mailpit and the shell opens;
  5. staging is turned off;
  6. nothing is left at the provider, and only the password of Mailpit is left in GitHub.
