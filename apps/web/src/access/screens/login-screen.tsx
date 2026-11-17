import { CircleCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"

import { FramedPane } from "../../shared/components/framed-pane"

/** What the login screen may be told by the screen the person comes from. */
export const loginNotices = ["setup_concluded"] as const

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
      <p className="text-muted-foreground">{t("login.notBuilt")}</p>
    </FramedPane>
  )
}
