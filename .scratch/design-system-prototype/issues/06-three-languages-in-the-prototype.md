# 06 - Three languages in the prototype

Status: ready-for-agent
Blocked by: 05

Read `../spec.md` and `docs/adr/0005-multilingual-from-the-start.md` first.

## What to build

The prototype in English (`en-US`, default), Brazilian Portuguese (`pt-BR`) and Spanish (`es`), so the maintainer sees how each layout behaves in each language.

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

- [ ] Switching language updates the navigation, the sample content in the gallery and the sample work screen without a reload.
- [ ] No translated string is hardcoded in components or screens of the prototype. The gallery's explanatory text is kept apart from the translation files, in one place per route.
- [ ] Removing a key from `pt-BR` or `es` makes the check fail and names the missing key.
- [ ] Translation keys are type-checked: using a key that does not exist fails the type check.
- [ ] Portuguese uses correct accents and Brazilian vocabulary; Spanish avoids regionalisms.
- [ ] Nothing overflows or overlaps in any of the three languages, at desktop and 375px widths.
