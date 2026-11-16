---
status: accepted
---

# Frontend organized by domain area, tested through whole screens in a real browser

The backend is split by domain area (ADR-0007). The frontend follows the same split, so a term from the glossary leads to one folder on each side, and an agent working on an area finds everything about it in one place.

- **One folder per domain area** in `apps/web/src`, named as in the backend and the glossary. It holds the area's screens, components, data module and translation files, and an `index.ts` with what the rest of the app may import.
- **Route files are thin.** A file in `routes/` declares the path, the search parameters and the loader, and opens a screen from an area.
- **Only the data module calls the API.** It is the one place in an area that uses the generated client, and it exposes the area's reads and writes under glossary names. Screens and components never see the client or a URL.
- **Dependencies point one way:** `routes` → an area → `shared` → `packages/ui`. An area uses another only through its `index.ts`. `shared` imports no area.
- **Translation files live in their area**, one per language. This is how the "one file per language and screen area" of ADR-0005 is laid out.
- **The default test renders a whole screen.** It runs in Vitest's Browser Mode, in a real Chromium driven by Playwright, with the API simulated at the network layer by MSW, and it acts as a user does: clicks, types, reads. A few end-to-end tests in Playwright cover critical flows against the real backend.
- **The prototype is outside all of this.** `apps/prototype` uses React, TanStack Router, react-i18next and `packages/ui`, with fake data in plain modules.

## Considered Options

- **Everything beside its route file.** Easy to find, but code used by two routes has no home and the structure follows navigation instead of the domain.
- **Folders by technical type** (`components/`, `hooks/`, `stores/`, `api/`). The layout models reproduce most, with every area mixed and no boundary to check.
- **Screens calling the generated client.** Less code, but fetching, invalidation and optimistic edits spread over the screens and each session does them differently.
- **A central `locales/` folder.** The usual react-i18next layout, and a whole language is visible at once. Deleting or moving an area would leave its texts behind.
- **The generated client in its own workspace package.** Worth it only when a second app consumes the API; the prototype has no backend.
- **jsdom with Testing Library.** Faster to start and better known to models, but with no real layout or focus, where the Base UI components behave differently.
- **End-to-end tests only.** The highest fidelity, slow and brittle, and every test needs the whole environment running.
- **Tests per isolated component.** Many small tests that break on every refactor and do not prove a screen works.
- **The product's stack in the prototype.** Approved screens would sit closer to final code, but ADR-0006 rebuilds them precisely to keep simulation code out of the product.

## Consequences

- A data module that only forwards a call is correct, for the same reason a pass-through use case is in the backend.
- Renaming a domain area changes the backend and the frontend together.
- The import rules are checked by tooling from the first frontend code. The tool is not chosen yet.
- Screen tests need a Chromium installed through Playwright and start slower than jsdom tests.
- The simulated API in tests comes from the responses orval generates, so it follows the contract.
