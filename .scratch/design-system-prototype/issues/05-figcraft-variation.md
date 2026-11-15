# 05 - FigCraft variation

Status: ready-for-agent
Blocked by: 04

Read `../spec.md` and section 2A of `docs/design/references.md` first.

## What to build

The `figcraft` style variation, following the `figcraft` column of the table in the spec. The model is the ArtCraft website, not the FigCraft application.

- Zero corner radius everywhere.
- One ink colour at fixed alphas for text, muted text and lines, over warm paper in light and cool near-black in dark.
- Primary button as an inverted ink block; outline buttons invert on hover.
- Expanded heavy display type, Inter for body, and small labels and buttons in uppercase mono with wide tracking.
- Flush panes in a hairline grid, no shadows.
- Crosshair corner ticks and numbered section strips as page framing only, never inside components.
- The green accent used sparingly: one highlighted element at a time, not as the button colour.

The reference is a marketing site with large type and roomy cells. Keep its character but use the compact control scale from the spec, since this is a work tool.

## Acceptance criteria

- [ ] `figcraft` appears in the variation selector and renders the gallery and the sample work screen in light and dark.
- [ ] It matches the spec table: zero radius, alpha-ink neutrals, inverted primary button, the three type roles, uppercase mono labels, flush panes, no shadows.
- [ ] Uppercase labels stay legible and do not overflow with the long-label samples.
- [ ] No component or screen file mentions the variation by name.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] Text contrast meets WCAG AA in both themes, including muted text built from alpha.
