# 01: Register the domain and move staging off sslip.io

**What to build:** Staging's host names are under `sslip.io` until the system has a domain of its own. That is provisional. `sslip.io` is not on the Public Suffix List, so its certificates share one Let's Encrypt limit with every user of it, and that limit has run out in some weeks. Until the domain, `ACME_EMAIL` stays empty, as the maintainer decided on 2026-11-20: Caddy 2.11 tries ZeroSSL only when it has an address for the account, so a week in which that limit runs out leaves staging without a certificate, and the publish's wait for `/up` reports it. The risk ends with the domain, whose limit at Let's Encrypt is shared with no one. The real login providers, Google, Microsoft and GitHub, also need a fixed address to register. The maintainer chooses a registrar and a DNS host and registers the domain. An agent then moves staging, and later production, to host names under it. See ADR-0019 and `docs/providers.md`, Domain and DNS.

**Blocked by:** None (can start immediately). It is in a folder of its own so that it outlives the effort of staging, whose folder is removed when that effort ends.

**Status:** resolved

- [x] The maintainer has chosen a registrar and a DNS host and registered the domain. `docs/providers.md` records them, with the yearly cost.
- [x] Decided with the maintainer: how staging's host names follow a server whose address changes each time it is turned on. One option is a reserved address with records set once by hand. Another is a DNS host with an API, updated by the tool on the machine, never by the pipeline.
- [x] Decided with the maintainer: how Caddy's certificate data outlives a staging server once the host names are fixed. Let's Encrypt refuses more than five certificates for the same set of names in a week, and every new server would ask for the same ones. One option is a private bucket, saved before the server is destroyed and restored on the next one, as the first version of ADR-0019 planned. Another is a provider volume moved from one server to the next. Either is done by the tool on the machine, never by the pipeline.
- [x] Turning staging off and on again more than five times in a week keeps HTTPS working, without a new certificate for the same names.
- [x] Staging's host names are under the domain, and nothing refers to `sslip.io` any more.
- [x] ADR-0019 no longer calls staging's address provisional.

## Comments

### The domain and its DNS (2026-11-25)

The maintainer registered `forgespark.dev` at GoDaddy on 2026-11-24 and moved its DNS to Cloudflare, on the Free plan, on 2026-11-25. The registration moves to Cloudflare Registrar on 2027-01-23 ([ticket 02](02-transfer-the-domain-to-cloudflare-registrar.md)). `docs/providers.md` records both, with their costs.

Cloudflare has an API for DNS records, and the tool on the machine could set staging's records with it on each turn-on, without a reserved address. It also gives free certificates to the host names it proxies: a Cloudflare Origin CA certificate on the server, valid for up to 15 years, would serve every staging server, with no new Let's Encrypt certificate each time. That needs host names one level under the domain, such as `app-staging.forgespark.dev`. These are options for the two decisions above, not decided yet.

### Decided while building (2026-11-25)

The maintainer chose, through the questions of the session:

- **A reserved address, with records set by hand.** Hetzner keeps one primary IPv4 address for staging, named `staging`, which every staging server gets. The tool creates it on the first turn-on that finds none, and turning off leaves it. The A records point to it and are set once, by hand or by an agent through the Cloudflare MCP; the tool and the pipeline set none. It costs US$ 0.60 a month, on or off.
- **The host names** `app.staging.forgespark.dev`, `files.staging.forgespark.dev` and `mail.staging.forgespark.dev`, DNS only. Two levels under the domain, so Cloudflare's free proxy certificate would not cover them; the proxy is not used.
- **Caddy's certificates on a Hetzner volume**, named `staging`, 10 GB in `fsn1`, kept like the address and mounted by the cloud-init file at `/srv/caddy-data`, which Caddy's `/data` is. Chosen over a copy on the maintainer's machine (free, but a step on each turn-on and turn-off, and lost when turning off cannot reach the server), over ZeroSSL as a fallback, and over Cloudflare's proxy with an Origin CA certificate. It costs US$ 0.77 a month, on or off.

ADR-0019, Servers in Environments of `CODING_STANDARDS.md` and `docs/providers.md` keep these decisions.

The address, `167.233.239.56`, and the volume were created on 2026-11-25 through Hetzner's API, with what the turn-on gives them, because staging was on at the time with a server of the tool before this change, publishing pull request #17. Left for the real run: turning staging off and on again, so that a server gets them; the three A records at Cloudflare; a publish over HTTPS at the host names under the domain; and a second turn-on finding the certificates of the first.

### The records of staging (2026-11-25)

The three A records were created at Cloudflare through the Cloudflare MCP, at the maintainer's request: `app.staging.forgespark.dev`, `files.staging.forgespark.dev` and `mail.staging.forgespark.dev`, each pointing to `167.233.239.56`, DNS only, with Cloudflare's automatic TTL. They resolve to that address at `cloudflare-dns.com` and at `dns.google`. They are the only records of the zone, as Domain and DNS in `docs/providers.md` records them.

Outside this folder, `sslip.io` is named only by ADR-0019, among the options it considered, as the second version of the decision.

### The real run, first server (2026-11-26)

`task staging:on` created a server that got the address `167.233.239.56` and the volume, mounted at `/srv/caddy-data` and empty but for `lost+found`. Pull request #21 was published to it. Caddy obtained three certificates from Let's Encrypt (issuer `YE2`), one for each host name, and kept them on the volume, under `/data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/`. Each host name answered over HTTPS: the app with 200 at `/up`, Garage with 403 at `/`, and Mailpit with 401 without its password. The certificates it served:

| Host name | Serial number | Valid until |
| --- | --- | --- |
| `app.staging.forgespark.dev` | `06541E455558020C838ECCABE757FA2A4F22` | 2027-02-24 |
| `files.staging.forgespark.dev` | `0622A37070A004B82010253AD9E7A7CEE87D` | 2027-02-24 |
| `mail.staging.forgespark.dev` | `06DAE4B0C4F118E9A07E669D93E56A423417` | 2027-02-24 |

`task staging:off` then destroyed the server and kept the address and the volume, and `task staging:on` created a second server with both, booted at 02:10:59 UTC. Pull request #21 was merged before its second publish, so the second server is published by the next pull request, which records what it served.

### What was built

Staging's host names are under the domain, `app.`, `files.` and `mail.staging.forgespark.dev`, pointing to the address Hetzner keeps for staging, and Caddy keeps its certificates on the volume Hetzner keeps for it (#18). The three DNS records were set at Cloudflare through its MCP. Production's record waits for its server, in Records to create of `docs/providers.md`.

### What was checked

- The second server, published by pull request #22, served the same three certificates as the first, with the same serial numbers and SHA-256 fingerprints, over HTTPS at the three host names: the app with 200 at `/up`, Garage with 403 at `/`, and Mailpit with 401 without its password. Its Caddy logged no `obtaining certificate`, no `certificate obtained successfully` and no new ACME account: it loaded the three certificates from the volume and asked Let's Encrypt for none.
- The criterion of more than five turn-ons in a week was checked with one turn-off and turn-on, as the maintainer chose: a server that finds the certificates on the volume asks for none, so the sixth turn-on of a week asks for none, as the second did. Let's Encrypt has been asked for one certificate for each host name, on 2026-11-26, valid until 2027-02-24.
- `task staging:off` after the second server left Hetzner with the address and the volume only, and `task staging:status` reports staging off.

ADR-0019, Servers in Environments of `CODING_STANDARDS.md` and `docs/providers.md` keep the decisions; the joint is marked as proven in the list of joints of `CODING_STANDARDS.md`.

### What was not checked

- The renewal of the certificates, which Caddy starts when a third of their 90 days is left, around 2027-01-25. The first server on after that date renews them onto the volume: one new certificate for each host name, well within the limit.
