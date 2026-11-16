# 07 - Herdr variation

Status: ready-for-agent
Blocked by: 05

Read `../spec.md`, "Style variations" in `DESIGN.md` and section 3B of `docs/design/references.md` first.

## What to build

The `herdr` style variation, following the `herdr` column of Traits and "Colours of `herdr`" in `DESIGN.md`. The model is the Herdr application, not its website, so no pill buttons and no large marketing shadows.

- Panes as separate rounded cards with 8px gaps on a darker canvas; the focused pane's border takes the pale-blue accent.
- Rounded controls (8px), panes (10px), tabs (7px) and badges (5px).
- No brand colour: the primary button is an inverted neutral, hover and selection are white-alpha fills, and colour is reserved for status.
- Geist for interface and display, JetBrains Mono for metadata.
- One shadow recipe, a 1px light ring plus a large soft shadow, used only on floating surfaces (dialog, popover, command palette).
- The size, line height and weight of each type role for `herdr`, recorded in `DESIGN.md`.

The reference has no official light theme. Derive one from the light column of "Colours of `herdr`" in `DESIGN.md`, fill in the values that column marks "set when built", record them there, and label the light theme as derived in the gallery.

## Acceptance criteria

- [ ] `herdr` appears in the style variation selector and renders the gallery and the sample work screen in light and dark.
- [ ] It matches `DESIGN.md`: carded panes with gaps, the radius steps, neutral primary button, white-alpha hover and selection, status-only colour.
- [ ] The focused pane in the sample work screen shows the accent border, by keyboard and by pointer.
- [ ] The gallery states that this style variation's light theme is derived, not taken from the reference.
- [ ] No component or screen file mentions the style variation by name.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] The dark values stay those of the reference. Every colour pair below 4.5:1 is marked as failing on the Foundations route.
- [ ] `DESIGN.md` has no value left as "set when built", and records every style variable this issue added.
