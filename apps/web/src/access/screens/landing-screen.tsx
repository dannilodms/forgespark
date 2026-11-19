import { useQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { LoaderCircleIcon, LogOutIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"

import { ErrorAlert } from "../../shared/errors/error-alert"
import { loggedInUserQuery, useLogOut } from "../data/session"

/**
 * Where a login lands, provisional: one sentence that says who is logged
 * in, and the logout. The shell of the app replaces it (ticket 06 of the
 * login), with the logout in the menu of the user. It has no language
 * selector: a logged-in user reads the app in their preferred language.
 */
export function LandingScreen() {
  const { t } = useTranslation("access")
  const navigate = useNavigate()
  const { data: user } = useQuery(loggedInUserQuery)
  const logOut = useLogOut()

  return (
    <main className="flex min-h-svh flex-col items-start gap-4 bg-background p-4">
      {user ? (
        <p className="type-body">
          {t("landing.loggedInAs", { name: user.name })}
        </p>
      ) : null}
      {logOut.error ? <ErrorAlert error={logOut.error} /> : null}
      <Button
        variant="outline"
        disabled={logOut.isPending}
        onClick={() =>
          logOut.mutate(undefined, {
            onSuccess: () =>
              void navigate({ to: "/login", search: { notice: "logged_out" } }),
          })
        }
      >
        {logOut.isPending ? (
          <LoaderCircleIcon data-icon="inline-start" className="animate-spin" />
        ) : (
          <LogOutIcon data-icon="inline-start" />
        )}
        {t("landing.logOut")}
      </Button>
    </main>
  )
}
