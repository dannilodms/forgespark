/**
 * The explanatory text every route of the gallery shares: the names of the
 * parts of a component block, in the order of DESIGN.md.
 *
 * Explanatory text is documentation, like DESIGN.md: it is in English only
 * and never enters the translation files. Each route keeps its own in the
 * file beside this one that carries its name.
 */
export const partTitle = {
  typical: "Typical",
  variants: "Variants",
  sizes: "Sizes",
  sides: "Sides",
  edgeCases: "Edge cases",
} as const

/** The part that pins the states of a component. */
export const pinnedStates = {
  title: "States",
  note: "Each state is pinned, so all of them are visible at once.",
} as const
