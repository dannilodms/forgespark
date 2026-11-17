---
status: accepted
---

# The design system lives in the product, and there is no prototype

ADR-0006 put the design system in `packages/ui`, shared by `apps/web` and by `apps/prototype`, a clickable app with fake data where every screen was designed and approved before `apps/web` rebuilt it. The design system, a gallery, a login screen and a sample work screen were made that way, while `apps/web` stayed empty. On 2026-11-16 the maintainer dropped the prototype: the step was taking a long time and brought no real gain.

- **No prototype.** `apps/prototype` is removed, with its screens, its fake data and the gallery. A screen is built in `apps/web`, against the real API, in the work that delivers its feature. The maintainer sees it in the running app, and what has to change is changed there.
- **The design system is part of `apps/web`.** With one app there is nothing to share, so `packages/ui` is removed. Its components and its stylesheet moved to `apps/web/src/ui`, with nothing changed but the paths of their imports.
- **One JavaScript package.** The repository is no longer a pnpm workspace. `apps/web` holds its own `package.json`, catalog, lockfile and Vite+ configuration, and the root has none.
- **`DESIGN.md` is the reference.** With no gallery, that document alone says how a screen looks. A component is seen on the screens that use it.
- **What stays from ADR-0006.** shadcn/ui components on Base UI primitives and Tailwind CSS, the Zed style, light and dark themes, and desktop first.

## Considered Options

- **Keeping the prototype first.** The earlier decision. A screen is seen and changed before any backend exists, but every screen is specified, built, approved by hand in each scenario, language, theme and width, and then built again.
- **Keeping the prototype for the gallery alone.** The reference stays, at the price of a second app, a shared package and a workspace kept for one set of pages.
- **A gallery inside `apps/web`.** The reference stays without a second app, but the gallery and the stylesheet that pins its states would sit in production code. It can come back as its own decision if `DESIGN.md` proves not to be enough.
- **Moving the screens of the prototype into `apps/web`.** ADR-0006 refused it, and the reason holds: code written to simulate behaviour would become production code.

## Consequences

- The coverage check of the gallery, the prototype bar with its scenarios and `task prototype` are gone. `task dev` starts the app.
- Nothing measures the contrast of the colour pairs any more. A value added to `DESIGN.md` is measured when it is added.
- The four recipes of `DESIGN.md` have no code. The first screen that needs one writes it in `apps/web`.
- The login screen and the sample work screen of the prototype remain only in the history of the repository.
- A mistake in the look or in the flow is found in the product. When fixing it changes the API, it costs more than it would have in the prototype.
- ADR-0014 was edited: it no longer describes the prototype, and its dependency rule ends in `ui` instead of `packages/ui`.
- A second app that needs the design system brings the shared package back, with a decision of its own.
