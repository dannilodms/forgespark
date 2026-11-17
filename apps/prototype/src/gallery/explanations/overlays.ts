/**
 * The explanatory text of the Overlays route, in English only. The label of a
 * button that opens an example is explanatory too: it says what the gallery
 * shows, and is not part of the component.
 */
export const overlays = {
  lead: "What opens over the page. Each one opens from a button that says what it opens, and closes with Escape.",

  dialog: {
    about: {
      title: "Dialog",
      note: "A question or a short form that has to be answered before the work goes on. It holds the focus until it closes, and Escape closes it.",
    },
    openRename: "Open the rename dialog",
    confirming: {
      title: "Confirming something that cannot be undone",
      note: "The title asks the question and names the thing. The button repeats the verb, and the way out is the quiet one.",
    },
    openDelete: "Open the delete confirmation",
    edgeCases: {
      title: "Edge cases",
      note: "A long title and a long text wrap. A dialog taller than the screen scrolls inside itself, and its buttons stack on a narrow screen.",
    },
    openLong: "Open a dialog with a long text",
  },

  dropdownMenu: {
    about: {
      title: "DropdownMenu",
      note: "The actions on one thing, behind one button. The arrow keys move through it, a letter jumps to an item, and Escape closes it.",
    },
    openRequest: "Open the request menu",
    options: {
      title: "Options that stay chosen",
      note: "An item can be a checkbox or one of a group of radios, under a label that names the group.",
    },
    openView: "Open the view menu",
    edgeCases: {
      title: "Edge cases",
      note: "A long item is cut with an ellipsis inside the width of the menu.",
    },
    openLong: "Open a menu with long items",
  },

  popover: {
    about: {
      title: "Popover",
      note: "A small panel beside the button that opens it, for a few fields or a detail. The page behind stays usable, and Escape closes it.",
    },
    openEstimate: "Open the estimate popover",
    sides: {
      title: "Sides",
      note: "A popover opens under its button by default, and moves when there is no room.",
    },
    openOn: {
      top: "Open a popover on the top",
      right: "Open a popover on the right",
      bottom: "Open a popover on the bottom",
      left: "Open a popover on the left",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long text wraps inside the popover, which never grows wider than the screen.",
    },
    openLong: "Open a popover with a long text",
  },

  tooltip: {
    about: {
      title: "Tooltip",
      note: "The name of a control that shows only an icon, with its shortcut. It opens on hover and on keyboard focus, and never holds something a person has to read to go on.",
    },
    typical: {
      title: "Typical",
      note: "Move the pointer over the button, or reach it with Tab.",
    },
    sides: {
      title: "Sides",
      note: "A tooltip opens above its control by default, and moves when there is no room.",
    },
    /** The button of each side, and what its tooltip says. */
    side: {
      top: { trigger: "Tooltip on the top", content: "Opens on the top" },
      right: { trigger: "Tooltip on the right", content: "Opens on the right" },
      bottom: {
        trigger: "Tooltip on the bottom",
        content: "Opens on the bottom",
      },
      left: { trigger: "Tooltip on the left", content: "Opens on the left" },
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long text wraps at the widest a tooltip gets.",
    },
    withLong: "Tooltip with a long text",
  },

  command: {
    about: {
      title: "Command",
      note: "A list of places and actions filtered by typing. As the command palette, it opens over any screen. The arrow keys move through it and Enter runs the item.",
    },
    typical: {
      title: "Typical",
      note: "The palette opens in a dialog, with the search field already focused.",
    },
    open: "Open the command palette",
    inPlace: {
      title: "In place",
      note: "The same list set into a pane, for a picker that stays on the screen.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A search that matches nothing says so, and a list with no item says there is nothing to choose. A long item is cut with an ellipsis.",
    },
  },

  sheet: {
    about: {
      title: "Sheet",
      note: "A panel that slides over the edge of the screen and keeps the page behind it. The sidebar becomes one on a narrow screen.",
    },
    openRequest: "Open the request sheet",
    sides: {
      title: "Sides",
      note: "A sheet comes from the right by default. Navigation comes from the left, and a short choice from the top or the bottom.",
    },
    /** The button of each side, and what its sheet says. */
    side: {
      left: {
        trigger: "Open a sheet from the left",
        content: "This sheet came from the left of the screen.",
      },
      top: {
        trigger: "Open a sheet from the top",
        content: "This sheet came from the top of the screen.",
      },
      bottom: {
        trigger: "Open a sheet from the bottom",
        content: "This sheet came from the bottom of the screen.",
      },
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long title wraps beside the close button, and a long text scrolls inside the sheet.",
    },
    openLong: "Open a sheet with a long text",
  },
} as const
