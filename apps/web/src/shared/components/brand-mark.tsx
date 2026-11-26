import type { ComponentProps } from "react"
import { useTranslation } from "react-i18next"

/**
 * The logo of the product in the header of the framed pane and of the shell:
 * the mark in `brand` and the product name beside it, in title type (Brand
 * in DESIGN.md). With `withSlogan`, the slogan goes under the name, in
 * caption type, as the framed pane shows it. A long name or slogan is cut
 * where the header has no room for it.
 */
export function ProductLogo({ withSlogan = false }: { withSlogan?: boolean }) {
  const { t } = useTranslation()
  const name = (
    <span className="min-w-0 truncate type-title">{t("productName")}</span>
  )

  return (
    <>
      <BrandMark className="size-7 shrink-0 text-brand" />
      {withSlogan ? (
        <span className="flex min-w-0 flex-col">
          {name}
          <span className="min-w-0 truncate type-caption text-muted-foreground">
            {t("productSlogan")}
          </span>
        </span>
      ) : (
        name
      )}
    </>
  )
}

/**
 * The mark of the product, a copy of `docs/design/brand/mark.svg` in one
 * colour: the current text colour, which the logo sets to `brand`. The spark
 * is a hole in the nut, drawn the other way round. It is decoration beside the
 * product name, so assistive technology skips it.
 */
function BrandMark(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 64 64" aria-hidden="true" focusable="false" {...props}>
      <path
        fill="currentColor"
        d="M59 26.8Q62 32 59 37.2L50 52.78Q47 57.98 41 57.98L23 57.98Q17 57.98 14 52.78L5 37.2Q2 32 5 26.8L14 11.22Q17 6.02 23 6.02L41 6.02Q47 6.02 50 11.22ZM32 15Q28.5 28.5 19 32Q28.5 35.5 32 49Q35.5 35.5 45 32Q35.5 28.5 32 15Z"
      />
    </svg>
  )
}
