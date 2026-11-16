---
status: accepted
---

# A public repository under AGPLv3

The repository is public on GitHub, on the free plan, and its code is licensed under the GNU Affero General Public License, version 3. A public repository gets, at no cost, what the pipeline depends on: secrets for each environment (ADR-0019), a protected `main` and unlimited minutes of GitHub Actions. On the free plan a private repository has no secrets for each environment. The AGPL was chosen because whoever offers a modified copy of the system as a service has to publish the changes.

Being public sets rules for everything that is committed and for the pipeline:

- **Nothing about a client is committed.** Issues, ADRs, seed data and examples name no real client, person or address.
- **A pull request from a fork gets no secret.** No workflow uses the `pull_request_target` event, a workflow of an outside contributor waits for approval, and only a branch of this repository is published to staging.
- **The logs of a workflow are public.** A publish prints no configuration, and production data is never restored on a GitHub runner (ADR-0020).
- **A secret is refused at the push.** GitHub's secret scanning with push protection is turned on.
- **Commits carry no personal address.** They use the maintainer's `noreply` address at GitHub. The commits made before the first push are rewritten with it.
- **The container images are public too**, in the GitHub container registry. They hold nothing that the repository does not.

## Considered Options

- **A private repository with a paid plan.** Code, decisions and history stay closed, for a monthly subscription.
- **A private repository on the free plan.** No subscription, but secrets would be shared by the whole repository, told apart by a prefix, and nothing would stop a push to `main`.
- **Public with no licence.** Anyone could read the code and nobody could use it. It can be opened later, but the maintainer chose to open it now.
- **MIT.** Anyone could use the code, a closed competing product included.

## Consequences

- A version published under the AGPL stays under it. The licence of later versions can change only while the maintainer holds the copyright of all the code.
- A contribution from outside belongs to its author. Whether contributors assign their rights or sign an agreement is decided before the first one is accepted.
- Every library added has to have a licence compatible with the AGPL. Garage, also under the AGPL, runs as a separate service and is not part of this code.
- Anyone can decode a public identifier and see roughly how many records of a type exist, as ADR-0010 already accepted.
- The files in `deploy/` show how the system is put together, and `docs/providers.md` lists the providers with the addresses, host names and DNS records set at each. Anyone can read them. No key, token or password is ever written there.
