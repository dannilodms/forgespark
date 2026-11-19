import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import {
  Sidebar,
  SidebarHeader,
  SidebarProvider,
  SidebarTrigger,
  useSidebar,
} from "#ui/components/sidebar"

import { BrandMark } from "./brand-mark"

/**
 * Every screen after the login: the header and the sidebar on the frame, and
 * the work surface beside the sidebar, which holds the screen. It is the App
 * shell recipe of DESIGN.md, cut down to what works: the header has the mark,
 * the product name and the menu of the user; the sidebar has the consultancy.
 * Under 960px the sidebar is a sheet, which a button at the start of the
 * header opens. The shell takes the height of the window and never scrolls:
 * the work surface scrolls inside itself.
 */
export function AppShell({
  consultancy,
  userMenu,
  children,
}: {
  /** What the top of the sidebar shows: the consultancy. */
  consultancy: ReactNode
  /** What the end of the header shows: the menu of the user. */
  userMenu: ReactNode
  /** The screen, on the work surface. */
  children: ReactNode
}) {
  const { t } = useTranslation()

  return (
    <SidebarProvider className="h-svh min-h-0 flex-col overflow-hidden bg-background [--header-height:--spacing(14)]">
      <header className="flex h-(--header-height) shrink-0 items-center gap-2.5 px-4">
        <SheetButton label={t("shell.openSidebar")} />
        <BrandMark className="size-7 shrink-0 text-brand" />
        <span className="min-w-0 truncate font-heading text-lg font-semibold">
          {t("productName")}
        </span>
        <div className="ml-auto flex shrink-0 items-center">{userMenu}</div>
      </header>

      <div className="flex min-h-0 flex-1">
        <Sidebar
          variant="inset"
          label={t("shell.sidebar")}
          className="top-(--header-height) h-[calc(100svh-var(--header-height))]!"
        >
          <SidebarHeader>{consultancy}</SidebarHeader>
        </Sidebar>
        <main className="mr-3 mb-3 ml-3 min-w-0 flex-1 overflow-auto rounded-(--radius-surface) border bg-card shell:ml-0">
          {children}
        </main>
      </div>
    </SidebarProvider>
  )
}

/** The button that opens the sidebar, where it is a sheet. */
function SheetButton({ label }: { label: string }) {
  const { isMobile } = useSidebar()
  if (!isMobile) {
    return null
  }
  return <SidebarTrigger aria-label={label} className="-ml-1.5" />
}
