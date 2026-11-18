---
status: accepted
---

# Frontend organized by domain area, tested through whole screens in a real browser

The backend is split by domain area (ADR-0007). The frontend follows the same split, so a term from the glossary leads to one folder on each side, and an agent working on an area finds everything about it in one place.

- **One folder per domain area** in `apps/web/src`, named as in the backend and the glossary. It holds the area's screens, components, data module and translation files, and an `index.ts` with what the rest of the app may import.
- **Route files are thin.** A file in `routes/` declares the path, the search parameters and the loader, and opens a screen from an area.
- **Only the data module calls the API.** It is the one place in an area that uses the generated client, and it exposes the area's reads and writes under glossary names. Screens and components never see the client or a URL.
- **Dependencies point one way:** `routes` → an area → `shared` → `ui`. An area uses another only through its `index.ts`. `shared` imports no area. `ui` is the design system, in `apps/web/src/ui` (ADR-0025).
- **Translation files live in their area**, one per language. This is how the "one file per language and screen area" of ADR-0005 is laid out.
- **The default test renders a whole screen.** It runs in Vitest's Browser Mode, in a real Chromium driven by Playwright, with the API simulated at the network layer by MSW, and it acts as a user does: clicks, types, reads. A few end-to-end tests in Playwright cover critical flows against the real backend.

## Considered Options

- **Everything beside its route file.** Easy to find, but code used by two routes has no home and the structure follows navigation instead of the domain.
- **Folders by technical type** (`components/`, `hooks/`, `stores/`, `api/`). The layout models reproduce most, with every area mixed and no boundary to check.
- **Screens calling the generated client.** Less code, but fetching, invalidation and optimistic edits spread over the screens and each session does them differently.
- **A central `locales/` folder.** The usual react-i18next layout, and a whole language is visible at once. Deleting or moving an area would leave its texts behind.
- **The generated client in a package of its own.** Worth it only when a second app consumes the API, and `apps/web` is the only app (ADR-0025).
- **jsdom with Testing Library.** Faster to start and better known to models, but with no real layout or focus, where the Base UI components behave differently.
- **End-to-end tests only.** The highest fidelity, slow and brittle, and every test needs the whole environment running.
- **Tests per isolated component.** Many small tests that break on every refactor and do not prove a screen works.

## Consequences

- A data module that only forwards a call is correct, for the same reason a pass-through use case is in the backend.
- Renaming a domain area changes the backend and the frontend together.
- The import rules are checked by tooling from the first frontend code: dependency-cruiser, chosen with the first screen (Enforcement in `CODING_STANDARDS.md`).
- Screen tests need a Chromium installed through Playwright and start slower than jsdom tests.
- The simulated API in tests comes from the responses orval generates, so it follows the contract.
- Screen tests in Vitest's Browser Mode with MSW serving the responses orval generates were tried with the setup screen (Vitest 5 inside Vite+, Playwright 1.65, MSW 3), and they hold. What they needed: MSW 3 serves its worker from its own Vite plugin, so no worker file is committed and the build emits none; `vitest-browser-react` is left out of the dependencies Vite bundles beforehand, or the bundle carries a second Vitest with no runner; every source file is scanned for dependencies at the start, or one found halfway reloads the page under the tests; and the app is rendered in StrictMode, as in the browser, or a fault that only the second mount of development shows passes the tests.
- The end-to-end tests run on the test runner of the same Playwright, through Caddy, with the API on a database of their own (Tests in the frontend rules of `CODING_STANDARDS.md`). The first is the setup, with the real API and storage.
