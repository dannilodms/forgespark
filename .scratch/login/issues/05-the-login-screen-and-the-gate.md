# 05: The login screen and the gate

**What to build:** A user opens the app, types their e-mail on the login screen, receives a code, types it on the same screen and lands in the app. The gate of the app now leads to the setup, the login or the app; a session that ends mid-work sends the user back to the login with a notice and returns them afterwards; the app speaks the user's preferred language once logged in. The landing after the login is provisional in this ticket: a route at the root with one sentence, which 06 replaces with the shell. Rules: State and Tests in the frontend rules of `CODING_STANDARDS.md`; the Framed pane recipe in `DESIGN.md`. Verifiable in the browser with Mailpit.

**Blocked by:** 04 (Checking the code: the session, who is acting, logging out and the origin).

**Status:** ready-for-agent

- [ ] The gate reads the state of the setup, then the logged-in user: the setup while not concluded, the login without a session, the app with one. A logged-in user at the login address is taken to the app.
- [ ] The login screen replaces the provisional one, in the narrow framed pane: a place for the provider buttons (empty until 08), a divider, the e-mail step with "send a code". Its flow is one reducer whose state is a union of named steps (`email`, `sending`, `code`); a page loaded again starts at `email`.
- [ ] The step of the code says a code was sent if the e-mail belongs to a user, takes six digits, shows a wrong code under the field with the attempts left, offers a new code 60 seconds after the last request, and returns to the e-mail step with a notice when the API says the code ended.
- [ ] Past the limit of codes for the address, the screen says to wait; past the limit of Caddy, it shows the 429 as the setup does.
- [ ] The notices come from the search parameter `notice`: `session_ended`, `logged_out`, and the notices of the setup with texts that now say the administrator logs in below; `login_not_concluded` and `no_active_user` are declared and shown by 08.
- [ ] A 401 answered to the app outside the login sends it to the login with `notice=session_ended` and `return_to`; after the login the app opens `return_to` when it is a path of the app, and the root otherwise.
- [ ] After the login the app switches to the user's preferred language; the selector stays only in the footer of the framed pane; logging out returns the app to the language of the selector and opens the login with `notice=logged_out`.
- [ ] Screen tests in Vitest's Browser Mode with MSW cover the two steps, the conditional message, a wrong code, the third attempt, the wait for a new code, the limit reached, each notice, the gate between the three places, a 401 mid-work with the return, and a logged-in user at the login address.
- [ ] The texts exist in the three supported languages and the missing-key test passes.
