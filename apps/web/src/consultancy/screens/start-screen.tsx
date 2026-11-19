import { useTranslation } from "react-i18next"

/**
 * Where a login lands when there is no address to return to: the work
 * surface of the shell, which says that the areas of work arrive in the
 * versions that come next, so an empty page does not look broken.
 */
export function StartScreen() {
  const { t } = useTranslation("consultancy")

  return (
    <div className="flex h-full items-center justify-center p-6 sm:p-8">
      <p className="max-w-prose text-center type-body text-muted-foreground">
        {t("start.areasOfWorkComeNext")}
      </p>
    </div>
  )
}
