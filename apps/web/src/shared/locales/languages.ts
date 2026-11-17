/**
 * The supported languages: the single list (ADR-0005). The texts the app
 * loads and the missing-key check both read it.
 *
 * Adding a language means adding it here and delivering its file,
 * `<code>.json`, in every `locales` folder.
 */
export const supportedLanguages = [
  { code: "en-US", name: "English" },
  { code: "pt-BR", name: "Português (Brasil)" },
  { code: "es", name: "Español" },
] as const

export type Language = (typeof supportedLanguages)[number]["code"]

/** English: its file is the source of truth for the keys. */
export const defaultLanguage: Language = "en-US"
