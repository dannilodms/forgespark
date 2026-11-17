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

- [ ] A person completes the first setup from the code to the login screen, and the consultancy and the administrator exist afterwards.
- [ ] A required field left empty, or a value in the wrong format, shows an error on its field and sends nothing.
- [ ] A 422 from the API shows on the field it names, and the screen moves to the step that holds that field.
- [ ] The names of countries, currencies and time zones follow the language of the screen, and the lists can be filtered by typing.
- [ ] Choosing a country suggests its currency, and a currency already changed by hand is not overwritten.
- [ ] Going back from the review to an earlier step and forward again loses nothing.
- [ ] With the language of the screen changed before concluding, the administrator is created with that preferred language.
- [ ] A setup code that expires on any step returns to the step of the code with the form kept, and a new code resumes where the person was.
- [ ] A second person who concludes after the first is told the setup is already concluded.
- [ ] Correct in light and dark, in the three languages, at desktop width and at 375px, and fully operable by keyboard.
- [ ] Screen tests cover the whole path, an error on its field, the code expiring halfway and the setup already concluded.
- [ ] The generated Zod schemas work in TanStack Form. What had to be adjusted is noted in ADR-0013.
