import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { Button } from "@workspace/ui/components/button"
import { useTranslation } from "react-i18next"

export const Route = createFileRoute("/signed-in")({ component: SignedIn })

/**
 * Where the login screen leads for now. It is replaced by the sample work
 * screen when that screen exists.
 */
function SignedIn() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  return (
    <main className="mx-auto flex max-w-xl flex-col gap-4 p-6">
      <h1 className="type-display">{t("signedIn.heading")}</h1>
      <p className="text-muted-foreground">{t("signedIn.body")}</p>
      <div>
        <Button
          variant="secondary"
          onClick={() => void navigate({ to: "/login" })}
        >
          {t("signedIn.logOut")}
        </Button>
      </div>
    </main>
  )
}
