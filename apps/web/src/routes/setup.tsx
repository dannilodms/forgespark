import { createFileRoute } from "@tanstack/react-router"

import { SetupScreen } from "../consultancy"

// The address `issue-setup-code` prints: APP_ORIGIN followed by /setup.
export const Route = createFileRoute("/setup")({ component: SetupScreen })
