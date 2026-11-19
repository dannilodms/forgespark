import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { LoginScreen, loginNotices } from "../access"

export const Route = createFileRoute("/login")({
  validateSearch: z.object({
    // What the screen the person comes from has to tell: `setup_concluded`
    // after concluding the setup, `session_ended` after a session ended
    // while they worked, and the others of `loginNotices`. A value it does
    // not know is no notice.
    notice: z.enum(loginNotices).optional().catch(undefined),
    // The address to return to after the login, which the gate opens when
    // it is a path of the app.
    return_to: z.string().optional().catch(undefined),
  }),
  component: Login,
})

function Login() {
  const { notice, return_to: returnTo } = Route.useSearch()
  return <LoginScreen notice={notice} returnTo={returnTo} />
}
