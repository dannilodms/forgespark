import { Outlet, createRootRoute } from "@tanstack/react-router"

import { PrototypeBar } from "#components/prototype-bar"

export const Route = createRootRoute({ component: RootLayout })

/**
 * The page scrolls inside its own area and the prototype bar sits below it,
 * so the bar never covers a control of the page. A page that fills the
 * screen uses `min-h-full`, not the height of the viewport.
 */
function RootLayout() {
  return (
    <div className="flex h-svh flex-col">
      <div className="min-h-0 flex-1 overflow-auto">
        <Outlet />
      </div>
      <div className="flex justify-center p-2">
        <PrototypeBar />
      </div>
    </div>
  )
}
