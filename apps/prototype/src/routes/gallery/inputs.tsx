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
import { inputs } from "../../gallery/explanations/inputs"
import { GalleryPage } from "../../gallery/parts"

export const Route = createFileRoute("/gallery/inputs")({ component: Inputs })

function Inputs() {
  return (
    <GalleryPage section="inputs" lead={inputs.lead}>
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
