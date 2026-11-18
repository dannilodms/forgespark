// Fails when a supported language is missing a key that English has, and
// names each missing key. The English file is the source of truth (ADR-0005).
//
// It reads every `locales` folder under `src`, so the texts of a domain area
// are checked as soon as the area has them. The languages come from the
// single list in src/shared/locales/languages.ts, so a language added there
// without its file fails here too.
//
// Run from `apps/web`: `vp node scripts/missing-keys.mjs`.
import { existsSync, readFileSync, readdirSync } from "node:fs"
import { join, relative, resolve } from "node:path"

import {
  defaultLanguage,
  supportedLanguages,
} from "../src/shared/locales/languages.ts"
import { missingKeys } from "../src/shared/locales/missing-keys.ts"

const appDir = resolve(import.meta.dirname, "..")

/** Every folder named `locales` under `dir`. */
function localesFolders(dir) {
  return readdirSync(dir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .flatMap((entry) => {
      const path = join(dir, entry.name)
      return entry.name === "locales" ? [path] : localesFolders(path)
    })
}

/** The texts of a language in a folder, or `null` when it has no file there. */
function textsOf(folder, code) {
  const file = join(folder, `${code}.json`)
  return existsSync(file) ? JSON.parse(readFileSync(file, "utf8")) : null
}

const translations = supportedLanguages
  .map(({ code }) => code)
  .filter((code) => code !== defaultLanguage)

const folders = localesFolders(join(appDir, "src"))
let failed = false

for (const folder of folders) {
  const name = relative(appDir, folder)
  const source = textsOf(folder, defaultLanguage)
  if (!source) {
    console.error(`${defaultLanguage} has no file in ${name}.`)
    failed = true
    continue
  }
  for (const code of translations) {
    const texts = textsOf(folder, code)
    if (!texts) {
      console.error(`${code} has no file in ${name}.`)
      failed = true
      continue
    }
    const missing = missingKeys(source, texts)
    if (missing.length > 0) {
      console.error(
        `${code} is missing ${missing.length} of the keys of ${defaultLanguage} in ${name}:\n` +
          missing.map((key) => `  ${key}`).join("\n")
      )
      failed = true
    }
  }
}

if (failed) {
  process.exit(1)
}

console.log(
  `Translations: every key of ${defaultLanguage} is in ${translations.join(", ")}, in ${folders.length} folder(s).`
)
