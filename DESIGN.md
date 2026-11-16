# Design system

How the product looks, and the rules every screen follows. This file is the source: the code in `packages/ui` implements it, and the gallery in `apps/prototype` renders it. When the two disagree with this file, this file is right and the code is fixed.

The style follows Zed, its site and its application. The evidence behind every value is in section 1 of [`docs/design/references.md`](docs/design/references.md). Why Zed, and what else was considered, is in ADR-0006. The work is tracked in `.scratch/design-system-prototype/`.

A change to a style variable, a component or a recipe updates this file in the same change.

## Vocabulary

**Style variable**:
A CSS custom property that holds one style value: a colour, a radius, a font, a spacing step, a shadow. Components read style variables and nothing else.

**Theme**:
Light or dark. Every style variable that changes with the theme has a value in each.

**Gallery**:
The section of the prototype that shows every style variable and every component of `packages/ui`, in every variant and state. It is permanent: it is the reference for every screen.

**Recipe**:
A named composition of components that several screens reuse, such as the app shell.

## Rules

The gallery renders each of these eight as a Do and Don't pair.

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
- A value taken from Zed that fails is changed by the smallest step that passes. The colour table then gives both the value of Zed and the value in use. The gallery measures every pair, so a failure is visible.
- Every interactive component is operable by keyboard and shows a visible focus state: a 3px outline with a 2px offset. A box of InputOTP has no offset, because its neighbours touch it.
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
| `border` | The line between surfaces. |
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
- `destructive` and the status colours: 4.5:1 as a word on the page, on a pane and on the subtle background of a badge.

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

### Type roles

Each role is a `type-*` utility that sets family, size, line height, weight and tracking together.

| Token | Use | Typeface | Size / line height / weight |
| --- | --- | --- | --- |
| `type-display` | A page title, the headline of an empty state. Used sparingly. | display | 28 to 36px / 1.15 / 300 italic, tracking -0.01em |
| `type-title` | The title of a section, a card or a dialog. | interface | 15px / 20px / 600 |
| `type-body` | Running text and descriptions. | interface | 13px / 20px / 400 |
| `type-ui` | Controls, rows, menus, tabs. | interface | 13px / 16px / 400 |
| `type-label` | A small label, in sentence case: table header, field group, section index. | interface | 12px / 16px / 500 |
| `type-caption` | Secondary information under or beside something. | interface | 12px / 16px / 400 |
| `type-code` | Shortcuts, counts, paths, identifiers, code. | mono | 12px / 16px / 400 |

The display size grows with the width of the viewport, between its two ends.

| Typeface | Style variable | Family | Notes |
| --- | --- | --- | --- |
| Display | `font-serif` | IBM Plex Serif | Light italic, weight 300 |
| Interface | `font-sans` | IBM Plex Sans | 12 to 13px |
| Mono | `font-mono` | Lilex | |

The three families are under the SIL Open Font License 1.1 and are installed from npm packages. A shadcn/ui component sets its text in `text-sm`, which is pinned to the interface size: 13px with an 18px line.

### Spacing and control sizes

The base unit is 4px. Spacing steps are multiples of it.

| Size | Use |
| --- | --- |
| 24px | Chip, small badge row |
| 27 to 28px | Title bar, status bar, toolbar, row, small button |
| 32px | Default control |
| 36px | Prominent control |
| 12 to 16px | Icon |

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
| Floating | Popover, dialog, menu, tooltip, command palette, toast | The one shadow recipe, in `shadow-floating` |

`shadow-floating` is `0 0 0 1px rgb(0 0 0 / 0.06), 0 8px 24px -8px rgb(0 0 0 / 0.28)`, in both themes. It is derived: the reference shows no floating surface. `shadow-xs`, which shadcn/ui puts on fields, is set to none.

A button has no shadow. Its depth is a 2px inset bottom edge, drawn by the `button-edge` and `button-edge-primary` utilities.

Behind a dialog or a sheet, the page is covered by `background` at 70% opacity. A tooltip has no arrow.

### Motion and icons

- Colour and opacity: 100 to 150ms, `cubic-bezier(.4, 0, .2, 1)`.
- Overlays: 200 to 250ms, opening from scale .96 to .98 along `ease-overlay`, which is `cubic-bezier(.2, .8, .2, 1)`.
- With `prefers-reduced-motion`, no animation and no transition runs: every state is reached at once.
- Icons are Lucide, at 12 to 16px.

## Traits

What makes the style recognisable, beyond the values above.

| Trait | How it shows |
| --- | --- |
| Primary button | Blue fill, inset darker bottom edge, and a shortcut chip inside when the action has a shortcut |
| Panes | Flush, split by 1px lines. No gap and no card between them |
| Depth | No shadow inside the work surface |
| Page framing | A 10px grid texture and diamond nodes where lines meet. A page opts in with the `page-framing` utility, and places each node with `frame-node` |
| Density | 27 to 28px bars, 32px controls, 12 to 13px interface text |
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

`packages/ui` holds shadcn/ui components on Base UI primitives. A component enters when a screen needs it, through the shadcn CLI, and it appears in the gallery in the same change. A check fails when a component of `packages/ui` is missing from the gallery.

| Gallery section | Components |
| --- | --- |
| Actions | Button (with icon, with shortcut chip), Badge, Kbd |
| Inputs | Field, Label, Input, InputOTP, Textarea, Select, Checkbox, RadioGroup, Switch, InputGroup |
| Navigation | Tabs, Breadcrumb, Sidebar |
| Data display | Card, Table, Avatar, Separator |
| Feedback | Alert, Toast, Skeleton, Empty, Status indicator |
| Overlays | Dialog, DropdownMenu, Popover, Tooltip, Command, Sheet |

The status indicator is this project's own component: a sign and a word in one of the four status colours. The sign belongs to the meaning: a dashed circle for working, a triangle for needs attention, a check for done, a cross for error. It is plain in a row, and a badge on the subtle background of its colour where it stands alone.

InputGroup and Sheet are in the list because Command and Sidebar are built on them. Command is built on the `cmdk` library, which brings its own primitives.

A component is restyled through style variables and through its variants.

## Recipes

The sample work screen of the prototype is built from these four.

| Recipe | What it is |
| --- | --- |
| App shell | A title bar, sidebar navigation, a top bar with breadcrumb and a command palette trigger, a main area of flush panes, and a status bar. |
| List pane | A filter, a list of rows that each carry a status indicator, and an empty state. Selecting a row drives the detail pane. |
| Detail pane | A header with the title and the actions of the selected item, and tabs that switch its content. |
| Form dialog | A dialog with fields, one primary action and a cancel action. It reports a validation error on the field. |

A recipe is added when a second screen needs the same composition.

## Gallery

The gallery has one route per section, with side navigation, in this order: Foundations, Actions, Inputs, Navigation, Data display, Feedback, Overlays, Do and Don't, Recipes.

- **Foundations** shows every style variable: colour roles as surface and text pairs with the measured contrast, type roles as a specimen table, spacing, radius, elevation, motion and icons.
- **A component block** shows, in order: the name with one line on when to use it, the typical case in real content, the variants, the sizes, the icon forms, the states, and the edge cases (long label, empty, error, loading, disabled).
- **States** are pinned side by side without interaction: default, hover, focus, active, disabled, plus loading for Button and invalid for fields.
- **Do and Don't** renders the eight rules as pairs.
- The theme and the language are chosen in the prototype bar, which applies to every route (see Prototype in `CODING_STANDARDS.md`).
- Navigation and the sample content inside components are in the three supported languages. Explanatory text is in English.

## Known gaps

- The display type of the reference has a weight between 320 and 390. IBM Plex Serif on npm has no weight between 300 and 400, so display type is set at 300.
- The hover of the primary button is its fill at 90%. The reference does not show one.
- The reference shows no command palette and no product table. Those components follow the rules above, with nothing to match.
- Hover, focus and active states come from the reference's CSS. None was seen in motion.
- The reference's light theme was read from its site and from an HTML mock of the application, never from the real application.
