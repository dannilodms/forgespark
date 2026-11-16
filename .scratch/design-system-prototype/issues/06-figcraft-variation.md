# 06 - FigCraft variation

Status: ready-for-agent
Blocked by: 05

Read `../spec.md`, "Style variations" in `DESIGN.md` and section 2A of `docs/design/references.md` first.

## What to build

The `figcraft` style variation, following the `figcraft` column of Traits and "Colours of `figcraft`" in `DESIGN.md`. The model is the ArtCraft website, not the FigCraft application.

- Zero corner radius everywhere.
- One ink colour at fixed alphas for text, muted text and lines, over warm paper in light and cool near-black in dark.
- Primary button as an inverted ink block; outline buttons invert on hover.
- Expanded heavy display type, Inter for body, and small labels and buttons in uppercase mono with wide tracking.
- Flush panes in a hairline grid, no shadows.
- Crosshair corner ticks and numbered section strips as page framing only, never inside components.
- The green accent used sparingly: one highlighted element at a time, not as the button colour.
- The size, line height and weight of each type role for `figcraft`, recorded in `DESIGN.md`.

The reference is a marketing site with large type and roomy cells. Keep its character but use the compact control scale from `DESIGN.md`, since this is a work tool.

## Acceptance criteria

- [ ] `figcraft` appears in the style variation selector and renders the gallery and the sample work screen in light and dark.
- [ ] It matches `DESIGN.md`: zero radius, alpha-ink neutrals, inverted primary button, the three typefaces, uppercase mono labels, flush panes, no shadows.
- [ ] Uppercase labels stay legible and do not overflow with the long-label samples.
- [ ] No component or screen file mentions the style variation by name.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] The values stay those of the reference, including muted text built from alpha. Every colour pair below 4.5:1 is marked as failing on the Foundations route.
- [ ] `DESIGN.md` records every value and style variable this issue added.
