# 08: The provider on the login screen

**What to build:** The login screen shows one button per login provider that is turned on, above the divider. The button takes the user to the provider and back into the app, at the page they were at when their session ended. When the provider's e-mail belongs to no active user, the screen says so, with the e-mail, and tells them to ask an administrator; when the login did not conclude, it says so and lets them try again. Rules: the Framed pane recipe in `DESIGN.md`; the login state read of 07.

**Blocked by:** 05 (The login screen and the gate), 07 (Dex and the login through a provider in the API).

**Status:** resolved

- [x] The screen reads the login state and draws one button per provider, with its display name; with no provider on, the divider and the buttons are absent.
- [x] A button navigates to the start operation of that provider with the current `return_to` when there is one.
- [x] The notice `no_active_user` shows the e-mail from the login state and tells the person to ask an administrator; the same text for an unknown and a deactivated user.
- [x] The notice `login_not_concluded` says the login did not conclude and that the person can try again.
- [x] Screen tests cover one and several providers, none, the two notices, and the return path passed to the start operation.
- [ ] In development, logging in through Dex as the seed's administrator lands in the shell, and as the seed's developer shows the notice with their e-mail.
- [x] The texts exist in the three supported languages.

## Comments

### What was built

- `access/data/session.ts`: `loginStateQuery` (key `["session", "login"]`, `tellLoginState`), the type `LoginProvider` (a provider turned on, as the login state names it), and `providerLoginAddress(key, returnTo)`, the address of the button of a provider, built with the generated `getStartProviderLoginUrl`. `access/index.ts` exports `loginStateQuery`.
- `routes/login.tsx`: a loader that reads the login state (`ensureQueryData`) before the login screen opens. The gate of the root runs first, so a logged-in user, led away from the login, never reads it.
- `access/components/login-providers.tsx`: one link per provider turned on, drawn as an outline button the width of the pane, "Log in with <display name>", and under them the divider, a line with "or" in its middle; nothing at all with no provider on.
- `access/components/login-notice.tsx`: the texts and signs of `login_not_concluded` and `no_active_user`; the second shows the e-mail of the refusal, which the screen passes from the login state.
- `access/screens/login-screen.tsx`: reads the login state (`useSuspenseQuery`, filled by the loader) and draws the notice with the refused e-mail and the buttons, above the step of the e-mail or of the code, in place of the comment ticket 05 left.
- Texts in `en-US`, `pt-BR` and `es`: `login.providers.logInWith`, `login.providers.or`, `login.notices.loginNotConcluded`, `login.notices.noActiveUser`.
- Tests: `access/screens/login-screen.test.tsx` gains the providers and the two notices, in place of the test that showed nothing for them (36 tests, 25 before; the opener takes the login state, no provider by default). `testing/api.ts` gains `noLoginProviderIsOn()`, which every test that reaches the login screen now hands the API beside `nobodyIsLoggedIn()`: `login-gate.test.tsx` and the four setup test files whose openers end on the login screen.
- Records: Data module in `CODING_STANDARDS.md` (the loader of the login state, the notice without a refusal, the button as a link whose address the data module builds); the login screen line of the Framed pane recipe in `DESIGN.md` (the button and the divider); `README.md` and `README.pt-BR.md` (the button of Dex on the login screen, in place of "comes next").

### Decided while building

- **The login state is read by the route, before the screen opens**, as the gate reads the setup and the user: the screen draws once, with its buttons, and a failure of the read shows the error screen with its trace identifier, as a failure of the gate does. The login with a code could have stayed usable then, but a failure of a read that needs no user is a failure of the API or its database, which the login code needs too.
- **The button of a provider is a link**, `<a href>` with the look of an outline button: it leaves the app for the provider, as the start operation must be navigated to and never fetched, so it works without the router and opens in a new tab like any link. Screen tests read where it leads (`href`), since a click would take the test page away. The role is `link`, its name "Log in with <display name>".
- **Which `return_to` the button passes.** The one the login was given, when it is a path of the app (`whereTheLoginLeads`), not the root, and within the 2048 characters the contract takes (the generated `StartProviderLoginQueryParams`); otherwise none, and the login returns to the root. The start is a navigation, so a value the API refuses would show its 422 as a page.
- **`no_active_user` with no refusal in the login state shows nothing.** The e-mail comes only from the login state; without it, the address was not the one the refusal led to (another browser, a later login that forgot it), and the notice would claim something the session does not hold. Its text is one for an unknown and a deactivated e-mail, since the API does not tell them apart.
- **The buttons stay above both steps**, the e-mail and the code, as the Framed pane recipe draws the login screen: a person waiting for a code can still go through a provider.
- **No mark on the button.** The local login provider has none; the marks of Google, Microsoft and GitHub, with the colours their owners require (the exception to rules 2 and 3 of `DESIGN.md`), come with those providers.
- **The e-mail of the notice wraps as a whole** (`wrap-anywhere`): `break-all`, which the step of the code uses, broke `seed+developer@...` inside its first word in the browser check.
- **The key of the login state is `["session", "login"]`**, under the key of who is logged in, like its path. With TanStack Query's default it is read again when the person comes back to the tab, so a provider turned on or off meanwhile shows.

### What was checked

- Screen tests (Vitest Browser Mode, MSW): one provider, its button above the divider leading to `/api/session/login_providers/local/start`; three providers, in the order of the login state, each to its own start; none, with neither a button nor the divider; `return_to=%2Fprojects%2F42%3Ftab%3D2` passed on to the start; `//elsewhere.example/`, `https://elsewhere.example/` and `/` passed as none; a `return_to` longer than 2048 characters passed as none; the error screen with the trace identifier when the login state answers 500; `login_not_concluded` with its title and "You can try again below", the button keeping `return_to=%2Fprojects`; `no_active_user` with the e-mail of the login state and "Ask an administrator for access", and an e-mail in the address not shown; `no_active_user` in `pt-BR`, the button "Entrar com Local login provider"; `no_active_user` with no refusal and an unknown value showing no notice. The tests of ticket 05 pass unchanged but for the login state handed to the API.
- `task web:check` passes: generate check, `vp check` (format, lint, types), imports, knip, missing keys, 142 screen tests. tsgolint started without `OXLINT_TSGOLINT_PATH`.
- In Chromium (Playwright's) through Caddy at `http://localhost`, with the API and the worker built from this worktree on the development database (migration 25, seeded), Vite on 5173 and the Dex of the shared stack, driven by a script outside the repository: `/projects/42` led to `/login?return_to=%2Fprojects%2F42`, with the button "Log in with Local login provider" leading to the start with that `return_to`, and the divider; through Dex as `seed+admin@fabrica.example` the browser came back logged in and landed on the provisional landing of ticket 05 at `/` ("Você entrou como Ana Exemplo.", `/projects/42` being an address the app does not have), `GET /api/session` answering 200; "Sair" led to `/login?notice=logged_out`. In a new browser, as `seed+developer@fabrica.example` from `/login?return_to=%2Fprojects`, the browser came back to `/login?notice=no_active_user&return_to=%2Fprojects` with "No active user has this e-mail" and `seed+developer@fabrica.example` in the notice, the e-mail in no address, the button keeping `return_to`, and the notice in Portuguese after choosing it in the selector. A callback with `error=access_denied` led to `/login?notice=login_not_concluded` with its notice. At 375px in the dark theme, in Spanish with that notice, nothing overflowed sideways, the button wrapped its name on two lines inside the pane, and it was the first stop of the Tab key. The API, the worker and Vite were stopped after; the development database keeps the `logged_in` entries of the check.

### What was not checked

- **Landing in the shell** (the sixth criterion, unticked): ticket 06 had not merged, so the login through Dex as the administrator landed on the provisional landing of ticket 05, not the shell. The part on the developer was seen as the criterion says. Once 06 merges, the same login lands where the root leads; ticket 12's end-to-end test logs in through Dex to the shell.
- `task e2e`, which belongs to ticket 12. The Chrome DevTools MCP was not connected; the browser check used Playwright's Chromium.

### Found, and left for a decision

- **The step of the code and the review of the setup break an e-mail inside a word** (`break-all` in `code-step.tsx` and `review-step.tsx`), as the notice did before this ticket. `wrap-anywhere` keeps the e-mail whole when it fits the line; changing them is a small change outside this ticket.
