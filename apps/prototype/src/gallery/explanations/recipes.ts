/** The explanatory text of the Recipes route, in English only. */
export const recipes = {
  lead: "A recipe is a named composition of components that several screens reuse. Each of the four is shown here on its own, over sample data.",

  /** One sentence, around the link to the sample work screen. */
  together: {
    before: "The four are put together in the ",
    link: "sample work screen",
    after: ", where the login screen leads.",
  },

  appShell: {
    about: {
      title: "App shell",
      note: "Use it around every screen of the product: it holds what stays in place while the panes change.",
    },
    typical: {
      title: "Typical",
      note: "A title bar with the user menu, the sidebar navigation, a top bar with the breadcrumb and the command palette trigger, the main area of flush panes, and a status bar. The button of the top bar collapses the sidebar, and Ctrl K opens the command palette.",
    },
    /** What stands in for each pane of the main area. */
    listPane: "List pane",
    detailPane: "Detail pane",
  },

  listPane: {
    about: {
      title: "List pane",
      note: "Use it for a set of items a person scans, filters and picks from, beside the detail of the one picked.",
    },
    typical: {
      title: "Typical",
      note: "A filter, rows that each carry a status indicator, and an empty state when the filter leaves nothing. The arrow keys move between the rows, and Enter selects one.",
    },
  },

  detailPane: {
    about: {
      title: "Detail pane",
      note: "Use it for the one item picked in a list pane: what it is, and what can be done to it.",
    },
    typical: {
      title: "Typical",
      note: "A header with the identifier, the status, the actions and the title of the item, and tabs that switch its content.",
    },
  },

  formDialog: {
    about: {
      title: "Form dialog",
      note: "Use it to create or change one thing with a few fields, without leaving the screen.",
    },
    typical: {
      title: "Typical",
      note: "Fields, one primary action and a cancel action. Create with no title to see the validation error on the field.",
    },
    open: "Open the form dialog",
  },
} as const
