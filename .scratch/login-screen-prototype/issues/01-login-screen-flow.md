# 01 - Login screen: layout and flow

Status: ready-for-agent
Blocked by: issue 03 of `.scratch/design-system-prototype/`

Read `../spec.md`, `DESIGN.md` and `docs/adr/0011-login-through-external-providers-with-a-server-side-session.md` first.

## What to build

The login screen with its two steps, in the default scenarios only: everything turned on, and every attempt succeeds.

- A login route in `apps/prototype`, laid out as "Layout" and "Order" in `../spec.md` describe: title bar, status bar, the page with its framing, and the bordered pane in the centre.
- The Start step and the Code step of "The flow" in `../spec.md`, as one reducer whose state is a union of named steps.
- The components the screen uses, added to `packages/ui`, to the gallery and to `DESIGN.md`: Field, Label, Input, InputOTP, Select and Separator.
- The marks of the three login providers, as the technical notes of `../spec.md` say.
- The language selector in the status bar.
- Every text of the screen in the three languages, from the prototype's translation files. The English texts are in "Texts" in `../spec.md`.
- Logging in opens a provisional page with an action that logs out and returns to the login route. Issue 05 of `.scratch/design-system-prototype/` replaces that page with the sample work screen and its user menu.
- A link to the login route in the prototype's navigation.

## Acceptance criteria

- [x] With an e-mail typed, the primary button leads to the Code step, which shows the conditional sentence with that e-mail.
- [x] An empty or malformed e-mail shows an error on the field and asks for nothing.
- [x] The button that logs in is disabled until six digits are typed; pasting six digits fills the boxes; typing the sixth digit sends nothing.
- [x] The action that asks for a new code is disabled for one minute and shows the time left.
- [x] The action that changes the e-mail returns to the Start step with the e-mail kept in the field.
- [x] Any six digits, and any login provider, open the provisional page; logging out there returns to the login route.
- [x] The screen has one primary button per step, and no control that exists only because it is in the prototype.
- [x] The language selector changes every text of the screen without a reload, and the choice survives a reload.
- [x] Correct in light and dark, in the three languages, at desktop and 375px widths, with no overflow.
- [x] Fully operable by keyboard, with focus moving to the first box when the Code step opens.
- [x] The coverage check passes with the new components in the gallery.

## Comments

Built ahead of the gallery, as "Built early" in `../spec.md` says. Checked in a browser: both steps, the three languages through the selector, both themes, 1280px and 375px. The order of focus was read from the page: the three login providers, the e-mail, the primary button, the language selector. A paste was checked as six digits arriving at once, which is what the browser hands the field; a real paste from the clipboard was not exercised.
