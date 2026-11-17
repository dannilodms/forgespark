/** The explanatory text of the Actions route, in English only. */
export const actions = {
  lead: "What a person presses to make something happen.",

  button: {
    about: {
      title: "Button",
      note: "Starts an action. One button in a group is primary; the others are secondary or ghost.",
    },
    when: {
      default: "The one primary action of a group.",
      secondary: "An action beside the primary one.",
      outline: "An action that stands alone on a surface.",
      ghost: "The quiet way out, and actions in a toolbar.",
      destructive: "An action that removes something.",
      link: "An action that reads as part of a sentence.",
    },
    withIcon: {
      title: "With an icon or a shortcut",
      note: "An icon leads or trails the label. A button with an icon alone names its action for a screen reader. The shortcut of an action is a chip at the end.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long label is cut with an ellipsis inside the width it is given. It never pushes the layout.",
    },
  },

  badge: {
    about: {
      title: "Badge",
      note: "A short word or a count beside the thing it describes. For the status of a record, use the status indicator.",
    },
    when: {
      default: "The one thing on the screen to notice first.",
      secondary: "A category or a tag. The usual choice.",
      outline: "A quiet fact beside a title.",
      destructive: "Something is wrong. The word says what.",
      ghost: "A tag that should almost disappear.",
      link: "A tag that leads somewhere.",
    },
    withIcon: {
      title: "With an icon or a count",
      note: "An icon leads or trails the word. A count is set in mono.",
    },
    states: {
      title: "States",
      note: "A badge has states only when it is a link. Each state is pinned, so all of them are visible at once.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long word is cut with an ellipsis inside the width it is given.",
    },
  },

  kbd: {
    about: {
      title: "Kbd",
      note: "A key or a shortcut, named where the action it triggers is shown.",
    },
    severalKeys: {
      title: "A shortcut of several keys",
      note: "Each key is its own chip, in the order they are pressed.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A key with a long name keeps its name whole and grows to fit it.",
    },
  },
} as const
