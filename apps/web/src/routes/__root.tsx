import { Outlet, createRootRoute } from "@tanstack/react-router"

import { TooltipProvider } from "#ui/components/tooltip"

export const Route = createRootRoute({ component: RootLayout })

/** The tooltip provider lets a second tooltip open at once after the first. */
function RootLayout() {
  return (
    <TooltipProvider>
      <Outlet />
    </TooltipProvider>
  )
}
