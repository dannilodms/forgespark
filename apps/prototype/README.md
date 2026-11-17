# Prototype

The clickable prototype of My Software Factory Manager. Screens are designed and approved here, with fake data and no backend, before `apps/web` builds them. The gallery is its permanent part: it shows every style variable and every component of `packages/ui`.

The rules are in "Prototype" in [`CODING_STANDARDS.md`](../../CODING_STANDARDS.md), and the look is in [`DESIGN.md`](../../DESIGN.md). Why a prototype exists is in ADR-0006.

## Running it

From the repository root:

| Command | What it does |
| --- | --- |
| `task prototype` | Starts the prototype at `http://localhost:5173` |
| `task check` | Runs every check below |

The prototype bar, at the bottom of every route, holds the theme, the language and the scenarios of the current screen. Its labels are in English.

## Languages

The prototype is in English (`en-US`, the default), Brazilian Portuguese (`pt-BR`) and Spanish (`es`), as ADR-0005 decides. The language chosen in the prototype bar is stored in the browser. With none stored, the browser's language applies when it is supported, otherwise English.

### Where each text lives

| Text | Where | Languages |
| --- | --- | --- |
| The navigation, every text of a screen, the sample content inside a component of the gallery | `src/locales/<code>.json` | The three |
| The explanatory text of the gallery: when to use a component, the name of a part, the caption of a Do and Don't pair, the label of a button that opens an example | `src/gallery/explanations/<route>.ts`, one file per route | English |
| The names and roles of the style variables | `src/gallery/style-variables.ts` | English |
| What a person would have typed: names, titles, descriptions | `src/gallery/free-text.ts` for the gallery, `src/work/sample-data.ts` for the sample work screen | The one it was typed in |
| The labels of the prototype bar and of the scenarios | Where they are declared | English |

A component or a screen writes no text of its own. It reads a translated text with `t("…")` or `<Trans>`, and writes a date or a number with `useFormat`, from `src/stores/language.ts`, so both follow the language. The name of a key, an identifier and a URL are the same in every language and are written as they are.

`packages/ui` has no translation: a component takes each word as a prop, and the screen supplies it. That includes a name read only by assistive technology, such as the label of a close button.

### Keys

The English file is the source of truth. A key is type-checked: `t` accepts only a key of `en-US.json`, so `task check` fails on a key that does not exist.

`task check` also runs `scripts/missing-keys.mjs`, which fails when another language lacks a key that English has, and names each missing key. Whoever adds a text delivers it in the three languages in the same change.

A text that counts something has one key for each plural form (`_one`, `_other`), with `count` as its number.

### Adding a language

1. Add its code and its own name to the list in `src/locales/languages.ts`. This is the single list: the prototype bar, the texts the app loads and the check all read it.
2. Deliver its file, `src/locales/<code>.json`, with every key of `en-US.json`.

`task check` fails until the file has every key.

## Checks

`task check` runs, in this order:

1. `vp check`: format, lint and types, for the whole JavaScript workspace.
2. `scripts/gallery-coverage.mjs`: fails when a component of `packages/ui` is shown nowhere in the gallery, or a style variable is not registered in the Foundations route.
3. `scripts/missing-keys.mjs`: fails when a language is missing a translation key.
4. The tests, which sit beside the module they cover. The prototype has no screen tests: a reducer, the fake data and a module with no screen have tests.
