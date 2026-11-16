# Login screen in the prototype

Status: ready-for-agent

## Goal

The login screen is the first real screen of the product. It is designed and approved in the prototype, with fake data and no backend, before `apps/web` builds it against the API. This effort ends when the maintainer approves the screen in every scenario, in the three languages and in both themes.

## Decisions already made

Do not reopen these.

- How a person logs in: `docs/adr/0011-login-through-external-providers-with-a-server-side-session.md` and "Login" in `CODING_STANDARDS.md`. The terms login code and login provider are in `GLOSSARY.md`.
- An approved screen is rebuilt in `apps/web`, never moved, and its texts in the three languages are copied across then: `docs/adr/0006-design-system-proven-in-a-disposable-prototype.md`.
- The look, the components and the rules: `DESIGN.md`.
- **Layout.** The screen is framed like the application: a title bar with the product name, a status bar, and between them the page with one bordered pane in its centre. The page carries the page framing of `DESIGN.md`, with diamond nodes on the corners of the pane.
- **Order.** The login providers come first, stacked, as secondary buttons that each show the provider's mark and name. A separator follows, then the e-mail field with the only primary button of the screen.
- **Login code.** Six boxes and a button. Nothing is sent when the sixth digit is typed: the person presses the button. Pasting six digits fills the six boxes.
- **The same answer for any e-mail.** After a code is asked for, the screen says the code arrives if the address belongs to a registered person. It never says a code was sent.
- **Only what is turned on.** Each login provider, the local provider and the login code show only when the configuration turns them on.
- **One extra.** The status bar holds a language selector and nothing else. The screen has no theme toggle, no version number and no link to the source code.
- **Scenarios** are chosen in the prototype bar (see Prototype in `CODING_STANDARDS.md`).
- **After logging in**, the prototype opens the sample work screen, and that screen's user menu logs out.

## The flow

The screen is one route. Its state is a union of named steps in one reducer, as `CODING_STANDARDS.md` asks of a flow with several steps, so that the rebuild in `apps/web` keeps the same shape.

| Step | What the pane shows |
| --- | --- |
| Start | A heading, one button per login provider, the separator, the e-mail field and the primary button that asks for a login code. |
| Code | The sentence about the code with the e-mail typed, the six boxes, the primary button that logs in, an action that asks for a new code, and an action that goes back to change the e-mail. |

- The e-mail is checked in the browser before anything is asked: an empty or malformed address is an error on the field.
- The button that logs in is disabled until the six boxes are filled.
- The action that asks for a new code is disabled for one minute after a code is asked for, and shows the time left.
- A button that waits for an answer shows a spinner and cannot be pressed again. The fake data answers after a short delay.
- Pressing a login provider shows the spinner on that button, then the outcome of the scenario. The prototype shows no imitation of the provider's own page.
- Every message that reports a problem is an Alert above the form or an error on its field, with an icon and words, never colour alone.

## Scenarios

Two groups, chosen independently in the prototype bar.

**Ways in**: what the configuration turns on.

| Scenario | The screen shows |
| --- | --- |
| Everything | The login code and the three login providers. This is the default. |
| One provider | The login code and GitHub. |
| Login code only | The e-mail field alone, with no separator. |
| Providers only | The three login providers alone, with no e-mail field and no separator. |
| Local provider | Everything, plus the local provider as the last of the provider buttons, with a badge that says it is for development. |

**Outcome**: how the fake data answers.

| Scenario | What happens |
| --- | --- |
| Success | Any six digits log in. Any login provider logs in. This is the default. |
| Wrong code | Every code is wrong. The first and the second attempt say so and say how many attempts are left. The third ends the code: the boxes are disabled and the screen asks the person to ask for a new code. |
| Expired code | Logging in answers that the code expired and that a new one has to be asked for. |
| Limit reached | Asking for a code, or for a new one, answers that too many were asked for and that the person should wait. |
| Account not registered | A login provider comes back to the Start step with an Alert: the account has no access, and a person who administers the system can give it. |
| Provider failed | A login provider comes back to the Start step with an Alert that names the provider and asks the person to try again. |
| Session ended | The Start step opens with an Alert: the session ended and the person logs in again. |

## Texts

English is the source. The Portuguese and Spanish versions are written with the screen, in the prototype's translation files.

| Where | Text |
| --- | --- |
| Title bar | My Software Factory Manager |
| Heading | Log in |
| Under the heading | Use an account that is already registered. |
| Provider button | Continue with Google (Microsoft, GitHub) |
| Separator | or |
| Field label | E-mail |
| Primary button, Start | Send login code |
| Sentence, Code step | If ana@example.com belongs to a registered person, the code arrives in a moment. |
| Boxes label | Login code |
| Primary button, Code | Log in |
| New code | Send a new code. While it waits: Send a new code in 0:42 |
| Back | Use another e-mail |
| Footer of the pane | Only registered people get in. Ask a person who administers the system for access. |

The messages of the scenarios are written with issue 02 and follow the same voice: what happened, then what to do.

## Out of scope

- `apps/web` and `apps/api`.
- An imitation of a login provider's page.
- Registering users, and the profile where a user picks a preferred language.
- How long a session lasts, which ADR-0011 leaves open.

## Technical notes

- The prototype rules apply: fake data in plain modules, no TanStack Query, no Zod, no screen tests (see `.scratch/design-system-prototype/spec.md`).
- The six boxes are shadcn/ui's InputOTP. It enters `packages/ui` through the shadcn CLI, with a block in the Inputs route of the gallery and a line in the component table of `DESIGN.md`, in the same change. Read the current shadcn/ui documentation for it before adding it.
- Rule 2 of `DESIGN.md` allows one accent colour, so the marks of the login providers are drawn in a single colour. Each provider publishes rules for the button that logs in with it. Read them: where a rule forbids a single-colour mark, keep the provider's own mark and record the exception under rule 2 in `DESIGN.md`.
- The language selector of the status bar and the one of the prototype bar set the same language.
- At a 375px-wide viewport the two bars stay, and the pane takes the width of the page.

## Notes for the rebuild in `apps/web`

These are what the prototype cannot show. They go to the issue that builds the real screen.

- The screen has to know which ways in are turned on before anyone is logged in, so the API needs an operation that tells it without a session.
- The screen cannot know that the e-mail server is down: sending is a job, and the answer is the same either way (ADR-0011). It can only know that the login code is turned off.
- The limit on asking for a code has to answer the same for a registered address and an unknown one, or it tells who is a user.

## Issues

1. `01-login-screen-flow.md`
2. `02-login-scenarios.md`
3. `03-approve-the-login-screen.md`
