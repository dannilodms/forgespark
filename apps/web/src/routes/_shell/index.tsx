import { createFileRoute } from "@tanstack/react-router"

import { StartScreen } from "../../consultancy"

// Where a login leads when there is no address to return to.
export const Route = createFileRoute("/_shell/")({ component: StartScreen })
