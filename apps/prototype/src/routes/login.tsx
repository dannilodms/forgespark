import { createFileRoute } from "@tanstack/react-router"

import { LoginScreen } from "../login/login-screen"

export const Route = createFileRoute("/login")({ component: LoginScreen })
