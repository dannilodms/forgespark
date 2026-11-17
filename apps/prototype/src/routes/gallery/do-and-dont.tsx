import { createFileRoute } from "@tanstack/react-router"

import { doAndDont } from "../../gallery/explanations/do-and-dont"
import { GalleryPage } from "../../gallery/parts"
import { Rules } from "../../gallery/rules"

export const Route = createFileRoute("/gallery/do-and-dont")({
  component: DoAndDont,
})

function DoAndDont() {
  return (
    <GalleryPage section="doAndDont" lead={doAndDont.lead}>
      <Rules />
    </GalleryPage>
  )
}
