import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { LoginScreen, loginNotices, loginStateQuery } from "../access"

export const Route = createFileRoute("/login")({
  validateSearch: z.object({
    // What the screen the person comes from has to tell: `setup_concluded`
    // after concluding the setup, `session_ended` after a session ended
    // while they worked, `no_active_user`, `another_account_linked` and
    // `login_not_concluded` after a login through a provider, and the others
    // of `loginNotices`. A value it does not know is no notice.
    notice: z.enum(loginNotices).optional().catch(undefined),
    // The address to return to after the login, which the gate opens when
    // it is a path of the app, and the button of a provider passes on.
    return_to: z.string().optional().catch(undefined),
  }),
  // The login state, read before the screen opens: the buttons of the
  // providers and the e-mail of a refused login. The gate runs first, so a
  // logged-in user, who is taken away from the login, never reads it.
  loader: ({ context: { queryClient } }) =>
    queryClient.ensureQueryData(loginStateQuery),
  component: Login,
})

function Login() {
  const { notice, return_to: returnTo } = Route.useSearch()
  return <LoginScreen notice={notice} returnTo={returnTo} />
}
