# 06 - Three languages in the prototype

Status: ready-for-agent
Blocked by: 05

Read `../spec.md` and `docs/adr/0005-multilingual-from-the-start.md` first.

## What to build

The prototype in English (`en-US`, default), Brazilian Portuguese (`pt-BR`) and Spanish (`es`), so the maintainer sees how each layout behaves in each language.

The login screen brought the base: react-i18next, the three translation files, the list of supported languages, the stored choice and the type-checked keys are in `apps/prototype` already. Build on them.

- react-i18next in `apps/prototype`, with the prototype's own translation files. They are not shared with any other package.
- A language selector in the prototype bar, next to the theme toggle. The choice survives a reload; with no stored choice, the browser's language applies when supported, otherwise English.
- What comes from the translation files:
  - the prototype's navigation and the gallery's navigation;
  - the sample content inside the components shown in the gallery (button labels, field labels, table headers, messages), since that is what tests a layout in each language;
  - every visible text of the sample work screen, including the fake data's labels and statuses.
- What stays in English only: the gallery's explanatory text, such as the line on when to use a component, the captions of the Do and Don't pairs and the names of style variables. It is documentation, like `DESIGN.md`. The labels of the prototype bar stay in English too: the bar is a tool of the prototype, not part of a screen.
- Sample content typed "by users" (item titles, descriptions) stays in one language, as free text does in the product.
- Dates and numbers are formatted for the selected language.
- English files are the source of truth. A check, run with the other checks, fails when `pt-BR` or `es` is missing a key that English has.
- A single list of supported languages. Adding a language means adding its code to that list and delivering its files; document this in the prototype's README.

## Acceptance criteria

- [x] Switching language updates the navigation, the sample content in the gallery and the sample work screen without a reload.
- [x] No translated string is hardcoded in components or screens of the prototype. The gallery's explanatory text is kept apart from the translation files, in one place per route.
- [x] Removing a key from `pt-BR` or `es` makes the check fail and names the missing key.
- [x] Translation keys are type-checked: using a key that does not exist fails the type check.
- [x] Portuguese uses correct accents and Brazilian vocabulary; Spanish avoids regionalisms.
- [x] Nothing overflows or overlaps in any of the three languages, at desktop and 375px widths.

## Comments

### What was built

- **Language selector.** The prototype bar holds the language next to the theme, with each language under its own name. The choice was already stored by the login effort; with none stored, the browser's language applies when supported (`es-MX` gives Spanish), otherwise English.
- **Single list.** `apps/prototype/src/locales/languages.ts` is the one list of supported languages. The bar reads it, the app loads the file of each language in it, and the check reads it. Adding a language is adding it there and delivering `<code>.json`. The README of the prototype, which is new, says so.
- **Missing-key check.** `apps/prototype/scripts/missing-keys.mjs` fails when `pt-BR` or `es` lacks a key that English has, or when a language of the list has no file, and names each key by its whole path. `task check` runs it. The comparison is `locales/missing-keys.ts`, built test first (7 tests). A key that only a translation has is left alone, since a language can have a plural form that English lacks.
- **Translated.** The home page of the prototype, the navigation of the gallery and the name of each of its routes, and the sample content inside every component block, the Do and Don't pairs, the type specimens of the Foundations route and the previews of the Recipes route. The translation files went from 90 to 337 keys per language.
- **Long labels.** The edge cases held one long label in Portuguese or in Spanish, hardcoded. Each is now a key with a long text in the three languages.
- **Explanatory text.** In English, in one module per route under `apps/prototype/src/gallery/explanations`. `shared.ts` holds the names of the parts that every block has ("Typical", "Variants", "States"). The names and roles of the style variables stay in `style-variables.ts`, the registry the coverage check reads.
- **Free text.** What a person would have typed (the names of people, clients and projects, the titles and descriptions of requests) is in `gallery/free-text.ts`, in one language, as `work/sample-data.ts` already was for the work screen. The long ones are in Portuguese.
- **Dates and numbers.** `useFormat`, in `stores/language.ts`, writes a number, a day and "12 days ago" with the browser's `Intl` for the language in force. The gallery uses it for every count and every decimal (`1,240.5`, `1.240,5`, `1240,5`), and the work screen for its days and its counts. `locales/format.ts` has its tests (4).
- **Sample work screen.** Issue 05 had translated its texts. This issue added the name of its breadcrumb and of the close button of its form dialog, and the formatting of its numbers.

### The three leftovers

- **The edge of the prototype bar.** The bar dropped its own border. Its edge is the one `shadow-floating` draws: a `border` of 0px and one 1px line, read from the computed style.
- **Hidden texts of `packages/ui`.** The sidebar takes the name it has as a sheet in a `label` prop, required unless it never collapses, and its trigger requires an `aria-label`. The description of the sheet, "Displays the mobile sidebar.", is gone. The same rule was applied to the other hidden English texts found in the package: the close button of a dialog and of a sheet (`closeLabel`, required unless `showCloseButton` is false), the close button of a toast (`closeLabel` on `Toaster`), the title and the description of the command palette (no default any more), and the label of a breadcrumb (`aria-label`, required). The ellipsis of a breadcrumb lost a hidden "More" that nothing could read, since the element is hidden from assistive technology. `DESIGN.md` and `CODING_STANDARDS.md` state the rule.
- **Command tables.** `README.md` and `README.pt-BR.md` say what `task check` runs now. `task check:fix` said "the same checks": it runs format, lint and types only, and says so.

### Decided while building

- **The name of a gallery route is translated**, since it is the word the navigation shows. The line under it is explanatory and stays in English.
- **A button that opens an example is explanatory**: "Open the rename dialog" is the gallery speaking, not the component. It stays in English, in the module of its route. The same holds for the sentence inside the sheets and the tooltips of the "Sides" parts.
- **The name of a key is not translated** (Ctrl, Shift, Backspace, Page Down): it is what the keyboard shows. An identifier, a version and a URL are written as they are.
- **A tag such as "Billing" is translated**: a category is configuration content, which has a version per language (ADR-0005).
- **One translation file per language** was kept, as the ticket describes them, with a group for each area (`prototype`, `login`, `work`, `gallery`).
- **A toast keeps the language it was shown in.** One that is on the screen when the language changes is not rewritten.
- **The check is a script**, like the coverage of the gallery, so `task check` lists it by name and its failure names the keys.

### What was checked in the browser

- **Without a reload.** With the keyboard, on `/gallery/data-display` and on `/work`: the select of the bar opens with Enter, the arrow keys move, Enter chooses, and the focus goes back to it with the 3px outline. The navigation, the title of the route, the sample content, the statuses, the counts and the dates changed in the same document, and the choice was in the storage of the browser.
- **By script**, on the nine routes of the gallery, in English, Portuguese and Spanish at 1280px and at 375px (three runs in dark, three in light): no horizontal overflow of the page, of the scrolling area or of a frame, no element past the right edge, no text cut without an ellipsis, and the prototype bar inside the viewport. The same on `/work` (Spanish at 1280px, Portuguese at 375px), `/login` (Portuguese and Spanish at 375px) and the home page (Spanish at 1280px).
- **Seen:** Actions in Portuguese (dark, 1280px), Data display in Portuguese (light, 375px and 1280px), Do and Don't in Spanish (light, 375px), Navigation in Spanish (dark, 1280px), the delete confirmation in Portuguese (dark, 375px), the sidebar of the work screen as a sheet in Portuguese (dark, 375px), the login screen and the home page in Portuguese (dark, 375px).
- **Names for assistive technology**, read from the page: the sheet of the sidebar is "Navegação principal", the close button of the form dialog is "Fechar", a toast closes with "Fechar esta mensagem", the breadcrumb of the work screen is "Ruta hasta esta página".
- **Overlays in Portuguese at 375px**, opened by script on the Overlays route: the three dialogs, the three menus, the six popovers, the five tooltips, the command palette and the five sheets each open inside the viewport with nothing wider than itself. Three stacked toasts stay inside it too, and the form dialog of the work screen shows its error on the field.
- **Numbers:** `6.0` and `1,240.5` in English, `6,0` and `1.240,5` in Portuguese, `6,0` and `1240,5` in Spanish; `13 de nov. de 2026` on the work screen.
- **The console** has no error and no warning on the routes visited. Chrome reports the same issue as before, "A form field element should have an id or name attribute", for sample fields that carry only an `aria-label`.
- **The checks:** a key removed from `pt-BR.json` and one from `es.json` made `task check` fail with `pt-BR is missing 1 of the keys of en-US: gallery.common.saveDraft` and the like. A key that does not exist, `t("gallery.common.saveChange")`, failed the type check.

### What was not checked

- Most combinations of route, language, theme and width were measured by script and not looked at.
- The overlays of the gallery were opened in Portuguese only, and measured, not looked at one by one. The delete confirmation was seen.
- No screen reader was run. The names were read from the page.
- The prototype was seen in Chrome only.
- The translations were written by the agent and read once. Nobody who writes Spanish every day reviewed them.

### Found, and left for a decision

- **The login screen has its own language selector**, and the bar now has one too. On `/login` the two sit one above the other. The one of the screen came with the login effort, as part of the screen: whether a person picks a language before logging in is a decision of that effort (its issue 03 approves the screen).
- **At 375px the prototype bar takes two lines**, and three on a route with two groups of scenarios, such as `/login` (92px). It stays inside the viewport and covers nothing, since the page scrolls above it.
- **The separator of the sidebar** ran 16px past its width, so the sidebar of the gallery could scroll sideways. It was a class of `packages/ui` that lost to the one of the separator, and it is fixed.
- **The caption of a table** scrolls sideways with the table at 375px, so its end is out of sight until the table is scrolled. It was so before this issue.
- **Section "Feedback" in Spanish** is "Avisos y estados": Spanish has no single neutral word for it.
