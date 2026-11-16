import { createFileRoute } from "@tanstack/react-router"

import { GalleryPage } from "../../gallery/parts"
import { Rules } from "../../gallery/rules"

export const Route = createFileRoute("/gallery/do-and-dont")({
  component: DoAndDont,
})

function DoAndDont() {
  return (
    <GalleryPage
      title="Do and Don't"
      lead="The eight rules of the design system, each shown followed and broken. The rule is under its name, and each example says why it is right or wrong."
    >
      <Rules />
    </GalleryPage>
  )
}
