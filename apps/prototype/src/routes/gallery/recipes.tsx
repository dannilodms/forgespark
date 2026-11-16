import { createFileRoute } from "@tanstack/react-router"

import { Recipes } from "../../gallery/recipes"

export const Route = createFileRoute("/gallery/recipes")({
  component: Recipes,
})
