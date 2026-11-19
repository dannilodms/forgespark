import { useNavigate } from "@tanstack/react-router"
import { useReducer } from "react"
import { useTranslation } from "react-i18next"

import { FramedPane } from "../../shared/components/framed-pane"
import { CodeStep } from "../components/code-step"
import { EmailStep } from "../components/email-step"
import { type LoginNotice, LoginNoticeAlert } from "../components/login-notice"
import { whereTheLoginLeads } from "../data/session"
import { loginReducer, startOfTheLogin } from "./login-flow"

/**
 * The login screen, in the narrow framed pane: the notice it was given, the
 * buttons of the login providers above a divider, and the login with a
 * login code, whose steps are one reducer. Once a user is logged in, the
 * app opens the address to return to, when it is a path of the app, or the
 * root.
 */
export function LoginScreen({
  notice,
  returnTo,
}: {
  notice?: LoginNotice
  returnTo?: string
}) {
  const { t } = useTranslation("access")
  const navigate = useNavigate()
  const [flow, dispatch] = useReducer(loginReducer, startOfTheLogin)

  return (
    <FramedPane>
      <h1 className="type-display">{t("login.heading")}</h1>
      <LoginNoticeAlert notice={notice} />
      {/*
        The buttons of the login providers turned on, and the divider under
        them, go here (ticket 08 of the login). With no provider on, both
        are absent, as now.
      */}
      {flow.step === "code" ? (
        <CodeStep
          email={flow.email}
          askedAt={flow.askedAt}
          onNewCode={() => dispatch({ type: "codeSent", at: Date.now() })}
          onEnded={() => dispatch({ type: "codeEnded" })}
          onBack={() => dispatch({ type: "wentBack" })}
          onLoggedIn={() =>
            void navigate({ href: whereTheLoginLeads(returnTo), replace: true })
          }
        />
      ) : (
        // One element for both steps, so that the form outlives the ask.
        <EmailStep
          email={flow.email}
          sending={flow.step === "sending"}
          codeEnded={flow.step === "email" && flow.codeEnded}
          onSending={(email) => dispatch({ type: "codeAskedFor", email })}
          onSent={() => dispatch({ type: "codeSent", at: Date.now() })}
          onRefused={() => dispatch({ type: "askRefused" })}
        />
      )}
    </FramedPane>
  )
}
