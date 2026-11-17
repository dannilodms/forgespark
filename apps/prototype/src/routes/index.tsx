import { Link, createFileRoute } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

export const Route = createFileRoute("/")({ component: Home })

/** Where the prototype leads. The name of each is its key in the translation files. */
const destinations = [
  { to: "/login", name: "login" },
  { to: "/work", name: "work" },
  { to: "/gallery", name: "gallery" },
] as const

function Home() {
  const { t } = useTranslation()

  return (
    <main className="mx-auto flex max-w-xl flex-col gap-6 p-6">
      <header className="flex flex-col gap-2">
        <h1 className="type-display">{t("prototype.title")}</h1>
        <p className="text-muted-foreground">{t("prototype.intro")}</p>
      </header>
      <nav
        aria-label={t("prototype.navigation")}
        className="flex flex-col border-t"
      >
        {destinations.map((destination) => (
          <Link
            key={destination.to}
            to={destination.to}
            className="flex flex-col gap-0.5 border-b px-2 py-3 hover:bg-accent"
          >
            <span className="type-title">
              {t(`prototype.destinations.${destination.name}.label`)}
            </span>
            <span className="type-caption text-muted-foreground">
              {t(`prototype.destinations.${destination.name}.about`)}
            </span>
          </Link>
        ))}
      </nav>
    </main>
  )
}
