import { Outlet, createRootRoute } from "@tanstack/react-router"
import { TooltipProvider } from "@workspace/ui/components/tooltip"

import { PrototypeBar } from "#components/prototype-bar"

export const Route = createRootRoute({ component: RootLayout })

/**
 * The page scrolls inside its own area and the prototype bar sits below it,
 * so the bar never covers a control of the page. A page that fills the
 * screen uses `min-h-full`, not the height of the viewport.
 *
 * The area is `relative`, so a visually hidden label, which is positioned
 * absolutely, scrolls with the page and does not make the document taller.
 *
 * The tooltip provider lets a second tooltip open at once after the first.
 */
function RootLayout() {
  return (
    <TooltipProvider>
      <div className="flex h-svh flex-col">
        <div className="relative min-h-0 flex-1 overflow-auto">
          <Outlet />
        </div>
        <div className="flex justify-center p-2">
          <PrototypeBar />
        </div>
      </div>
    </TooltipProvider>
  )
}
