---
status: accepted
---

# Backend organized by domain area, with dependencies pointing inward

AI agents write the backend and a single maintainer reviews it. Every session has to produce the same structure, business rules have to be testable without a database, HTTP or an LLM, and infrastructure has to stay replaceable. The backend therefore combines the strategic part of Domain-Driven Design, a selective tactical part, the dependency rule and use cases from Clean Architecture, and three rule groups from object calisthenics. The day-to-day rules live in `CODING_STANDARDS.md`; this record holds the shape and the reasons.

- **One package tree per domain area**, with `domain`, `application` and `adapters` inside it. The domain area is the deep module: its public surface is its use cases, and everything else is internal. One area reaches another only through that area's use cases and holds its aggregates by ID.
- **Dependencies point inward.** `domain` imports no HTTP, database, storage or LLM code. `application` declares the interfaces it needs, and `adapters` implement them.
- **Every operation goes through a use case**, including one that only passes a call through. The use case is also the default place where tests enter, with in-memory adapters.
- **API types are separate from domain types.** The Huma request and response types live in the HTTP adapter and remain the source of the OpenAPI spec, so ADR-0003 stands.
- **Writes and reads take different paths.** A write loads and saves a whole aggregate through its repository, inside one transaction per use case. A read goes through a query use case that returns data shaped for the caller without building aggregates.
- **Selective tactical DDD.** Value objects, aggregates and one repository per aggregate. A primitive becomes a value object only when it carries a rule or is an identifier.
- **One context.** The whole backend shares one `GLOSSARY.md`. A second context is created when the same term comes to mean two things.
- **Precedence when rules collide:** the dependency rule, then a small public surface for the domain area, then the modelling and control-flow rules.

## Considered Options

- **Packages by layer, with areas inside.** The classic Clean Architecture layout. One change to an area spreads over three folder trees, and the area stops being a module with a small surface.
- **One flat package per domain area.** The deepest module and the fewest folders, but the dependency rule can no longer be checked through imports, because everything sits in one package.
- **Use cases only where there is orchestration.** Less code, but every session would have to judge what counts as simple, and tests would enter at different places.
- **Exposing domain types directly in the API.** Less mapping code, but a domain change becomes a contract change and the domain depends on Huma.
- **Full tactical DDD from the start.** Domain events, CQRS, event sourcing and a strict one-aggregate-per-transaction rule. Deferred until a concrete case needs one of them.
- **All nine object calisthenics rules.** The numeric size limits (small types, two fields per type) push toward many shallow types, the opposite of a deep module. Go's conventional short names are kept for the same reason of fit.
- **Bounded contexts from day one.** The domain is barely modelled yet, so any split now would be a guess.

## Consequences

- Simple operations carry a pass-through use case and a mapping between API types and domain types. This repetition is accepted: it buys one place for logic and one place for tests.
- The LLM provider interface of ADR-0004 lives with the shared technical infrastructure, which `domain` never imports.
- The stable error codes of ADR-0005 are declared beside the rule that raises them, in `domain` or `application`. The HTTP adapter only translates them.
- An effect on another area or on an external service happens after the transaction commits. A case that needs both to succeed or fail together reopens the decision on domain events.
- Rules that tooling can check are automated from the first backend code. The tool is not chosen yet.
- The frontend takes the glossary's vocabulary and the names of the domain areas from this decision. Its own structure is decided in ADR-0014.
