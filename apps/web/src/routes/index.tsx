import { createFileRoute } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

export const Route = createFileRoute("/")({ component: Home })

/** What the app shows until its first screen is built. */
function Home() {
  const { t } = useTranslation()

  return (
    <main className="mx-auto flex min-h-svh max-w-xl flex-col justify-center gap-2 p-6">
      <h1 className="type-display">{t("home.title")}</h1>
      <p className="text-muted-foreground">{t("home.intro")}</p>
    </main>
  )
}
