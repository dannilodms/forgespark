# 04 - Gallery components and the Do and Don't route

Status: ready-for-agent
Blocked by: 03

Read `../spec.md`, `DESIGN.md` and `.claude/skills/shadcn/design-system-page.md` first.

## What to build

The components of the design system, each shown in the gallery, and the rules of `DESIGN.md` rendered as pairs.

Field, Label, Input, InputOTP, Select and Separator are already in `packages/ui`, with a block each on the Inputs and Data display routes: the login screen needed them first. Complete those routes; do not add those components again.

- Add to `packages/ui`, through the shadcn CLI, every component listed under "Components" in `DESIGN.md` that is not there yet. The gallery only composes them.
- The status indicator, this project's own component, with its four meanings: working, needs attention, done, error.
- The routes Actions, Inputs, Navigation, Data display, Feedback and Overlays, with the components `DESIGN.md` assigns to each. Every component gets a block built to the block anatomy in "Gallery" in `DESIGN.md`, following the Button block from issue 02.
- States pinned with `StateMatrix`: a variant by state matrix for Button, Badge, Input, Textarea, Select, Checkbox, RadioGroup, Switch and Tabs, with loading for Button and invalid for fields. The other components show single examples.
- Edge cases beside the component they stress: a long label on every component that carries text, an empty Table and an empty Command, a Field with an error, loading through Skeleton and through a Button with a spinner, an Avatar with no image, every field disabled.
- Each overlay opens from a real trigger labelled with what it opens.
- The Do and Don't route: the eight rules of `DESIGN.md`, each as a rendered pair with the rule as a one-line caption.

## Acceptance criteria

- [ ] Every component listed in `DESIGN.md` is in `packages/ui` and has a block in the gallery.
- [ ] The coverage check passes.
- [ ] Each of the eight rules has a Do and a Don't example.
- [ ] All routes are correct in light and dark.
- [ ] Nothing overflows or overlaps at a 375px-wide viewport or with the long-label samples.
- [ ] Every interactive component is reachable and operable by keyboard, with a visible focus state.
- [ ] A status never appears as colour alone.
