import i18n from "i18next"
import { initReactI18next, useTranslation } from "react-i18next"

import { type Language, defaultLanguage, supportedLanguages } from "./languages"

/** The texts of one namespace in one language: a text, or a group of texts. */
type Texts = Record<string, unknown>

/**
 * Every translation file of the app, by its path from `src`:
 * `./<namespace>/locales/<language>.json`. The namespace is the domain area,
 * or `shared`.
 */
type TranslationFiles = Record<string, Texts>

/** The supported language closest to what the browser asks for, or English. */
function closestLanguage(wanted: readonly string[]): Language {
  for (const tag of wanted) {
    const exact = supportedLanguages.find(({ code }) => code === tag)
    if (exact) {
      return exact.code
    }
    const primary = tag.split("-")[0]
    const close = supportedLanguages.find(
      ({ code }) => code.split("-")[0] === primary
    )
    if (close) {
      return close.code
    }
  }
  return defaultLanguage
}

/** The texts of every supported language, by namespace. */
function resourcesOf(files: TranslationFiles) {
  const resources: Record<string, Record<string, Texts>> = {}
  for (const { code } of supportedLanguages) {
    resources[code] = {}
  }
  for (const [path, texts] of Object.entries(files)) {
    const [, namespace, , file] = path.split("/")
    const language = file?.replace(/\.json$/, "") ?? ""
    const byNamespace = resources[language]
    if (namespace && byNamespace) {
      byNamespace[namespace] = texts
    }
  }
  return resources
}

/**
 * The language of the selector, which the screens before the login show:
 * the one closest to what the browser asks for, until the person chooses
 * another. The preferred language of a logged-in user does not change it, so
 * logging out returns to it.
 */
let selectedLanguage: Language = defaultLanguage

/**
 * Loads the texts and puts the screen in the supported language closest to
 * `wanted`, the languages the browser asks for. Call before the first
 * render. A later call only changes the language.
 */
export function startLanguage(
  wanted: readonly string[],
  files: TranslationFiles
) {
  const language = closestLanguage(wanted)
  selectedLanguage = language
  document.documentElement.lang = language
  if (i18n.isInitialized) {
    void i18n.changeLanguage(language)
    return
  }
  i18n.on("languageChanged", (code) => {
    document.documentElement.lang = code
  })
  const resources = resourcesOf(files)
  void i18n.use(initReactI18next).init({
    resources,
    lng: language,
    fallbackLng: defaultLanguage,
    defaultNS: "shared",
    ns: Object.keys(resources[defaultLanguage] ?? {}),
    interpolation: { escapeValue: false },
    initAsync: false,
  })
  // `{{max_size, fileSize}}`: a size in bytes, in megabytes, as the language
  // writes them (Values in CODING_STANDARDS.md).
  i18n.services.formatter?.add("fileSize", (bytes: number, code) =>
    new Intl.NumberFormat(code, {
      style: "unit",
      unit: "megabyte",
      maximumFractionDigits: 1,
    }).format(bytes / 2 ** 20)
  )
  // `{{wait, seconds}}`: a number of seconds, in words, as the language
  // writes them: "1 second", "42 segundos".
  i18n.services.formatter?.add("seconds", (seconds: number, code) =>
    new Intl.NumberFormat(code, {
      style: "unit",
      unit: "second",
      unitDisplay: "long",
    }).format(seconds)
  )
}

/**
 * The language the screen is in, and a way to change it at once: the
 * selector's, which is kept for when a user logs out.
 */
export function useLanguage() {
  const { i18n: instance } = useTranslation()
  return {
    language: instance.language as Language,
    choose: (language: Language) => {
      selectedLanguage = language
      void instance.changeLanguage(language)
    },
  }
}

/** Changes every text to `language`, unless the screen is already in it. */
function speak(language: Language) {
  if (i18n.language !== language) {
    void i18n.changeLanguage(language)
  }
}

/**
 * Puts the screen in the preferred language of the user logged in, whatever
 * the browser asks for or the selector shows. The selector keeps its own.
 */
export function speakThePreferredLanguage(language: Language) {
  speak(language)
}

/**
 * Puts the screen back in the language of the selector, as nobody is logged
 * in: after a logout, or a session that ended.
 */
export function speakTheSelectedLanguage() {
  speak(selectedLanguage)
}
