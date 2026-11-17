# 07 - Development behind Caddy, with the limit on the setup code

Status: ready-for-agent
Blocked by: 05

Read `../spec.md` and, in `CODING_STANDARDS.md`, Delivery, Health and Compose files. The reasons are in ADR-0009, ADR-0011 and ADR-0024.

## What to build

In development the browser reaches the system through Caddy, at one origin, and the operations that take a setup code are limited for each IP address.

- A Caddy image built in this repository with a rate-limit plugin, both at exact versions. Choosing the plugin is part of this ticket.
- The `Caddyfile` of development in `deploy/`, and Caddy in the development Compose file. One site on `localhost` forwards `/api` to the API on the machine and the rest to the Vite server on the machine.
- A limit, for each IP address, on the `POST` operations of the setup, starting at ten requests a minute. Over it, Caddy answers 429 in the error format of the API, with the code `too_many_requests`.
- `Cache-Control: no-store` on every response under `/api`.
- `/up`, answered by Caddy with the status code of readiness and no body. Liveness and readiness stay unreachable from outside.
- The Task target that starts the app in development starts the API and Vite together, and the address to open is the one of Caddy.
- `CODING_STANDARDS.md` records the plugin and the number of the limit, and the two places that said they were still to be chosen are updated. `README.md` and `README.pt-BR.md` give the address to open.

Compression, the security headers, the Brotli copies and the cache headers of the built app wait for the deploy.

## Acceptance criteria

- [ ] With the services and the app started, the address of Caddy shows the app, and a request under `/api` reaches the API.
- [ ] Requests past the limit to an operation of the setup answer 429 as `application/problem+json` with the code `too_many_requests`, and requests within it are untouched.
- [ ] `GET /api/setup` and every path outside the setup have no limit.
- [ ] A response under `/api` carries `Cache-Control: no-store`.
- [ ] `/up` answers 200 with no body when the API is ready and a failure status when it is not.
- [ ] Liveness and readiness cannot be reached through Caddy.
- [ ] Caddy reaches the API and Vite on the machine both under Docker Desktop and on the runner of the pipeline.
- [ ] The Caddy image and its plugin have exact versions, and the check of ticket 02 covers its Dockerfile.
- [ ] `CODING_STANDARDS.md` names the plugin and the limit.
