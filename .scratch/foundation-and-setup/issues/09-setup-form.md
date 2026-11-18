# 09 - The setup form: consultancy, administrator, review and concluding

Status: ready-for-agent
Blocked by: 08

Read `../spec.md`, `DESIGN.md` and, in `CODING_STANDARDS.md`, State, Forms and validation, Errors, Values and Components in the frontend rules.

## What to build

After the setup code, the person fills in the consultancy and its first administrator, reviews everything and concludes. The app then leads to the login screen. Images come in ticket 11.

- The step of the consultancy: the name it is known by, the legal name, the country, the tax ID, the time zone and the currency, which are required, and the e-mail, the telephone, the website and the address in two lines, city, region and postal code, which are not.
- The step of the administrator: the name, the e-mail, the country and the national ID, which are required, and the telephone, which is not.
- The country, the currency and the time zone as searchable lists whose names the browser writes in the language of the screen. The time zone starts at the one of the browser, and the currency at the one of the country chosen.
- The step of the review, which shows everything typed, the e-mail of the administrator in plain sight, with a way back to each step.
- Moving back and forth between the steps keeps what was typed.
- The forms on TanStack Form with the generated Zod schemas. A form checks only what the contract declares, and a 422 from the API shows on the field it names, on the step that holds it.
- Concluding sends everything once, with the language the screen is in as the preferred language of the administrator, and leads to the login screen with the notice that the setup is concluded.
- A request that answers that the code is not valid sends the person back to the step of the code and keeps everything typed. A valid code returns them to the step they were on.
- An answer that the setup is already concluded says so and leads to the login screen.
- A button that waits for an answer shows it and cannot be pressed again.
- Every text in the three supported languages. A component the screen needs and the design system lacks enters through the shadcn CLI, and `DESIGN.md` gains it in the same change.

## Acceptance criteria

- [x] A person completes the first setup from the code to the login screen, and the consultancy and the administrator exist afterwards.
- [x] A required field left empty, or a value in the wrong format, shows an error on its field and sends nothing.
- [x] A 422 from the API shows on the field it names, and the screen moves to the step that holds that field.
- [x] The names of countries, currencies and time zones follow the language of the screen, and the lists can be filtered by typing.
- [x] Choosing a country suggests its currency, and a currency already changed by hand is not overwritten.
- [x] Going back from the review to an earlier step and forward again loses nothing.
- [x] With the language of the screen changed before concluding, the administrator is created with that preferred language.
- [x] A setup code that expires on any step returns to the step of the code with the form kept, and a new code resumes where the person was.
- [x] A second person who concludes after the first is told the setup is already concluded.
- [x] Correct in light and dark, in the three languages, at desktop width and at 375px, and fully operable by keyboard.
- [x] Screen tests cover the whole path, an error on its field, the code expiring halfway and the setup already concluded.
- [x] The generated Zod schemas work in TanStack Form. What had to be adjusted is noted in ADR-0013.

## Comments

### What was built

- **The lists the API accepts, generated into the app.** A test of `apps/api/internal/kernel` (`lists_of_the_app_test.go`) writes the countries, the currencies and the time zones of the kernel as TypeScript constants with literal types into `apps/web/src/shared/places/lists.gen.ts` (`task api:lists`), and fails when the committed file differs (`task api:lists:check`, in `task check`, in `task api:test` and in the backend job). The app never types a code again and never takes one from `Intl.supportedValuesOf`. (After the review of the effort the test is in `apps/api/internal/consultancy/domain`, where the currencies and the time zones moved; the countries stay in `kernel`.)
- **`shared/places`.** `places.ts` builds each list once for each language: the countries and the currencies named by `Intl.DisplayNames` and sorted by the collator of the language, a currency followed by its code ("Brazilian Real (BRL)"), a time zone as the IANA database writes it with spaces for underscores, followed by the name `Intl.DateTimeFormat` gives it ("America/Sao Paulo (Brasilia Standard Time)"). `timeZoneOfTheBrowser()` is the zone of the browser when the API accepts it, and nothing otherwise. `currency-of-country.ts` is the currency of each country, from CLDR 49.0.0, typed against the generated lists.
- **The Combobox** of shadcn/ui (`ui/components/combobox.tsx`, a line in `DESIGN.md`), restyled to the design system, with its button out of the order of the Tab key.
- **`shared/components/form-fields.tsx`**: `TextField` and `ListField`, a field of TanStack Form with its label (and "(optional)"), its description, its problem, `aria-invalid`, `aria-required` and `aria-describedby`; `focusTheFirstProblem`.
- **`checkTheContract(schema, toBody)`**: the typed values are turned into the body before the generated schema checks them.
- **The setup form.** One TanStack Form beside the reducer (`components/setup-form.ts`), holding every field as a text named by its place in the body. The steps of the consultancy and of the administrator are `form.FormGroup`s that check their part of `ConcludeSetupBody` when the person moves on; the review shows everything, the e-mail of the administrator in plain sight, with a way back to each step; the title of each step takes the focus when it opens. The country suggests its currency until the currency is chosen by hand.
- **Concluding** (`useConcludeSetup` in `data/setup.ts`): one `POST /api/setup/conclude` with everything, the code as typed and the language of the screen as the preferred language; empty optional fields and an empty address are left out. On success the state of the gate is set to concluded and the screen leads to `/login?notice=setup_concluded`. A 422 on fields goes onto the fields (`form.setErrorMap`) and to the first step that holds one. `setup_code_invalid` reads the state of the setup again: concluded, it leads to `/login?notice=setup_already_concluded`; not concluded, to the step of the code with a notice, everything kept, and a new code returns to the step the person was on. A 409 leads to the same notice. Anything else shows above the actions with its `trace_id`. The button shows a spinner and cannot be pressed while it waits.
- **The login screen** gained the notice `setup_already_concluded`: nothing of what was typed was saved.
- **The alert of `ui`** no longer paints a status icon in the colour of its text.
- **Texts** in English, Portuguese and Spanish: the steps, the review, the notices, the field codes of the domain (`country_unknown`, `currency_unknown`, `time_zone_unknown`, `language_unsupported`, `email_invalid`, `telephone_invalid`, `website_invalid`) and `setup_already_concluded`.
- **Documents.** `CODING_STANDARDS.md` (Enforcement of the backend; Layout, Contract, Forms and validation, Values and Enforcement of the frontend), `DESIGN.md` (Combobox), ADR-0013, both READMEs, the comments of the three lists of the kernel.

### Decided while building

- **The lists come through a generated file, not the contract.** Making them an `enum` of the contract would answer `invalid` for a value outside them instead of `country_unknown`, which ticket 06 chose against, and would put 600 time zones in the spec. A test of `kernel` writes them because they are unexported, as the spec is written by a test of `cmd/api`. The check runs in the backend job, which has Go.
- **The currency of each country**: CLDR's `supplemental/currencyData.json`, release 49.0.0 (npm `cldr-core`), the first current legal tender of each country that the API accepts. Antarctica has none; seven countries list two and the first is kept. Its source and rule are written at the top of the file, in one place.
- **A time zone outside the list of the API starts empty**, and the person chooses one: the field is required, so moving on says so.
- **One form, steps as groups, no validator on the form.** TanStack Form runs the validators of the form, filtered, when a group submits, so an `onSubmitAsync` on the form would conclude from every step. Concluding is a mutation of the data module instead, and its answer reaches the fields through `setErrorMap`, because a field forgets its problems when it unmounts.
- **Expiry is noticed by a request**, as the spec says, not by a timer on `expires_at`, whose clock is the browser's.
- **A second person is told by reading the state again**: the API answers the second conclusion with `setup_code_invalid`, because the first one ended the code.
- **`setup_already_concluded` is a notice of the login screen**, a search parameter like `setup_concluded`, in amber with a triangle: what was typed was not saved.
- **The focus**: the title of a step that opens, or the first field with a problem; the code after a code stopped working.
- **Labels**: "(optional)" on each optional field, the required ones `aria-required`; a short description under the fields whose form is not obvious (telephone, website, tax ID, national ID).

### What was checked

- **The screen tests**, 58 in the app with the missing keys (23 new in `setup-steps.test.tsx`; the 3 of ticket 08 that expected the placeholder now expect the step of the consultancy): each required field of both steps empty, sending nothing, with the focus on the first; an e-mail, a telephone and a website out of their form and a tax ID too long; the countries in Portuguese filtered by "japao" to "Japão"; the currency suggested for Brazil and for Portugal, and not over one chosen by hand; the names of what was chosen again in Portuguese after the language changed; the time zone of the browser and a time zone the API does not know; the administrator kept when going back; the review with every value; Change and Continue through both steps; concluding with the exact body and the notice; Spanish chosen before concluding sent as `es`; the button disabled while it waits and a single request; a 422 on `administrator.email` on its field on its step, with the focus; a 422 on two steps, one after the other; a 500 with its `trace_id`; a 429; a code that stopped working, the code step with the focus, a new code back at the review and concluding with it; a second person by `setup_code_invalid` and by 409; the step of the consultancy filled in and left with the keys alone, a country chosen by typing and Enter.
- **That the tests bite**: without the update of the state of the gate, the test of concluding failed; without the check of a currency chosen by hand, its test failed.
- **The checks bite**: a hand-edited `lists.gen.ts` failed `task api:lists:check`; a currency outside the lists in `currency-of-country.ts` failed the type check.
- **The real API, through the shared Caddy, in two real Chromium contexts** (`fs-09-manual.sh`, a database of its own, dropped after): both people opened the setup with the code `issue-setup-code` printed and reached the review; the time zone started at São Paulo and Brazil suggested the real; a second code was issued, which ended the first; A concluded, was sent back to the step of the code with the notice, typed the second code in lower case and was back at the review with everything; A chose Spanish and concluded, landing on `/login?notice=setup_concluded` in Spanish; B concluded and landed on `/login?notice=setup_already_concluded`. The database then held the consultancy as typed (the telephone without spaces, the e-mail in lower case), one administrator with `preferred_language = es` and the administrator role, both roles, both codes ended, and the audit entry `setup_concluded`. The logs of the API held neither code.
- **The look**, by screenshots of Vitest's Browser Mode, looked at: the consultancy with every error in English, filled in Portuguese and Spanish, the administrator in Portuguese and Spanish, the review in the three languages, the review with a failure, the code stopped working in Portuguese, the list of countries open in Spanish, the list of time zones filtered, the login screen with the new notice, each in light and in dark, at 1280px and at 375px, with no horizontal overflow measured on any.
- **`task check`** passes on the branch. `feat/foundation-and-setup` at `28ff6e6` (tickets 08 and 10) was fast-forwarded into it before the first commit and had nothing newer when I finished. The first run failed in `cmd/api` and `platform/database` on `CREATE DATABASE ... timeout` under load; the two packages passed alone and the next run passed whole. The Go lint and the guard pass on the new test of `kernel`.

### What was not checked

- **A setup code that expires with time.** The API gives the same answer, `setup_code_invalid`, for an expired code and for one a new code ended; the real check used the second.
- **A field refused by the real API**: the contract and the steps refuse first, so the 422 on a field was simulated in the screen tests only.
- **A screen reader**, and browsers other than Chromium.
- **The pipeline.** Nothing was pushed: the new step of the backend job never ran.
- **The translations** were written by the agent and read once.
