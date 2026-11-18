import { CircleXIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"

import { useErrorTexts } from "./texts"

/**
 * A failure that is about no field, above the form: what happened, what to
 * do, and the trace identifier to pass on. Red, with a cross: never colour
 * alone.
 */
export function ErrorAlert({ error }: { error: unknown }) {
  const { t } = useTranslation()
  const { ofError } = useErrorTexts()
  const { title, todo, traceId } = ofError(error)

  return (
    <Alert variant="destructive">
      <CircleXIcon />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>
        <p>{todo}</p>
        {traceId ? (
          <p className="text-muted-foreground">
            {t("errors.traceId")}:
            <code className="block type-code break-all text-foreground">
              {traceId}
            </code>
          </p>
        ) : null}
      </AlertDescription>
    </Alert>
  )
}
