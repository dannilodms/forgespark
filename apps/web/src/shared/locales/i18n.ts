import i18n from "i18next"
import { initReactI18next } from "react-i18next"

import type enUS from "./en-US.json"
import { type Language, defaultLanguage, supportedLanguages } from "./languages"

// The keys of the English file are the keys of the app: `t` accepts no other.
declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "translation"
    resources: { translation: typeof enUS }
  }
}

/** The file of each language in this folder, by its path from this file. */
const files = import.meta.glob<Record<string, unknown>>("./*.json", {
  eager: true,
  import: "default",
})

/** The texts of every supported language. A language is loaded by being in the list. */
const resources = Object.fromEntries(
  supportedLanguages.map(({ code }) => [
    code,
    { translation: files[`./${code}.json`] ?? {} },
  ])
)

/** The supported language closest to what the browser asks for, or English. */
function browserLanguage(): Language {
  for (const wanted of navigator.languages) {
    const exact = supportedLanguages.find(({ code }) => code === wanted)
    if (exact) {
      return exact.code
    }
    const primary = wanted.split("-")[0]
    const close = supportedLanguages.find(
      ({ code }) => code.split("-")[0] === primary
    )
    if (close) {
      return close.code
    }
  }
  return defaultLanguage
}

/** Loads the texts in the language of the browser. Call once, before the first render. */
export function startLanguage() {
  const language = browserLanguage()
  document.documentElement.lang = language
  void i18n.use(initReactI18next).init({
    resources,
    lng: language,
    fallbackLng: defaultLanguage,
    interpolation: { escapeValue: false },
  })
}
