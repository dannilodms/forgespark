---
status: accepted
---

# Use cases served by several inbound adapters, REST and MCP first

The backend has to be reachable in more than one way: the React app calls a REST API, and the maintainer's coding agents (Claude Code, Codex, OpenCode, pi) operate the system to update project status, look up clients and hours, and fetch credentials. More ways may follow. Each of them is an **inbound adapter**: code that authenticates the caller, calls one use case and maps the result. The use cases of ADR-0007 stay the only way in, so a new inbound adapter adds no business logic.

Two inbound adapters are built first:

- **REST**, complete. Every use case has a REST operation, because the React app needs all of them. Streaming to the browser (LLM answers, progress of long tasks) is Server-Sent Events inside this adapter, described in the same OpenAPI spec, and is added with the first use case that needs it.
- **MCP**, a chosen subset. It runs in the same Go binary, built with the official Go MCP SDK and served over HTTP beside the REST API. It exposes tools only. A tool is designed for what an agent is trying to do and calls exactly one use case; a tool that needs a composed result or a different shape gets a new use case. A use case becomes a tool only when an agent has a real use for it.

Callers that are not a person at a screen authenticate with an **access token**: created by a user in the app, named, revocable, and carrying scopes chosen at creation. A scope is read or write in one domain area. The same access tokens serve the agents of ADR-0004 on the REST API. An inbound adapter only authenticates and tells the use case who is acting and with which scopes; the use case decides whether the action is allowed. An action taken with an access token is recorded as its user's, together with the token used.

Reading the value of a credential through MCP sends that value to the model provider and into session transcripts. It therefore needs a scope of its own that no token has by default, returns one credential per call, and records every read. Listings never include the value.

MCP has no frontend to translate error codes. It returns the stable code and the English message of ADR-0005, and tool descriptions are in English. Configuration content comes in the preferred language of the token's user, so an agent names a status the way the screen does.

Keeping the door open costs rules, not code. Nothing is written for an inbound adapter that does not exist yet.

## Considered Options

- **MCP tools generated from the OpenAPI spec.** No code per tool, but the tools mirror the endpoints one to one. Anthropic's guidance on writing tools for agents names this as a common error: many granular tools with verbose results, where a few shaped around the agent's task work better.
- **MCP as a separate service calling the REST API.** Isolates MCP, but adds a container to operate, a network hop and a second authentication layer, for a server with one user.
- **One registry of operations from which REST and MCP are generated.** Less repetition, but it forces every inbound adapter into the same shape and is an abstraction of our own to maintain.
- **Every use case exposed on every inbound adapter.** Predictable, but it fills the agent's context with tools and exposes through MCP operations that have no agent use.
- **OAuth for MCP from the start.** Needed for claude.ai and Claude Desktop connectors, which are not among the consumers. The Go MCP SDK is a resource server only, so it would also mean building or operating an authorization server.
- **MCP that never returns a credential's value.** Safest, but an agent could not use a credential by itself, which is one of the reasons for MCP here.
- **Authorization in each inbound adapter.** Simpler use cases, but each new adapter repeats the rules and can forget one.
- **A CLI.** Dropped: the agents that would use it all speak MCP.
- **WebSocket for all real-time traffic.** One mechanism for everything, but outside the OpenAPI spec, with no generated client and no support in Huma.

## Consequences

- GraphQL and gRPC stay rejected for the reasons in ADR-0003. Each one, like WebSocket, starts with a new ADR when a concrete case appears. The first expected case for WebSocket is live collaboration between people.
- Screens that update by themselves and live notifications need the backend to announce what changed, which reopens the decision on domain events deferred in ADR-0007.
- OAuth, MCP resources and MCP prompts are added when a consumer needs them.
- The list of MCP tools is not decided: the domain areas for clients, projects and hours are not modelled yet.
- How a person logs in to the React app is decided in ADR-0011. It produces the same "who is acting" that an access token does.
- Mounting the MCP handler on the same router as Huma has not been tried yet; it rests on the SDK exposing a standard `http.Handler`.
