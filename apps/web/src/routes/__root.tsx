import type { QueryClient } from "@tanstack/react-query"
import { createRootRouteWithContext, redirect } from "@tanstack/react-router"

import { leadsToTheSetup, setupStateQuery } from "../consultancy"
import { ErrorScreen } from "../shared/errors/error-screen"

type RouterContext = { queryClient: QueryClient }

export const Route = createRootRouteWithContext<RouterContext>()({
  // The gate: before any screen opens, the state of the setup decides
  // between the setup and the login screen, the only two that exist.
  beforeLoad: async ({ context, location }) => {
    const setup = await context.queryClient.ensureQueryData(setupStateQuery)
    const destination = leadsToTheSetup(setup, location.pathname === "/setup")
      ? "/setup"
      : "/login"
    if (location.pathname !== destination) {
      throw redirect({ to: destination, replace: true })
    }
  },
  // A failure of the gate, or while a screen renders, falls here.
  errorComponent: ErrorScreen,
})
