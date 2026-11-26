# Providers

Every outside service this system depends on: why it was chosen, what it costs, and what is configured there. The decisions behind them are in ADRs [0019](adr/0019-three-environments-with-staging-on-a-server-created-on-demand.md), [0020](adr/0020-postgresql-backed-up-by-pgbackrest-and-production-files-in-backblaze-b2.md) and [0021](adr/0021-a-public-repository-under-agplv3.md).

**No secret is written here.** No key, token, password or passphrase, and no value that works as one. Addresses, host names, DNS records, regions, plan names and bucket names are fine. Where secrets live is in Secrets, in [`CODING_STANDARDS.md`](../CODING_STANDARDS.md).

**Keep it true.** A change made at a provider (a new record, a new bucket, a change of plan) is written here in the same pull request, or right after when the change is made by hand.

Prices were read on the providers' own pages on 2026-11-16, without tax, and Hetzner's from its API on 2026-11-19 and 2026-11-25. The domain's are what the maintainer paid GoDaddy and the renewal price GoDaddy gave, both told on 2026-11-25. The GitHub repository and the Hetzner project of staging were created by hand, and the domain registered by hand; the GitHub Environment and the label of staging were made by the first `task staging:on`, and the packages of the images by the first publish to staging, on 2026-11-20, and again under the name ForgeSpark by the publish of pull request #13, on 2026-11-25. The address and the volume of staging were created on 2026-11-25 through Hetzner's API, with what `task staging:on` gives them, and the DNS records of staging at Cloudflare the same day. A value shown as _not created_ is filled in when the resource exists.

## Summary

| Provider | Used for | Environments | Approximate cost | State |
| --- | --- | --- | --- | --- |
| [GitHub](#github) | Repository, pipeline, container images, secrets | all | Free | Repository created and protected; the environment `staging`, the label `staging` and the packages of the three images created; the environment `production` and the secret `RESEND_TEST_API_KEY` not created yet |
| [Hetzner Cloud](#hetzner-cloud) | The servers | staging, production | US$ 0.0104 an hour while staging is on, and US$ 1.37 a month for its address and its volume; production _not created_ | Project of staging created, with the address and the volume of staging; its server exists only while staging is on |
| [Backblaze B2](#backblaze-b2) | Production files, database backups | production | Free under 10 GB, about US$ 0.63 a month at 100 GB | Decided, not created |
| [Domain and DNS](#domain-and-dns) | The addresses of the system | staging, production | R$ 96 the first year; R$ 142 a year from the first renewal; the DNS free | `forgespark.dev` registered at GoDaddy until 2027-01-23, then at Cloudflare ([ticket 02](../.scratch/domain/issues/02-transfer-the-domain-to-cloudflare-registrar.md)); DNS at Cloudflare, with the three records of staging |
| [Let's Encrypt](#lets-encrypt) | HTTPS certificates, through Caddy | staging, production | Free | Nothing to create |
| [Resend](#resend) | Sending e-mail | production | _not checked_ | Decided, not created |
| [Login providers](#login-providers) | Login through Google, Microsoft and GitHub | staging, production | Free | Decided, not created |
| [LLM provider](#llm-provider) | Calls to language models | production | By use | Not chosen |
| [Hosted monitor](#hosted-monitor) | Warning when production or a backup stops | production | Free plan intended | Not chosen |

Expected total while the system is small, from the prices below: about US$ 7.10 a month for the production server and its IPv4 address, US$ 1.37 a month for the address and the volume of staging, about US$ 0.25 for each day staging is on, and nothing for Backblaze B2 under 10 GB, plus the domain, R$ 142 a year from its first renewal, and what Resend and the LLM provider charge.

## GitHub

**Why.** The repository is public, and a public repository on the free plan has what the pipeline needs: secrets for each environment, a protected `main` and unlimited minutes of GitHub Actions (ADR-0021).

**Cost.** Free.

| Item | Value |
| --- | --- |
| Account or organisation | `dannilodms`, a personal account |
| Repository | [`dannilodms/forgespark`](https://github.com/dannilodms/forgespark), public. Named `software-factory-manager` until 2026-11-25, when the product became ForgeSpark; GitHub redirects the old address |
| Description | "ForgeSpark — From ideas to execution." |
| Social preview | [`social-preview.png`](design/brand/social-preview.png), uploaded by hand in the settings of the repository: GitHub has no API for it |
| Licence | AGPLv3 |
| Branch protection on `main` | The `main` ruleset: no direct push, a pull request with a squash merge, no force push, no deletion, and a merge only when the `Checks` check of GitHub Actions has passed. No approving review is required, and its pull request rule has "extra approval for unattributed changes" off (`require_extra_approval_for_unattributed_changes: false`), set through the API on 2026-11-25: GitHub turns it on when a ruleset is created without it, and once enforced it blocked every pull request with commits of an agent, which the maintainer opens and cannot approve |
| Merge | Squash only; the branch is deleted after the merge |
| Environment `staging` | Created by `task staging:on` when it does not exist, with no protection rule and no branch restriction, so that a run of a pull request can use it, the job `Publish to staging` of `ci.yml`, and so that the scheduled `staging-reminder.yml`, which reads `TURNED_ON_AT` there and records no deployment, neither waits nor fails; an environment that exists is not changed. A run of the reminder creates it the same way, empty, when it does not exist yet. While staging is on it holds the secrets `DEPLOY_KEY` (the private deploy key of the server) and `MAILPIT_PASSWORD` (the password of Mailpit), and the variables `SERVER_ADDRESS` (the IPv4 address of the server), `SERVER_HOST_KEY` (its host key, as a line of known hosts) and `TURNED_ON_AT` (when the server was created). `task staging:off` deletes all but `MAILPIT_PASSWORD`. What each one holds is in Secrets, in [`CODING_STANDARDS.md`](../CODING_STANDARDS.md) |
| Environment `production` | Restricted to `main`. _Not created_ |
| Secret scanning with push protection | On |
| Dependabot alerts | On. Its automatic pull requests are off |
| Private vulnerability reporting | On; the way to report is in [`SECURITY.md`](../SECURITY.md) |
| Actions | An action runs only when pinned to a full commit hash. The token of a workflow is read-only unless the workflow asks for more, and cannot approve a pull request |
| Workflows of outside contributors | Run only after approval, for every outside contributor |
| Tag protection | A ruleset on the tags that start with `v`: no update and no deletion (ADR-0023). _Not created_ |
| Releases | Created by the publish workflow, one for each tag, with the notes from `CHANGELOG.md` and no attached file. _None yet_ |
| Container registry (GHCR) | Public images, tagged with the commit hash: the packages `forgespark-api`, `forgespark-caddy` and `forgespark-seed` of `ghcr.io/dannilodms`, pushed by the job `Publish to staging` of `ci.yml` with the token of the job, and linked to the repository by the label `org.opencontainers.image.source` of each image. The server pulls them without logging in, so each must be public. The first publish to staging created the three of the first name, `software-factory-manager-api`, `-caddy` and `-seed`, on 2026-11-20, public from their first push, with nothing done by hand; the three of the name ForgeSpark were created by the publish of the rename itself, pull request #13, on 2026-11-25, public from their first push as well. The three of the first name are still there, holding the images of 2026-11-20, which nothing pulls any more, until they are deleted by hand, in their settings on GitHub. A new package that came up private would be made public by hand, once, in its settings on GitHub (Danger Zone, Change visibility). The images of a release also carry its number |
| Label that publishes a pull request to staging | `staging`, color `FBCA04`, described as "Publish this pull request to staging once its checks pass". Created, or given that description and color, by `task staging:on`. Adding it to an open pull request publishes that pull request, and each publish takes it off every other open pull request, with the token of the job |
| Commit address | The maintainer's `noreply` address |
| Repository secret `RESEND_TEST_API_KEY` | The key of Resend's test account that the scheduled workflow `resend.yml` gives the tests that call Resend (see Resend). _Not created_: until it exists those tests skip and the run passes |
| Token of `gh` on the maintainer's machine | The token with which `gh`, logged in as the maintainer, opens pull requests, reads the pipeline and turns staging on and off (the value is a secret). Fine-grained, with access to `dannilodms/forgespark` alone and no account permission. Expiration date: none, until the maintainer makes the token again. Its repository permissions: Metadata, read (required). Contents, read: compares branches when a pull request is opened, and reads releases; not write, because agents never merge, and tags and releases come from the workflow. Pull requests, read and write: opens and edits pull requests, puts the label on and takes it off, creates the label of the repository and comments on a pull request; those endpoints accept Issues or Pull requests write (`X-Accepted-GitHub-Permissions: issues=write; pull_requests=write`). Actions, read and write: runs, logs, rerun, cancel and `gh workflow run`; read also covers reading an environment. Secrets, read and write: the repository secret `RESEND_TEST_API_KEY`. Environments, read and write: the secrets and variables of `staging`, and later of `production`; without it GitHub answers 403 to the first secret, and `task staging:on` destroys the server it created. Administration, read and write: creates the environments `staging` and `production`; the maintainer accepted the risk that it also allows deleting the repository and changing its rulesets. Left out: Issues (Pull requests covers it), Workflows (git goes over SSH, and a token counts only for git over HTTPS), Variables (those of an environment come with Environments), Webhooks, and Packages (a fine-grained token cannot reach the GHCR; the pipeline uses `GITHUB_TOKEN`) |

## Hetzner Cloud

**Why.** The lowest price found for a server of this size that is billed by the hour, which is what lets the staging server exist only while it is used. The tool of the environments reaches its API through hcloud-go, from the maintainer's machine: GitHub Actions holds no credential of Hetzner and knows only a server it reaches over SSH (ADR-0019).

**Cost.** Read from Hetzner's API on 2026-11-19 and 2026-11-25, in US dollars, without VAT; the account is billed in dollars. CX23 (2 vCPU, 4 GB, 40 GB): US$ 0.0104 an hour, up to US$ 6.49 a month. A primary IPv4 address: US$ 0.001 an hour, up to US$ 0.60 a month, whether a server has it or not. A volume: US$ 0.0767 a gigabyte a month, so US$ 0.77 for the 10 GB of staging. The server of staging costs about US$ 0.25 a day while it is on; its address and its volume, US$ 1.37 a month together, on or off. Production is priced when its server is created.

**To check before ordering.** A new account has its resources limited in its first month. On 2026-11-19 the CX23 was offered in `fsn1`, `nbg1` and `hel1`; moving staging to another of them is a change to `deploy/server.staging.yaml`.

| Item | Value |
| --- | --- |
| Project of staging | `forgespark-staging`, for staging alone; named `software-factory-manager-staging` until 2026-11-25, and renamed by hand in Hetzner's console, which keeps its token. Its API token, which needs read and write permission, is `HETZNER_TOKEN` in the `.env` of the maintainer's main checkout, and nowhere else |
| Region | `fsn1` (Falkenstein, Germany), a location: Hetzner's API no longer names a datacenter |
| Staging server | `staging`, plan CX23, system image `ubuntu-26.04`, labelled `environment=staging`, with the address and the volume of staging. Created by `task staging:on` and destroyed by `task staging:off`, from the choices in `deploy/server.staging.yaml`, and prepared by `deploy/cloud-init.yaml`. It exists only while staging is on |
| Address of staging | `staging`, the primary IPv4 address `167.233.239.56` in `fsn1`, labelled `environment=staging`, with auto delete off, so that deleting a server keeps it. Created on 2026-11-25 through Hetzner's API, as `task staging:on` creates it when it finds none. Every staging server gets it, and no IPv6. The A records of the host names of staging point to it (see Domain and DNS). Only Hetzner's console deletes it; a turn-on that finds none creates another, with a new address, and says so |
| Volume of staging | `staging`, 10 GB in `fsn1`, formatted by Hetzner with ext4, labelled `environment=staging`. Created on 2026-11-25 through Hetzner's API, as `task staging:on` creates it when it finds none. Attached to every staging server, which mounts it at `/srv/caddy-data` (`deploy/cloud-init.yaml`), where Caddy keeps its certificates. Hetzner detaches it when the server is deleted, and keeps it. Only Hetzner's console deletes it; a turn-on that finds none creates another, which costs one new certificate for each host name |
| SSH key of the project | `staging`, the deploy public key of the server, labelled `environment=staging`, registered with the server so that Hetzner e-mails no root password, and removed with it. A new one each turn-on |
| Firewall | None at Hetzner: the firewall is inside the server, ufw, which lets in 22, 80 and 443 (`deploy/cloud-init.yaml`) |
| SSH | Key login as the deploy user; no password, no root. The host key is the one the tool gave the server |
| Metadata service | `169.254.169.254`, which gives anything on the server its user data, the rendered cloud-init file with the private host key. The server reads it; no container reaches it, by a rule of `deploy/cloud-init.yaml` (Servers in [`CODING_STANDARDS.md`](../CODING_STANDARDS.md)) |
| Production server | Plan CX23; project, name and IP address _not created_ |
| Restore-test server | _Decided with production_: a task of the maintainer's machine (ADR-0019) |

## Backblaze B2

**Why.** The cheapest S3-compatible storage with object versioning: 10 GB free for good, the lowest price for each gigabyte, and free egress up to three times what is stored. It is a different company from the one that hosts the servers, so losing that account does not lose the backups (ADR-0020).

**Cost.** US$ 6.95 for each terabyte a month, with the first 10 GB free. About US$ 0.63 a month for 100 GB stored, 20 GB uploaded and 50 GB downloaded. Requests are free.

**Not proven yet.** The AWS SDK for Go v2 with its default checksum, `CopyObject`, which confirming an upload uses (ADR-0017), an upload in parts with signed URLs, and pgBackRest against it. The work that builds the deploy tries all four, before production uses it. If one fails, the storage moves to Scaleway or OVHcloud.

| Item | Value |
| --- | --- |
| Account region | _not chosen_; fixed for the account when it is created |
| S3 endpoint | _not created_ |
| Bucket of production files | _not created_; private, versioning on, old versions removed after 30 days, cross-origin requests from the app's origin only |
| Bucket of database backups | _not created_; private, written by pgBackRest, 30 days kept |
| Application keys | One for each bucket, limited to it (the values are secrets) |

## Domain and DNS

**Why.** The system needs addresses of its own: production needs one, and so does each application that staging and production register with a login provider. The host names of staging are under the domain, and their records point to the address Hetzner keeps for staging, so they are set once, by hand, and stay while staging is off (ADR-0019).

The DNS is at Cloudflare, which has an API for its records and gives free certificates to the host names it proxies. Staging uses neither: its records are DNS only, set by hand or by an agent through the Cloudflare MCP (`docs/agents/dns-tools.md`), and Caddy gets its certificates from Let's Encrypt. Cloudflare Registrar sells `.dev` at cost, so the registration moves there too ([ticket 02](../.scratch/domain/issues/02-transfer-the-domain-to-cloudflare-registrar.md)).

**Cost.** R$ 96 for the first year, paid at registration. R$ 142 a year from the first renewal, GoDaddy's renewal price. After the transfer, Cloudflare's price, about US$ 12.20 a year. Cloudflare's DNS is free, on its Free plan.

| Item | Value |
| --- | --- |
| Registrar | GoDaddy, until the registration moves to Cloudflare Registrar, planned for 2027-01-23, the 60th day after registration, or the day after if it is refused as too early ([ticket 02](../.scratch/domain/issues/02-transfer-the-domain-to-cloudflare-registrar.md)). GoDaddy then leaves this list |
| DNS host | Cloudflare, on the Free plan, since 2026-11-25: name servers `henrik.ns.cloudflare.com` and `luciane.ns.cloudflare.com` |
| Records | The three of staging, below. The records of GoDaddy's parking page were deleted when the DNS moved |
| DNSSEC | Off. It can be turned on at Cloudflare once the registration is there |
| Domain | `forgespark.dev` |
| HTTPS only | `.dev` is on the HSTS preload list, so browsers open every host name under it over HTTPS only |

Records, each DNS only, with Cloudflare's automatic TTL:

| Host name | Type | Points to | Serves |
| --- | --- | --- | --- |
| `app.staging.forgespark.dev` | A | `167.233.239.56`, the address of staging | The app, the API and MCP of the pull request |
| `files.staging.forgespark.dev` | A | `167.233.239.56` | Garage, for uploads and downloads |
| `mail.staging.forgespark.dev` | A | `167.233.239.56` | Mailpit, behind a password |

Records to create:

| Host name | Type | Points to | Serves |
| --- | --- | --- | --- |
| _app, production_ | A | The production server | The app, the API and MCP |
| _the records Resend asks for_ | TXT, MX, CNAME | Resend | Proof that the domain may send e-mail |

In production the browser reaches the files at Backblaze's own host name, so there is no record for them.

## Let's Encrypt

**Why.** Caddy gets and renews the HTTPS certificates by itself, at no cost.

**Cost.** Free.

**Limit that matters.** No more than five certificates for the same host names in a week. Every staging server has the same host names, so Caddy keeps its certificates on the volume of staging, which outlives each server: a new server finds them and asks for none (ADR-0019). Only a new volume makes staging ask again, once for each host name.

Nothing is configured there. The certificates are for the host names listed in Domain and DNS.

**ZeroSSL.** Caddy asks ZeroSSL when Let's Encrypt refuses a certificate only when it has an address for the account, `ACME_EMAIL` (`deploy/Caddyfile.staging`). Staging gives none, so it asks Let's Encrypt alone; with its certificates kept on the volume, the limit above is not reached.

## Resend

**Why.** It sends the e-mail of production and reports a bounce to the API. Development, tests and staging use Mailpit and send nothing real.

**Cost.** _Not checked._

| Item | Value |
| --- | --- |
| Account | _not created_ |
| Sending domain | _not created_; verified by the DNS records above |
| Sender address | _not chosen_ |
| API key of production | One, which the worker sends with (`RESEND_API_KEY`; the value is a secret) |
| Webhook for bounces | `POST /api/resend/bounces` on the production address, for the event `email.bounced` only. Resend signs it with Svix, and the API checks the signature with the signing secret of the webhook (`RESEND_WEBHOOK_SECRET`; the value is a secret). _Not created_ |
| Test API key | A key of its own, apart from the one of production, which only the scheduled workflow `resend.yml` uses, from the repository secret `RESEND_TEST_API_KEY` (see GitHub); its tests send to Resend's test addresses (`delivered@resend.dev`, `bounced@resend.dev`), which deliver nothing. Which Resend account holds it is decided with the deploy. _Not created_ |
| Used by Grafana for alerts | Not tried (ADR-0012) |

## Login providers

**Why.** People log in with an account they already have, and the system stores no password (ADR-0011).

**Cost.** Free.

Each provider needs two registered applications, because staging and production have different addresses. Staging registers its own at its host names, which no longer change; until they are registered it logs in with a login code only (ADR-0019). Development and the pipeline use Dex and register nothing. The redirect address each application registers is the origin of the app followed by the callback of the provider, `<APP_ORIGIN>/api/session/login_providers/<key>/callback`, where the key names the provider in the configuration (such as `google`).

| Provider | Application for | Redirect address |
| --- | --- | --- |
| Google | staging | _not created_ |
| Google | production | _not created_ |
| Microsoft | staging | _not created_ |
| Microsoft | production | _not created_ |
| GitHub | staging | _not created_ |
| GitHub | production | _not created_ |

## LLM provider

**Why.** The backend calls models through one interface in the Chat Completions format, so the provider is configuration: a base address, a key and a model (ADR-0004). Only production calls a real provider. Development and the pipeline use the stand-in written in this repository, and staging does once a feature calls an LLM; until then it declares none.

**Cost.** By use, in production only. Nothing automated calls a real provider anywhere else (ADR-0019).

| Item | Value |
| --- | --- |
| Provider | _not chosen_ |
| Base address | _not chosen_ |
| Default model | _not chosen_ |
| Keys | One, for production (the value is a secret) |

## Hosted monitor

**Why.** Production runs no telemetry yet, so something outside the server has to warn when the system stops answering or a backup does not finish (ADR-0012).

**Cost.** A free plan is intended.

| Item | Value |
| --- | --- |
| Service | _not chosen_ |
| HTTP check | `/up` on the production address |
| Backup signal | Sent by each backup when it ends; a missing one warns |
| Warning | By e-mail, to the maintainer |
