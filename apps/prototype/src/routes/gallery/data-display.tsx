import { createFileRoute } from "@tanstack/react-router"

import { AvatarBlock, CardBlock } from "../../gallery/blocks/card"
import { SeparatorBlock } from "../../gallery/blocks/select"
import { TableBlock } from "../../gallery/blocks/table"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/data-display")({
  component: DataDisplay,
})

function DataDisplay() {
  return (
    <GalleryPage
      title="Data display"
      lead="What shows information without asking for anything."
    >
      <CardBlock />
      <TableBlock />
      <AvatarBlock />
      <SeparatorBlock />
    </GalleryPage>
  )
}
