import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

import { LoginScreen, loginNotices } from "../access"

export const Route = createFileRoute("/login")({
  // What the screen the person comes from has to tell: `setup_concluded`
  // after concluding the setup, `administrator_registered` after the way
  // back in.
  validateSearch: z.object({ notice: z.enum(loginNotices).optional() }),
  component: Login,
})

function Login() {
  const { notice } = Route.useSearch()
  return <LoginScreen notice={notice} />
}
