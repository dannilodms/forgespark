import { createFileRoute } from "@tanstack/react-router"

import { loginScenarioGroups } from "../login/fake-login"
import { LoginScreen } from "../login/login-screen"

export const Route = createFileRoute("/login")({
  component: LoginScreen,
  // The prototype bar lists these on this route.
  staticData: { scenarioGroups: loginScenarioGroups },
})
