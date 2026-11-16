# 05 - Sample work screen and Recipes

Status: ready-for-agent
Blocked by: 04

Read `../spec.md` and `DESIGN.md` first.

## What to build

Some traits of the style only show in a full work surface: flush panes, density, and how status colour reads in a list. Build one clickable sample screen to judge them in context, out of the four recipes in `DESIGN.md`.

- The four recipes of `DESIGN.md`: app shell, list pane, detail pane and form dialog.
- The sample work screen, composed only of those recipes: a title bar, sidebar navigation, a top bar with breadcrumb and a command-palette trigger, a list pane and a detail pane, and a status bar.
- Fake in-memory data, clearly labelled as sample data in the code. Use generic items such as projects and requests. This does not define the product's domain or its vocabulary.
- Enough interaction to feel real: selecting a list row updates the detail pane, tabs in the detail pane switch content, the form dialog creates an item that appears in the list, filtering the list works, the command palette opens from the keyboard and navigates.
- Items carry the four statuses (working, needs attention, done, error).
- Panes are flush and split by 1px lines, with no shadow inside the work surface.
- The Recipes route of the gallery: each recipe shown on its own in a framed preview, with one line on when to use it, and a link to the sample work screen.
- The login screen leads here. Replace its provisional page (the `signed-in` route) with this screen, and add a user menu to the app shell with an action that logs out and returns to the login route. Record the user menu in the App shell recipe of `DESIGN.md`.

## Acceptance criteria

- [x] The screen is reachable from the prototype's navigation and built only from the recipes and `packages/ui` components.
- [x] Each interaction listed above works; created items last until the page is reloaded.
- [x] The Recipes route shows the four recipes, and each matches its description in `DESIGN.md`.
- [x] Correct in light and dark, and usable at a 375px-wide viewport (panes stack or collapse).
- [x] Fully operable by keyboard.
- [x] The prototype bar covers no control of the screen.

## Comments

### What was built

- The four recipes are in `apps/prototype/src/recipes`: `app-shell.tsx` (with the user menu), `list-pane.tsx`, `detail-pane.tsx` and `form-dialog.tsx`. They take their texts and their content from the screen, and use only components of `packages/ui`.
- The sample work screen is the route `/work`, in `apps/prototype/src/work`. It has two places, Requests and Projects, with 12 and 5 sample items that carry the four statuses. `sample-data.ts` says at its top that the data is sample data and defines nothing of the product.
- The place and the selected item are in the address (`/work?place=projects&item=PRJ-0012`), so the browser goes back through them. What is typed in the filter and the open dialog are state of the screen. The items are held in memory by `work-store.ts` and go back to the sample when the page reloads.
- The detail pane also changes the status of the item, through a menu. It was not asked for: it lets the four status colours be judged moving in the list and in the status bar.
- The login screen leads to `/work`, the `signed-in` route and its texts are gone, and the home page of the prototype lists the screen. The user menu sits at the end of the title bar, and its last action logs out and leads to `/login`.
- The Recipes route (`/gallery/recipes`) shows each recipe in a framed preview, with one line on when to use it and a link to the screen. The app shell is shown with stand-ins where its panes go.
- Every visible text of the screen comes from the translation files, in the three languages. Titles and descriptions of the sample items stay in English, as text typed by a person. Dates are written for the language in force.
- What the screen does to its items is plain functions in `work-items.ts`, built test first: filtering, checking a draft, adding an item, changing a status, counting by status. `sample-data.test.ts` checks that each place shows the four statuses. `task check` runs them, since the login scenarios wired the tests into it.

### The two known gaps, settled first

- **A status word on a hovered or a selected row.** The plain status indicator now keeps the colour on its sign and sets its word in `foreground`, at the weight of text. A sign needs 3:1, and every status colour has it on `accent`, on `selected` and on `chrome` in both themes (the lowest is amber on `selected` in light, at 3.06:1). The colours were not darkened again, and Zed shows a status in a list this way. A badge keeps the word in the colour, on its subtle background. The colour of the sign is set on its stroke, so a highlighted menu item, which recolours its content, leaves it alone.
- **The edge of a floating surface in dark.** The 1px edge of `shadow-floating` is drawn in `border` in place of black at 6%, so it shows in both themes. Sheet dropped the border it drew itself.
- The Foundations route measures the new pairs: 12 status signs (four statuses on a hovered row, a selected row and a bar, all above 3:1) and three lines between surfaces, which have no level to reach and show their ratio. It measures 57 pairs, and none fails in either theme.

### What was checked in the browser

- `/work` and `/gallery/recipes`, in light and dark, at 1280px and 375px wide, and `/work` at 900px: no horizontal overflow of the page or of the scrolling area (measured by script), and nothing overlapping in what was looked at. The console has no error, no warning and no issue.
- With the keyboard only: Tab reaches every control and each shows the 3px outline (measured by script on 25 elements); the arrow keys, Home and End move between the rows and Enter selects; Ctrl K opens the command palette, typing filters it and Enter goes to a place, opens an item or opens the form dialog; Enter in the empty form shows the error on the title and puts the focus there, and Enter with a title creates the item, which shows first in the list and selected; the status menu opens with Enter, moves with the arrow keys, and closes on a choice with the focus back on its button; the user menu logs out to `/login`; Ctrl B collapses the sidebar to icons, and each icon names its place in a tooltip; at 375px the button of the top bar opens the sidebar as a sheet, and going to a place closes it; Escape closes the palette, the dialog and the menus.
- With a real pointer: a hovered row takes `accent` and the selected one `selected`, and the sign of the status keeps its colour on both.
- A row, the title bar, the top bar and the status bar are 28px high. Inside the screen the only shadows are the inset edges of buttons.
- A created item and a changed status last through navigation between places and to the login and back, and are gone after a reload.
- The login screen, after the merge with its scenarios, leads to `/work`.
- In Portuguese at 375px and in Spanish at 1280px, nothing overflows.
- In dark, the command palette, the form dialog, the menus, the tooltips and the sheet show their 1px edge.

### What was not checked

- The arrow keys on the tabs of the detail pane were not pressed here: issue 04 checked them on the component.
- No screen reader was run. The accessibility tree was read: the two navigations, the two panes and each row have a name.
- `prefers-reduced-motion` was not looked at on this screen.
- The screen was seen in Chrome only.

### Found, and left for a decision

- **The prototype bar shows a 2px edge.** It has a `border` of its own and `shadow-floating`, which now draws the edge. Taking `border` out of `components/prototype-bar.tsx` fixes it. It was left alone because the login scenarios were changing that file at the same time.
- **At 375px a row gives the status word up to half of its width**, and a long word such as "Waiting for the client" leaves the title about 90px. The word could be cut first, or the row could take two lines on a narrow screen.
- **After an item is created from the command palette, the focus is on the page**, not on a control, because the palette that opened the dialog is closed by then. Created from the button of the list pane, the focus goes back to that button.
- The sidebar as a sheet keeps "Sidebar" and "Toggle Sidebar" in English for assistive technology: they are inside `packages/ui`. The visible texts and the label of the button are translated.
- Under 1024px the first item is not selected by default, since there the list shows alone. The width is read in the screen with a media query, beside the `lg` classes of the layout.
