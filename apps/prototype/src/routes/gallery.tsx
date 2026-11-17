import { Link, Outlet, createFileRoute } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { gallerySections } from "../gallery/sections"

export const Route = createFileRoute("/gallery")({ component: GalleryLayout })

/** The shell of the gallery: the navigation beside the route, or above it on a narrow screen. */
function GalleryLayout() {
  const { t } = useTranslation()

  return (
    <div className="flex h-full flex-col md:flex-row">
      <nav
        aria-label={t("gallery.navigation.label")}
        className="flex shrink-0 items-center gap-1 overflow-x-auto border-b bg-sidebar p-2 md:w-48 md:flex-col md:items-stretch md:overflow-y-auto md:border-r md:border-b-0 md:p-3"
      >
        <Link
          to="/"
          className="rounded-md px-2 py-1 type-label whitespace-nowrap text-muted-foreground hover:bg-accent hover:text-foreground md:mb-2"
        >
          {t("gallery.navigation.prototype")}
        </Link>
        {gallerySections.map((section) => (
          <Link
            key={section.to}
            to={section.to}
            className="rounded-md px-2 py-1 type-ui whitespace-nowrap text-muted-foreground hover:bg-accent hover:text-foreground"
            activeProps={{ className: "bg-selected text-foreground" }}
          >
            {t(`gallery.navigation.sections.${section.name}`)}
          </Link>
        ))}
      </nav>
      {/* `relative` keeps a visually hidden label inside the scrolling area. */}
      <div className="relative min-h-0 min-w-0 flex-1 overflow-auto">
        <Outlet />
      </div>
    </div>
  )
}
