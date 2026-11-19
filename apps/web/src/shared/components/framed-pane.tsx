import type { ReactNode } from "react"

import { cn } from "#ui/lib/utils"

import { ProductLogo } from "./brand-mark"
import { LanguageSelect } from "./language-select"

/**
 * A screen of one pane, seen before anyone is logged in: the header with the
 * mark and the product name, the frame with the pane in its centre, and the
 * footer with the language selector and nothing else. It is the Framed pane
 * recipe of DESIGN.md. The pane is narrow for a screen of few controls, such
 * as the login screen, and wide for a form, such as the setup, whose fields
 * go side by side. At a narrow width the pane takes the width of the page.
 */
export function FramedPane({
  width = "narrow",
  children,
}: {
  width?: "narrow" | "wide"
  children: ReactNode
}) {
  return (
    <div className="flex min-h-svh flex-col bg-background">
      <header className="flex h-14 shrink-0 items-center gap-2.5 px-4">
        <ProductLogo />
      </header>

      <main className="flex flex-1 items-center justify-center px-4 py-2">
        <div
          className={cn(
            "w-full rounded-(--radius-surface) border bg-card p-6 sm:p-8",
            width === "wide" ? "max-w-180" : "max-w-sm"
          )}
        >
          <div className="flex flex-col gap-6">{children}</div>
        </div>
      </main>

      <footer className="flex h-14 shrink-0 items-center justify-end px-4">
        <LanguageSelect />
      </footer>
    </div>
  )
}
