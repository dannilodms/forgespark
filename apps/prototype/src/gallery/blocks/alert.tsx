import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@workspace/ui/components/alert"
import { Button } from "@workspace/ui/components/button"
import { CircleXIcon, InfoIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { feedback } from "../explanations/feedback"
import { partTitle } from "../explanations/shared"
import { Frame, Part, Section } from "../parts"

export function AlertBlock() {
  const { t } = useTranslation()
  const text = feedback.alert

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame className="max-w-xl">
          <Alert>
            <InfoIcon />
            <AlertTitle>{t("gallery.alert.billingCloses.title")}</AlertTitle>
            <AlertDescription>
              {t("gallery.alert.billingCloses.description")}
            </AlertDescription>
          </Alert>
        </Frame>
      </Part>

      <Part {...text.variants}>
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch">
          <Alert>
            <InfoIcon />
            <AlertTitle>{t("gallery.alert.archived.title")}</AlertTitle>
            <AlertDescription>
              {t("gallery.alert.archived.description")}
            </AlertDescription>
          </Alert>
          <Alert variant="destructive">
            <CircleXIcon />
            <AlertTitle>{t("gallery.common.exportFailed")}</AlertTitle>
            <AlertDescription>
              {t("gallery.alert.exportFailedDescription")}
            </AlertDescription>
          </Alert>
        </Frame>
      </Part>

      <Part {...text.carries}>
        <Frame className="max-w-xl flex-col flex-nowrap items-stretch">
          <Alert>
            <InfoIcon />
            <AlertTitle>{t("gallery.alert.twoEditing")}</AlertTitle>
          </Alert>
          <Alert>
            <AlertDescription>{t("gallery.alert.hoursToday")}</AlertDescription>
          </Alert>
          <Alert variant="destructive">
            <CircleXIcon />
            <AlertTitle>{t("gallery.common.exportFailed")}</AlertTitle>
            <AlertDescription>
              {t("gallery.alert.fileNotCreated")}
            </AlertDescription>
            <AlertAction>
              <Button variant="outline" size="xs">
                {t("gallery.alert.tryAgain")}
              </Button>
            </AlertAction>
          </Alert>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame className="max-w-xs">
          <Alert variant="destructive">
            <CircleXIcon />
            <AlertTitle>{t("gallery.alert.long.title")}</AlertTitle>
            <AlertDescription>
              {t("gallery.alert.long.description")}
            </AlertDescription>
          </Alert>
        </Frame>
      </Part>
    </Section>
  )
}
