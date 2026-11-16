// Fails when a supported language is missing a key that English has, and
// names each missing key. The English file is the source of truth (ADR-0005).
//
// The languages come from the single list in src/locales/languages.ts, so a
// language added there without its file fails here too.
//
// Run from anywhere: `vp node apps/prototype/scripts/missing-keys.mjs`.
import { existsSync, readFileSync } from "node:fs"
import { join, resolve } from "node:path"

import {
  defaultLanguage,
  supportedLanguages,
} from "../src/locales/languages.ts"
import { missingKeys } from "../src/locales/missing-keys.ts"

const localesDir = resolve(import.meta.dirname, "../src/locales")

/** The texts of a language, or `null` when it has no file. */
function textsOf(code) {
  const file = join(localesDir, `${code}.json`)
  return existsSync(file) ? JSON.parse(readFileSync(file, "utf8")) : null
}

const source = textsOf(defaultLanguage)
if (!source) {
  console.error(`${defaultLanguage} has no file in apps/prototype/src/locales.`)
  process.exit(1)
}

const translations = supportedLanguages
  .map(({ code }) => code)
  .filter((code) => code !== defaultLanguage)

let failed = false

for (const code of translations) {
  const texts = textsOf(code)
  if (!texts) {
    console.error(`${code} has no file in apps/prototype/src/locales.`)
    failed = true
    continue
  }
  const missing = missingKeys(source, texts)
  if (missing.length > 0) {
    console.error(
      `${code} is missing ${missing.length} of the keys of ${defaultLanguage}:\n` +
        missing.map((key) => `  ${key}`).join("\n")
    )
    failed = true
  }
}

if (failed) {
  process.exit(1)
}

console.log(
  `Translations: every key of ${defaultLanguage} is in ${translations.join(", ")}.`
)
