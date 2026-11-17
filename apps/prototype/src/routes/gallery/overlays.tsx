import { createFileRoute } from "@tanstack/react-router"

import { CommandBlock } from "../../gallery/blocks/command"
import { DialogBlock, SheetBlock } from "../../gallery/blocks/dialog"
import {
  DropdownMenuBlock,
  PopoverBlock,
  TooltipBlock,
} from "../../gallery/blocks/menu"
import { overlays } from "../../gallery/explanations/overlays"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/overlays")({
  component: Overlays,
})

function Overlays() {
  return (
    <GalleryPage section="overlays" lead={overlays.lead}>
      <DialogBlock />
      <DropdownMenuBlock />
      <PopoverBlock />
      <TooltipBlock />
      <CommandBlock />
      <SheetBlock />
    </GalleryPage>
  )
}
