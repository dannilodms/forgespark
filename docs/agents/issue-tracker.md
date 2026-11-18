# Issue tracker: Local Markdown

Issues and specs for this repo live as markdown files in `.scratch/`.

## Conventions

- One feature per directory: `.scratch/<feature-slug>/`
- The spec is `.scratch/<feature-slug>/spec.md`
- Implementation issues are one file per ticket at `.scratch/<feature-slug>/issues/<NN>-<slug>.md`, numbered from `01`, never a single combined tickets file
- Triage state is recorded as a `Status:` line near the top of each issue file (see `triage-labels.md` for the role strings)
- Comments and conversation history append to the bottom of the file under a `## Comments` heading

## When a skill says "publish to the issue tracker"

Create a new file under `.scratch/<feature-slug>/` (creating the directory if needed).

## When a skill says "fetch the relevant ticket"

Read the file at the referenced path. The user will normally pass the path or the issue number directly.

## When a skill says "close" or "resolve" a ticket

Follow these steps for any ticket whose work is built and checked, whether or not a skill asked. Make the changes in the same change as the ticket's code.

1. Tick each acceptance criterion the work meets (`- [x]`).
2. Append these sections under `## Comments`, leaving out any section with nothing to say:
   - `### What was built`
   - `### Decided while building`: each decision the ticket left open, with its reason.
   - `### What was checked`
   - `### What was not checked`: each criterion left unticked, and why.
   - `### Found, and left for a decision`: whatever the user still has to settle.
3. Move each decision that should outlast the effort to the file that keeps it: an ADR in `docs/adr/`, a term in `GLOSSARY.md`, a rule in `CODING_STANDARDS.md` or `DESIGN.md`. Point the comment at it. The spec stays as written.
4. Set `Status: resolved`.

Done when every criterion is either ticked or named under "What was not checked", and `Status:` reads `resolved`.

## When a spec ends

A spec and its tickets are a snapshot of what was known before the build. The first change after it makes them stale, and an agent that reads a stale spec trusts it. When every ticket of a spec is `resolved`:

1. Settle each item under "Found, and left for a decision" in its tickets. Ask the user. Then put the answer into the code or into the file that keeps it. If the user defers it, it goes into the spec of the next effort or into a new ticket.
2. Check that `.scratch/<feature-slug>/` holds nothing that should last: each lasting decision is in an ADR, `GLOSSARY.md`, `CODING_STANDARDS.md` or `DESIGN.md`.
3. Set `Status: resolved` in the spec.
4. After the effort's code review, remove `.scratch/<feature-slug>/` in a commit of its own, as the last commit before the effort's pull request is merged. The folder then leaves with the effort that used it. Git keeps it: `git log --diff-filter=D --name-only -- .scratch/` lists the removed files, and `git show <commit>^:<path>` reads one back.

Done when the folder is gone and nothing in it was the only record of a decision.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a file with one **child** file per ticket.

- **Map**: `.scratch/<effort>/map.md` (the Notes / Decisions-so-far / Fog body).
- **Child ticket**: `.scratch/<effort>/issues/NN-<slug>.md`, numbered from `01`, with the question in the body. A `Type:` line records the ticket type (`research`/`prototype`/`grilling`/`task`); a `Status:` line records `claimed`/`resolved`.
- **Blocking**: a `Blocked by: NN, NN` line near the top. A ticket is unblocked when every file it lists is `resolved`.
- **Frontier**: scan `.scratch/<effort>/issues/` for files that are open, unblocked, and unclaimed; first by number wins.
- **Claim**: set `Status: claimed` and save before any work.
- **Resolve**: append the answer under an `## Answer` heading, set `Status: resolved`, then append a context pointer (gist + link) to the map's Decisions-so-far in `map.md`.
