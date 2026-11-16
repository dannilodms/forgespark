import { createFileRoute } from "@tanstack/react-router"

import { SeparatorBlock } from "../../gallery/blocks/select"
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
      <SeparatorBlock />
    </GalleryPage>
  )
}
