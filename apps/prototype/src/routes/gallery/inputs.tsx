import { createFileRoute } from "@tanstack/react-router"

import {
  CheckboxBlock,
  RadioGroupBlock,
  SwitchBlock,
} from "../../gallery/blocks/choice"
import { FieldBlock, LabelBlock } from "../../gallery/blocks/field"
import { InputBlock, InputOTPBlock } from "../../gallery/blocks/input"
import { SelectBlock } from "../../gallery/blocks/select"
import { InputGroupBlock, TextareaBlock } from "../../gallery/blocks/textarea"
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
      <TextareaBlock />
      <SelectBlock />
      <CheckboxBlock />
      <RadioGroupBlock />
      <SwitchBlock />
      <InputGroupBlock />
    </GalleryPage>
  )
}
