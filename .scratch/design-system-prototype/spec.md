# Design system and clickable prototype

Status: ready-for-agent

## Goal

Before any product feature exists, the maintainer wants to see and click the interface, compare visual directions, and change them. This effort delivers the design system package, a clickable prototype app, and a gallery where the same components are shown under three style variations, one per visual reference. It ends when the maintainer has chosen the final style.

## Decisions already made

Do not reopen these; they are recorded in the ADRs.

- React single-page app, Vite+ (`vp` CLI) and TanStack Router: `docs/adr/0002-react-spa-without-a-full-stack-framework.md`.
- Three languages (`en-US` default, `pt-BR`, `es`) with react-i18next; English files are the source of truth and a check fails on missing keys: `docs/adr/0005-multilingual-from-the-start.md`.
- Design system in `packages/ui` on shadcn/ui and Tailwind CSS; prototype in `apps/prototype` with fake in-memory data and no backend; approved screens are rebuilt in `apps/web`, never moved: `docs/adr/0006-design-system-proven-in-a-disposable-prototype.md`.
- Light and dark themes from the start. Desktop first, adapting down to phone widths.
- The prototype has all three languages with a selector, in its own translation files.
- Reference analysis with measured values: `docs/design/references.md`. Read it before touching any style value.

## Out of scope

- `apps/web`, `apps/api`, Docker Compose, Caddy.
- Real product screens. The domain is not modelled yet; the sample work screen uses placeholder data and does not define entities or their names.
- The final style. It is chosen by the maintainer in issue 08.

## Style variations

Each variation is a complete set of style variables in light and dark, applied to the same components. Switching variation never changes component code.

| Trait | `zed` (site and app) | `figcraft` (site) | `herdr` (app) |
| --- | --- | --- | --- |
| Light surfaces | Cream paper page `#f5f4f3`; app steps `#fafafa`, `#ebebec`, `#dcdcdd` | `#f2f1ee`, raised `#faf9f7`, sunken `#e9e8e4` | Derived, no official light theme: `#f6f7f9`, surface `#ffffff`, border `#e5e7ec` |
| Dark surfaces | `#282c33`, `#2f343e`, `#3b414d` | `#121316`, raised `#101014`, sunken `#060607` | Canvas `#131518`, chrome `#101114`, pane `#17191d`, raised `#2c2e31` |
| Neutrals | Cool gray ramp; border `#c9c9ca` / `#464b57` | One ink (`#101014` / `#f2f1ee`) at fixed alphas: 62%, 42%, 16%, 40% | White alpha: hover 5.5%, selected 9%, border 9.4% |
| Accent | Blue `#1348dc`; on dark `#74ade8` | Green `#02da96`, used sparingly | Pale blue `#96b2d7`, for focus only |
| Primary button | Blue fill, inset darker bottom edge, shortcut chip inside | Inverted ink block, uppercase mono label | Inverted neutral (text colour fill) |
| Corner radius | 4px; shortcut chips 2px | 0 everywhere | Controls 8px, panes 10px, tabs 7px, badges 5px |
| Display type | IBM Plex Serif, light italic | Archivo, weight 620 at 118% width, tight tracking | Geist, weight 560, tight tracking |
| Interface type | IBM Plex Sans, 12 to 13px | Inter | Geist, 13px |
| Mono | Lilex | Geist Mono | JetBrains Mono |
| Small labels | Sentence case | Mono, 11px, uppercase, +0.15em tracking | Sentence case |
| Panes | Flush, split by 1px lines | Flush, in a hairline grid | Separate cards with 8px gaps on a darker canvas; focused pane gets the accent border |
| Depth | No shadows inside; inset "keycap" edges on buttons | None | 1px light ring plus a large soft shadow, floating surfaces only |
| Decoration | 10px grid texture and diamond nodes, page framing only | Crosshair corner ticks and numbered section strips, page framing only | None |
| Density | 27 to 28px bars, 32px controls | 36px controls, 11px labels | 30px rows, 26px icon buttons |

Shared by all three: structure from 1px borders, mono for metadata and shortcuts, Lucide icons at 12 to 16px, 100 to 150ms colour transitions, and one status palette (blue working, amber needs attention, green done, red error).

The `herdr` light theme is not evidenced by the reference. Mark it as derived in the gallery.

## Technical notes

- Frontend tooling is Vite+; use `vp` for install, dev, build, test and checks. It is newer than plain Vite, so read its documentation at viteplus.dev instead of assuming Vite or Vitest commands.
- Follow shadcn/ui's documented monorepo layout: a `components.json` in each workspace, shared code imported from the UI package, and the `tailwind.config` field left empty for Tailwind CSS v4. Do not keep Turborepo if the shadcn CLI scaffolds it: this repository uses no monorepo orchestrator (ADR-0006).
- Components are generated on Base UI primitives, shadcn/ui's current default, not on Radix. Base UI is newer than what you remember: read the current shadcn/ui and Base UI documentation.
- The prototype uses React, TanStack Router, react-i18next and `packages/ui`, with fake data in plain modules. A Zustand store may hold the theme and the style variation. No TanStack DB, TanStack Query, TanStack Form, Zod, orval or screen tests: those belong to `apps/web` (ADR-0013, ADR-0014).
- React Compiler is enabled in the build, through the Babel preset that react.dev documents. Do not write `useMemo`, `useCallback` or `memo`.
- `vp check` type-checks with TypeScript 7. Its type-aware linter was reported not to support `baseUrl` in `tsconfig`; confirm in the Vite+ documentation before configuring path aliases.
- Every colour, radius, font, spacing step and shadow is a CSS custom property. Components never hardcode a value.
- Theme (light or dark) and style variation are two independent switches on the root element. Both persist across reloads.
- Fonts are open-licensed and installed from npm packages, not loaded from a third-party CDN at runtime. Confirm each licence when adding it.
- Code identifiers are in English.

## Issues

1. `01-frontend-workspace-and-first-component.md`
2. `02-component-gallery.md`
3. `03-sample-work-screen.md`
4. `04-style-variation-switch-and-zed.md`
5. `05-figcraft-variation.md`
6. `06-herdr-variation.md`
7. `07-three-languages-in-the-prototype.md`
8. `08-choose-the-final-style.md`
