# 05 - Style variation switch and the Zed variation

Status: ready-for-agent
Blocked by: 04

Read `../spec.md`, "Style variations" in `DESIGN.md` and section 1 of `docs/design/references.md` first.

## What to build

The mechanism for switching style variations, proven with the first one.

- A style variation is a named set of values for the design system's style variables, in light and dark. It is selected on the root element, independently of the light/dark theme, and it can also be applied to any element, so that a part of a page renders under another style variation. Issue 08 depends on this.
- A single list of style variations. The selector and every page that shows more than one read from it.
- A selector in the prototype's top bar, visible on every page, switches style variation. The choice survives a reload.
- The `zed` style variation, following the `zed` column of Traits and "Colours of `zed`" in `DESIGN.md`. It draws on both the site (cream paper light theme, light italic serif display type, grid texture and diamond nodes as page framing) and the app (flush panes split by 1px lines, small interface text, 4px radius, no shadows inside the work surface).
- The size, line height and weight of each type role for `zed`, recorded in `DESIGN.md`.
- Wherever a trait needs a style variable that `DESIGN.md` does not list yet, add it to the design system and to `DESIGN.md`. Do not branch on the name of the style variation inside components.

## Acceptance criteria

- [ ] Switching style variation changes the gallery and the sample work screen without a reload and without layout breakage.
- [ ] Style variation and theme are independent: both `zed` light and `zed` dark render correctly.
- [ ] A style variation set on an element inside the page applies to that element's content only.
- [ ] No component or screen file mentions a style variation by name.
- [ ] The `zed` style variation matches `DESIGN.md`: surfaces, accent, radius, the three typefaces, flush panes, button edge and shortcut chip, page-level decoration only.
- [ ] Fonts are installed from npm packages, with their licences confirmed.
- [ ] The values stay those of the reference. Every colour pair below 4.5:1 is marked as failing on the Foundations route.
- [ ] `DESIGN.md` records every value and style variable this issue added.
