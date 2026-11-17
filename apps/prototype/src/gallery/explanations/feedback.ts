/** The explanatory text of the Feedback route, in English only. */
export const feedback = {
  lead: "What tells a person the state of things: a message, a status, something loading, nothing to show.",

  alert: {
    about: {
      title: "Alert",
      note: "A message that stays on the screen, beside what it is about. For a message about something that just happened, use a toast.",
    },
    variants: {
      title: "Variants",
      note: "The destructive variant reports an error. It always carries an icon, so it does not depend on colour.",
    },
    carries: {
      title: "What it can carry",
      note: "A title alone, a description alone, and an action at the end.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long title and a long description wrap inside the alert.",
    },
  },

  toast: {
    about: {
      title: "Toast",
      note: "A short message about something that just happened. It leaves by itself, so it never holds the only copy of an error the person has to fix.",
    },
    kinds: {
      title: "Kinds",
      note: "Each kind has its own sign. A status colour never appears without it.",
    },
    /** The label of the button that shows each kind, and when the kind is used. */
    kind: {
      plain: {
        label: "Show a plain toast",
        when: "Something happened and nothing has to be done.",
      },
      success: {
        label: "Show a success toast",
        when: "What the person asked for is done.",
      },
      warning: {
        label: "Show a warning toast",
        when: "It worked, and something needs attention.",
      },
      error: {
        label: "Show an error toast",
        when: "It did not work. The toast says what to do next.",
      },
      info: {
        label: "Show an information toast",
        when: "A fact the person may want to know.",
      },
    },
    loading: {
      title: "Loading and an action",
      note: "A toast follows work in progress and changes when it ends. A toast can carry one action, such as undoing what was just done.",
    },
    showLoading: "Show a loading toast",
    showAction: "Show a toast with an action",
    edgeCases: {
      title: "Edge cases",
      note: "A long title and a long description wrap inside the toast, and several toasts stack.",
    },
    showLong: "Show a toast with a long text",
  },

  skeleton: {
    about: {
      title: "Skeleton",
      note: "The shape of content that is still loading. It takes the place and the size of what will arrive, so nothing jumps.",
    },
    list: {
      title: "A list that is loading",
      note: "One skeleton per row, with widths that differ as real text does.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A skeleton shrinks with its container and never pushes the layout.",
    },
  },

  empty: {
    about: {
      title: "Empty",
      note: "What a list, a table or a pane shows when it has nothing. It says why, and offers the next step when there is one.",
    },
    afterSearch: {
      title: "After a search, and with two actions",
      note: "A search that finds nothing says what was searched and offers the way back. One action is primary.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long headline and a long description wrap inside a narrow pane.",
    },
  },

  statusIndicator: {
    about: {
      title: "Status indicator",
      note: "The status of a record: a sign in one of the four status colours, and a word. The sign belongs to the meaning, and the word comes from the screen.",
    },
    meanings: {
      title: "The four meanings",
      note: "A colour has one meaning everywhere. Which status of a record shows which meaning is decided with each screen.",
    },
    when: {
      working: "Someone is on it. Nothing is asked of the reader.",
      attention: "It is stopped until someone acts.",
      done: "It ended as expected.",
      error: "It ended wrong, or it cannot go on.",
    },
    variants: {
      title: "Variants",
      note: "Plain in a row or a table, where the line already separates: the sign carries the colour and the word is set in the text colour. A badge beside a title, where the status stands alone: the sign and the word take the colour, on its subtle background.",
    },
    inList: {
      title: "In a list",
      note: "Each row carries its status at the end, so a column of them can be read down.",
    },
    onRow: {
      title: "On a hovered and a selected row",
      note: "A row changes its fill, and the status on it stays readable: the word is text, and the sign keeps 3:1 on both fills.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long word is cut with an ellipsis. The sign is never cut.",
    },
  },
} as const
