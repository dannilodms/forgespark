---
status: accepted
---

# Design system in a shared package, proven in a disposable clickable prototype

How the product looks matters a great deal to the maintainer, who wants to click through screens and change them before any feature is built. The design system lives in `packages/ui` (shadcn/ui components on Tailwind CSS, light and dark themes, designed for desktop first and adapting down to phones) and is used by two apps: `apps/web`, the real product, and `apps/prototype`, a clickable app with fake in-memory data and no backend.

A screen approved in the prototype is rebuilt in `apps/web` against the real API. It is not moved there. Its texts in the three supported languages are copied across at that moment. This keeps fake data and screens still under discussion out of production code, and the shared package is what guarantees the product looks like what was approved.

The look comes from three references: Zed, FigCraft and Herdr, analysed in `docs/design/references.md`. A gallery page in the prototype shows the same components under one style variation per reference. The maintainer compares them and picks or mixes traits into the final style.

The design system is written in `DESIGN.md` at the repository root: its rules, the style variables and their values, the components and the recipes. The gallery renders that document and stays after the final style is chosen, as the reference for every screen. What is disposable in the prototype is its screens, not the gallery.

## Considered Options

- **Prototyping inside `apps/web` with a mocked API.** Nothing is rebuilt, but unapproved screens and mocking code live next to production code.
- **Moving approved prototype screens into `apps/web`.** Less rework, but code written quickly to simulate behaviour becomes production code.
- **A pre-styled component library (Mantine, MUI).** Faster to start, but the look is the library's, and bending it towards three references means overriding much of it.
- **Unstyled primitives with hand-written CSS.** Full freedom, with far more work before a first set of components exists.
- **Storybook for the gallery.** A dedicated tool that is one more app to configure and update; a gallery page in the prototype is enough.
- **One translation package shared by prototype and product.** Texts from screens under discussion or abandoned would mix into the production catalog.

## Consequences

- The frontend becomes a pnpm workspace with three packages (`apps/web`, `apps/prototype`, `packages/ui`), run through Vite+. There is still no separate monorepo orchestrator.
- Rebuilding approved screens is deliberate duplicated effort. Do not "fix" it by importing prototype screens into the product.
- The gallery has to stay complete to be a reference. A check fails when a component of `packages/ui` is missing from it.
