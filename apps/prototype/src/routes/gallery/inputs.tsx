import { createFileRoute } from "@tanstack/react-router"

import { FieldBlock, LabelBlock } from "../../gallery/blocks/field"
import { InputBlock, InputOTPBlock } from "../../gallery/blocks/input"
import { SelectBlock } from "../../gallery/blocks/select"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/inputs")({ component: Inputs })

function Inputs() {
  return (
    <GalleryPage
      title="Inputs"
      lead="What a person types into or chooses from."
    >
      <FieldBlock />
      <LabelBlock />
      <InputBlock />
      <InputOTPBlock />
      <SelectBlock />
    </GalleryPage>
  )
}
