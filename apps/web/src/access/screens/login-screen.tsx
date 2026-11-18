import { CircleCheckIcon, TriangleAlertIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"

import { FramedPane } from "../../shared/components/framed-pane"

/**
 * What the login screen may be told by the screen the person comes from:
 * the setup was concluded, by them, by someone before them, or by one of the
 * two after an attempt of theirs got no answer, or a new administrator was
 * registered, the way back in.
 */
export const loginNotices = [
  "setup_concluded",
  "setup_already_concluded",
  "setup_concluded_without_answer",
  "administrator_registered",
] as const

type LoginNotice = (typeof loginNotices)[number]

/**
 * The login screen, provisional: the login comes with the second effort,
 * which replaces this screen. Until then it says so, and shows the notice it
 * was given.
 */
export function LoginScreen({ notice }: { notice?: LoginNotice }) {
  const { t } = useTranslation("access")

  return (
    <FramedPane>
      <h1 className="type-display">{t("login.heading")}</h1>
      {notice === "setup_concluded" ? (
        <Alert>
          <CircleCheckIcon className="text-status-done" />
          <AlertTitle>{t("login.setupConcluded.title")}</AlertTitle>
          <AlertDescription>{t("login.setupConcluded.todo")}</AlertDescription>
        </Alert>
      ) : null}
      {notice === "setup_already_concluded" ? (
        <Alert>
          <TriangleAlertIcon className="text-status-attention" />
          <AlertTitle>{t("login.setupAlreadyConcluded.title")}</AlertTitle>
          <AlertDescription>
            {t("login.setupAlreadyConcluded.todo")}
          </AlertDescription>
        </Alert>
      ) : null}
      {notice === "setup_concluded_without_answer" ? (
        <Alert>
          <TriangleAlertIcon className="text-status-attention" />
          <AlertTitle>
            {t("login.setupConcludedWithoutAnswer.title")}
          </AlertTitle>
          <AlertDescription>
            {t("login.setupConcludedWithoutAnswer.todo")}
          </AlertDescription>
        </Alert>
      ) : null}
      {notice === "administrator_registered" ? (
        <Alert>
          <CircleCheckIcon className="text-status-done" />
          <AlertTitle>{t("login.administratorRegistered.title")}</AlertTitle>
          <AlertDescription>
            {t("login.administratorRegistered.todo")}
          </AlertDescription>
        </Alert>
      ) : null}
      <p className="text-muted-foreground">{t("login.notBuilt")}</p>
    </FramedPane>
  )
}
