import { CircleAlertIcon } from "lucide-react"

import { FieldError } from "#ui/components/field"

import type { FieldProblem } from "./problems"
import { useErrorTexts } from "./texts"

/**
 * What is wrong with a field, under it: a sign and a sentence that says what
 * to do, in the error colour (Form dialog in DESIGN.md). The field names it
 * with `aria-describedby={id}`, so it is read with the field.
 */
export function FieldProblemMessage({
  id,
  problem,
}: {
  id: string
  problem: FieldProblem
}) {
  const { ofField } = useErrorTexts()

  return (
    <FieldError id={id} className="flex items-start gap-1">
      <CircleAlertIcon className="mt-0.5 size-3.5 shrink-0" />
      {ofField(problem)}
    </FieldError>
  )
}
