# 06: The label follows one pull request

**What to build:** Staging shows one pull request at a time. Adding the `staging` label to a pull request that is already open publishes it. Publishing a pull request takes the label off any other open pull request that has it, with a comment there saying which pull request staging now shows, so that a push to the old one never publishes it back. See the spec, section The pipeline.

**Blocked by:** 05.

**Status:** ready-for-agent

- [ ] The workflow also runs when a label is added to a pull request. Adding `staging` to an open pull request runs its checks again and publishes it.
- [ ] After a publish succeeds, the label comes off every other open pull request that has it, and each of them gets a comment naming the pull request staging now shows.
- [ ] A push to a pull request whose label was taken off publishes nothing.
- [ ] Adding another label does not publish a pull request without `staging`.
- [ ] Tried for real with two pull requests, or recorded under "What was not checked" with the reason.
