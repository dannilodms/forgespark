# 04: The Compose file and Caddy of staging

**What to build:** The description of what runs on the staging server. It is a Compose file that extends the base file and runs PostgreSQL, Garage with its cross-origin rule, Mailpit, Caddy, the migration and the seed as steps that run once, and the API and the worker. Caddy serves three host names over HTTPS: the app with the built web app, `/api` and `/up`; the files, forwarded to Garage; and Mailpit behind a password. It keeps the same rules of delivery as development. See the spec, section The staging Compose file.

**Blocked by:** 03.

**Status:** ready-for-agent

- [ ] The staging Compose file extends the base file and declares only what the spec lists: no Dex, no River panel, no telemetry, no LLM stand-in.
- [ ] The host names of the app, the files and Mailpit, and the origin of the app, come from configuration, so that the publish derives them from the server's address.
- [ ] Caddy gets its certificates by itself. Its data is in a volume that the publish keeps, while the volumes of PostgreSQL and Garage are the ones the publish recreates.
- [ ] Mailpit's host name asks the browser for a password, taken from configuration.
- [ ] The rules of delivery keep one home for development and staging: compression, cache headers, security headers, the rate limit on the setup code and the login code, and the masking of addresses in the logs.
- [ ] A test through HTTP, as `caddy_test.go` does, starts Caddy with the staging configuration and its own local certificates. It proves the forwarding of the three host names, the password of Mailpit, the rate limit and the headers.
- [ ] The checks validate the staging Compose file with `docker compose config`.
- [ ] `CODING_STANDARDS.md` (Compose files, Delivery) matches what was built.

## Comments

### From ticket 01

- Docker publishes the port of a container past ufw, the firewall that `deploy/cloud-init.yaml` sets inside the server: it writes rules of its own in iptables. The staging Compose file publishes Caddy's 80 and 443 and nothing else, or what it publishes is open to the internet whatever ufw says. The firewall lets in 443/udp too, for HTTP/3 if Caddy keeps it. The rule is in Ports, in Servers of `CODING_STANDARDS.md`.
- The host names are `app.`, `files.` and `mail.` before the address of the server written with dashes, then `.sslip.io` (`app.203-0-113-10.sslip.io`), as `hostNamesOf` in `apps/api/tools/environment/status.go` writes them and `task staging:status` prints them. The publish derives them the same way, from that function.
