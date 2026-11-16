# 04 - Sample work screen and Recipes

Status: ready-for-agent
Blocked by: 03

Read `../spec.md` and `DESIGN.md` first.

## What to build

Several traits that differ between the references only show in a full work surface: flush panes versus carded panes, density, and how status colour reads in a list. Build one clickable sample screen to judge them in context, out of the four recipes in `DESIGN.md`.

- The four recipes of `DESIGN.md`: app shell, list pane, detail pane and form dialog.
- The sample work screen, composed only of those recipes: sidebar navigation, a top bar with breadcrumb and a command-palette trigger, a list pane and a detail pane, and a status bar.
- Fake in-memory data, clearly labelled as sample data in the code. Use generic items such as projects and requests. This does not define the product's domain or its vocabulary.
- Enough interaction to feel real: selecting a list row updates the detail pane, tabs in the detail pane switch content, the form dialog creates an item that appears in the list, filtering the list works, the command palette opens from the keyboard and navigates.
- Items carry the four statuses (working, needs attention, done, error).
- The shell's pane arrangement is driven by the style variables `pane-gap`, `pane-radius` and `pane-border`, so a style variation can render the panes flush or as separate cards without changing the screen's code.
- The Recipes route of the gallery: each recipe shown on its own in a framed preview, with one line on when to use it, and a link to the sample work screen.

## Acceptance criteria

- [ ] The screen is reachable from the prototype's navigation and built only from the recipes and `packages/ui` components.
- [ ] Each interaction listed above works; created items last until the page is reloaded.
- [ ] Changing the pane variables switches between flush and carded panes with no code change in the screen.
- [ ] The Recipes route shows the four recipes, and each matches its description in `DESIGN.md`.
- [ ] Correct in light and dark, and usable at a 375px-wide viewport (panes stack or collapse).
- [ ] Fully operable by keyboard.
