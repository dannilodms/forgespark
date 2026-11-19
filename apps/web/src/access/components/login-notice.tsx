import { CircleCheckIcon, InfoIcon, TriangleAlertIcon } from "lucide-react"
import type { ReactNode } from "react"
import { Trans, useTranslation } from "react-i18next"

import { Alert, AlertDescription, AlertTitle } from "#ui/components/alert"

/**
 * What the login screen may be told by the screen or the request the person
 * comes from, in the search parameter `notice`:
 *
 * - the setup was concluded, by them, by someone before them, or by one of
 *   the two after an attempt of theirs got no answer;
 * - a new administrator was registered, the way back in, or may have been,
 *   by an attempt of theirs that got no answer;
 * - their session ended while they worked, or they logged out;
 * - a login through a login provider did not conclude, the provider's
 *   verified e-mail belongs to no active user, or the account is not the one
 *   linked to the user who has that e-mail. The callback of the provider
 *   sends these three; the e-mail comes from the login state, never from the
 *   address.
 */
export const loginNotices = [
  "setup_concluded",
  "setup_already_concluded",
  "setup_concluded_without_answer",
  "administrator_registered",
  "administrator_registered_without_answer",
  "session_ended",
  "logged_out",
  "login_not_concluded",
  "no_active_user",
  "another_account_linked",
] as const

export type LoginNotice = (typeof loginNotices)[number]

const done = <CircleCheckIcon className="text-status-done" />
const attention = <TriangleAlertIcon className="text-status-attention" />
const information = <InfoIcon className="text-status-working" />

/**
 * What each notice shows: its texts, a title and what to do, in `access`,
 * and its sign, in a feedback colour: done, attention or information (rule 9
 * of DESIGN.md: never colour alone, the title says it).
 */
const notices = {
  setup_concluded: { texts: "login.notices.setupConcluded", sign: done },
  setup_already_concluded: {
    texts: "login.notices.setupAlreadyConcluded",
    sign: attention,
  },
  setup_concluded_without_answer: {
    texts: "login.notices.setupConcludedWithoutAnswer",
    sign: attention,
  },
  administrator_registered: {
    texts: "login.notices.administratorRegistered",
    sign: done,
  },
  administrator_registered_without_answer: {
    texts: "login.notices.administratorRegisteredWithoutAnswer",
    sign: attention,
  },
  session_ended: { texts: "login.notices.sessionEnded", sign: attention },
  logged_out: { texts: "login.notices.loggedOut", sign: information },
  login_not_concluded: {
    texts: "login.notices.loginNotConcluded",
    sign: attention,
  },
  no_active_user: { texts: "login.notices.noActiveUser", sign: attention },
  another_account_linked: {
    texts: "login.notices.anotherAccountLinked",
    sign: attention,
  },
} as const satisfies Record<LoginNotice, { texts: string; sign: ReactNode }>

/**
 * The notice the screen was given, above the ways to log in. The notice
 * `no_active_user` shows the e-mail the last login through a provider in
 * this browser was refused for, as the login state tells it; the same text
 * whether the e-mail is unknown or of a deactivated user, which the API does
 * not tell apart. With no refusal in the login state, the address was not
 * the one the refusal led to, and the notice has nothing to say.
 */
export function LoginNoticeAlert({
  notice,
  refusedEmail,
}: {
  notice?: LoginNotice
  /** The e-mail of the refused login, from the login state. */
  refusedEmail?: string
}) {
  const { t } = useTranslation("access")
  if (
    notice === undefined ||
    (notice === "no_active_user" && refusedEmail === undefined)
  ) {
    return null
  }

  return (
    <Alert>
      {notices[notice].sign}
      <AlertTitle>{t(`${notices[notice].texts}.title`)}</AlertTitle>
      <AlertDescription>
        <Trans
          t={t}
          i18nKey={`${notices[notice].texts}.todo`}
          values={{ email: refusedEmail }}
          // The e-mail goes whole to the next line, and breaks only where it
          // is longer than the line.
          components={{ email: <strong className="wrap-anywhere" /> }}
        />
      </AlertDescription>
    </Alert>
  )
}
