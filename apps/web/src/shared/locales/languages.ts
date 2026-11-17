// The one exception to "only the data module uses the generated client" (see
// Dependency rule in CODING_STANDARDS.md): the list of supported languages is
// the backend's, and reaches the app through the contract.
import { SetupAdministratorPreferredLanguage } from "../api/generated/model/setupAdministratorPreferredLanguage.ts"

/**
 * A supported language, by its BCP 47 tag. The single list (ADR-0005) is
 * `supportedLanguages` in `apps/api/internal/kernel/language.go`: the
 * contract carries it as the values of `preferred_language`, and orval
 * writes them into the generated client.
 */
export type Language = SetupAdministratorPreferredLanguage

/**
 * The name of each language in that language, which the backend cannot
 * know. A language on one side only fails the type check.
 */
const names = {
  "en-US": "English",
  es: "Español",
  "pt-BR": "Português (Brasil)",
} satisfies Record<Language, string>

/**
 * The supported languages, with their names. The texts the app loads and
 * the missing-key check both read this list, so a language added in Go
 * fails the build until it has a name here and a file, `<code>.json`, in
 * every `locales` folder.
 */
export const supportedLanguages = Object.values(
  SetupAdministratorPreferredLanguage
).map((code) => ({ code, name: names[code] }))

/** English: its file is the source of truth for the keys. */
export const defaultLanguage: Language = "en-US"
