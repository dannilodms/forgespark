# Design system

How the product looks, and the rules every screen follows. This file is the source: the code in `apps/web/src/ui` implements it. When the code disagrees with this file, this file is right and the code is fixed.

The style follows Zed, its site and its application. The evidence behind every value is in section 1 of [`docs/design/references.md`](docs/design/references.md). Why Zed, and what else was considered, is in ADR-0006. Why the design system lives in the app, with no prototype and no gallery, is in ADR-0025.

The size of the text does not follow Zed. Zed sets its interface in 12 to 13px; here it is 16px, and every control is taller to hold it. The maintainer found 13px too small to read, and chose 16px over 14 and 15px from sample pages of the setup and of the app shell.

A change to a style variable, a component or a recipe updates this file in the same change.

## Vocabulary

**Style variable**:
A CSS custom property that holds one style value: a colour, a radius, a font, a spacing step, a shadow. Components read style variables and nothing else.

**Theme**:
Light or dark. Every style variable that changes with the theme has a value in each.

**Recipe**:
A named composition of components that several screens reuse, such as the app shell.

## Rules

1. **Borders carry the structure.** Surfaces are separated by 1px lines. A shadow appears only on a floating surface: popover, dialog, command palette, toast.
2. **One accent, and colour means status.** The one accent colour is blue. Any other colour is a status colour with a fixed meaning: blue is working, amber needs attention, green is done, red is an error. A status colour is never decoration.
3. **Mono is for metadata.** Shortcuts, counts, paths and identifiers are set in the mono typeface. Running text is set in the interface typeface.
4. **Decoration frames the page.** The grid texture and the diamond nodes belong to the page framing. The inside of a component is plain.
5. **Motion is short and plain.** A colour change takes 100 to 150ms. A control moves at most 1 to 2px. An overlay opens in 200 to 250ms.
6. **Every value is a style variable.** A component takes each colour, radius, font, spacing step and shadow from a style variable, and hardcodes none.
7. **One primary action per group.** In a group of actions, one button is primary. The others are secondary or ghost.
8. **Status is colour plus a sign.** A status always shows an icon or a word together with its colour.

One exception to rule 2: the mark of a login provider keeps the colours its owner requires. Google forbids a single-colour "G" and asks for it on white, and Microsoft asks that its logo is not altered. GitHub's mark takes the colour of the text.

## Accessibility and layout

- The style meets WCAG AA: 4.5:1 for text, 3:1 for large text, for the boundary of a control and for the focus indicator.
- A value taken from Zed that fails is changed by the smallest step that passes. The colour table then gives both the value of Zed and the value in use. Nothing measures the pairs automatically: a value is measured when it is added or changed.
- Every interactive component is operable by keyboard and shows a visible focus state: a 3px outline with a 2px offset. A box of InputOTP has no offset, because its neighbours touch it. A row of a list pane and a control in a 32px bar draw the outline inside themselves, because the pane or the bar would cut it.
- Motion respects `prefers-reduced-motion`.
- Layouts are designed for desktop first and adapt down to a 375px-wide viewport with no overflow.
- Text runs longer in Portuguese and Spanish than in English. Every component that carries text holds a long label without overflowing.

## Style variables

The names and roles below are the contract. The colour names are those of shadcn/ui, with the additions marked.

### Colour roles

| Style variable | Role |
| --- | --- |
| `background` / `foreground` | The page, behind and between panes. Default text. |
| `card` / `card-foreground` | A pane or a card. |
| `sidebar` / `sidebar-foreground` | The sidebar. |
| `chrome` (added) | The title bar and the status bar, one step stronger than `sidebar`. |
| `popover` / `popover-foreground` | A floating surface. |
| `primary` / `primary-foreground` | The primary button. |
| `secondary` / `secondary-foreground` | A raised control: secondary button, active tab, chip. |
| `muted` / `muted-foreground` | A sunken or quiet fill. Secondary text. |
| `accent` / `accent-foreground` | The hover wash. This is shadcn/ui's name; it is not the accent colour of rule 2. |
| `selected` (added) | The fill of a selected row or item, one step stronger than `accent`. |
| `faint-foreground` (added) | Tertiary text: placeholder, disabled text, an index. |
| `border` | The line between surfaces, and the edge of a floating surface. |
| `input` | The border of a field. It reaches 3:1 against the surface, which `border` does not have to. |
| `border-subtle` (added) | A lighter line inside a surface, between rows or groups. |
| `ring` | The focus indicator. |
| `brand` / `brand-foreground` (added) | The accent colour of rule 2 where it is not a button: a link, an icon, a highlighted word. |
| `destructive` | A destructive action and a validation error. |
| `status-working`, `status-attention`, `status-done`, `status-error` (added) | The four status colours of rule 2. |
| `status-*-subtle` (added) | The background of a status badge. |
| `button-edge`, `button-edge-primary` (added) | The inset bottom edge of a button, and that of the primary button. |
| `page-grid` (added) | The lines of the grid texture of the page framing. |

The four status colours are visual meanings. Which status of a record shows which meaning is decided with each screen.

### Colours

The values in use. Where a value of Zed failed WCAG AA, the value in use is followed by the one it replaces.

| Style variable | Light | Dark |
| --- | --- | --- |
| `background` | `#f5f4f3` | `#282c33` |
| `card` | `#fafafa` | `#282c33` |
| `popover` | `#ebebec` | `#2f343e` |
| `sidebar` | `#ebebec` | `#2f343e` |
| `chrome` | `#dcdcdd` | `#3b414d` |
| `secondary` | `#ebebec` | `#2e343e` |
| `muted` | `#ebebec` | `#2f343e` |
| `accent` | `#dfdfe0` | `#363c46` |
| `selected` | `#cacaca` | `#454a56` |
| `foreground` | `#242529` | `#dce0e5` |
| `muted-foreground` | `#555557` (Zed: `#58585a`) | `#b3b9c6` (Zed: `#a9afbc`) |
| `faint-foreground` | `#696a70` (Zed: `#7e8086`) | `#989ba9` (Zed: `#878a98`) |
| `border` | `#c9c9ca` | `#464b57` |
| `input` | `#8d8d8e` (Zed: `#c9c9ca`) | `#6f7582` (Zed: `#464b57`) |
| `border-subtle` | `#dfdfe0` | `#363c46` |
| `primary` | `#1348dc` | `#1348dc` |
| `primary-foreground` | `#ffffff` | `#ffffff` |
| `brand` | `#1348dc` | `#74ade8` |
| `brand-foreground` | `#ffffff` | `#1b1e23` |
| `ring` | `#094ece` at 65% (Zed: at 50%) | `#74ade8` (Zed: `#094ece` at 50%) |
| `destructive` | `#b14234` (Zed: `#d36151`) | `#e48488` (Zed: `#d07277`) |
| `button-edge` | `hsl(218 13% 50% / 0.1)` | the same |
| `button-edge-primary` | `hsl(219 93% 30%)` | the same |
| `page-grid` | `foreground` at 5% | `foreground` at 4.5% |

The text colours of a surface (`card-foreground`, `popover-foreground`, `secondary-foreground`, `accent-foreground`, `sidebar-foreground`) are `foreground`. The `sidebar-*` variables that shadcn/ui expects take the values of their plain counterparts.

What each adjustment had to reach:

- `muted-foreground`: 4.5:1 on a selected row, where it was 4.33:1 in light and 4.03:1 in dark.
- `faint-foreground`: 4.5:1 on the page, a pane and the sidebar. It was 3.31:1 at the lowest.
- `input`: 3:1 on the page and on a pane. The line of Zed is 1.5:1, which is enough between surfaces and not for the border of a field.
- `ring`: 3:1 on the page, a pane and the sidebar. In dark the blue of Zed cannot reach it at any opacity, so the focus takes the accent colour as text.
- `destructive` and the status colours: 4.5:1 as a word on the page, on a pane and on the subtle background of a badge, and 3:1 as a sign on a hovered row, on a selected row and on a bar.

`popover`, `muted` and `page-grid` have no value in the reference. `popover` and `muted` take Zed's panel surface, and `page-grid` is derived.

### Status colours

| Style variable | Light | Dark |
| --- | --- | --- |
| `status-working` | `#455dc5` (Zed: `#5c78e2`) | `#74ade8` |
| `status-attention` | `#886c00` (Zed: `#a48819`) | `#dec184` |
| `status-done` | `#3d7430` (Zed: `#669f59`) | `#a1c181` |
| `status-error` | `#b14234` (Zed: `#d36151`) | `#e48488` (Zed: `#d07277`) |
| `status-working-subtle` | `#e2e2fa` | the status colour at 10% opacity |
| `status-attention-subtle` | `#faf2e6` | the status colour at 10% opacity |
| `status-done-subtle` | `#dfeadb` | the status colour at 10% opacity |
| `status-error-subtle` | `#fbdfd9` | the status colour at 10% opacity |

`destructive` and `status-error` are the same red.

A status colour is a word only where it reaches 4.5:1: on the page, on a pane and in a badge. On a hovered or a selected row it does not: the lowest is amber on `selected` in light, at 3.06:1. The colours were not darkened a second time, which would take them further from Zed. In a row and in a bar the sign carries the colour, where a sign needs 3:1 and has it on every fill, and the word beside it is set in `foreground`, at the weight of text. Zed shows a status in a list the same way: a coloured sign beside plain text.

### Type roles

Each role is a `type-*` utility that sets family, size, line height, weight and tracking together.

| Token | Use | Typeface | Size / line height / weight |
| --- | --- | --- | --- |
| `type-display` | A page title, the headline of an empty state. Used sparingly. | display | 28 to 36px / 1.15 / 300 italic, tracking -0.01em |
| `type-title` | The title of a section, a card or a dialog. | interface | 18px / 26px / 600 |
| `type-body` | Running text and descriptions. | interface | 16px / 24px / 400 |
| `type-ui` | Controls, rows, menus, tabs. | interface | 16px / 20px / 400 |
| `type-label` | A small label, in sentence case: table header, field group, section index. | interface | 14px / 20px / 500 |
| `type-caption` | Secondary information under or beside something. | interface | 14px / 20px / 400 |
| `type-code` | Shortcuts, counts, paths, identifiers, code. | mono | 14px / 20px / 400 |

The display size grows with the width of the viewport, between its two ends.

| Typeface | Style variable | Family | Notes |
| --- | --- | --- | --- |
| Display | `font-serif` | IBM Plex Serif | Light italic, weight 300 |
| Interface | `font-sans` | IBM Plex Sans | 14 to 18px |
| Mono | `font-mono` | Lilex | |

The three families are under the SIL Open Font License 1.1 and are installed from npm packages. A shadcn/ui component sets its text in `text-sm`, which is pinned to the interface size, 16px with a 24px line, and its smaller text in `text-xs`, which is pinned to the size of a label, 14px with a 20px line.

### Spacing and control sizes

The base unit is 4px. Spacing steps are multiples of it.

| Size | Use |
| --- | --- |
| 24px | Shortcut chip, inside a button |
| 28px | Badge, extra-small button, a control inside a bar |
| 32px | Title bar, status bar, top bar, toolbar, small button, an item of a menu, the header of a table |
| 36px | Row of a list pane, of the sidebar or of the command palette |
| 40px | Default control |
| 44px | Prominent control |
| 14 to 16px | Icon: 16px, and 14px inside a chip, a badge or a small control |

Button, input, select and tab heights move together, so a form stays aligned.

### Radius

| Role | Value | Used by |
| --- | --- | --- |
| `radius-chip` | 2px | Badge, shortcut chip, checkbox |
| `radius-tab` | 4px | Tab, menu item |
| `radius-control` | 4px | Button, input, select |
| `radius-pane` | 4px | Pane, card |
| `radius-floating` | 6px | Dialog, popover, menu, tooltip, command palette, toast |

### Elevation

| Level | Surfaces | Treatment |
| --- | --- | --- |
| Flat | Page, chrome | None |
| Hairline | Pane, card, control | 1px border |
| Floating | Popover, dialog, menu, tooltip, command palette, toast, sheet | A 1px edge and the one shadow recipe, both in `shadow-floating` |

`shadow-floating` is `0 0 0 1px var(--border), 0 8px 24px -8px rgb(0 0 0 / 0.28)`. It is derived: the reference shows no floating surface. `shadow-xs`, which shadcn/ui puts on fields, is set to none.

The edge of a floating surface is a line in `border`, as rule 1 asks, so it shows in both themes: an edge of black at 6% did not show in dark. The utility draws it, and a floating surface adds no border of its own. Like every line between surfaces, the edge has no contrast level to reach.

A button has no shadow. Its depth is a 2px inset bottom edge, drawn by the `button-edge` and `button-edge-primary` utilities.

Behind a dialog or a sheet, the page is covered by `background` at 70% opacity. A tooltip has no arrow.

### Motion and icons

- Colour and opacity: 100 to 150ms, `cubic-bezier(.4, 0, .2, 1)`.
- Overlays: 200 to 250ms, opening from scale .96 to .98 along `ease-overlay`, which is `cubic-bezier(.2, .8, .2, 1)`.
- With `prefers-reduced-motion`, no animation and no transition runs: every state is reached at once.
- Icons are Lucide, at 14 to 16px.

## Traits

What makes the style recognisable, beyond the values above.

| Trait | How it shows |
| --- | --- |
| Primary button | Blue fill, inset darker bottom edge, and a shortcut chip inside when the action has a shortcut |
| Panes | Flush, split by 1px lines. No gap and no card between them |
| Depth | No shadow inside the work surface |
| Page framing | A 10px grid texture and diamond nodes where lines meet. A page opts in with the `page-framing` utility, and places each node with `frame-node` |
| Density | Flush panes and 1px lines, with 32px bars, 36px rows, 40px controls and 16px interface text: denser in layout than a web page, larger in text than Zed |
| Display type | A light italic serif, used sparingly against the sans of the interface |

## Brand mark

The mark is a software factory in one solid shape: a building with a sawtooth roof and two chimneys, the code symbol `</>` cut out of its wall, standing on a conveyor belt with three rollers. It is drawn on a 64-unit square.

- The mark has one colour: `brand` on a surface, or `primary-foreground` on a `primary` tile. It follows rule 2.
- The logo is the mark and the product name. The name is set in IBM Plex Sans at weight 600, in `foreground`, and converted to outlines, so the file needs no font.
- The smallest size of the mark is 16px. The space around it is at least a quarter of its width.
- The mark is never stretched, rotated, outlined or given a shadow.

The files are in [`docs/design/brand/`](docs/design/brand/). They are the source: an app serves copies of them.

| File | Use |
| --- | --- |
| `mark.svg`, `mark-on-dark.svg` | The mark alone, on a light and on a dark surface |
| `logo.svg`, `logo-on-dark.svg` | The mark and the product name, on a light and on a dark surface |
| `favicon.svg` | The browser tab. It changes colour with the theme of the browser |
| `favicon.ico` | The browser tab where an SVG is not accepted: 16, 32 and 48px |
| `apple-touch-icon.png` | The home screen of iOS: 180px, the mark in white on `primary` |
| `icon-192.png`, `icon-512.png` | The icons of the web app manifest |
| `icon-maskable-512.png` | The maskable icon of the web app manifest, with the mark inside the safe area |
| `social-preview.png` | The preview of a shared link and of the repository: 1280 by 640px |
| `source/` | What the image models returned: the concept of the mark and the background of the preview |

The shape of the mark and the background of the preview came from image models. The mark was then redrawn on the 64-unit square, and the name was set over the background.

## Components

`apps/web/src/ui` holds shadcn/ui components on Base UI primitives. A component enters when a screen needs it, through the shadcn CLI, and this table gains it in the same change.

| Group | Components |
| --- | --- |
| Actions | Button (with icon, with shortcut chip), Badge, Kbd |
| Inputs | Field, Label, Input, InputOTP, Textarea, Select, Combobox, Checkbox, RadioGroup, Switch, InputGroup |
| Navigation | Tabs, Breadcrumb, Sidebar |
| Data display | Card, Table, Avatar, Separator |
| Feedback | Alert, Toast, Skeleton, Empty, Status indicator |
| Overlays | Dialog, DropdownMenu, Popover, Tooltip, Command, Sheet |

The status indicator is this project's own component: a sign in one of the four status colours, and a word. The sign belongs to the meaning: a dashed circle for working, a triangle for needs attention, a check for done, a cross for error. It is plain in a row, where the sign carries the colour and the word is set in `foreground`, at the weight of text. It is a badge where it stands alone, with the sign and the word in the colour, on its subtle background. A badge is not put on a row that can be hovered or selected.

InputGroup and Sheet are in the list because Command and Sidebar are built on them. Command is built on the `cmdk` library, which brings its own primitives.

Combobox is a field whose value is chosen from a long list filtered by typing, such as a country: Select is for a short list read whole. Its list floats like a menu. The input is its one stop of the keyboard, and the arrow keys open the list from it; the button beside it is for the pointer and is out of the order of the Tab key. That button, the one that clears the field and the one that removes a chip are icons alone, so each requires its name as a prop, as the close button of a dialog does: `triggerLabel` and `clearLabel` on `ComboboxInput`, `removeLabel` on `ComboboxChip`. `ListField` names the button after its field ("Country: show the list").

A component is restyled through style variables and through its variants.

A component holds no text of its own. Every word comes from the screen, in the language of the person, and so does a name that only assistive technology reads: the label of the close button of a dialog, a sheet or a toast, of the buttons of a combobox, the name of the sidebar when it is a sheet and of the button that toggles it, the title and the description of the command palette, the label of a breadcrumb. A component that needs such a name requires it as a prop.

## Recipes

Five compositions are described here. The framed pane has code, `FramedPane` in `apps/web/src/shared/components`. The others have none yet: the first screen that needs one writes it in `apps/web`.

| Recipe | What it is |
| --- | --- |
| App shell | A title bar with the user menu, sidebar navigation, a top bar with breadcrumb and a command palette trigger, a main area of flush panes, and a status bar. |
| List pane | A filter, a list of rows that each carry a status indicator, and an empty state. Selecting a row drives the detail pane. |
| Detail pane | A header with the title and the actions of the selected item, and tabs that switch its content. |
| Form dialog | A dialog with fields, one primary action and a cancel action. It reports a validation error on the field. |
| Framed pane | A title bar with the product name, the framed page with one bordered pane in its centre, and a status bar that holds the language selector and nothing else. The screens seen before a login: the setup and the login screen. |

A recipe is added when a second screen needs the same composition.

**App shell.** The title bar and the status bar are 32px high, on `chrome`. The top bar is 32px high, on `card`, and its controls are 28px high. The shell takes the height it is given and never scrolls: each pane scrolls inside itself.

- The user menu sits at the end of the title bar. Its button shows the name of the person who is logged in. The menu shows the name and the e-mail, then the actions on the session. Logging out is the last of them, and it leads to the login screen.
- The button at the start of the top bar, or Ctrl B, collapses the sidebar to its icons. Under 768px the sidebar is a sheet that the same button opens.
- The command palette opens from its trigger in the top bar, or with Ctrl K. It lists the places to go, the actions of the screen and the items to open.
- The status bar holds counts: a status there is its sign and a number, with the word for whoever does not see the sign.

**List pane.** A row is 36px high: an identifier in mono, the title, and a plain status indicator at the end. A hovered row takes `accent` and the selected one takes `selected`.

- The filter keeps the rows that hold every word typed, whatever the case and the accents. The one action of the pane, creating an item, sits beside it.
- The arrow keys, Home and End move between the rows, and Enter selects one.
- When the filter leaves nothing, the empty state says what was typed and offers to clear it.

**Detail pane.** The header holds the identifier in mono, the status as a badge, the actions at its end, and the title under them, which wraps. The tabs are the `line` variant, across the pane. Under 1024px the list and the detail do not fit side by side: the screen shows the list, then the detail of the item picked, with a button that leads back.

**Framed pane.** The title bar and the status bar are 32px high, on `chrome`: the title bar shows the product name in caption type, and the status bar holds the language selector at its end. The page between them carries the page framing, with a diamond node on each corner of the pane. The pane is on `card`, with a 1px border and 24px of padding, in the centre of the page; at a narrow width it takes the width of the page, less 16px on each side. Its heading is display type. The screen has no theme toggle, no version number and no link to the source code.

- The pane has two widths. A screen of few controls, such as the login screen, takes the narrow pane, at most 384px wide. A form, such as the setup, takes the wide pane, at most 720px wide, so that its fields go side by side and the form stays short.
- In the wide pane the fields go in pairs, or in threes when they are short, such as city, region and postal code. A field with a long value, such as a website, takes the whole line. Where the group of fields is narrower than 512px, they go one under the other.
- A step of a form ends on a line above a 1px `border-subtle`: the way back to the step before at the start, when there is one, and the one primary action at the end. On a narrow screen the primary action takes the rest of the line.

**Form dialog.** The fields, then a footer with the cancel action and the primary action. Enter in a field submits. A validation error shows under its field, in the error colour, with a sign and a sentence that says what to do; the field is marked and takes the focus. A dialog that closes keeps nothing of what was typed.

## Known gaps

- The display type of the reference has a weight between 320 and 390. IBM Plex Serif on npm has no weight between 300 and 400, so display type is set at 300.
- The hover of the primary button is its fill at 90%. The reference does not show one.
- The reference shows no command palette and no product table. Those components follow the rules above, with nothing to match.
- Hover, focus and active states come from the reference's CSS. None was seen in motion.
- The reference's light theme was read from its site and from an HTML mock of the application, never from the real application.
