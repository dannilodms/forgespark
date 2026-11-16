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

- [x] Every component listed in `DESIGN.md` is in `packages/ui` and has a block in the gallery.
- [x] The coverage check passes.
- [x] Each of the eight rules has a Do and a Don't example.
- [x] All routes are correct in light and dark.
- [x] Nothing overflows or overlaps at a 375px-wide viewport or with the long-label samples.
- [x] Every interactive component is reachable and operable by keyboard, with a visible focus state.
- [x] A status never appears as colour alone.

## Comments

### What was built

- 23 components came from the shadcn CLI and were restyled: Badge, Kbd, Textarea, Checkbox, RadioGroup, Switch, InputGroup, Tabs, Breadcrumb, Sidebar, Card, Table, Avatar, Alert, Toast, Skeleton, Empty, Dialog, DropdownMenu, Popover, Tooltip, Command, Sheet. The status indicator is this project's own. `packages/ui` holds 31 components, and each has a block.
- InputGroup and Sheet are not in the first list of `DESIGN.md`: Command is built on InputGroup and Sidebar on Sheet, so the CLI brought them. They are in the list now, on the Inputs and the Overlays routes.
- Command is built on the `cmdk` library (MIT), which depends on Radix primitives. The catalog and `packages/ui` name it.
- New in the style: the style variable `ease-overlay`, an opening of 200ms from a scale of .97 with no slide (Select included), nothing animated under `prefers-reduced-motion`, `background` at 70% behind a dialog or a sheet, and a tooltip with no arrow. `DESIGN.md` records each.
- Left out of what the CLI generated: the badge of an avatar (a dot of colour would be a status with no sign), the cookie of the sidebar, and its `useMemo` and `useCallback`.
- The routes Navigation, Feedback, Overlays and Do and Don't are new. Actions, Inputs and Data display are complete.

### What was checked in the browser

- Light and dark, 1280px and 375px wide, on the seven routes: no horizontal overflow of the page or of a frame (measured by script on every route at both widths), and nothing overlapping in the sections looked at.
- The console has no error and no warning on any route. Chrome reports one issue, "A form field element should have an id or name attribute", for sample fields that carry only an `aria-label`.
- Keyboard: every element that takes focus on the seven routes shows the 3px outline (measured by script, 254 elements; a box of InputOTP shows it on the active box). Opened and closed with the keyboard: dialog, sheet, command palette (arrow keys move the selection), popover, tooltip (opens on focus), menu (arrow keys, the submenu with the right arrow), the sidebar as a sheet at 375px. A checkbox and a switch toggle with Space, and the arrow keys move a radio group and the tabs. Focus returns to the button that opened an overlay.
- An overlay opens in 0.2s along `cubic-bezier(0.2, 0.8, 0.2, 1)`, with a 6px radius and `shadow-floating` (read from the computed style of the dialog).
- With `prefers-reduced-motion`, an animation and a transition last 0.01ms.
- A table row is 28px high.

### What was not checked

- The letter jump of a menu and the checkbox and radio items of a menu were not pressed.
- Every section was looked at in one theme at least, and most in both. Seen in light only: Card, Avatar, Alert, Skeleton, Empty, Kbd, Command. Seen in dark only: Breadcrumb, Popover, Tooltip. The overflow, console and focus checks ran on every route.
- The hover of a row, a menu item and a tooltip under a real pointer: hover was seen only as the pinned state of a matrix.
- No screen reader was run.
- The sample content is in English, with the long labels in Portuguese and Spanish. Moving it to the translation files is issue 06.

### Found, and left for a decision

- Status colours on a selected row or under the hover wash are below 4.5:1 in places (lowest 3.06:1, amber on `selected` in light). The list pane of issue 05 puts a status on a selected row. It is in "Known gaps" of `DESIGN.md`.
- In dark, a floating surface stands out by its fill and a soft shadow alone. A tooltip beside a button is hard to tell from it. A border on floating surfaces would follow rule 1. Also in "Known gaps".
- Run in a worktree under `.claude/worktrees`, the shadcn CLI wrote the generated files to the root of the main checkout instead of `packages/ui/src/components`. They were moved into the worktree, and the main checkout was left as it was.
