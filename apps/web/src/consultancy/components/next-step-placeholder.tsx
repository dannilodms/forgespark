import { CircleCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"

/**
 * Stands in for the registration of a new administrator, which ticket 12 of
 * the effort builds on the step of the administrator and the review: it says
 * what the setup code opened.
 */
export function NextStepPlaceholder() {
  const { t } = useTranslation("consultancy")

  return (
    <Alert>
      <CircleCheckIcon className="text-status-done" />
      <AlertTitle>{t("setup.opened.newAdministrator")}</AlertTitle>
      <AlertDescription>{t("setup.opened.toCome")}</AlertDescription>
    </Alert>
  )
}
