/**
 * Every style variable of packages/ui, as the Foundations route shows it.
 *
 * This is the one place to register a style variable. The coverage check
 * fails when the stylesheet of packages/ui declares one that is not named
 * here, so write each name in full, in quotes.
 */

export type PairSpec = {
  name: string
  background: string
  foreground: string
  minimum?: 3 | 4.5 | null
  kind?: "text" | "line" | "sign"
}

export type PairGroup = { title: string; note: string; pairs: PairSpec[] }

export const pairGroups: PairGroup[] = [
  {
    title: "Surfaces and their text",
    note: "Each surface with the text colour that belongs on it. Text has to reach 4.5:1.",
    pairs: [
      { name: "Page", background: "background", foreground: "foreground" },
      {
        name: "Pane or card",
        background: "card",
        foreground: "card-foreground",
      },
      {
        name: "Floating surface",
        background: "popover",
        foreground: "popover-foreground",
      },
      {
        name: "Primary button",
        background: "primary",
        foreground: "primary-foreground",
      },
      {
        name: "Raised control",
        background: "secondary",
        foreground: "secondary-foreground",
      },
      {
        name: "Quiet fill",
        background: "muted",
        foreground: "muted-foreground",
      },
      {
        name: "Hover wash",
        background: "accent",
        foreground: "accent-foreground",
      },
      {
        name: "Selected row",
        background: "selected",
        foreground: "foreground",
      },
      {
        name: "Sidebar",
        background: "sidebar",
        foreground: "sidebar-foreground",
      },
      {
        name: "Sidebar, primary item",
        background: "sidebar-primary",
        foreground: "sidebar-primary-foreground",
      },
      {
        name: "Sidebar, hovered item",
        background: "sidebar-accent",
        foreground: "sidebar-accent-foreground",
      },
      {
        name: "Title bar and status bar",
        background: "chrome",
        foreground: "foreground",
      },
      {
        name: "Accent fill",
        background: "brand",
        foreground: "brand-foreground",
      },
    ],
  },
  {
    title: "Secondary and tertiary text",
    note: "Secondary text also appears in a bar and on a selected row. Tertiary text is a placeholder or an index.",
    pairs: [
      {
        name: "Secondary on the page",
        background: "background",
        foreground: "muted-foreground",
      },
      {
        name: "Secondary on a pane",
        background: "card",
        foreground: "muted-foreground",
      },
      {
        name: "Secondary on the sidebar",
        background: "sidebar",
        foreground: "muted-foreground",
      },
      {
        name: "Secondary on a bar",
        background: "chrome",
        foreground: "muted-foreground",
      },
      {
        name: "Secondary on a selected row",
        background: "selected",
        foreground: "muted-foreground",
      },
      {
        name: "Tertiary on the page",
        background: "background",
        foreground: "faint-foreground",
      },
      {
        name: "Tertiary on a pane",
        background: "card",
        foreground: "faint-foreground",
      },
      {
        name: "Tertiary on the sidebar",
        background: "sidebar",
        foreground: "faint-foreground",
      },
    ],
  },
  {
    title: "Accent and error as text",
    note: "A link or a highlighted word in the accent colour, and a validation error.",
    pairs: [
      {
        name: "Accent on the page",
        background: "background",
        foreground: "brand",
      },
      { name: "Accent on a pane", background: "card", foreground: "brand" },
      {
        name: "Error on the page",
        background: "background",
        foreground: "destructive",
      },
      {
        name: "Error on a pane",
        background: "card",
        foreground: "destructive",
      },
    ],
  },
  {
    title: "Status colours",
    note: "Each status colour as a word, on the page, on a pane and on its own subtle background.",
    pairs: [
      {
        name: "Working on the page",
        background: "background",
        foreground: "status-working",
      },
      {
        name: "Working on a pane",
        background: "card",
        foreground: "status-working",
      },
      {
        name: "Working badge",
        background: "status-working-subtle",
        foreground: "status-working",
      },
      {
        name: "Attention on the page",
        background: "background",
        foreground: "status-attention",
      },
      {
        name: "Attention on a pane",
        background: "card",
        foreground: "status-attention",
      },
      {
        name: "Attention badge",
        background: "status-attention-subtle",
        foreground: "status-attention",
      },
      {
        name: "Done on the page",
        background: "background",
        foreground: "status-done",
      },
      { name: "Done on a pane", background: "card", foreground: "status-done" },
      {
        name: "Done badge",
        background: "status-done-subtle",
        foreground: "status-done",
      },
      {
        name: "Error on the page",
        background: "background",
        foreground: "status-error",
      },
      {
        name: "Error on a pane",
        background: "card",
        foreground: "status-error",
      },
      {
        name: "Error badge",
        background: "status-error-subtle",
        foreground: "status-error",
      },
    ],
  },
  {
    title: "Status signs",
    note: "In a row and in a bar the sign of a status carries the colour, and the word beside it is set in the text colour. A sign has to reach 3:1 on every fill it sits on.",
    pairs: [
      {
        name: "Working sign on a hovered row",
        background: "accent",
        foreground: "status-working",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Working sign on a selected row",
        background: "selected",
        foreground: "status-working",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Working sign on a bar",
        background: "chrome",
        foreground: "status-working",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Attention sign on a hovered row",
        background: "accent",
        foreground: "status-attention",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Attention sign on a selected row",
        background: "selected",
        foreground: "status-attention",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Attention sign on a bar",
        background: "chrome",
        foreground: "status-attention",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Done sign on a hovered row",
        background: "accent",
        foreground: "status-done",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Done sign on a selected row",
        background: "selected",
        foreground: "status-done",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Done sign on a bar",
        background: "chrome",
        foreground: "status-done",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Error sign on a hovered row",
        background: "accent",
        foreground: "status-error",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Error sign on a selected row",
        background: "selected",
        foreground: "status-error",
        minimum: 3,
        kind: "sign",
      },
      {
        name: "Error sign on a bar",
        background: "chrome",
        foreground: "status-error",
        minimum: 3,
        kind: "sign",
      },
    ],
  },
  {
    title: "Control borders and focus",
    note: "The border of a field and the focus indicator have to reach 3:1 against the surface they sit on.",
    pairs: [
      {
        name: "Field border on the page",
        background: "background",
        foreground: "input",
        minimum: 3,
        kind: "line",
      },
      {
        name: "Field border on a pane",
        background: "card",
        foreground: "input",
        minimum: 3,
        kind: "line",
      },
      {
        name: "Focus on the page",
        background: "background",
        foreground: "ring",
        minimum: 3,
        kind: "line",
      },
      {
        name: "Focus on a pane",
        background: "card",
        foreground: "ring",
        minimum: 3,
        kind: "line",
      },
      {
        name: "Focus on the sidebar",
        background: "sidebar",
        foreground: "sidebar-ring",
        minimum: 3,
        kind: "line",
      },
    ],
  },
]

/**
 * The line that splits two surfaces, measured against each. A line between
 * surfaces has no level to reach, unlike the border of a field.
 */
export const surfaceLines: PairGroup = {
  title: "Lines between surfaces",
  note: "The 1px line that splits panes and draws the edge of a floating surface. It has no level to reach, and it has to show in both themes.",
  pairs: [
    {
      name: "Line on the page",
      background: "background",
      foreground: "border",
      minimum: null,
      kind: "line",
    },
    {
      name: "Line on a pane",
      background: "card",
      foreground: "border",
      minimum: null,
      kind: "line",
    },
    {
      name: "Edge of a floating surface",
      background: "popover",
      foreground: "border",
      minimum: null,
      kind: "line",
    },
  ],
}

/** Colours that separate or decorate. None of them carries text, so none has a contrast to reach. */
export const lines = [
  { variable: "border", role: "The line between surfaces" },
  { variable: "border-subtle", role: "A lighter line inside a surface" },
  { variable: "sidebar-border", role: "The line of the sidebar" },
  { variable: "button-edge", role: "The inset bottom edge of a button" },
  { variable: "button-edge-primary", role: "The edge of the primary button" },
  { variable: "page-grid", role: "The grid texture of the page framing" },
]

export const typefaces = [
  { variable: "font-serif", name: "Display" },
  { variable: "font-sans", name: "Interface" },
  { variable: "font-mono", name: "Mono" },
]

/**
 * `specimen` names the sample text of the role. The Foundations route writes
 * it in the language of the prototype.
 */
export const typeRoles = [
  {
    token: "type-display",
    use: "A page title, the headline of an empty state",
    specimen: "display",
  },
  {
    token: "type-title",
    use: "The title of a section, a card or a dialog",
    specimen: "title",
  },
  {
    token: "type-body",
    use: "Running text and descriptions",
    specimen: "body",
  },
  {
    token: "type-ui",
    use: "Controls, rows, menus, tabs",
    specimen: "ui",
  },
  {
    token: "type-label",
    use: "A small label: table header, field group",
    specimen: "label",
  },
  {
    token: "type-caption",
    use: "Secondary information under or beside something",
    specimen: "caption",
  },
  {
    token: "type-code",
    use: "Shortcuts, counts, paths, identifiers",
    specimen: "code",
  },
] as const

/** `text-sm` is what shadcn/ui components set their text in. It is pinned to the interface size. */
export const pinnedSizes = [
  { variable: "text-sm", role: "The size of text in a component" },
  { variable: "text-sm--line-height", role: "Its line height" },
]

export const radiusRoles = [
  { variable: "radius-chip", usedBy: "Badge, shortcut chip, checkbox" },
  { variable: "radius-tab", usedBy: "Tab, menu item" },
  { variable: "radius-control", usedBy: "Button, input, select" },
  { variable: "radius-pane", usedBy: "Pane, card" },
  {
    variable: "radius-floating",
    usedBy: "Dialog, popover, menu, tooltip, command palette, toast",
  },
  { variable: "radius", usedBy: "The base step shadcn/ui expects to exist" },
]

export const shadows = [
  {
    variable: "shadow-floating",
    usedBy: "Popover, dialog, menu, tooltip, command palette, toast",
  },
  { variable: "shadow-xs", usedBy: "Nothing: a control has no shadow" },
]

export const spacingSteps = [
  { size: "4px", className: "w-1" },
  { size: "8px", className: "w-2" },
  { size: "12px", className: "w-3" },
  { size: "16px", className: "w-4" },
  { size: "24px", className: "w-6" },
  { size: "32px", className: "w-8" },
]

export const controlSizes = [
  { size: "24px", className: "h-6", use: "Chip, small badge row" },
  {
    size: "28px",
    className: "h-7",
    use: "Title bar, status bar, row, small button",
  },
  { size: "32px", className: "h-8", use: "Default control" },
  { size: "36px", className: "h-9", use: "Prominent control" },
]

export const motion = [
  {
    what: "Colour and opacity",
    duration: "100 to 150ms",
    easing: "cubic-bezier(.4, 0, .2, 1)",
  },
  {
    what: "An overlay opening",
    duration: "200 to 250ms",
    easing: "cubic-bezier(.2, .8, .2, 1)",
  },
  {
    what: "A control when pressed",
    duration: "at once",
    easing: "moves 1px down",
  },
]

export const easings = [
  { variable: "ease-overlay", role: "The curve an overlay opens along" },
]
