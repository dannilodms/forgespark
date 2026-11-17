import { createFileRoute } from "@tanstack/react-router"

import { AvatarBlock, CardBlock } from "../../gallery/blocks/card"
import { SeparatorBlock } from "../../gallery/blocks/select"
import { TableBlock } from "../../gallery/blocks/table"
import { dataDisplay } from "../../gallery/explanations/data-display"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/data-display")({
  component: DataDisplay,
})

function DataDisplay() {
  return (
    <GalleryPage section="dataDisplay" lead={dataDisplay.lead}>
      <CardBlock />
      <TableBlock />
      <AvatarBlock />
      <SeparatorBlock />
    </GalleryPage>
  )
}
