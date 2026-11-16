import { createFileRoute } from "@tanstack/react-router"

import { BadgeBlock, KbdBlock } from "../../gallery/blocks/badge"
import { ButtonBlock } from "../../gallery/blocks/button"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/actions")({ component: Actions })

function Actions() {
  return (
    <GalleryPage
      title="Actions"
      lead="What a person presses to make something happen."
    >
      <ButtonBlock />
      <BadgeBlock />
      <KbdBlock />
    </GalleryPage>
  )
}
