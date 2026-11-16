# Design system and clickable prototype

Status: ready-for-agent

## Goal

Before any product feature exists, the maintainer wants to see and click the interface, and change it. This effort delivers the design system package, a clickable prototype app, and a gallery that shows every style variable and component in the Zed style. It ends when the gallery and the sample work screen exist in that style, in the three languages.

The design system is written in `DESIGN.md` at the repository root. The gallery renders that document and stays after this effort, as the reference for every screen.

## Decisions already made

Do not reopen these; they are recorded in the ADRs and in `DESIGN.md`.

- React single-page app, Vite+ (`vp` CLI) and TanStack Router: `docs/adr/0002-react-spa-without-a-full-stack-framework.md`.
- Three languages (`en-US` default, `pt-BR`, `es`) with react-i18next; English files are the source of truth and a check fails on missing keys: `docs/adr/0005-multilingual-from-the-start.md`.
- Design system in `packages/ui` on shadcn/ui and Tailwind CSS; prototype in `apps/prototype` with fake in-memory data and no backend; approved screens are rebuilt in `apps/web`, never moved; the gallery is permanent: `docs/adr/0006-design-system-proven-in-a-disposable-prototype.md`.
- The style follows Zed. The plan to build one style variation per reference and compare them was dropped (ADR-0006). Do not build a second set of values or a switch between sets.
- `DESIGN.md` is the source of the rules, the style variables and their values, the components and the recipes. Read it before touching any style value. An issue that adds or settles a value records it there in the same change.
- Light and dark themes from the start. Desktop first, adapting down to phone widths.
- The gallery has one route per section.
- The style meets WCAG AA. The gallery measures the contrast of every colour pair, and a value of Zed that fails is adjusted by the smallest step that passes.
- A component enters `packages/ui` when a screen needs it. The first set is the list in `DESIGN.md`.
- The prototype has all three languages, in its own translation files. Navigation and sample content are translated; the gallery's explanatory text is in English.
- The theme, the language and the scenarios of a screen are chosen in the prototype bar, a floating bar outside every screen and every gallery route (see Prototype in `CODING_STANDARDS.md`). No selector lives in a screen's own header because of the prototype.
- Reference analysis with measured values: `docs/design/references.md`. Section 1 is the evidence behind `DESIGN.md`.

## Out of scope

- `apps/web`, `apps/api`, Docker Compose, Caddy.
- Real product screens. The sample work screen uses placeholder data and does not define entities or their names. The login screen is its own effort, in `.scratch/login-screen-prototype/`.

## Technical notes

- Frontend tooling is Vite+; use `vp` for install, dev, build, test and checks. It is newer than plain Vite, so read its documentation at viteplus.dev instead of assuming Vite or Vitest commands.
- Follow shadcn/ui's documented monorepo layout: a `components.json` in each workspace, shared code imported from the UI package, and the `tailwind.config` field left empty for Tailwind CSS v4. Do not keep Turborepo if the shadcn CLI scaffolds it: this repository uses no monorepo orchestrator (ADR-0006).
- Components are generated on Base UI primitives, shadcn/ui's current default, not on Radix. Base UI is newer than what you remember: read the current shadcn/ui and Base UI documentation.
- The prototype uses React, TanStack Router, react-i18next and `packages/ui`, with fake data in plain modules. A Zustand store may hold the theme and the scenario. No TanStack DB, TanStack Query, TanStack Form, Zod, orval or screen tests: those belong to `apps/web` (ADR-0013, ADR-0014).
- React Compiler is enabled in the build, through the Babel preset that react.dev documents. Do not write `useMemo`, `useCallback` or `memo`.
- `vp check` type-checks with TypeScript 7. Its type-aware linter was reported not to support `baseUrl` in `tsconfig`; confirm in the Vite+ documentation before configuring path aliases.
- Every colour, radius, font, spacing step and shadow is a CSS custom property. Components never hardcode a value.
- The theme (light or dark) is a switch on the root element. It persists across reloads.
- The gallery follows `.claude/skills/shadcn/design-system-page.md` for how each foundation and component is shown. Where that file lists every shadcn/ui component, this project shows the ones in `DESIGN.md`.
- The stylesheet that pins hover, focus and active for the gallery's state matrices belongs to `apps/prototype`. It must not reach `packages/ui`, which `apps/web` also uses.
- A coverage check, run with the other checks, fails when a component of `packages/ui` is shown nowhere in the gallery.
- Fonts are open-licensed and installed from npm packages, not loaded from a third-party CDN at runtime. Confirm each licence when adding it.
- Code identifiers are in English.

## Issues

1. `01-frontend-workspace-and-first-component.md`
2. `02-gallery-shell-and-foundations.md`
3. `03-zed-style.md`
4. `04-gallery-components-and-rules.md`
5. `05-sample-work-screen-and-recipes.md`
6. `06-three-languages-in-the-prototype.md`
