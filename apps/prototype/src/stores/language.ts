import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import { create } from "zustand"
import { persist } from "zustand/middleware"

import type enUS from "../locales/en-US.json"
import { formatFor } from "../locales/format"
import {
  type Language,
  defaultLanguage,
  supportedLanguages,
} from "../locales/languages"

// The keys of the English file are the keys of the app: `t` accepts no other.
declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "translation"
    resources: { translation: typeof enUS }
  }
}

/** The file of each language in `locales`, by its path from this file. */
const files = import.meta.glob<Record<string, unknown>>("../locales/*.json", {
  eager: true,
  import: "default",
})

/** The texts of every supported language. A language is loaded by being in the list. */
const resources = Object.fromEntries(
  supportedLanguages.map(({ code }) => [
    code,
    { translation: files[`../locales/${code}.json`] ?? {} },
  ])
)

type LanguageState = {
  /** The language chosen by the person. `null` follows the browser. */
  chosen: Language | null
  choose: (language: Language) => void
}

export const useLanguageStore = create<LanguageState>()(
  persist(
    (set) => ({
      chosen: null,
      choose: (language) => set({ chosen: language }),
    }),
    { name: "prototype-language" }
  )
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

function currentLanguage(): Language {
  return useLanguageStore.getState().chosen ?? browserLanguage()
}

function applyLanguage() {
  const language = currentLanguage()
  document.documentElement.lang = language
  void i18n.changeLanguage(language)
}

/** Loads the texts and keeps the language in force. Call once, before the first render. */
export function startLanguage() {
  void i18n.use(initReactI18next).init({
    resources,
    lng: currentLanguage(),
    fallbackLng: defaultLanguage,
    interpolation: { escapeValue: false },
  })
  document.documentElement.lang = currentLanguage()
  useLanguageStore.subscribe(applyLanguage)
}

/** The language in force: the chosen one, or the browser's when none was chosen. */
export function useLanguage(): Language {
  const chosen = useLanguageStore((state) => state.chosen)
  return chosen ?? browserLanguage()
}

/** Dates and numbers written for the language in force. */
export function useFormat() {
  return formatFor(useLanguage())
}
