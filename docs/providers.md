# Providers

Every outside service this system depends on: why it was chosen, what it costs, and what is configured there. The decisions behind them are in ADRs [0019](adr/0019-three-environments-with-staging-on-a-server-created-on-demand.md), [0020](adr/0020-postgresql-backed-up-by-pgbackrest-and-production-files-in-backblaze-b2.md) and [0021](adr/0021-a-public-repository-under-agplv3.md).

**No secret is written here.** No key, token, password or passphrase, and no value that works as one. Addresses, host names, DNS records, regions, plan names and bucket names are fine. Where secrets live is in Secrets, in [`CODING_STANDARDS.md`](../CODING_STANDARDS.md).

**Keep it true.** A change made at a provider (a new record, a new bucket, a change of plan) is written here in the same pull request, or right after when the change is made by hand.

Prices were read on the providers' own pages on 2026-11-16, without tax. Nothing has been created yet: a value shown as _not created_ is filled in when the resource exists.

## Summary

| Provider | Used for | Environments | Approximate cost | State |
| --- | --- | --- | --- | --- |
| [GitHub](#github) | Repository, pipeline, container images, secrets | all | Free | Decided, not created |
| [Hetzner Cloud](#hetzner-cloud) | The servers | staging, production | € 5.99 a month, plus cents for staging | Decided, not created |
| [Backblaze B2](#backblaze-b2) | Production files, database backups | production, staging (certificates only) | Free under 10 GB, about US$ 0.63 a month at 100 GB | Decided, not created |
| [Domain and DNS](#domain-and-dns) | The addresses of the system | staging, production | _not known_ | Registrar not chosen |
| [Let's Encrypt](#lets-encrypt) | HTTPS certificates, through Caddy | staging, production | Free | Nothing to create |
| [Resend](#resend) | Sending e-mail | production | _not checked_ | Decided, not created |
| [Login providers](#login-providers) | Login through Google, Microsoft and GitHub | staging, production | Free | Decided, not created |
| [LLM provider](#llm-provider) | Calls to language models | production | By use | Not chosen |
| [Hosted monitor](#hosted-monitor) | Warning when production or a backup stops | production | Free plan intended | Not chosen |

Expected total while the system is small: about € 7 to € 9 a month, plus the domain.

## GitHub

**Why.** The repository is public, and a public repository on the free plan has what the pipeline needs: secrets for each environment, a protected `main` and unlimited minutes of GitHub Actions (ADR-0021).

**Cost.** Free.

| Item | Value |
| --- | --- |
| Account or organisation | _not created_ |
| Repository | _not created_, public |
| Licence | AGPLv3 |
| Branch protection on `main` | No direct push; merge only with the checks passing |
| Environments | `staging` and `production`; `production` restricted to `main` |
| Secret scanning with push protection | On |
| Workflows of outside contributors | Run only after approval |
| Container registry (GHCR) | Public images, tagged with the commit hash |
| Label that creates the staging server | `staging` |
| Commit address | The maintainer's `noreply` address |

## Hetzner Cloud

**Why.** The lowest price found for a server of this size that is billed by the hour, which is what lets the staging server exist only while it is used. One account and one command-line tool cover both servers (ADR-0019).

**Cost.** CX23 (2 vCPU, 4 GB, 40 GB): € 5.99 a month, or € 0.0096 an hour. Staging at about 40 hours a month is about € 0.40. The monthly restore test costs cents. The price of the reserved IP address and of the production IPv4 address was not confirmed.

**To check before ordering.** The plan showed limited stock, and a new account has its resources limited in its first month. If the CX23 cannot be ordered, the next plans are CAX11 (ARM, € 6.49) and CPX22 (€ 16.99).

| Item | Value |
| --- | --- |
| Project | _not created_ |
| Region | _not chosen_ (Germany or Finland) |
| Production server | Plan CX23; name and IP address _not created_ |
| Staging server | Plan CX23, created by the pipeline for a labelled pull request, destroyed within 24 hours |
| Restore-test server | Plan CX23, created and destroyed by the monthly workflow |
| Reserved IP address (staging) | _not created_ |
| Firewall | Ports 22, 80 and 443 |
| System image | _not chosen_; recorded in `deploy/cloud-init.yaml` |
| SSH | Key login as the deploy user; no password, no root |

## Backblaze B2

**Why.** The cheapest S3-compatible storage with object versioning: 10 GB free for good, the lowest price for each gigabyte, and free egress up to three times what is stored. It is a different company from the one that hosts the servers, so losing that account does not lose the backups (ADR-0020).

**Cost.** US$ 6.95 for each terabyte a month, with the first 10 GB free. About US$ 0.63 a month for 100 GB stored, 20 GB uploaded and 50 GB downloaded. Requests are free.

**Not proven yet.** The AWS SDK for Go v2 with its default checksum, an upload in parts with signed URLs, and pgBackRest against it. The walking skeleton tries all three. If one fails, the storage moves to Scaleway or OVHcloud.

| Item | Value |
| --- | --- |
| Account region | _not chosen_; fixed for the account when it is created |
| S3 endpoint | _not created_ |
| Bucket of production files | _not created_; private, versioning on, old versions removed after 30 days, cross-origin requests from the app's origin only |
| Bucket of database backups | _not created_; private, written by pgBackRest, 30 days kept |
| Bucket of staging certificates | _not created_; private, holds Caddy's certificate data between staging servers |
| Application keys | One for each bucket, limited to it (the values are secrets) |

## Domain and DNS

**Why.** The system needs addresses of its own. The staging server keeps a reserved IP address, so the records are set once, by hand, and the registrar needs no API (ADR-0019).

**Cost.** _Not known_: the yearly fee of the domain.

| Item | Value |
| --- | --- |
| Registrar | _not chosen_ |
| DNS host | _not chosen_ |
| Domain | _not registered_ |

Records to create:

| Host name | Type | Points to | Serves |
| --- | --- | --- | --- |
| _app, production_ | A | The production server | The app, the API and MCP |
| _app, staging_ | A | The reserved IP address | The app, the API and MCP of the branch |
| _files, staging_ | A | The reserved IP address | Garage, for uploads and downloads |
| _mail, staging_ | A | The reserved IP address | Mailpit, behind a password |
| _the records Resend asks for_ | TXT, MX, CNAME | Resend | Proof that the domain may send e-mail |

In production the browser reaches the files at Backblaze's own host name, so there is no record for them.

## Let's Encrypt

**Why.** Caddy gets and renews the HTTPS certificates by itself, at no cost.

**Cost.** Free.

**Limit that matters.** No more than five identical certificates in a week. Each staging server would ask for the same ones, so Caddy's certificate data is kept in a bucket between one staging server and the next (ADR-0019).

Nothing is configured there. The certificates are for the host names listed in Domain and DNS.

## Resend

**Why.** It sends the e-mail of production and reports a bounce to the API. Development, tests and staging use Mailpit and send nothing real.

**Cost.** _Not checked._

| Item | Value |
| --- | --- |
| Account | _not created_ |
| Sending domain | _not created_; verified by the DNS records above |
| Sender address | _not chosen_ |
| Webhook for bounces | An operation under `/api` on the production address |
| Used by Grafana for alerts | Not tried (ADR-0012) |

## Login providers

**Why.** People log in with an account they already have, and the system stores no password (ADR-0011).

**Cost.** Free.

Each provider needs two registered applications, because staging and production have different addresses. Development and the pipeline use Dex and register nothing.

| Provider | Application for | Redirect address |
| --- | --- | --- |
| Google | staging | _not created_ |
| Google | production | _not created_ |
| Microsoft | staging | _not created_ |
| Microsoft | production | _not created_ |
| GitHub | staging | _not created_ |
| GitHub | production | _not created_ |

## LLM provider

**Why.** The backend calls models through one interface in the Chat Completions format, so the provider is configuration: a base address, a key and a model (ADR-0004). Only production calls a real provider. Development, the pipeline and staging use the stand-in written in this repository.

**Cost.** By use. The scheduled tests of the LLM adapter spend a little every day, with a key of their own.

| Item | Value |
| --- | --- |
| Provider | _not chosen_ |
| Base address | _not chosen_ |
| Default model | _not chosen_ |
| Keys | One for production and one for the scheduled tests (the values are secrets) |

## Hosted monitor

**Why.** Production runs no telemetry yet, so something outside the server has to warn when the system stops answering or a backup does not finish (ADR-0012).

**Cost.** A free plan is intended.

| Item | Value |
| --- | --- |
| Service | _not chosen_ |
| HTTP check | `/up` on the production address |
| Backup signal | Sent by each backup when it ends; a missing one warns |
| Warning | By e-mail, to the maintainer |
