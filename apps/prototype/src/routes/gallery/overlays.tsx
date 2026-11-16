import { createFileRoute } from "@tanstack/react-router"

import { CommandBlock } from "../../gallery/blocks/command"
import { DialogBlock, SheetBlock } from "../../gallery/blocks/dialog"
import {
  DropdownMenuBlock,
  PopoverBlock,
  TooltipBlock,
} from "../../gallery/blocks/menu"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/overlays")({
  component: Overlays,
})

function Overlays() {
  return (
    <GalleryPage
      title="Overlays"
      lead="What opens over the page. Each one opens from a button that says what it opens, and closes with Escape."
    >
      <DialogBlock />
      <DropdownMenuBlock />
      <PopoverBlock />
      <TooltipBlock />
      <CommandBlock />
      <SheetBlock />
    </GalleryPage>
  )
}
