import { CircleAlertIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { FieldError } from "#ui/components/field"

import type { FieldProblem } from "./problems"
import { useErrorTexts } from "./texts"

/**
 * What is wrong with a field, under it: a sign and a sentence that says what
 * to do, in the error colour (Form dialog in DESIGN.md). A problem the API
 * answered also shows the trace identifier of the answer, to be passed on.
 * The field names it with `aria-describedby={id}`, so it is read with the
 * field.
 */
export function FieldProblemMessage({
  id,
  problem,
}: {
  id: string
  problem: FieldProblem
}) {
  const { t } = useTranslation()
  const { ofField } = useErrorTexts()

  return (
    <FieldError id={id} className="flex items-start gap-1">
      <CircleAlertIcon className="mt-0.5 size-3.5 shrink-0" />
      <span className="min-w-0">
        {ofField(problem)}
        {problem.traceId ? (
          <span className="block text-muted-foreground">
            {t("errors.traceId")}:{" "}
            <code className="type-code break-all text-foreground">
              {problem.traceId}
            </code>
          </span>
        ) : null}
      </span>
    </FieldError>
  )
}
