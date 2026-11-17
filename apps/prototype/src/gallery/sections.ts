/**
 * The routes of the gallery, in the order of DESIGN.md. A new route is added
 * here, with its name in the translation files under
 * `gallery.navigation.sections`.
 */
export const gallerySections = [
  { to: "/gallery/foundations", name: "foundations" },
  { to: "/gallery/actions", name: "actions" },
  { to: "/gallery/inputs", name: "inputs" },
  { to: "/gallery/navigation", name: "navigation" },
  { to: "/gallery/data-display", name: "dataDisplay" },
  { to: "/gallery/feedback", name: "feedback" },
  { to: "/gallery/overlays", name: "overlays" },
  { to: "/gallery/do-and-dont", name: "doAndDont" },
  { to: "/gallery/recipes", name: "recipes" },
] as const

export type GallerySection = (typeof gallerySections)[number]["name"]
