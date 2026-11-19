import type { QueryClient } from "@tanstack/react-query"
import {
  Navigate,
  createRootRouteWithContext,
  redirect,
} from "@tanstack/react-router"

import { loggedInUserQuery, returnPathOf, whereTheLoginLeads } from "../access"
import { leadsToTheSetup, setupStateQuery } from "../consultancy"
import { ErrorScreen } from "../shared/errors/error-screen"
import { speakIn, speakTheSelectedLanguage } from "../shared/locales/i18n"

type RouterContext = { queryClient: QueryClient }

export const Route = createRootRouteWithContext<RouterContext>()({
  // The gate: before any screen opens, it decides between the setup, the
  // login and the app. The state of the setup first: every address leads to
  // the setup until it is concluded. Then who is logged in: nobody leads to
  // the login, which keeps the address to return to; a user leads away from
  // the login, to that address or the root, and the app speaks their
  // preferred language.
  beforeLoad: async ({ context: { queryClient }, location }) => {
    const setup = await queryClient.ensureQueryData(setupStateQuery)
    const atTheSetup = location.pathname === "/setup"
    if (leadsToTheSetup(setup, atTheSetup)) {
      if (!atTheSetup) {
        throw redirect({ to: "/setup", replace: true })
      }
      return
    }
    const user = await queryClient.ensureQueryData(loggedInUserQuery)
    const atTheLogin = location.pathname === "/login"
    if (!user) {
      speakTheSelectedLanguage()
      if (!atTheLogin) {
        throw redirect({
          to: "/login",
          search: { return_to: returnPathOf(location) },
          replace: true,
        })
      }
      return
    }
    speakIn(user.preferred_language)
    if (atTheLogin || atTheSetup) {
      const { return_to: returnTo } = location.search as { return_to?: unknown }
      throw redirect({ href: whereTheLoginLeads(returnTo), replace: true })
    }
  },
  // A failure of the gate, or while a screen renders, falls here.
  errorComponent: ErrorScreen,
  // An address the app does not have: the gate let a logged-in user through,
  // and the root is where they land.
  notFoundComponent: () => <Navigate to="/" replace />,
})
