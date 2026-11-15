# 02 - Component gallery

Status: ready-for-agent
Blocked by: 01

Read `../spec.md` first.

## What to build

A gallery section in the prototype where the maintainer reviews the design system in isolation.

- A page for the style variables: colour roles as swatches with their names, the type scale, radius steps, spacing steps, and shadows.
- A page, or one page per group, for the components. Each component appears in every variant, size and state (default, hover, focus, active, disabled, loading and error where they apply).
- Components, at least: Button (with icon and with a shortcut chip), Input, Textarea, Select, Checkbox, Radio, Switch, Tabs, Badge, status indicator (working, needs attention, done, error), Card, Table, Dialog, Dropdown menu, Popover, Tooltip, Command palette, Keyboard shortcut chip, Avatar, Breadcrumb, Toast, Skeleton and an empty state.
- Each text-bearing component is also shown with a long label, to stand in for Portuguese and Spanish, which run longer than English.
- Gallery navigation that works at desktop and phone widths.

Add the components to `packages/ui` through shadcn/ui; the gallery only composes them.

## Acceptance criteria

- [ ] Every listed component is in `packages/ui` and shown in the gallery in all its variants and states.
- [ ] Every style variable defined in `packages/ui` appears on the variables page; adding one makes it appear without editing the page by hand, or the page documents the single place to register it.
- [ ] All pages are correct in light and dark.
- [ ] Nothing overflows or overlaps at a 375px-wide viewport or with the long-label samples.
- [ ] Every interactive component is reachable and operable by keyboard, with a visible focus state.
