---
description: The question that opens a pull request also offers staging; with staging, turn it on before the push and label the pull request only once it is on; staging ends only when the maintainer says so
---

# Every pull request gets the choice of staging

Whenever an agent is about to push a branch and open a pull request in this repository (`dannilodms/forgespark` on GitHub, named `software-factory-manager` until 2026-11-25), worktrees of it included, it asks the maintainer one question with three answers: open it and publish it to staging, open it without staging, or do not open it.

Reason: staging is a server that exists only while it is on, and it costs by the hour. The pipeline publishes a pull request that has the `staging` label once its checks pass, and publishes nothing while staging is off. Turning staging on before the push, and putting the label on only once it is on, keeps every labelled pull request with a staging to go to, and a failed turn-on leaves nothing half-done on GitHub. The pipeline cannot reach the provider, so staging ends when the maintainer ends it.

Apply proactively — without the maintainer having to ask.

## The question

1. Ask it with AskUserQuestion, in Portuguese, the recommended answer first and marked "(Recomendado)". It is the question that asks to push and open the pull request; commits still follow the user rule `~/.claude/rules/no-unsolicited-commits.md`.
2. Recommend "without staging" for a pull request that changes only documents, as the job `changes` of `.github/workflows/ci.yml` judges them, and "with staging" for any other.
3. The answer "with staging" carries the hourly price of the staging server, read today as [provider-cost-approval](provider-cost-approval.md) says; this question is the approval that rule asks for.

```text
Abrir o pull request de <branch> para <base>?
- Abrir e publicar na homologação (Recomendado): liga a homologação, faz o push e abre o pull request com o rótulo staging. A homologação custa <preço por hora> até ser desligada (preço lido em <link>, em <data>); se já mostra outro pull request, ele perde o rótulo.
- Abrir sem homologação: faz o push e abre o pull request sem o rótulo.
- Não abrir: não faz o push nem abre o pull request.
```

## With staging

1. Run `task staging:on` in the main checkout or in a worktree of it: either way the tool reads the `.env` of the main checkout and keeps the deploy key there. It takes up to about ten minutes, so run it in the background and wait for it to exit.
2. An exit status other than 0 means staging is not on: report what the command printed and stop before the push. The command has already destroyed what it created; what comes next is the maintainer's call.
3. With exit status 0, push, then open the pull request with `STAGING_QUESTION=with gh pr create --label staging`, the label in that same command: the concurrency of `ci.yml` counts on it to publish the pull request once. The hook `.claude/hooks/staging-question.js` refuses a `gh pr create` of this repository that does not declare the answer this way (in PowerShell, `$env:STAGING_QUESTION='with'; ` before the command); declare it only once the maintainer answered.
4. Follow the publish. `gh run list --workflow ci.yml --branch <branch>` shows the run of `opened` cancelled by the run of `labeled`. Follow the newer one with `gh run watch <id> --exit-status`, in the background, until its job `Publish to staging` ends, about twenty minutes after the push.
5. Report how it ended:
   - Published: the links of the comment the job wrote on the pull request (`gh pr view <n> --comments`): the app, and Mailpit, whose user is `maintainer` and whose password is `STAGING_MAILPIT_PASSWORD` in the `.env` of the main checkout, given by that name; and the user of the seed to log in as, `seed+admin@fabrica.example`, with a login code read in Mailpit.
   - Not published: the check that failed, the job's notice that staging is off, or the step that failed, from `gh run view <id> --log-failed`.

## Without staging

Push, and open the pull request with `STAGING_QUESTION=without gh pr create`, without the label.

## Ending staging

1. Staging stays on until the maintainer says to turn it off: run `task staging:off` only after the maintainer said yes to it.
2. When the maintainer says the pull request was merged or closed, ask with AskUserQuestion, in Portuguese, whether to turn staging off. Run `task staging:status` first, and recommend turning it off unless another open pull request has the label.

`gh` is at `C:\Program Files\GitHub CLI\gh.exe` when the PATH of the shell lacks it.
