---
status: accepted
---

# Go for the backend

The backend is written in Go. The code is written by AI agents and reviewed by a single maintainer, and shipping comes before learning a stack. Go is a small language that agents write reliably, it compiles in seconds so the agent's edit-build-test loop stays short, and it ships as one small binary in the single Docker Compose instance this system runs on.

## Considered Options

- **TypeScript.** The language models know best, the same one as the frontend (shared types with no code generation), readable by the maintainer, and the only one of the three with the Claude Agent SDK. Not chosen; picking Go means giving those up.
- **Rust.** Listed in the first README draft. Its compiler catches the most mistakes before the code runs, but builds are slow, agents stall more often on lifetimes and async, Anthropic publishes no official Rust SDK, and the extra performance does not show at the scale of one consultancy.

## Consequences

- Backend and frontend are different languages, so the frontend's API client is generated from an OpenAPI spec (ADR-0003).
- The Claude Agent SDK exists only for TypeScript and Python. Agents built on it run as separate services, not inside the backend (ADR-0004).
