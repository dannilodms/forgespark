import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { LanguageSelect } from "./language-select"

/**
 * A screen of one pane, seen before anyone is logged in: the title bar with
 * the product name, the framed page with the pane in its centre, and the
 * status bar with the language selector and nothing else. It is the Framed
 * pane recipe of DESIGN.md. At a narrow width the pane takes the width of
 * the page.
 */
export function FramedPane({ children }: { children: ReactNode }) {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex h-7 shrink-0 items-center border-b bg-chrome px-3 type-caption text-muted-foreground">
        {t("productName")}
      </header>

      <main className="flex flex-1 items-center justify-center page-framing p-4">
        <div className="relative w-full max-w-sm rounded-xl border bg-card p-6">
          <span className="absolute top-0 left-0 -translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute top-0 right-0 translate-x-1/2 -translate-y-1/2 frame-node" />
          <span className="absolute bottom-0 left-0 -translate-x-1/2 translate-y-1/2 frame-node" />
          <span className="absolute right-0 bottom-0 translate-x-1/2 translate-y-1/2 frame-node" />
          <div className="flex flex-col gap-6">{children}</div>
        </div>
      </main>

      <footer className="flex h-7 shrink-0 items-center justify-end border-t bg-chrome px-2">
        <LanguageSelect />
      </footer>
    </div>
  )
}
