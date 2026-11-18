import type { ErrorComponentProps } from "@tanstack/react-router"
import { RotateCwIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"

import { FramedPane } from "../components/framed-pane"
import { ErrorAlert } from "./error-alert"

/**
 * Where a failure while opening or rendering a screen falls: what failed,
 * its trace identifier when the API gave one, and a reload.
 */
export function ErrorScreen({ error }: ErrorComponentProps) {
  const { t } = useTranslation()

  return (
    <FramedPane>
      <h1 className="type-display">{t("errorScreen.heading")}</h1>
      <ErrorAlert error={error} />
      <Button onClick={() => window.location.reload()}>
        <RotateCwIcon data-icon="inline-start" />
        {t("errorScreen.reload")}
      </Button>
    </FramedPane>
  )
}
