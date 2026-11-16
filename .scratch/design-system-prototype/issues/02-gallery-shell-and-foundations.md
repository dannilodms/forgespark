# 02 - Gallery shell and Foundations

Status: ready-for-agent
Blocked by: 01

Read `../spec.md`, `DESIGN.md` and `.claude/skills/shadcn/design-system-page.md` first.

## What to build

The gallery as a section of the prototype, with its first two routes and the pieces every later route reuses.

- The gallery shell: one route per section and a side navigation that works at desktop and phone widths. The navigation lists the routes that exist; later issues add theirs. The theme is chosen in the prototype bar from issue 01; the gallery has no selector of its own.
- The showcase helpers from `.claude/skills/shadcn/assets/showcase/`, copied into the prototype: `ColorPair`, which measures contrast live, and `StateMatrix` with `previewState`, which pin hover, focus and active side by side. The stylesheet that pins the states belongs to the prototype, never to `packages/ui`.
- The Foundations route, following "Foundations" in the skill file and "Style variables" in `DESIGN.md`:
  - Colour roles as surface and text pairs, each with its style variable names, its resolved value, the measured contrast and the WCAG level. A pair below 4.5:1 is marked as failing.
  - The remaining colours as swatches: borders, focus, accent, status colours and their subtle backgrounds.
  - Type roles as a specimen table in real copy, and one paragraph of running text with a link and inline code.
  - Spacing steps and control sizes, the radius roles, the elevation levels, motion and icons.
- The Actions route with one component block, Button, built to the block anatomy in "Gallery" in `DESIGN.md`. It sets the pattern the other blocks follow.
- A coverage check, run with the other checks, that fails when a component file of `packages/ui` is imported by no gallery file. `.claude/skills/shadcn/scripts/showcase-coverage.mjs` is a starting point.

## Acceptance criteria

- [x] The gallery is reachable from the prototype's navigation, and each section has its own address.
- [x] Every style variable defined in `packages/ui` appears on the Foundations route; adding one makes it appear without editing the page by hand, or the page documents the single place to register it.
- [x] Every colour pair shows its measured contrast, and the measure updates when the theme changes.
- [x] The Button block shows variants, sizes, icon forms, the state matrix with loading, and a long label.
- [x] Removing Button from the gallery makes the coverage check fail and name the component.
- [x] Both routes are correct in light and dark.
- [x] Nothing overflows or overlaps at a 375px-wide viewport.
- [x] The gallery contains no lorem ipsum, and no state name is written inside a component.
