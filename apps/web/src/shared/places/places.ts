import { currencyOfCountry } from "./currency-of-country"
import { countries, currencies, timeZones } from "./lists.gen"

// The countries, the currencies and the time zones the API accepts, named
// by the browser in the language of the screen (Values in
// CODING_STANDARDS.md). The codes are the API's own lists, written into
// `lists.gen.ts` by `task api:lists`: a code the browser happens to know and
// the API does not is never offered.

/** One entry of a list: the code the API takes, and its name on the screen. */
export type ListItem = { value: string; label: string }

const lists = new Map<string, ListItem[]>()

/** A list, built once for each language: naming 600 time zones takes time. */
function cached(key: string, build: () => ListItem[]) {
  const known = lists.get(key)
  if (known) {
    return known
  }
  const built = build()
  lists.set(key, built)
  return built
}

/** The items in the order of their names, as the language sorts them. */
function byName(items: ListItem[], language: string) {
  const collator = new Intl.Collator(language)
  return items.sort((one, other) => collator.compare(one.label, other.label))
}

/** The countries, by their names in `language`. */
export function countryItems(language: string): ListItem[] {
  return cached(`countries:${language}`, () => {
    const names = new Intl.DisplayNames([language], {
      type: "region",
      fallback: "code",
    })
    return byName(
      countries.map((code) => ({ value: code, label: names.of(code) ?? code })),
      language
    )
  })
}

/**
 * The currencies, by their names in `language`, each followed by its code:
 * several currencies share a name ("dollar") in some languages.
 */
export function currencyItems(language: string): ListItem[] {
  return cached(`currencies:${language}`, () => {
    const names = new Intl.DisplayNames([language], {
      type: "currency",
      fallback: "code",
    })
    return byName(
      currencies.map((code) => {
        const name = names.of(code) ?? code
        return {
          value: code,
          label: name === code ? code : `${name} (${code})`,
        }
      }),
      language
    )
  })
}

/**
 * The name of a time zone in `language`, such as "Brasilia Time": the
 * browser has no list of the names of time zones, but writes the one of each
 * zone in a date. A zone the browser does not know has none.
 */
function timeZoneName(timeZone: string, language: string) {
  try {
    return new Intl.DateTimeFormat(language, {
      timeZone,
      timeZoneName: "longGeneric",
    })
      .formatToParts()
      .find((part) => part.type === "timeZoneName")?.value
  } catch {
    return undefined
  }
}

/**
 * The time zones, in the order of their names in the IANA database, each
 * written as the database writes it, with spaces for underscores, and
 * followed by its name in `language` when the browser knows one:
 * "America/Sao Paulo (Brasilia Time)".
 */
export function timeZoneItems(language: string): ListItem[] {
  return cached(`timeZones:${language}`, () =>
    timeZones.map((timeZone) => {
      const place = timeZone.replaceAll("_", " ")
      const name = timeZoneName(timeZone, language)
      return { value: timeZone, label: name ? `${place} (${name})` : place }
    })
  )
}

/**
 * The time zone of the browser, when the API accepts it. Otherwise nothing,
 * and the person chooses one.
 */
export function timeZoneOfTheBrowser(): string {
  const { timeZone } = Intl.DateTimeFormat().resolvedOptions()
  return (timeZones as readonly string[]).includes(timeZone) ? timeZone : ""
}

/** The currency a country uses, or nothing for one that has none. */
export function currencyOf(country: string): string {
  return Object.hasOwn(currencyOfCountry, country)
    ? currencyOfCountry[country as keyof typeof currencyOfCountry]
    : ""
}
