# 04 - Style variation switch and the Zed variation

Status: ready-for-agent
Blocked by: 03

Read `../spec.md` and section 1 of `docs/design/references.md` first.

## What to build

The mechanism for switching style variations, proven with the first one.

- A style variation is a named set of values for the design system's style variables, in light and dark. It is selected on the root element, independently of the light/dark theme.
- A selector in the prototype's top bar, visible on every page, switches variation. The choice survives a reload.
- The `zed` variation, following the `zed` column of the table in the spec. It draws on both the site (cream paper light theme, light italic serif display type, grid texture and diamond nodes as page framing) and the app (flush panes split by 1px lines, small interface text, 4px radius, no shadows inside the work surface).
- Wherever a trait needs a variable that does not exist yet (label casing and tracking, button edge treatment, page decoration), add the variable to the design system. Do not branch on the variation name inside components.

## Acceptance criteria

- [ ] Switching variation changes the gallery and the sample work screen without a reload and without layout breakage.
- [ ] Variation and theme are independent: both `zed` light and `zed` dark render correctly.
- [ ] No component or screen file mentions a variation by name.
- [ ] The `zed` variation matches the spec table: surfaces, accent, radius, the three type roles, flush panes, button edge and shortcut chip, page-level decoration only.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] Text contrast meets WCAG AA in both themes.
