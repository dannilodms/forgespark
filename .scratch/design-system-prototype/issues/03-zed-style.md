# 03 - The Zed style

Status: ready-for-agent
Blocked by: 02

Read `../spec.md`, `DESIGN.md` and section 1 of `docs/design/references.md` first.

## What to build

The values of the style, in place of the placeholders from issue 01, checked on the Foundations route.

- Every style variable takes its value from "Colours", "Status colours", "Radius", "Elevation" and "Motion and icons" in `DESIGN.md`, in light and dark.
- The three typefaces: IBM Plex Serif for display, IBM Plex Sans for the interface, Lilex for mono. The type roles are built as `type-*` utilities, and the size, line height and weight of each are set inside the ranges of `DESIGN.md` and recorded there.
- The values `DESIGN.md` lists under "Known gaps" as having no value in the reference (`muted`, `popover`, `shadow-floating`) are derived and recorded there.
- The Button in the Zed style: blue fill with the inset bottom edge on the primary variant, the lighter edge on the others, and a shortcut chip inside when the action has a shortcut.
- The page framing: the 10px grid texture and the diamond nodes, as something a page opts into. It never appears inside a component.
- The contrast pass. Read the measured contrast of every colour pair on the Foundations route, in light and dark. Where a pair fails WCAG AA, change the value by the smallest step that passes, and record in the colour table of `DESIGN.md` both the value of Zed and the value in use.

Do not add a second set of values or a way to switch between sets (see `../spec.md`).

## Acceptance criteria

- [ ] The Foundations route and the Button block match `DESIGN.md`: surfaces, accent, status colours, radius, the three typefaces, the button edge and the shortcut chip.
- [ ] No colour pair on the Foundations route fails WCAG AA, in light or in dark.
- [ ] Every adjusted value is recorded in `DESIGN.md` beside the value of Zed it replaces.
- [ ] The focus indicator is a 3px outline with a 2px offset, and it passes 3:1 against the surfaces it appears on.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] No component file holds a colour, a radius, a font or a shadow as a literal.
- [ ] The items this issue settles are removed from "Known gaps" in `DESIGN.md`.
