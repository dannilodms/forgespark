# Design system and clickable prototype

Status: ready-for-agent

## Goal

Before any product feature exists, the maintainer wants to see and click the interface, compare visual directions, and change them. This effort delivers the design system package, a clickable prototype app, and a gallery where the same components are shown under three style variations, one per visual reference. It ends when the maintainer has chosen the final style.

The design system is written in `DESIGN.md` at the repository root. The gallery renders that document and stays after this effort, as the reference for every screen.

## Decisions already made

Do not reopen these; they are recorded in the ADRs and in `DESIGN.md`.

- React single-page app, Vite+ (`vp` CLI) and TanStack Router: `docs/adr/0002-react-spa-without-a-full-stack-framework.md`.
- Three languages (`en-US` default, `pt-BR`, `es`) with react-i18next; English files are the source of truth and a check fails on missing keys: `docs/adr/0005-multilingual-from-the-start.md`.
- Design system in `packages/ui` on shadcn/ui and Tailwind CSS; prototype in `apps/prototype` with fake in-memory data and no backend; approved screens are rebuilt in `apps/web`, never moved; the gallery is permanent: `docs/adr/0006-design-system-proven-in-a-disposable-prototype.md`.
- `DESIGN.md` is the source of the rules, the style variables and their values, the components and the recipes. Read it before touching any style value. An issue that adds or settles a value records it there in the same change.
- Light and dark themes from the start. Desktop first, adapting down to phone widths.
- The gallery has one route per section, and a Compare route that shows the three style variations side by side until the final style is chosen.
- A style variation keeps the values of its reference. The gallery measures the contrast of every colour pair and marks the ones below 4.5:1; only the final style has to meet WCAG AA.
- A component enters `packages/ui` when a screen needs it. The first set is the list in `DESIGN.md`.
- The prototype has all three languages with a selector, in its own translation files. Navigation and sample content are translated; the gallery's explanatory text is in English.
- Reference analysis with measured values: `docs/design/references.md`. It is the evidence behind `DESIGN.md`.

## Out of scope

- `apps/web`, `apps/api`, Docker Compose, Caddy.
- Real product screens. The domain is not modelled yet; the sample work screen uses placeholder data and does not define entities or their names.
- The final style. It is chosen by the maintainer in issue 10.

## Style variations

Each style variation is a complete set of style variables in light and dark, applied to the same components. Switching style variation never changes component code.

The traits and the values of `zed`, `figcraft` and `herdr` are under "Style variations" in `DESIGN.md`.

## Technical notes

- Frontend tooling is Vite+; use `vp` for install, dev, build, test and checks. It is newer than plain Vite, so read its documentation at viteplus.dev instead of assuming Vite or Vitest commands.
- Follow shadcn/ui's documented monorepo layout: a `components.json` in each workspace, shared code imported from the UI package, and the `tailwind.config` field left empty for Tailwind CSS v4. Do not keep Turborepo if the shadcn CLI scaffolds it: this repository uses no monorepo orchestrator (ADR-0006).
- Components are generated on Base UI primitives, shadcn/ui's current default, not on Radix. Base UI is newer than what you remember: read the current shadcn/ui and Base UI documentation.
- The prototype uses React, TanStack Router, react-i18next and `packages/ui`, with fake data in plain modules. A Zustand store may hold the theme and the style variation. No TanStack DB, TanStack Query, TanStack Form, Zod, orval or screen tests: those belong to `apps/web` (ADR-0013, ADR-0014).
- React Compiler is enabled in the build, through the Babel preset that react.dev documents. Do not write `useMemo`, `useCallback` or `memo`.
- `vp check` type-checks with TypeScript 7. Its type-aware linter was reported not to support `baseUrl` in `tsconfig`; confirm in the Vite+ documentation before configuring path aliases.
- Every colour, radius, font, spacing step and shadow is a CSS custom property. Components never hardcode a value.
- Theme (light or dark) and style variation are two independent switches on the root element. Both persist across reloads. A style variation can also be set on any element, which the Compare route uses.
- The gallery follows `.claude/skills/shadcn/design-system-page.md` for how each foundation and component is shown. Where that file speaks of a single `DESIGN.md` style, this project has three style variations; where it lists every shadcn/ui component, this project shows the ones in `DESIGN.md`.
- The stylesheet that pins hover, focus and active for the gallery's state matrices belongs to `apps/prototype`. It must not reach `packages/ui`, which `apps/web` also uses.
- A coverage check, run with the other checks, fails when a component of `packages/ui` is shown nowhere in the gallery.
- Fonts are open-licensed and installed from npm packages, not loaded from a third-party CDN at runtime. Confirm each licence when adding it.
- Code identifiers are in English.

## Issues

1. `01-frontend-workspace-and-first-component.md`
2. `02-gallery-shell-and-foundations.md`
3. `03-gallery-components-and-rules.md`
4. `04-sample-work-screen-and-recipes.md`
5. `05-style-variation-switch-and-zed.md`
6. `06-figcraft-variation.md`
7. `07-herdr-variation.md`
8. `08-compare-route.md`
9. `09-three-languages-in-the-prototype.md`
10. `10-choose-the-final-style.md`
