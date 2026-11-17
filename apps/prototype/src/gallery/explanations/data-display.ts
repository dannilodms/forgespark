/** The explanatory text of the Data display route, in English only. */
export const dataDisplay = {
  lead: "What shows information without asking for anything.",

  card: {
    about: {
      title: "Card",
      note: "A group of facts and actions about one thing, on a page that shows several. Panes of a work screen are not cards: they are flush and split by lines.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long title and a long description wrap. A card that is loading keeps its shape with skeletons.",
    },
  },

  table: {
    about: {
      title: "Table",
      note: "Records of one kind, one per row, compared column by column. An identifier and a number are set in mono, and a number is aligned to the right.",
    },
    typical: {
      title: "Typical",
      note: "The second row is selected. A row shows the hover wash under the pointer.",
    },
    withTotal: {
      title: "With a total and a caption",
      note: "The footer holds what sums the column. The caption says what the table is, for everyone.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long text is cut inside its cell, and the table scrolls sideways inside its frame when the columns do not fit. With no rows, the table keeps its header and shows an empty state. While it loads, each cell is a skeleton.",
    },
  },

  avatar: {
    about: {
      title: "Avatar",
      note: "The picture of a person, or the initials when there is no picture. It always sits beside the name or carries it for a screen reader.",
    },
    group: {
      title: "A group",
      note: "People on the same thing overlap. The ones that do not fit become a count.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "With no picture, or with one that fails to load, the initials show. A long name beside the avatar is cut with an ellipsis.",
    },
  },

  separator: {
    about: {
      title: "Separator",
      note: "A line between two groups inside one surface. Between surfaces, the border of the surface does this.",
    },
    directions: { title: "Horizontal and vertical" },
  },
} as const
