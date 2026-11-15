# 06 - Herdr variation

Status: ready-for-agent
Blocked by: 04

Read `../spec.md` and section 3B of `docs/design/references.md` first.

## What to build

The `herdr` style variation, following the `herdr` column of the table in the spec. The model is the Herdr application, not its website, so no pill buttons and no large marketing shadows.

- Panes as separate rounded cards with 8px gaps on a darker canvas; the focused pane's border takes the pale-blue accent.
- Rounded controls (8px), panes (10px), tabs (7px) and badges (5px).
- No brand colour: the primary button is an inverted neutral, hover and selection are white-alpha fills, and colour is reserved for status.
- Geist for interface and display, JetBrains Mono for metadata.
- One shadow recipe, a 1px light ring plus a large soft shadow, used only on floating surfaces (dialog, popover, command palette).

The reference has no official light theme. Derive one from the values in the spec table and label it as derived in the gallery.

## Acceptance criteria

- [ ] `herdr` appears in the variation selector and renders the gallery and the sample work screen in light and dark.
- [ ] It matches the spec table: carded panes with gaps, the radius steps, neutral primary button, white-alpha hover and selection, status-only colour.
- [ ] The focused pane in the sample work screen shows the accent border, by keyboard and by pointer.
- [ ] The gallery states that this variation's light theme is derived, not taken from the reference.
- [ ] No component or screen file mentions the variation by name.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] Text contrast meets WCAG AA in both themes.
