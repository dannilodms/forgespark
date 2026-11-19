# 02: Staging's values in GitHub

**What to build:** Turning staging on also gives GitHub what the pipeline needs to reach the server, and nothing else. In the GitHub Environment `staging` it writes the deploy key and the password of Mailpit as secrets, and the address, the host key and the time staging was turned on as variables. It also makes sure the environment and the `staging` label exist. Turning off, or a failed turn-on, takes those values back out. The status names the pull request that has the label. See the spec, sections The tool and The pipeline.

**Blocked by:** 01.

**Status:** ready-for-agent

- [ ] GitHub is an interface: make sure the environment and the label exist, set and delete a secret or a variable of the environment, and find the open pull request with the label. It has an in-memory implementation for the tests and an adapter that calls `gh`.
- [ ] The `gh` adapter passes a secret value through standard input, never as an argument. A test against a stand-in executable that records its arguments and its input proves it.
- [ ] Turning on writes the deploy private key and the password of Mailpit as secrets of `staging`, and the address, the host key as a line of known hosts and the time staging was turned on as variables. Turning on when staging is already on writes them again.
- [ ] Turning off, and a turn-on that fails after writing, delete the address, the host key, the deploy key and the time. The password of Mailpit stays.
- [ ] The password of Mailpit is generated the first time staging is turned on and kept in the `.env` of the main checkout, under a name of its own that is not in `.env.example`. The next turn-on reuses it.
- [ ] The status shows the open pull request that has the `staging` label, or none.
- [ ] The token of the provider never reaches GitHub: a test checks that nothing written to the fake GitHub holds it.
- [ ] After a real turn-on, `gh secret list --env staging` and `gh variable list --env staging` show the names above. After a real turn-off, only the password of Mailpit is left.
- [ ] `docs/providers.md` lists the environment `staging`, the label, and the names of its secrets and variables, with no value.
