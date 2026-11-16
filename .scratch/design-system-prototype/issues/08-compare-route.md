# 08 - Compare route

Status: ready-for-agent
Blocked by: 06, 07

Read `../spec.md` and `DESIGN.md` first.

## What to build

A gallery route where the maintainer sees the three style variations at once, instead of switching and comparing from memory.

- The Compare route, last in the gallery navigation. It shows one column per style variation, read from the single list of style variations, each column rendered under its own style variation and under the current theme.
- The same examples in every column, in the same order, chosen to expose the traits listed in issue 10:
  - a type specimen with the display, interface and mono typefaces and a small label;
  - buttons: primary, secondary and ghost, one with a shortcut chip;
  - a field with a label, an input and an error;
  - tabs, a badge and the four status indicators;
  - a card and a table with a selected row;
  - two panes side by side, one focused;
  - a floating surface, shown open.
- Each column is headed by the name of its style variation. The `herdr` column states that its light theme is derived.
- At phone widths the columns stack, or scroll sideways inside the page without the page itself scrolling sideways.

This route is temporary. It is removed with the three style variations when the final style is built (see issue 10).

## Acceptance criteria

- [ ] The three columns render side by side at desktop width, each under its own style variation.
- [ ] Switching theme changes all three columns; the style variation selector does not change the columns.
- [ ] Adding a style variation to the list adds a column with no other change to the route.
- [ ] Every trait listed in issue 10 can be judged from at least one example.
- [ ] The page does not scroll sideways at a 375px-wide viewport.
