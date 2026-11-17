/**
 * The explanatory text of the Foundations route, in English only. The names
 * and the roles of the style variables are in `style-variables.ts`, the
 * registry the coverage check reads.
 */
export const foundations = {
  lead: "Every style variable of the design system, with the value it has in the current theme. Colours are shown on the surface they are used on, with the contrast measured here.",

  /** What a measured colour pair says under its sample. */
  contrast: {
    on: "on",
    noLevel: "no level to reach",
    fails: "fails",
    passes: "passes",
  },

  lines: {
    title: "Lines and edges",
    note: "Colours that separate or decorate. None of them carries text.",
  },

  type: {
    about: {
      title: "Type",
      note: "Three typefaces and seven roles. A role sets family, size, line height, weight and tracking together.",
    },
    columns: {
      role: "Role",
      size: "Size / line / weight",
      use: "Use",
      specimen: "Specimen",
    },
  },

  spacing: {
    about: {
      title: "Spacing and control sizes",
      note: "The base unit is 4px, and every spacing step is a multiple of it. Controls of the same kind share a height, so a form stays aligned.",
    },
    steps: "Spacing steps",
    controls: "Control sizes",
  },

  radius: {
    title: "Radius",
    note: "Five roles. A corner is never rounder than its role.",
  },

  elevation: {
    about: {
      title: "Elevation",
      note: "Borders carry the structure. A shadow appears only on a floating surface.",
    },
    flat: { name: "Flat", note: "Page and chrome. No border and no shadow." },
    hairline: {
      name: "Hairline",
      note: "Pane, card and control. A 1px border.",
    },
    floating: {
      name: "Floating",
      note: "Popover, dialog, command palette and toast. A 1px edge and the one shadow.",
    },
  },

  pageFraming: {
    title: "Page framing",
    note: "The grid texture and the diamond nodes frame a page. A page opts in with page-framing and frame-node; the inside of a component stays plain.",
  },

  motion: {
    about: { title: "Motion", note: "Motion is short and plain." },
    columns: { what: "What moves", duration: "Duration", easing: "Easing" },
  },

  icons: { title: "Icons", note: "Lucide, at 12, 14 and 16px." },
} as const
