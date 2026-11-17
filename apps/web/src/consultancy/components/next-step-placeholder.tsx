import { CircleCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"

import type { OpenedSetupCode } from "../data/setup"

/**
 * Stands in for the steps after the setup code, which ticket 09 of the
 * effort builds: it says which setup the code opened.
 */
export function NextStepPlaceholder({ opened }: { opened: OpenedSetupCode }) {
  const { t } = useTranslation("consultancy")

  return (
    <Alert>
      <CircleCheckIcon className="text-status-done" />
      <AlertTitle>
        {opened.opens === "first_setup"
          ? t("setup.opened.firstSetup")
          : t("setup.opened.newAdministrator")}
      </AlertTitle>
      <AlertDescription>{t("setup.opened.toCome")}</AlertDescription>
    </Alert>
  )
}
