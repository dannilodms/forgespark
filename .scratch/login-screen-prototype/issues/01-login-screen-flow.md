# 01 - Login screen: layout and flow

Status: ready-for-agent
Blocked by: issues 05 and 06 of `.scratch/design-system-prototype/`

Read `../spec.md`, `DESIGN.md` and `docs/adr/0011-login-through-external-providers-with-a-server-side-session.md` first.

## What to build

The login screen with its two steps, in the default scenarios only: everything turned on, and every attempt succeeds.

- A login route in `apps/prototype`, laid out as "Layout" and "Order" in `../spec.md` describe: title bar, status bar, the page with its framing, and the bordered pane in the centre.
- The Start step and the Code step of "The flow" in `../spec.md`, as one reducer whose state is a union of named steps.
- InputOTP added to `packages/ui`, to the gallery and to `DESIGN.md`, as the technical notes of `../spec.md` say.
- The marks of the three login providers, as the technical notes of `../spec.md` say.
- The language selector in the status bar.
- Every text of the screen in the three languages, from the prototype's translation files. The English texts are in "Texts" in `../spec.md`.
- Logging in opens the sample work screen. Add a user menu to the app shell with an action that logs out and returns to the login route, and record the user menu in the App shell recipe of `DESIGN.md`.
- A link to the login route in the prototype's navigation.

## Acceptance criteria

- [ ] With an e-mail typed, the primary button leads to the Code step, which shows the conditional sentence with that e-mail.
- [ ] An empty or malformed e-mail shows an error on the field and asks for nothing.
- [ ] The button that logs in is disabled until six digits are typed; pasting six digits fills the boxes; typing the sixth digit sends nothing.
- [ ] The action that asks for a new code is disabled for one minute and shows the time left.
- [ ] The action that changes the e-mail returns to the Start step with the e-mail kept in the field.
- [ ] Any six digits, and any login provider, open the sample work screen; logging out there returns to the login route.
- [ ] The screen has one primary button per step, and no control that exists only because it is in the prototype.
- [ ] The language selector changes every text of the screen without a reload, and agrees with the prototype bar.
- [ ] Correct in light and dark, in the three languages, at desktop and 375px widths, with no overflow.
- [ ] Fully operable by keyboard, with focus moving to the first box when the Code step opens.
- [ ] The coverage check passes with InputOTP in the gallery.
