# Design system

How the product looks, and the rules every screen follows. This file is the source: the code in `apps/web/src/ui` implements it. When the code disagrees with this file, this file is right and the code is fixed.

The style follows monday.com and its design system, Vibe, adjusted to WCAG AA. The evidence behind every value is in section 4 of [`docs/design/references.md`](docs/design/references.md), and the other styles that were compared are in its section 5. Why monday.com is in ADR-0027. Why the design system lives in the app, with no prototype in the repository and no gallery, is in ADR-0025.

The size of the text does not follow Vibe. Vibe sets its interface in 14px; here it is 16px, and every control is taller to hold it. The maintainer found smaller text hard to read, and chose 16px from sample pages (PR #7).

A change to a style variable, a component or a recipe updates this file in the same change.

## Vocabulary

**Style variable**:
A CSS custom property that holds one style value: a colour, a radius, a font, a spacing step, a shadow. Components read style variables and nothing else.

**Theme**:
Light or dark. Every style variable that changes with the theme has a value in each.

**Frame**:
The blue-grey ground of the app, on `background`. The header and the sidebar sit on it.

**Work surface**:
The white panel on `card` that holds a screen, inset in the frame with rounded corners. It scrolls inside itself.

**Label**:
A filled block of colour with a word in it, that shows the value of a status or a category: "Done", "Stuck", "Consultancy". In a table it fills its cell from edge to edge.

**Group**:
A set of rows of a table under one title, such as the tasks of a sprint. The group has a colour, which its title and a strip on the left edge of its rows take.

**Recipe**:
A named composition of components that several screens reuse, such as the app shell.

## Rules

1. **The work surface sits on the frame.** The header and the sidebar are on the frame; a screen is on the work surface. Inside the work surface, 1px lines separate rows, columns and sections. A shadow appears only on a floating surface: popover, menu, dialog, side panel, command palette, toast.
2. **One colour for action.** Blue, `primary`, marks the primary action, a link, the focus and the selected item. No other colour is used for an action.
3. **Colour labels a value.** A status or a category shows as a label, whose colour comes from the label palette. Each screen decides which value takes which colour, and the same value keeps its colour on every screen. Feedback (an alert, a toast, a problem in a field) uses the four feedback colours instead.
4. **Groups carry a colour.** The title of a group and the strip on the left edge of its rows share one group colour. Colour never decorates anything else.
5. **Headings in Poppins, everything else in Figtree.** A page title, the title of a card, of a dialog and of a group are set in Poppins. Running text, controls, tables and labels are set in Figtree. Mono is only for what is copied or compared character by character: codes, identifiers, secrets, addresses.
6. **Motion is short and plain.** A colour change takes 100 to 150ms. An overlay opens in 200 to 250ms. Nothing bounces.
7. **Every value is a style variable.** A component takes each colour, radius, font, spacing step and shadow from a style variable, and hardcodes none.
8. **One primary action per group.** In a group of actions, one button is primary. The others are outline or ghost.
9. **Status is colour plus a word.** A label always holds its word; a feedback colour always comes with an icon or a word.

One exception to rules 2 and 3: the mark of a login provider keeps the colours its owner requires. Google forbids a single-colour "G" and asks for it on white, and Microsoft asks that its logo is not altered. GitHub's mark takes the colour of the text.

## Accessibility and layout

- The style meets WCAG AA: 4.5:1 for text, 3:1 for large text, for the boundary of a control and for the focus indicator.
- A value taken from Vibe that fails is changed by the smallest step that passes. The colour tables then give both the value of Vibe and the value in use. Nothing measures the pairs automatically: a value is measured when it is added or changed.
- A label takes the ink, `#323338`, or white, whichever reaches 4.5:1 on its fill. The label palette lists the one each colour takes.
- Every interactive component is operable by keyboard and shows a visible focus state: a 3px outline with a 2px offset, in `ring`. A box of InputOTP has no offset, because its neighbours touch it. A row of a table and a control in a bar draw the outline inside themselves, because the table or the bar would cut it.
- Motion respects `prefers-reduced-motion`.
- Layouts are designed for desktop first and adapt down to a 375px-wide viewport with no overflow. A wide table scrolls sideways inside its own container.
- Text runs longer in Portuguese and Spanish than in English. Every component that carries text holds a long label without overflowing.

## Style variables

The names and roles below are the contract. The colour names are those of shadcn/ui, with the additions marked.

### Colour roles

| Style variable | Role |
| --- | --- |
| `background` / `foreground` | The frame. Default text. |
| `card` / `card-foreground` | The work surface, a card or a pane inside it. |
| `sidebar` / `sidebar-foreground` | The sidebar, on the frame. |
| `popover` / `popover-foreground` | A floating surface. |
| `primary` / `primary-foreground` | The primary button. |
| `primary-hover` (added) | The primary button under the pointer. |
| `secondary` / `secondary-foreground` | A quiet filled control: a secondary button, the active tab of a default tab list, a chip. |
| `muted` / `muted-foreground` | A sunken or quiet fill, such as the footer of a group. Secondary text. |
| `accent` / `accent-foreground` | The hover wash. This is shadcn/ui's name; it is not the colour of rule 2. |
| `selected` (added) | The fill of a selected row or item. |
| `faint-foreground` (added) | Placeholder and disabled text. Vibe has no third step of text, so it is `muted-foreground`. |
| `border` | The line between areas of the work surface, and the edge of a floating surface. |
| `input` | The border of a field and of an outline button. It reaches 3:1 against the surface, which `border` does not have to. |
| `border-subtle` (added) | A lighter line inside a card, between rows of a list. |
| `ring` | The focus indicator. |
| `brand` / `brand-foreground` (added) | The colour of rule 2 where it is not a button: a link, an icon, the mark. |
| `destructive` | A destructive action and a validation error. |
| `status-working`, `status-attention`, `status-done`, `status-error` (added) | The four feedback colours of rule 3: blue informs or is in progress, amber needs attention, green is done, red is an error. |
| `status-*-subtle` (added) | The background of a feedback badge or alert. |
| `group-blue`, `group-green`, `group-purple`, `group-orange`, `group-gray`, `group-red` (added) | The title of a group. Its strip takes the label colour of the same name. |
| `label-*` and `label-*-foreground` (added) | The label palette: the fill of a label and the text on it. The same in both themes. |
| `scrim` (added) | What covers the page behind a dialog or a side panel. |

### Colours

The values in use. Where a value of Vibe failed WCAG AA, the value in use is followed by the one it replaces.

| Style variable | Light | Dark |
| --- | --- | --- |
| `background` | `#eceff8` | `#181b34` |
| `card` | `#ffffff` | `#30324e` |
| `popover` | `#ffffff` | `#30324e` |
| `sidebar` | `#eceff8` | `#181b34` |
| `secondary` | `#f6f7fb` | `#292b45` |
| `muted` | `#f6f7fb` | `#292b45` |
| `accent` | `#676879` at 10% | `#676879` at 20% (Vibe: at 30%) |
| `selected` | `#e5f0ff` (Vibe: `#cce5ff`) | `#133774` |
| `foreground` | `#323338` | `#d5d8df` |
| `muted-foreground` | `#626370` (Vibe: `#676879`) | `#a6a9b6` (Vibe: `#9699a6`) |
| `faint-foreground` | `#626370` (Vibe: `#676879`) | `#a6a9b6` (Vibe: `#9699a6`) |
| `border` | `#d0d4e4` | `#4b4e69` |
| `input` | `#7f8292` (Vibe: `#c3c6d4`) | `#8a8ea3` (Vibe: `#797e93`) |
| `border-subtle` | `#e6e9ef` | `#3d405c` |
| `primary` | `#0073ea` | `#0073ea` |
| `primary-hover` | `#0060b9` | `#0060b9` |
| `primary-foreground` | `#ffffff` | `#ffffff` |
| `brand` | `#0060b9` (Vibe: `#1f76c2`) | `#7db3f2` (Vibe: `#69a7ef`) |
| `brand-foreground` | `#ffffff` | `#181b34` |
| `ring` | `#0073ea` | `#69a7ef` |
| `destructive` | `#c4213b` (Vibe: `#d83a52`) | `#ff8591` (Vibe: `#d83a52`) |
| `scrim` | `#292f4c` at 45% | black at 55% |

The text colours of a surface (`card-foreground`, `popover-foreground`, `secondary-foreground`, `accent-foreground`, `sidebar-foreground`) are `foreground`. The `sidebar-*` variables that shadcn/ui expects take the values of their plain counterparts.

What each adjustment had to reach:

- `muted-foreground`: 4.5:1 on the frame under the hover wash, where the sidebar shows counts. It was 4.23:1 in light. In dark, 4.5:1 on the work surface, where it was 4.38:1.
- `accent` in dark: at 30% the wash took `muted-foreground`, `brand` and the red below 4.5:1. At 20% they pass.
- `selected` in light: `#cce5ff` takes `muted-foreground` to 4.23:1. `#e5f0ff` keeps it at 4.76:1.
- `input`: 3:1 on the work surface and on the frame. The line of Vibe is 1.7:1, which is enough between areas and not for the border of a field.
- `brand`: 4.5:1 as a link on every fill of the work surface, including the hover wash and a selected row.
- `destructive` and the feedback colours: 4.5:1 as a word on the work surface, on the frame and on their subtle background.

`selected` is Vibe's `primary-selected`; `accent` is its hover colour; `muted` is its grey background.

### Feedback colours

| Style variable | Light | Dark |
| --- | --- | --- |
| `status-working` | `#0060b9` (Vibe: `#0073ea`) | `#7db3f2` (Vibe: `#0073ea`) |
| `status-attention` | `#965500` (Vibe: `#ffcb00`) | `#fdab3d` (Vibe: `#ffcb00`) |
| `status-done` | `#007038` (Vibe: `#00854d`) | `#33d391` (Vibe: `#00854d`) |
| `status-error` | `#c4213b` (Vibe: `#d83a52`) | `#ff8591` (Vibe: `#d83a52`) |
| `status-working-subtle` | `#e5f0ff` | `#1d3566` |
| `status-attention-subtle` | `#fff4d6` | `#4a3a1a` |
| `status-done-subtle` | `#e3f7ec` | `#173d33` |
| `status-error-subtle` | `#fde8eb` | `#4a2030` |

`destructive` and `status-error` are the same red. Vibe's yellow `#ffcb00` reaches 1.5:1 as a word on white, so the word of attention is a dark amber; the yellow stays in the label palette, behind ink.

### Label palette

The same in both themes. A screen picks a colour for each value it labels. Monday's own meanings are a good default: green is done, orange is in progress, red is stuck, grey is not started.

| Colour | Fill | Text | Contrast |
| --- | --- | --- | --- |
| `label-green` | `#00c875` | ink | 5.71:1 |
| `label-orange` | `#fdab3d` | ink | 6.64:1 |
| `label-yellow` | `#ffcb00` | ink | 8.28:1 |
| `label-red` | `#df2f4a` | white | 4.52:1 |
| `label-purple` | `#9d50dd` | white | 4.54:1 |
| `label-blue` | `#5c9ffc` (Vibe: `#579bfc`) | ink | 4.69:1 |
| `label-gray` | `#c4c4c4` | ink | 7.23:1 |
| `label-royal` | `#784bd1` | white | 5.64:1 |
| `label-indigo` | `#5559df` | white | 5.41:1 |
| `label-dark-purple` | `#401694` | white | 11.93:1 |
| `label-navy` | `#225091` | white | 8.00:1 |
| `label-grass` | `#037f4c` | white | 5.07:1 |

The ink is `#323338`, the `foreground` of the light theme. `label-blue` was moved one step, because Vibe's blue reaches 4.50:1 with the ink only after rounding.

### Group colours

The strip takes the label colour; the title takes the shade below, which reaches 4.5:1 on the work surface.

| Group | Strip | Title, light | Title, dark |
| --- | --- | --- | --- |
| Blue | `label-blue` | `#0060b9` | `#7db3f2` |
| Green | `label-green` | `#007038` | `#33d391` |
| Purple | `label-purple` | `#7a3bb8` | `#c78ef5` |
| Orange | `label-orange` | `#965500` | `#fdab3d` |
| Grey | `label-gray` | `#626370` | `#a6a9b6` |
| Red | `label-red` | `#c4213b` | `#ff8591` |

### Type roles

Each role is a `type-*` utility that sets family, size, line height, weight and tracking together.

| Token | Use | Typeface | Size / line height / weight |
| --- | --- | --- | --- |
| `type-display` | A page title, the name on a record page, the headline of an empty state. | heading | 28 to 32px / 1.2 / 500, tracking -0.01em |
| `type-title` | The title of a section, a card, a group or a dialog. | heading | 18px / 26px / 500 |
| `type-body` | Running text and descriptions. | interface | 16px / 24px / 400 |
| `type-ui` | Controls, rows, menus, tabs. | interface | 16px / 20px / 400 |
| `type-label` | A small label, in sentence case: table header, field group, section index. | interface | 14px / 20px / 500 |
| `type-caption` | Secondary information under or beside something. | interface | 14px / 20px / 400 |
| `type-code` | Codes, identifiers, secrets, addresses. | mono | 14px / 20px / 400 |

The display size grows with the width of the viewport, between its two ends.

| Typeface | Style variable | Family | Notes |
| --- | --- | --- | --- |
| Heading | `font-heading` | Poppins | Weights 500 and 600 |
| Interface | `font-sans` | Figtree | Variable, 14 to 18px |
| Mono | `font-mono` | JetBrains Mono | Variable. Vibe defines no mono |

The three families are under the SIL Open Font License 1.1 and are installed from npm packages. A shadcn/ui component sets its text in `text-sm`, which is pinned to the interface size, 16px with a 24px line, and its smaller text in `text-xs`, which is pinned to the size of a label, 14px with a 20px line.

### Spacing and control sizes

The base unit is 4px. Spacing steps are multiples of it.

| Size | Use |
| --- | --- |
| 24px | Small label, shortcut chip |
| 28px | Badge, extra-small button, a control inside a bar |
| 32px | Small button, an item of a menu |
| 36px | Search field of the header, a tool of a toolbar |
| 40px | Default control, item of the sidebar, header of a table |
| 44px | Row of a table, prominent control |
| 56px | The header of the app |
| 960px | `breakpoint-shell`: under it the sidebar of the app shell is a sheet |
| 16 to 20px | Icon: 20px in the header and the sidebar, 16px in a control, 14px inside a chip or a badge |

Button, input, select and tab heights move together, so a form stays aligned.

### Radius

| Role | Value | Used by |
| --- | --- | --- |
| `radius-chip` | 4px | Label, badge, shortcut chip, checkbox |
| `radius-tab` | 4px | Tab, menu item, item of the sidebar |
| `radius-control` | 4px | Button, input, select |
| `radius-pane` | 8px | Card, a pane inside the work surface |
| `radius-floating` | 8px | Popover, menu, tooltip, command palette, toast |
| `radius-modal` | 16px | Dialog |
| `radius-surface` | 12px | The work surface, the pane of the framed pane |

### Elevation

| Level | Surfaces | Treatment |
| --- | --- | --- |
| Frame | Header, sidebar | None |
| Surface | The work surface | Its own colour against the frame, and a 1px `border` where it meets the frame |
| Hairline | Card, pane, control, table | 1px line |
| Floating | Popover, menu, tooltip, command palette, toast, side panel | `shadow-floating` |
| Modal | Dialog | `shadow-modal` |

`shadow-floating` is `0 0 0 1px var(--border), 0 6px 20px rgb(0 0 0 / 0.2)`, Vibe's medium shadow with an edge. `shadow-modal` is `0 0 0 1px var(--border), 0 15px 50px rgb(0 0 0 / 0.3)`, its large one. The edge is a line in `border`, so a floating surface on `card` shows in dark, where both are `#30324e`. `shadow-xs`, which shadcn/ui puts on fields, is set to none.

A button has no shadow. Behind a dialog or a side panel, the page is covered by `scrim`. A tooltip has no arrow.

### Motion and icons

- Colour and opacity: 100 to 150ms, `cubic-bezier(.4, 0, .2, 1)`.
- Overlays: 200 to 250ms, opening from scale .96 to .98 along `ease-overlay`, which is `cubic-bezier(.2, .8, .2, 1)`. A side panel slides in from its edge.
- With `prefers-reduced-motion`, no animation and no transition runs: every state is reached at once.
- Icons are Lucide, at 14 to 20px.

## Traits

What makes the style recognisable, beyond the values above.

| Trait | How it shows |
| --- | --- |
| Frame and work surface | A blue-grey frame holds the header and the sidebar; the screen is a white panel inset in it, with 12px corners |
| Labels | A status fills its cell from edge to edge, in a vivid colour, with its word centred |
| Groups | Rows grouped under a coloured title, with a strip of the same colour on their left edge, and sums and a bar of statuses at the foot of each group |
| Primary button | Blue fill that darkens under the pointer; a split button, "New item" plus a chevron, opens the creation of a screen |
| Headings | Poppins at weight 500 over Figtree |
| People | Coloured circles with initials |

## Brand mark

The product is ForgeSpark, one word with two capitals, and its slogan is "From ideas to execution.", with its full stop. Both stay in English in every language of the app: they are the brand, not text to translate.

The mark is a nut with a spark: a hexagon with rounded corners, flat at the top and the bottom like a nut or a forged part, with a four-pointed spark cut out of its middle. The part is the execution and the spark the idea. It is drawn on a 64-unit square: the hexagon has a radius of 30 units and corners rounded by 6, and the spark is 26 units wide and 34 high.

- The mark has one colour: `brand` on a surface, or `primary-foreground` on a `primary` tile. It follows rule 2.
- The logo is the mark and the product name. In `logo.svg` the name is drawn in Poppins at weight 600, in `foreground`, its capitals 45% as high as the mark. In the app the name is set in title type (`type-title`), in `foreground`, beside the mark in `brand`: `ProductLogo` in `apps/web/src/shared/components/brand-mark.tsx`, whose mark is a copy of `mark.svg`, draws both, in the header of the framed pane and of the shell.
- The slogan shows on the screens seen before a login: in the header of the framed pane, under the product name, in caption type (`type-caption`), in `muted-foreground`. The shell shows the name alone, because the work needs no slogan.
- The smallest size of the mark is 16px, where the spark closes to a round hole; from 32px it reads as a spark. The space around it is at least a quarter of its width.
- The mark is never stretched, rotated, outlined or given a shadow.

The files are in [`docs/design/brand/`](docs/design/brand/). They are the source: the web app keeps no copy of them. `apps/web/index.html` names the favicon, the icon of iOS and the web app manifest, and `apps/web/brand-files.ts` serves them at the root of the origin, read from here, with `manifest.webmanifest`, which names the app, its slogan and its three icons.

| File | Use |
| --- | --- |
| `mark.svg`, `mark-on-dark.svg` | The mark alone, on a light and on a dark surface |
| `logo.svg`, `logo-on-dark.svg` | The mark and the product name, on a light and on a dark surface |
| `favicon.svg` | The browser tab. It changes colour with the theme of the browser |
| `favicon.ico` | The browser tab where an SVG is not accepted: 16, 32 and 48px |
| `apple-touch-icon.png` | The home screen of iOS: 180px, the mark in white on a square of `primary`, which iOS rounds |
| `icon-192.png`, `icon-512.png` | The icons of the web app manifest: the mark in white on a tile of `primary` with corners of an eighth of its width |
| `icon-maskable-512.png` | The maskable icon of the web app manifest: the mark in white on `primary` to the edges, inside the safe area |
| `social-preview.png` | The preview of a shared link and of the repository: `logo-on-dark.svg` and the slogan in Figtree, in `muted-foreground` of the dark theme, over the background, 1280 by 640px |
| `source/` | What an image model returned for the background of the preview |

The mark was drawn by hand on the 64-unit square; the background of the preview came from an image model. The PNG files and `favicon.ico` are drawn from the path of `mark.svg`, in `primary` and in `brand`, the colours of the current style.

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

The status indicator is this project's own component, for feedback: a sign in one of the four feedback colours, and a word. The sign belongs to the meaning: a dashed circle for working, a triangle for needs attention, a check for done, a cross for error. It is plain beside text, where the sign carries the colour and the word is set in `foreground`. It is a badge where it stands alone, with the sign and the word in the colour, on its subtle background. The label of a record's status is not a status indicator: it comes with the first table that shows one (see Board table).

InputGroup and Sheet are in the list because Command and Sidebar are built on them. Command is built on the `cmdk` library, which brings its own primitives.

Combobox is a field whose value is chosen from a long list filtered by typing, such as a country: Select is for a short list read whole. Its list floats like a menu. The input is its one stop of the keyboard, and the arrow keys open the list from it; the button beside it is for the pointer and is out of the order of the Tab key. That button, the one that clears the field and the one that removes a chip are icons alone, so each requires its name as a prop, as the close button of a dialog does: `triggerLabel` and `clearLabel` on `ComboboxInput`, `removeLabel` on `ComboboxChip`. `ListField` names the button after its field ("Country: show the list").

A component is restyled through style variables and through its variants. The Button variants are `default` (primary), `outline` (a border in `input`, the secondary action), `secondary` (a quiet fill), `ghost`, `destructive` and `link`.

A component holds no text of its own. Every word comes from the screen, in the language of the person, and so does a name that only assistive technology reads: the label of the close button of a dialog, a sheet or a toast, of the buttons of a combobox, the name of the sidebar when it is a sheet and of the button that toggles it, the title and the description of the command palette, the label of a breadcrumb. A component that needs such a name requires it as a prop.

## Recipes

Seven compositions are described here. Two have code, in `apps/web/src/shared/components`: the framed pane, `FramedPane`, with `StepActions` for the end of a step, which the setup and the login screen use; and the app shell, `AppShell`, which holds every screen after the login. The others have none yet: the first screen that needs one writes it in `apps/web`, following the clickable prototype the maintainer approved on 2026-11-18 (ADR-0027).

| Recipe | What it is |
| --- | --- |
| App shell | The header and the sidebar on the frame, and the work surface beside the sidebar. |
| Board table | A table of rows in coloured groups, with labels in its status columns and sums at the foot of each group. |
| Record page | The page of one record, such as a client or a project: a header, a row of figures, and view tabs. |
| Side panel | The detail of an item opened from a table, sliding in from the right, over the screen. |
| Form page | A long form on its own page, in sections. |
| Form dialog | A dialog with fields, one primary action and a cancel action. It reports a validation error on the field. |
| Framed pane | The header with the mark, the product name and the slogan, the frame with one pane in its centre, and a footer that holds the language selector and nothing else. The screens seen before a login: the setup and the login screen. |

A recipe is added when a second screen needs the same composition.

**App shell.** The header is 56px high, on the frame: the mark and the product name at its start, then the search field (Ctrl K opens the command palette), the notifications, help and the user menu at its end. The sidebar is 256px wide, on the frame: the consultancy at its top, then the items of the menu, 40px high with a 20px icon, then favourites, and the settings, with a gear, at its foot. An item with a submenu opens it in place. The selected item takes `selected`. The work surface fills the rest, less 12px at its right and bottom edges, with 12px corners. The shell takes the height it is given and never scrolls: the work surface scrolls inside itself.

- The user menu shows the name and the e-mail, then the actions on the session. Logging out is the last of them, and it leads to the login screen.
- Under 960px the sidebar is a sheet that a button at the start of the header opens, and the search field gives way to its icon.
- The command palette lists the places to go, the actions of the screen and the items to open.

The code, `AppShell`, draws the recipe cut down to what works, because a control that does nothing is not shown: the header has the mark, the product name and the user menu; the sidebar has the consultancy alone; the work surface holds the screen. The search field and the command palette, the notifications, help, the items of the menu, favourites and the settings are left out until the feature each one opens exists: there is no area of work, no notification and no setting yet. Each arrives with its feature, in the place this recipe gives it.

- The consultancy at the top of the sidebar is its logo, or its initials, in a 32px square with 4px corners, where the whole logo shows, beside its name.
- The button of the user menu is the photo of the user, or their initials, in a 32px circle. The menu shows them again at 40px, with the name and the e-mail, and one action, logging out, which leads to the login screen with its notice. The menu stays open while the logout is sent, and shows its failure.
- Initials are the first letters of the first and the last words of the name. They take the `muted` fill for now: the colour of a person's circle comes with the first screen that tells people apart, such as the board table.
- The photo and the logo come from the file storage, through URLs that work for 15 minutes; the app reads the user and the consultancy again every ten minutes, and keeps showing an image until the picture of its new URL has loaded, so an image never breaks nor blinks to the initials.
- The sidebar lies on the frame, below the header, with no line of its own. The work surface keeps 12px from the edge of the window on its left too, while the sidebar is a sheet.
- The width at which the sidebar becomes a sheet is `--breakpoint-shell`, 960px, whose `shell:` classes the Sidebar of `ui` uses in place of shadcn/ui's `md:` (768px), and `useIsMobile` in `ui/hooks` holds the same width.

**Board table.** Each group starts with its title in `type-title`, in the group colour, with a chevron that folds it and its count. Its table has a header of 40px, with the titles of the columns in `type-label`, centred but for the first. Each row is 44px high; the first cell holds the name and starts with a 6px strip in the group colour. A status column is a label that fills its cell. A person is a coloured circle with initials. Numbers align to the right. The last row of a group adds an item. The foot of a group sums its numbers and shows a bar of its labels, each colour as wide as its share. A hovered row takes `muted`, the selected one takes `selected`. Under the toolbar of the screen (the split button that creates an item, the search field, and the tools Person, Filter, Sort and Group), the table scrolls sideways inside its own container when it is wider than the work surface.

**Record page.** A breadcrumb; a header with the logo or the initials of the record, its name in display type, its labels and the chain it belongs to, such as consultancy and client, and its actions at the end; a row of figures, each a small card; then view tabs in the `line` variant, whose content fills the rest of the page.

**Side panel.** Up to 640px wide, over the screen, with `scrim` behind it and `shadow-floating`. A header with the context in caption type, the title and its labels, and the close button; a body that scrolls; a footer with the actions. Escape and the scrim close it.

**Form page.** Sections one under the other, each on two columns: its title and a sentence on the left, its fields on the right, in a grid of six columns where a field takes two, three, four or six. A line of `border` ends each section. A choice between few options with a sentence each is a set of option cards rather than a select. The cancel and the primary action stay at the bottom of the work surface while it scrolls. Under 900px a section is one column.

**Framed pane.** The header and the footer are 56px high, on the frame: the header shows the logo, the mark in `brand` and the product name in title type, with the slogan under the name in caption type (`ProductLogo` with `withSlogan`), and the footer holds the language selector at its end. Between them, the pane is on `card`, with a 1px border, 12px corners and 24px of padding, 32px from 640px up, in the centre of the frame; at a narrow width it takes the width of the page, less 16px on each side. Its heading is display type. The screen has no theme toggle, no version number and no link to the source code.

- The pane has two widths. A screen of few controls, such as the login screen, takes the narrow pane, at most 384px wide. A form, such as the setup, takes the wide pane, at most 720px wide, so that its fields go side by side and the form stays short.
- In the wide pane the fields go in pairs, or in threes when they are short, such as city, region and postal code. A field with a long value, such as a website, takes the whole line. Where the group of fields is narrower than 512px, they go one under the other.
- A step of a form ends on a line above a 1px `border-subtle`: the way back to the step before at the start, when there is one, and the one primary action at the end. On a narrow screen the primary action takes the rest of the line.
- The login screen is the narrow pane: its heading, the notice it was given, the buttons of the login providers above a divider, absent while no provider is on, then the step of the e-mail or of the login code. A button of a provider is an outline button the width of the pane, "Log in with" and the display name of the provider, one under the other in the order the API gives them; it is a link, because it leaves the app for the provider, and a long name wraps inside it on a narrow screen. The divider is a 1px line in `border` with "or" in its middle, in caption type, in `muted-foreground`. The primary action stays the one of the step. The six digits of a login code are an InputOTP of six boxes, in mono. The way to a new code is a link button under the field, closed for a minute after the last code, with the seconds left in caption type under it.

**Form dialog.** The fields, then a footer with the cancel action and the primary action. Enter in a field submits. A validation error shows under its field, in the error colour, with a sign and a sentence that says what to do; the field is marked and takes the focus. A dialog that closes keeps nothing of what was typed.

## Known gaps

- The frame colour comes from Vibe's tokens, but where monday.com puts it, behind the header and the sidebar, was read from its product, not from documentation.
- Vibe documents no mono typeface; JetBrains Mono is this project's choice.
- Hover, focus and active states come from Vibe's CSS. None was seen in motion.
- Vibe's focus ring is a 3px glow at half opacity, which does not reach 3:1; the focus keeps the solid outline of the previous style.
