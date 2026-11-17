/** The texts of one language: a text, or a group of texts, under each key. */
export type Texts = { [key: string]: string | Texts }

/**
 * The keys of `source` that `translation` lacks, each by its whole path, such
 * as `login.code.label`.
 */
export function missingKeys(source: Texts, translation: Texts): string[] {
  return Object.entries(source).flatMap(([key, text]) => {
    const translated = translation[key]
    if (typeof text === "string") {
      return typeof translated === "string" ? [] : [key]
    }
    const group = typeof translated === "object" ? translated : {}
    return missingKeys(text, group).map((path) => `${key}.${path}`)
  })
}
