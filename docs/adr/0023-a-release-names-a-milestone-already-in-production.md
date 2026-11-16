---
status: accepted
---

# A release names a milestone already in production

Every merge into `main` is published to production (ADR-0019), so a release publishes nothing new. It gives a number and notes to a milestone the maintainer chose. It has two readers: the maintainer and the agents, who get named points in the history, and whoever reads the public repository or installs the system from it (ADR-0021). The number lives in a `VERSION` file and a release is a pull request that changes it, so the build knows its own version and the pipeline creates the tag.

- **One version for the whole system.** The API, the worker, the web client and the images come from one commit and are published together.
- **Semantic Versioning, below 1.0.0 for now.** The number reads `0.MILESTONE.FIX`: a milestone raises the middle number, and a release that only fixes raises the last. 1.0.0 is the first release the consultancy uses with real clients. What raises the first number after that is decided then.
- **The maintainer decides when.** A release is made when a deliverable is complete in production. It does not follow each merge or a calendar.
- **A release is a pull request.** It changes `VERSION` and `CHANGELOG.md` and nothing else, so it marks code that is already running in production. An agent opens it when asked, proposing the number and the notes, and the maintainer merges it.
- **The pipeline creates the tag.** After a publish to production succeeds, when no tag exists for the number in `VERSION`, the pipeline creates the tag `v` plus the number on that commit, and a GitHub Release whose text is that version's section of `CHANGELOG.md`. It also adds the number as a second tag to the images of that commit, without building them again.
- **A failed publish gets no tag.** The tag goes to the next commit published with that number, so a tag never points at code that did not run.
- **Between releases.** A build shows the number of the last release it contains. Its commit, shown beside the number, tells one build from another.
- **Notes.** `CHANGELOG.md` is in English. The agent writes each section from the pull requests merged since the last release: what changed for someone who uses the system and, apart from that, what someone who installs it has to do.
- **Only forward.** A tag is never moved or deleted, and a ruleset on GitHub refuses both. A fix goes into `main` and becomes a new release. There are no release branches, no fixes to an older version and no pre-releases.
- **The first release.** 0.1.0 is the walking skeleton running in production. `VERSION` starts at 0.0.0 with the first code, and nothing is tagged before that.

## Considered Options

- **A tag that publishes to production.** A merge would stop publishing. ADR-0019 already dropped a manual promotion.
- **A tag created by hand, with the version read from Git.** No file and no pull request. The image in production is built before the tag exists, so it would show the release before and a commit hash.
- **A tag that builds and publishes the same commit again.** The build would know its number, but nothing would review the number or the notes before the tag exists.
- **A release for every merge that changes what a user sees.** Made from the commit titles, with nothing to decide. It gives a release for almost every deploy, which the commit hash already names.
- **A release on a calendar.** One for each month, whatever was merged in it.
- **A version by date, or a plain sequence.** A date says when and nothing else. A sequence has no place for a fix. Neither tells a milestone from a fix.
- **A version for each part.** It fits parts that are distributed apart, and these are not.
- **The version in the root `package.json`, or as the first heading of the changelog.** The first makes Go and the pipeline read a file of another toolchain. The second makes the build depend on the layout of a document.
- **A development version between releases, such as 0.4.0-dev.** A build after a release would say so, at the price of a second pull request for each release.
- **Notes only in the GitHub Release, made from the pull request titles.** No file to keep, but nobody reviews them, and an agent would have to leave the repository to read them.
- **Notes in English and in Portuguese.** Two texts to keep equal at every release.
- **Skipping a number whose publish failed.** Simpler to state, but it leaves a section in the changelog with no tag.
- **A release pull request that also carries code.** One merge fewer, and a release that marks code production has not run yet.

## Consequences

- A release is one more publish: the release pull request reaches production like any other, with the same code and a new number.
- `VERSION` is not a document. A pull request that changes it runs the builds and the tests.
- Between releases the number does not identify a build. The commit does.
- The job that creates the tag and the GitHub Release writes to the repository and to the container registry. No other job gets that permission.
- The workflow that goes back to an earlier version takes a release as well as a commit.
- "Release" is a term of the pipeline and stays out of `GLOSSARY.md`. The projects of clients may get versions and deliveries of their own, and the term is left free for them.
- `CHANGELOG.md` starts at 0.1.0. What came before is in the ADRs and in the Git history.
- Adding a tag to an image that is already in the registry, and the ruleset that protects the tags, were not tried. Both are proven with the walking skeleton, before 0.1.0.
