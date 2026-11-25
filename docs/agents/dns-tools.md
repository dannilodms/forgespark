# DNS tools

## Cloudflare API MCP

The [Cloudflare API MCP server](https://developers.cloudflare.com/agents/model-context-protocol/cloudflare/servers-for-cloudflare/) lets an agent read and change the DNS records of `forgespark.dev`, whose DNS is at Cloudflare (see Domain and DNS in `docs/providers.md`). It has three tools: `search()` finds an endpoint of the Cloudflare API, `execute()` calls it, and `docs()` searches Cloudflare's documentation.

### Configuration

Claude Code reads `.mcp.json` at the repository root. Other agents do not read that file: register the same address in the agent's own MCP configuration, as an HTTP server.

```
https://mcp.cloudflare.com/mcp
```

Cloudflare hosts the server, so there is no version to pin.

### Login

The server needs a login once on each machine. In Claude Code, run `/mcp`, pick `cloudflare` and authenticate. Cloudflare opens in the browser:

1. Log in with the maintainer's Cloudflare account.
2. On the consent page, select **Edit Permissions**. Every optional permission comes selected: leave only DNS Read, DNS Write and the read of zones.
3. Authorize.

Claude Code keeps the login. No token goes in the repository or in `.env`. A tool call that needs a permission left out fails; authenticate again and grant it, if the maintainer agrees.

### Rules

- Reading records needs no one's approval.
- An agent creates, changes or deletes a record only when the maintainer asks for that record. The same change writes it in `docs/providers.md`, as the records of the domain say there.
- An agent changes nothing else at Cloudflare, such as the zone's settings, the SSL/TLS mode or DNSSEC, unless the maintainer asks.
- The pipeline never uses this server. A tool of the repository that sets records calls the API with a token of its own, kept as Secrets in `CODING_STANDARDS.md` says.
