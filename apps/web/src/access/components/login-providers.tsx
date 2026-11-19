import { useTranslation } from "react-i18next"

import { buttonVariants } from "#ui/components/button"
import { Separator } from "#ui/components/separator"
import { cn } from "#ui/lib/utils"

import { type LoginProvider, providerLoginAddress } from "../data/session"

/**
 * The buttons of the login providers turned on, one each with its display
 * name, above the divider that sets them apart from the login with a login
 * code. Each is a link, because the browser leaves the app: it goes to the
 * start of the login through that provider, which sends it to the provider,
 * and the provider brings it back logged in, at the address to return to.
 * With no provider on, nothing is drawn, the divider included.
 */
export function LoginProviders({
  providers,
  returnTo,
}: {
  providers: readonly LoginProvider[]
  /** The address to return to once logged in, as the login was given it. */
  returnTo?: string
}) {
  const { t } = useTranslation("access")
  if (providers.length === 0) {
    return null
  }

  return (
    <>
      <div className="flex flex-col gap-3">
        {providers.map((provider) => (
          <a
            key={provider.key}
            href={providerLoginAddress(provider.key, returnTo)}
            // A long name in Portuguese or Spanish wraps inside the button.
            className={cn(
              buttonVariants({ variant: "outline" }),
              "h-auto min-h-10 w-full py-2 text-center whitespace-normal"
            )}
          >
            {t("login.providers.logInWith", { name: provider.display_name })}
          </a>
        ))}
      </div>
      <div className="relative flex justify-center">
        <Separator className="absolute inset-x-0 top-1/2" />
        <span className="relative bg-card px-2 type-caption text-muted-foreground">
          {t("login.providers.or")}
        </span>
      </div>
    </>
  )
}
