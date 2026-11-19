# 07: The reminder after 24 hours

**What to build:** Nothing turns staging off by itself, so a scheduled workflow warns the maintainer when staging has been on for more than 24 hours. It reads only the variable holding the time staging was turned on, needs no secret and no credential of the provider, and fails when that time is too old, which is how a scheduled workflow warns. See the spec, section The pipeline.

**Blocked by:** 02.

**Status:** ready-for-agent

- [ ] A scheduled workflow runs every six hours and on a manual trigger.
- [ ] It fails, with a message giving how long staging has been on, when the time variable of `staging` is more than 24 hours old. It passes when the variable is empty or recent.
- [ ] It reads no secret, asks for no permission beyond reading, and blocks no pull request.
- [ ] actionlint and zizmor pass.
- [ ] `CODING_STANDARDS.md` (Scheduled workflows) describes the reminder in place of the sweep.
