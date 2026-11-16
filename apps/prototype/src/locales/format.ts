import type { Language } from "./languages"

/**
 * Dates and numbers written for a language, by the browser's `Intl`. A screen
 * and a block of the gallery write none by hand.
 */
export function formatFor(language: Language) {
  return {
    /** A number, with as many decimal places as asked for. */
    number: (value: number, decimals = 0) =>
      new Intl.NumberFormat(language, {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals,
      }).format(value),

    /** A day such as `2026-11-02`. */
    day: (day: string) =>
      new Intl.DateTimeFormat(language, { dateStyle: "medium" }).format(
        new Date(`${day}T00:00:00`)
      ),

    /** How long ago something happened, such as "12 days ago". */
    ago: (amount: number, unit: "minute" | "day") =>
      new Intl.RelativeTimeFormat(language).format(-amount, unit),
  }
}
