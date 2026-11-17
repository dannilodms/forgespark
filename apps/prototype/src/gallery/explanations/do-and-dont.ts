/**
 * The explanatory text of the Do and Don't route, in English only: the eight
 * rules as DESIGN.md states them, and one line on why each example is right
 * or wrong.
 */
export const doAndDont = {
  lead: "The eight rules of the design system, each shown followed and broken. The rule is under its name, and each example says why it is right or wrong.",

  /** The captions of a pair. */
  caption: { do: "Do", dont: "Don't" },

  borders: {
    name: "Borders carry the structure",
    rule: "Surfaces are separated by 1px lines. A shadow appears only on a floating surface: popover, dialog, command palette, toast.",
    do: "Two panes side by side, flush, with one line between them.",
    dont: "The same panes as cards with a gap and a shadow. The shadow says they float, and they do not.",
  },
  accent: {
    name: "One accent, and colour means status",
    rule: "The one accent colour is blue. Any other colour is a status colour with a fixed meaning: blue is working, amber needs attention, green is done, red is an error.",
    do: "Tags are neutral. Colour appears on the primary action and on the status, where it means something.",
    dont: "A colour for each tag and a green button. Amber and green no longer mean attention and done.",
  },
  mono: {
    name: "Mono is for metadata",
    rule: "Shortcuts, counts, paths and identifiers are set in the mono typeface. Running text is set in the interface typeface.",
    do: "The identifier, the path, the count and the shortcut are mono. The sentence is not.",
    dont: "The sentence is mono and the identifier is not. Nothing tells the reader which part is data.",
  },
  decoration: {
    name: "Decoration frames the page",
    rule: "The grid texture and the diamond nodes belong to the page framing. The inside of a component is plain.",
    do: "The texture is behind the pane and the nodes sit on its corners. The pane itself is plain.",
    dont: "The texture fills the pane and the button, and a node decorates a title. The content has to fight it.",
  },
  motion: {
    name: "Motion is short and plain",
    rule: "A colour change takes 100 to 150ms. A control moves at most 1 to 2px. An overlay opens in 200 to 250ms.",
    do: "The button changes colour at once under the pointer and moves 1px when pressed. The popover opens in 200ms.",
    dont: "The badge never stops moving, and the button grows and turns slowly under the pointer. Motion that says nothing takes the eye from the work.",
    openPopover: "Open the popover",
  },
  variables: {
    name: "Every value is a style variable",
    rule: "A component takes each colour, radius, font, spacing step and shadow from a style variable, and hardcodes none.",
    do: "Every value comes from a style variable, so the pane follows the theme. Change the theme in the prototype bar.",
    dont: "The colours, the radius and the shadow are written as literals, so the pane stays white in the dark theme.",
  },
  primary: {
    name: "One primary action per group",
    rule: "In a group of actions, one button is primary. The others are secondary or ghost.",
    do: "One primary button says what the group is for. The others step back.",
    dont: "Three primary buttons. Nothing says which one the person came here to press.",
  },
  status: {
    name: "Status is colour plus a sign",
    rule: "A status always shows an icon or a word together with its colour.",
    do: "Each status has its sign and its word. It reads without the colour.",
    dont: "A dot of colour and nothing else. Someone who does not tell red from green cannot read it, and nobody knows what amber means.",
  },
} as const
