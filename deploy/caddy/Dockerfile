# The Caddy image of every environment: Caddy with a rate-limit plugin, which
# Caddy does not have. The plugin and the versions are pinned here, and
# `task versions:check` reads both. Rules: Rate limit in Delivery and Versions
# in CODING_STANDARDS.md.

FROM caddy:2.11.9-builder-alpine AS builder

# The builder image names its own Caddy version, so xcaddy builds 2.11.9.
RUN xcaddy build --with github.com/mholt/caddy-ratelimit@v0.1.0

FROM caddy:2.11.9-alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
