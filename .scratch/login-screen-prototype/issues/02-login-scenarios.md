# 02 - Login screen: scenarios

Status: ready-for-agent
Blocked by: 01

Read `../spec.md` and "Prototype" in `CODING_STANDARDS.md` first.

## What to build

Every condition the login screen can be in, reachable from the prototype bar.

- Scenarios in the prototype bar: a screen declares its scenarios in groups, the bar lists the groups of the current route, and the fake data of the screen answers according to the scenario chosen in each group. A route that declares none shows no scenario control. The choice lasts until the page is reloaded.
- The two groups of the login screen and every scenario in them, as "Scenarios" in `../spec.md` lists: Ways in and Outcome.
- The message of each scenario, in the three languages. Each says what happened and then what to do, and is shown as an Alert above the form or as an error on its field, with an icon.
- The local provider's button, last of the provider buttons, with a badge that says it is for development.

## Acceptance criteria

- [x] The prototype bar shows the two groups on the login route and no scenario control on a route that declares none.
- [x] Each scenario of Ways in shows exactly the ways the table lists, with no separator when only one kind of way is on.
- [x] Wrong code: the first two attempts say how many are left, and the third disables the boxes and leaves the action that asks for a new code.
- [x] Expired code, Limit reached, Account not registered, Provider failed and Session ended each behave as the table says.
- [x] Changing a scenario while on the Code step leaves the screen in a state the flow allows.
- [x] No message reports a problem by colour alone.
- [x] No component or screen file reads a scenario by name outside the fake data of the screen.
- [x] Every scenario is correct in light and dark, in the three languages, at desktop and 375px widths.

## Comments

### What was built

- `apps/prototype/src/stores/scenarios.ts` holds the scenario chosen in each group, in memory only. A route carries the groups of its screen in `staticData`, and the prototype bar shows one select for each group of the current route.
- `login/fake-login.ts` declares the two groups and is the only file that reads a scenario by name. The screen reads from it which ways in are on and whether a session ended, and gets an answer from each operation.
- The reducer gained the answers that refuse, and one action for "what the screen reads changed while it was open".
- The reducer and the fake data have tests beside them (36), written first. `task check` now runs them, and "Prototype" in `CODING_STANDARDS.md` says so.

### Decided while building

- An expired code disables the boxes, as an ended one does: neither can be typed again.
- The one-minute wait is kept after a code is over. The action that asks for a new code stays on the screen and shows the time left, as the limit of ADR-0011 would make it.
- When a scenario changes, the alert, the awaited answer and a code that was over go away, and the count of wrong attempts starts again. The Code step stays, with empty boxes, unless the login code is off or the session ended: then the screen goes back to the Start step and keeps the e-mail.
- An answer follows the scenario chosen when the button was pressed.
- The end of a session is a neutral Alert with the sign of a notice. The three refusals are red Alerts with a cross. An error on the code is on its field, with an icon, and the boxes take the red border.
- The local provider shows a terminal icon and is named "the local provider". The names of the providers moved to the translation files, so the Alert of a failed provider can name it in a sentence.
- Focus returns to the boxes each time they can be typed in again: after a wrong code and after a new one.

### What was checked in the browser

- By script, in light and dark, in the three languages, at 1280px and 375px: the five scenarios of Ways in, the four Alerts of the Start step, and the Code step plain, with two attempts left, with one left, ended and expired. No element leaves the pane, the page has no horizontal overflow, the prototype bar stays inside the viewport, and every message carries an icon. The texts of the three languages were read from the page.
- Seen: Session ended (dark, English, 1280px), a wrong code with two attempts left (dark, English, 1280px), an ended code (light, English, 1280px), the local provider (light, English, 1280px), a failed local provider (dark, Spanish, 375px), Limit reached on the Code step (light, Portuguese, 375px).
- Limit reached when asking for a new code, after waiting out the minute: the Alert shows above the sentence and the digits typed stay.
- On the Code step: choosing Providers only goes back to the Start step with the e-mail kept; choosing Session ended does the same and shows the Alert; choosing another Outcome keeps the step with empty boxes.
- Session ended: a provider logs in, and logging out opens the login route with the Alert again. A reload puts both groups back on their defaults.
- The home route, the gallery and the provisional page show no scenario control.
- Keyboard: the two selects of the bar take focus with the 3px outline, open with Enter and choose with the arrow keys. A wrong code typed on the keyboard gives focus back to the boxes.
- The console has no error and no warning from the screen. The one error is the missing `favicon.ico`, which was there before.

### What was not checked

- No screen reader was run. The Alerts and the field errors have `role="alert"`.
- Most combinations of scenario, language, theme and width were measured by script and not looked at.
- At 375px the badge of the local provider goes to a second line inside its button, in the three languages.

### Found, and left

- The "wrong" InputOTP sample of the gallery passes `aria-invalid` to the field only, so its boxes keep the plain border. The login screen passes it to each box.
- The table of commands in the two README files does not mention what `task check` runs beyond format, lint and types.
