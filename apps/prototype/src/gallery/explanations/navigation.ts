/** The explanatory text of the Navigation route, in English only. */
export const navigation = {
  lead: "What takes a person from one place to another, and shows where they are.",

  tabs: {
    about: {
      title: "Tabs",
      note: "Views of the same thing, one at a time. The arrow keys move between the tabs, and Tab goes into the view.",
    },
    when: {
      default:
        "Views of one thing inside a pane. The active tab is a raised control.",
      line: "Views that span the width of a pane, under its header.",
    },
    withIcon: {
      title: "With an icon or a count, and vertical",
      note: "An icon leads the label and a count trails it, in mono. A vertical list sits beside its view.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "Long labels wrap onto a second line of tabs before they push the layout, and a single label can be cut with an ellipsis.",
    },
  },

  breadcrumb: {
    about: {
      title: "Breadcrumb",
      note: "Where the current page sits, as a path of links back up. The last item is the page itself and is not a link.",
    },
    longPath: {
      title: "A long path, and another separator",
      note: "The middle of a long path collapses into an ellipsis. A slash can stand in for the chevron.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long name is cut with an ellipsis, and a path wider than its container wraps onto a second line.",
    },
  },

  sidebar: {
    about: {
      title: "Sidebar",
      note: "The navigation of the app: the places a person can go, in groups. It sits flush against the page, split from it by a line.",
    },
    typical: {
      title: "Typical",
      note: "The current place takes the selected fill. A count sits at the end of its item, in mono.",
    },
    collapsing: {
      title: "Collapsing",
      note: "The button, or Ctrl B, collapses the sidebar to its icons, and each icon then names its place in a tooltip. On a narrow screen the sidebar is a sheet that the same button opens.",
    },
    states: {
      title: "States of an item",
      note: "Each state is pinned, so all of them are visible at once.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long label is cut with an ellipsis and never widens the sidebar. An item can hold a second level. While the places load, each is a skeleton.",
    },
    /** What stands in for the page in a preview. */
    page: "The page beside the sidebar.",
  },
} as const
