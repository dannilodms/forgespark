---
status: accepted
---

# LLM providers behind one interface; agents as separate HTTP services

LLM features are central to the product, and the maintainer wants to switch freely between OpenRouter, OpenAI, Anthropic and others. The backend calls models through one internal Go interface whose first adapter speaks the OpenAI Chat Completions format, so changing provider is configuration: base URL, key and model. A provider-native adapter is added only when a feature exclusive to that provider is needed.

Custom agents, built later with the Anthropic or OpenAI SDKs in whatever language suits them, run as separate containers reached over HTTP. They act on the system only through the backend API, using a token with limited scope, and never touch the database.

The exact contract between backend and agents is deliberately left open until the first agent exists. A Chat Completions-shaped endpoint fits short conversational work; an agent that runs for minutes, reports progress and can be cancelled needs an asynchronous task contract (create, follow events, fetch the result).

## Considered Options

- **Native adapters for every provider from the start.** Full feature access, at the cost of writing and maintaining adapters before any is needed.
- **A LiteLLM gateway container.** Translates to each native API and centralises keys and cost, but adds a Python service to operate.
- **OpenRouter only, no interface.** Fastest to ship, but ties the system to one intermediary.
- **A TypeScript worker with the Claude Agent SDK from day one.** A ready agent harness, but a second backend language and a queue before there is a concrete agent to run.

## Consequences

Features that exist only in a provider's native API, such as fine control of prompt caching, are unavailable through the first adapter.
