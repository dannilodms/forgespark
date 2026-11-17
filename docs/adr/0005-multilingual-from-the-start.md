---
status: accepted
---

# Multilingual from the start, with two kinds of content

The system must work in several languages, with English as the default, and adding a language must stay cheap. It ships with English (`en-US`), Brazilian Portuguese (`pt-BR`) and a neutral Spanish (`es`); a reader using another regional variant gets the closest one. Each user picks a preferred language in their profile; before login the browser's language applies, falling back to English. The language never appears in the URL.

Everything the system itself says is translated: interface text, error and validation messages, e-mails, notifications, and LLM answers, which come back in the user's language. What people type is split in two, because the two behave differently:

- **Configuration content** (templates, names of categories, statuses and services) keeps one version per supported language. It can be written in any supported language; the LLM proposes the others as automatic drafts, which stay marked as such until a person reviews them. A reader with no version in their language sees the English one, and failing that the original.
- **Free text** (request descriptions, comments, messages, documents) is kept only in the language it was written in, with that language recorded. A reader can ask for a translation, which the LLM produces, the system caches, and the screen labels as automatic. The original remains the only official text.

## How it is built

- The interface uses react-i18next, with one file per language and screen area. English files are the source of truth, and an automated check fails when another language is missing a key, so whoever adds a text delivers all translations in the same change.
- The API returns a stable error code, its parameters and an English message. The frontend translates by code; AI agents consuming the API read the English. The backend keeps its own catalog only for what it sends itself, such as e-mails and notifications.
- Adding a language is a small code change followed by a deploy: add its code to the single list of supported languages and deliver its interface files. It must never require a database schema change.
- The single list is `supportedLanguages` in `apps/api/internal/kernel/language.go`. The contract carries it as the values of `preferred_language`, and the app reads the codes from the client generated from the contract (`apps/web/src/shared/locales/languages.ts`), which adds only the name of each language in its own language. A language on one side only fails the type check, and one without its interface files fails the missing-key check, so `task check` fails until a new language is complete.

## Considered Options

- **Translating only the interface.** Rejected: e-mails, notifications and LLM answers would stay in English for people who do not read it.
- **Stored versions for every text field.** Rejected: each edit leaves the other versions stale, and nearly every table would need translation structure.
- **Translating all free text automatically on save.** Rejected: LLM cost on every text, and a mistranslation would pass as official text.
- **The backend translating API messages.** Rejected: the frontend needs its own catalog anyway, and agents would receive messages that vary by language.
- **Adding languages from an admin screen, without a deploy.** Rejected: it means moving all interface text into the database and building translation management screens.
- **Paraglide JS, Lingui or FormatJS.** Paraglide checks keys more strictly, but react-i18next is the library models know best, which matters because agents write the code.

## Consequences

How versions of configuration content are stored is not decided yet. Whatever is chosen has to hold, per version, its language, whether it is the original, and whether it is still an automatic draft, without a schema change when a language is added.
