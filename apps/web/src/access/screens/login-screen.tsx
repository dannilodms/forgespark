import { useSuspenseQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useReducer } from "react"
import { useTranslation } from "react-i18next"

import { FramedPane } from "../../shared/components/framed-pane"
import { CodeStep } from "../components/code-step"
import { EmailStep } from "../components/email-step"
import { type LoginNotice, LoginNoticeAlert } from "../components/login-notice"
import { LoginProviders } from "../components/login-providers"
import { loginStateQuery, whereTheLoginLeads } from "../data/session"
import { loginReducer, startOfTheLogin } from "./login-flow"

/**
 * The login screen, in the narrow framed pane: the notice it was given, the
 * buttons of the login providers turned on above a divider, and the login
 * with a login code, whose steps are one reducer. The providers and the
 * e-mail of a refused login come from the login state, which the route
 * reads before the screen opens. Once a user is logged in, the app opens
 * the address to return to, when it is a path of the app, or the root; a
 * login through a provider returns there by the API.
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
  const { data: loginState } = useSuspenseQuery(loginStateQuery)
  const [flow, dispatch] = useReducer(loginReducer, startOfTheLogin)

  return (
    <FramedPane>
      <h1 className="text-center type-display">{t("login.heading")}</h1>
      <LoginNoticeAlert
        notice={notice}
        refusedEmail={loginState.refused_email}
      />
      <LoginProviders
        providers={loginState.login_providers}
        returnTo={returnTo}
      />
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
