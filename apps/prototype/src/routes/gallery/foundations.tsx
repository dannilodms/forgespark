import { createFileRoute } from "@tanstack/react-router"

import { Foundations } from "../../gallery/foundations"

export const Route = createFileRoute("/gallery/foundations")({
  component: Foundations,
})
