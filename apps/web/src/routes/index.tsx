import { createFileRoute } from "@tanstack/react-router"

import { LandingScreen } from "../access"

// Where a login leads when there is no address to return to.
export const Route = createFileRoute("/")({ component: LandingScreen })
