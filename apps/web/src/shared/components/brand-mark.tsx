import type { ComponentProps } from "react"
import { useTranslation } from "react-i18next"

/**
 * The logo of the product in the header of the framed pane and of the shell:
 * the mark in `brand` and the product name beside it, in title type (Brand
 * in DESIGN.md). A long name is cut where the header has no room for it.
 */
export function ProductLogo() {
  const { t } = useTranslation()

  return (
    <>
      <BrandMark className="size-7 shrink-0 text-brand" />
      <span className="min-w-0 truncate type-title">{t("productName")}</span>
    </>
  )
}

/**
 * The mark of the product, a copy of `docs/design/brand/mark.svg` in one
 * colour: the current text colour, which the logo sets to `brand`. It is
 * decoration beside the product name, so assistive technology skips it.
 */
function BrandMark(props: ComponentProps<"svg">) {
  return (
    <svg viewBox="0 0 64 64" aria-hidden="true" focusable="false" {...props}>
      <path
        fill="currentColor"
        d="M8 48L8 13L21 24L21 13L34 24L35.5 24L36.75 3L43.25 3L44.5 24L47 24L48.25 3L54.75 3L56 24L56 48zM21 42.5L23.5 40L19 35.5L23.5 31L21 28.5L14 35.5zM27.25 43.5L31.75 43.5L36.75 27.5L32.25 27.5zM50 35.5L43 28.5L40.5 31L45 35.5L40.5 40L43 42.5zM10 46H54A8 8 0 0 1 54 62H10A8 8 0 0 1 10 46zM9.5 54a3.5 3.5 0 1 0 7 0a3.5 3.5 0 1 0 -7 0zM28.5 54a3.5 3.5 0 1 0 7 0a3.5 3.5 0 1 0 -7 0zM47.5 54a3.5 3.5 0 1 0 7 0a3.5 3.5 0 1 0 -7 0z"
      />
    </svg>
  )
}
