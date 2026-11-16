# Design system

How the product looks, and the rules every screen follows. This file is the source: the code in `packages/ui` implements it, and the gallery in `apps/prototype` renders it. When the two disagree with this file, this file is right and the code is fixed.

The final style is not chosen yet. Until it is, the design system carries three style variations, one per visual reference, and this file gives the values of each. The evidence behind every value is in [`docs/design/references.md`](docs/design/references.md). The work is tracked in `.scratch/design-system-prototype/`.

A change to a style variable, a component or a recipe updates this file in the same change.

## Vocabulary

**Style variable**:
A CSS custom property that holds one style value: a colour, a radius, a font, a spacing step, a shadow. Components read style variables and nothing else.

**Theme**:
Light or dark. Every style variable that changes with the theme has a value in each.

**Style variation**:
A named, complete set of values for the style variables, in both themes. There are three: `zed`, `figcraft` and `herdr`. Theme and style variation are independent. Style variations exist only until the final style is chosen.

**Final style**:
The single set of values the product ships with, picked or mixed by the maintainer from the three style variations.

**Gallery**:
The section of the prototype that shows every style variable and every component of `packages/ui`, in every variant and state. It is permanent: it stays after the final style is chosen, as the reference for every screen.

**Recipe**:
A named composition of components that several screens reuse, such as the app shell.

## Rules

These eight hold in every style variation and in the final style. The gallery renders each as a Do and Don't pair.

1. **Borders carry the structure.** Surfaces are separated by 1px lines. A shadow appears only on a floating surface: popover, dialog, command palette, toast.
2. **One accent, and colour means status.** At most one accent colour. Any other colour is a status colour with a fixed meaning: blue is working, amber needs attention, green is done, red is an error. A status colour is never decoration.
3. **Mono is for metadata.** Shortcuts, counts, paths and identifiers are set in the mono typeface. Running text is set in the interface typeface.
4. **Decoration frames the page.** Grid texture, corner ticks and section strips belong to the page framing. The inside of a component is plain.
5. **Motion is short and plain.** A colour change takes 100 to 150ms. A control moves at most 1 to 2px. An overlay opens in 200 to 250ms.
6. **Every value is a style variable.** A component takes each colour, radius, font, spacing step and shadow from a style variable. A trait that differs between style variations becomes a new style variable, so that component code is the same under all of them.
7. **One primary action per group.** In a group of actions, one button is primary. The others are secondary or ghost.
8. **Status is colour plus a sign.** A status always shows an icon or a word together with its colour.

## Accessibility and layout

- The final style meets WCAG AA: 4.5:1 for text, 3:1 for large text, for the boundary of a control and for the focus indicator.
- A style variation keeps the values of its reference even when a pair of colours fails. The gallery measures every pair and marks each failure, so the failure informs the choice of the final style.
- Every interactive component is operable by keyboard and shows a visible focus state.
- Motion respects `prefers-reduced-motion`.
- Layouts are designed for desktop first and adapt down to a 375px-wide viewport with no overflow.
- Text runs longer in Portuguese and Spanish than in English. Every component that carries text holds a long label without overflowing.

## Style variables

The names and roles below are the contract. They are the same in every style variation; only the values change. The colour names are those of shadcn/ui, with the additions marked.

### Colour roles

| Style variable | Role |
| --- | --- |
| `background` / `foreground` | The page, behind and between panes. Default text. |
| `card` / `card-foreground` | A pane or a card. |
| `sidebar` / `sidebar-foreground` | Chrome: sidebar, top bar, status bar. |
| `popover` / `popover-foreground` | A floating surface. |
| `primary` / `primary-foreground` | The primary button. |
| `secondary` / `secondary-foreground` | A raised control: secondary button, active tab, chip. |
| `muted` / `muted-foreground` | A sunken or quiet fill. Secondary text. |
| `accent` / `accent-foreground` | The hover wash. This is shadcn/ui's name; it is not the accent colour of rule 2. |
| `selected` (added) | The fill of a selected row or item, one step stronger than `accent`. |
| `faint-foreground` (added) | Tertiary text: placeholder, disabled text, an index. |
| `border`, `input` | The hairline between surfaces. The border of a control. |
| `border-strong` (added) | A stronger outline, for an outlined button or an emphasised edge. |
| `ring` | The focus indicator. |
| `brand` / `brand-foreground` (added) | The one accent colour of rule 2. |
| `destructive` | A destructive action and a validation error. |
| `status-working`, `status-attention`, `status-done`, `status-error` (added) | The four status colours of rule 2. |
| `status-*-subtle` (added) | The background of a status badge: the status colour at 10 to 15% opacity. |

The four status colours are visual meanings. Which status of a record shows which meaning is decided with each screen.

### Type roles

Each role is a `type-*` utility that sets family, size, line height, weight and tracking together.

| Token | Use | Typeface |
| --- | --- | --- |
| `type-display` | A page title, the headline of an empty state. Used sparingly. | display |
| `type-title` | The title of a section, a card or a dialog. | interface |
| `type-body` | Running text and descriptions. | interface |
| `type-ui` | Controls, rows, menus, tabs. | interface |
| `type-label` | A small label: table header, field group, section index. | per style variation |
| `type-caption` | Secondary information under or beside something. | interface |
| `type-code` | Shortcuts, counts, paths, identifiers, code. | mono |

`type-label` takes its typeface, casing and tracking from the style variables `label-font`, `label-transform` and `label-tracking`.

### Spacing and control sizes

The base unit is 4px. Spacing steps are multiples of it.

| Size | Use |
| --- | --- |
| 24px | Chip, small badge row |
| 28 to 30px | Row, small button, toolbar, status bar |
| 32px | Default control |
| 36px | Prominent control |
| 12 to 16px | Icon |

Each style variation sets its own heights inside this scale (see Traits). Button, input, select and tab heights move together, so a form stays aligned.

### Radius

| Role | Used by |
| --- | --- |
| `radius-chip` | Badge, shortcut chip |
| `radius-tab` | Tab, menu item |
| `radius-control` | Button, input, select |
| `radius-pane` | Pane, card |
| `radius-floating` | Dialog, popover, command palette |

### Elevation

| Level | Surfaces | Treatment |
| --- | --- | --- |
| Flat | Page, chrome | None |
| Hairline | Pane, card, control | 1px border |
| Floating | Popover, dialog, command palette, toast | The one shadow recipe of the style variation, in `shadow-floating` |

### Trait variables

These exist because a trait differs between style variations (rule 6).

| Style variable | Trait |
| --- | --- |
| `label-font`, `label-transform`, `label-tracking` | Small labels: sentence case, or uppercase mono with wide tracking |
| `button-edge` | The inset bottom edge of a button, or none |
| `pane-gap`, `pane-radius`, `pane-border` | Panes flush and split by lines, or separate cards with a gap |
| `pane-focus-border` | The border of the focused pane |
| `page-decoration` | The page framing: grid and nodes, ticks and strips, or none |

### Motion and icons

- Colour and opacity: 100 to 150ms, `cubic-bezier(.4, 0, .2, 1)`.
- Overlays: 200 to 250ms, `cubic-bezier(.2, .8, .2, 1)`, opening from scale .96 to .98.
- Icons are Lucide, at 12 to 16px.

## Style variations

Hex values come from the references. A value marked *derived* has no evidence in its reference and was worked out for this project.

### Traits

| Trait | `zed` | `figcraft` | `herdr` |
| --- | --- | --- | --- |
| Model | Zed's site and application | ArtCraft's site, not the FigCraft application | Herdr's application, not its site |
| Primary button | Blue fill, inset darker bottom edge, shortcut chip inside | Inverted ink block, uppercase mono label | Inverted neutral: the text colour as fill |
| Radius | Chip 2px; tab, control and pane 4px; floating 6px | 0 everywhere | Chip 5px, tab 7px, control 8px, pane 10px, floating 14px |
| Display typeface | IBM Plex Serif, light italic, weight 320 to 390 | Archivo, weight 620 at 118% width, tracking -0.035em | Geist, weight 560, tracking -0.035em |
| Interface typeface | IBM Plex Sans, 12 to 13px | Inter | Geist, 13px |
| Mono typeface | Lilex | Geist Mono | JetBrains Mono |
| Small labels | Sentence case | Mono, 11px, weight 500, uppercase, tracking +0.15em | Sentence case |
| Panes | Flush, split by 1px lines | Flush, in a hairline grid | Separate cards with 8px gaps on a darker page; the focused pane takes the accent border |
| Depth | No shadow inside the work surface; inset bottom edge on buttons | None | 1px light ring plus a large soft shadow, on floating surfaces only |
| Decoration | 10px grid texture and diamond nodes | Crosshair corner ticks and numbered section strips | None |
| Density | 27 to 28px bars, 32px controls | 36px controls, 11px labels | 30px rows, 26px icon buttons |
| Focus indicator | 3px outline, 2px offset | 2px outline, 2px offset | The accent colour; on a pane, as its border |

`figcraft` follows a marketing site with large type and roomy cells. It keeps that character on the compact control scale above, because this is a work tool.

### Colours of `zed`

| Style variable | Light | Dark |
| --- | --- | --- |
| `background` | `#f5f4f3` | `#282c33` |
| `card` | `#fafafa` | `#282c33` |
| `sidebar` | `#ebebec`; title and status bars `#dcdcdd` | `#2f343e`; title and status bars `#3b414d` |
| `secondary` | `#ebebec` | `#2e343e` |
| `accent` | `#dfdfe0` | `#363c46` |
| `selected` | `#cacaca` | `#454a56` |
| `foreground` | `#242529` | `#dce0e5` |
| `muted-foreground` | `#58585a` | `#a9afbc` |
| `faint-foreground` | `#7e8086` | `#878a98` |
| `border` | `#c9c9ca`; subtle `#dfdfe0` | `#464b57`; subtle `#363c46` |
| `primary`, `brand` | `#1348dc` | `#1348dc`; as text `#74ade8` |
| `primary-foreground` | `#ffffff` | `#ffffff` |
| `ring` | `#094ece` at 50% | `#094ece` at 50% |
| `button-edge` | `hsl(219 93% 30%)`, 2px inset, on the primary button; `hsl(218 13% 50% / 0.1)` on the others | the same |

### Colours of `figcraft`

Neutrals are one ink colour at fixed opacities, with no gray ramp.

| Style variable | Light | Dark |
| --- | --- | --- |
| `background`, `sidebar` | `#f2f1ee` | `#121316` |
| `card`, `popover` | `#faf9f7` | `#101014` |
| `muted`, `accent` | `#e9e8e4` | `#060607` |
| `foreground` | `#101014` | `#f2f1ee` |
| `muted-foreground` | ink at 62% | ink at 60% |
| `faint-foreground` | ink at 42% | ink at 38% |
| `border` | ink at 16% | ink at 15% |
| `border-strong` | ink at 40% | ink at 40% |
| `primary` / `primary-foreground` | `#101014` / `#f2f1ee` | `#ffffff` / `#0b0b0c` |
| `brand` | `#02da96`; as text `#017a54` | `#02da96`; as text `#3fe6ae` |
| `ring` | `#1659c4` | `#74aaff` |
| `destructive` | `#b3261e` | `#ff6259` |

The green highlights one element at a time. It is never the button colour. An outlined button inverts to the ink fill on hover, in 120ms.

### Colours of `herdr`

The reference has no official light theme. The light column is *derived* from its "Board" sample look, and the gallery says so.

| Style variable | Light (*derived*) | Dark |
| --- | --- | --- |
| `background` | `#f6f7f9` | `#131518` |
| `sidebar` | set when built | `#101114` |
| `card` | `#ffffff` | `#17191d` |
| `secondary` | set when built | `#2c2e31` |
| `accent` | set when built | white at 5.5% |
| `selected` | set when built | white at 9% |
| `border` | `#e5e7ec` | white at 9.4% |
| `foreground` | `#252b37` | `#eeeeef` |
| `muted-foreground` | `#89909d` | `#98999b` |
| `faint-foreground` | set when built | `#696a6d` |
| `primary` / `primary-foreground` | `foreground` / `background` | `foreground` / `background` |
| `brand`, `ring`, `pane-focus-border` | set when built | `#96b2d7` |
| `shadow-floating` | set when built | `0 0 0 1px rgb(255 255 255 / .08), 0 40px 120px -30px rgb(0 0 0 / .6)` |

There is no brand colour: the accent is used for focus only, and colour is reserved for status.

### Status colours

Shared by the three style variations.

| Style variable | Light | Dark |
| --- | --- | --- |
| `status-working` | `#5c78e2` | `#76a8ff` |
| `status-attention` | `#a48819` | `#f2a33c` |
| `status-done` | `#669f59` | `#4cc48c` |
| `status-error` | `#d36151` | `#d07277` |

The light values and the dark error are Zed's; the other dark values are Herdr's.

## Components

`packages/ui` holds shadcn/ui components on Base UI primitives. A component enters when a screen needs it, through the shadcn CLI, and it appears in the gallery in the same change. A check fails when a component of `packages/ui` is missing from the gallery.

| Gallery section | Components |
| --- | --- |
| Actions | Button (with icon, with shortcut chip), Badge, Kbd |
| Inputs | Field, Label, Input, Textarea, Select, Checkbox, RadioGroup, Switch |
| Navigation | Tabs, Breadcrumb, Sidebar |
| Data display | Card, Table, Avatar, Separator |
| Feedback | Alert, Toast, Skeleton, Empty, Status indicator |
| Overlays | Dialog, DropdownMenu, Popover, Tooltip, Command |

The status indicator is this project's own component: a sign and a word in one of the four status colours.

A component is restyled through style variables and through its variants. Its code never reads the name of a style variation.

## Recipes

The sample work screen of the prototype is built from these four.

| Recipe | What it is |
| --- | --- |
| App shell | Sidebar navigation, a top bar with breadcrumb and a command palette trigger, a main area of panes, and a status bar. The pane arrangement follows `pane-gap`, `pane-radius` and `pane-border`. |
| List pane | A filter, a list of rows that each carry a status indicator, and an empty state. Selecting a row drives the detail pane. |
| Detail pane | A header with the title and the actions of the selected item, and tabs that switch its content. |
| Form dialog | A dialog with fields, one primary action and a cancel action. It reports a validation error on the field. |

A recipe is added when a second screen needs the same composition.

## Gallery

The gallery has one route per section, with side navigation, in this order: Foundations, Actions, Inputs, Navigation, Data display, Feedback, Overlays, Do and Don't, Recipes, Compare.

- **Foundations** shows every style variable: colour roles as surface and text pairs with the measured contrast, type roles as a specimen table, spacing, radius, elevation, motion and icons.
- **A component block** shows, in order: the name with one line on when to use it, the typical case in real content, the variants, the sizes, the icon forms, the states, and the edge cases (long label, empty, error, loading, disabled).
- **States** are pinned side by side without interaction: default, hover, focus, active, disabled, plus loading for Button and invalid for fields.
- **Do and Don't** renders the eight rules as pairs.
- **Compare** shows the three style variations in columns over the same examples. It is removed with the style variations when the final style is built.
- The theme selector and the style variation selector apply to every route.
- Navigation and the sample content inside components are in the three supported languages. Explanatory text is in English.

## Known gaps

- The size, line height and weight of each type role are set per style variation when it is built, inside the ranges in Traits, and recorded here.
- The values marked "set when built" in the `herdr` light theme are derived when that style variation is built, and recorded here.
- The references show no command palette and no product table. Those components follow the rules above, with no reference to match.
- Hover, focus and active states come from the references' CSS. None was seen in motion.
