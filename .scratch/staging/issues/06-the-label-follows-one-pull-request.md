# 06: The label follows one pull request

**What to build:** Staging shows one pull request at a time. Adding the `staging` label to a pull request that is already open publishes it. Publishing a pull request takes the label off any other open pull request that has it, with a comment there saying which pull request staging now shows, so that a push to the old one never publishes it back. See the spec, section The pipeline.

**Blocked by:** 05.

**Status:** ready-for-agent

- [ ] The workflow also runs when a label is added to a pull request. Adding `staging` to an open pull request runs its checks again and publishes it.
- [ ] After a publish succeeds, the label comes off every other open pull request that has it, and each of them gets a comment naming the pull request staging now shows.
- [ ] A push to a pull request whose label was taken off publishes nothing.
- [ ] Adding another label does not publish a pull request without `staging`.
- [ ] Tried for real with two pull requests, or recorded under "What was not checked" with the reason.

## Comments

### From ticket 05

- The publish ends in `announce` (`apps/api/tools/environment/publish.go`), which comments on the published pull request through the GitHub interface once `/up` answered. Taking the label off the other open pull requests, with a comment there, fits right after it: `LabelledPullRequests` exists, and the interface needs a way to remove a label (`gh pr edit <n> --remove-label staging`, or `DELETE /repos/{owner}/{repo}/issues/{n}/labels/staging` through `gh api`), with its fake in `githubtest`. The job already has `pull-requests: write`.
- The job `staging` of `ci.yml` runs only when `github.event.pull_request.labels` has `staging`, which on a `labeled` event includes the label just added. Its first step skips the rest when staging is off.
- The workflow's `concurrency` no longer cancels the run before when the pull request has the label. With `labeled` among the triggers, adding any label to a pull request without `staging` would still cancel its checks in flight, and adding `staging` starts a run that waits for the one in flight.
- Until `labeled` is a trigger, a labelled pull request is published by a push to it, or by closing and reopening it.
