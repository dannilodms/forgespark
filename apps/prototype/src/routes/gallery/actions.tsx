import { createFileRoute } from "@tanstack/react-router"

import { BadgeBlock, KbdBlock } from "../../gallery/blocks/badge"
import { ButtonBlock } from "../../gallery/blocks/button"
import { actions } from "../../gallery/explanations/actions"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/actions")({ component: Actions })

function Actions() {
  return (
    <GalleryPage section="actions" lead={actions.lead}>
      <ButtonBlock />
      <BadgeBlock />
      <KbdBlock />
    </GalleryPage>
  )
}
