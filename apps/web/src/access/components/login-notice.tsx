import { CircleCheckIcon, InfoIcon, TriangleAlertIcon } from "lucide-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

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
 * - a login through a login provider did not conclude, or the provider's
 *   verified e-mail belongs to no active user. The login through a provider
 *   shows these two (ticket 08 of the login): the e-mail comes from the
 *   login state, never from the address.
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
] as const

export type LoginNotice = (typeof loginNotices)[number]

/** The notices this screen writes: a title and what to do, in `access`. */
const texts = {
  setup_concluded: "login.notices.setupConcluded",
  setup_already_concluded: "login.notices.setupAlreadyConcluded",
  setup_concluded_without_answer: "login.notices.setupConcludedWithoutAnswer",
  administrator_registered: "login.notices.administratorRegistered",
  administrator_registered_without_answer:
    "login.notices.administratorRegisteredWithoutAnswer",
  session_ended: "login.notices.sessionEnded",
  logged_out: "login.notices.loggedOut",
} as const

/**
 * The sign of each notice, in a feedback colour: done, attention or
 * information (rule 9 of DESIGN.md: never colour alone, the title says it).
 */
const signs: Record<keyof typeof texts, ReactNode> = {
  setup_concluded: <CircleCheckIcon className="text-status-done" />,
  setup_already_concluded: (
    <TriangleAlertIcon className="text-status-attention" />
  ),
  setup_concluded_without_answer: (
    <TriangleAlertIcon className="text-status-attention" />
  ),
  administrator_registered: <CircleCheckIcon className="text-status-done" />,
  administrator_registered_without_answer: (
    <TriangleAlertIcon className="text-status-attention" />
  ),
  session_ended: <TriangleAlertIcon className="text-status-attention" />,
  logged_out: <InfoIcon className="text-status-working" />,
}

function isWrittenHere(notice: LoginNotice): notice is keyof typeof texts {
  return Object.hasOwn(texts, notice)
}

/** The notice the screen was given, above the ways to log in. */
export function LoginNoticeAlert({ notice }: { notice?: LoginNotice }) {
  const { t } = useTranslation("access")
  if (notice === undefined || !isWrittenHere(notice)) {
    return null
  }

  return (
    <Alert>
      {signs[notice]}
      <AlertTitle>{t(`${texts[notice]}.title`)}</AlertTitle>
      <AlertDescription>{t(`${texts[notice]}.todo`)}</AlertDescription>
    </Alert>
  )
}
