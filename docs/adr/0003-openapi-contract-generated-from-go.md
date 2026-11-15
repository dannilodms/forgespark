---
status: accepted
---

# REST API with the OpenAPI contract generated from Go code

The API is a first-class contract: the React app consumes it, and so do AI agents operating the system. It is a REST API whose Go types are the source of truth. Huma, running on the standard library router, generates the OpenAPI 3.1 spec from those types and validates requests against them, and the frontend's TypeScript client is generated from that spec. With one source, the spec cannot drift from the implementation.

## Considered Options

- **OpenAPI written first.** An explicit contract the maintainer can review without reading Go, but every change touches the spec and then the code.
- **ConnectRPC with Protobuf.** Generated code on both sides and native streaming, but less familiar, harder to inspect with plain HTTP tools, and less direct for agents to consume.
- **GraphQL.** Resolver and caching complexity that a single frontend does not justify.
- **Gin with swaggo.** The Go framework models know best, but the spec comes from comment annotations, which fall out of sync with the code, and it emits the older OpenAPI 2.0.
