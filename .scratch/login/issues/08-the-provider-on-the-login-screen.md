# 08: The provider on the login screen

**What to build:** The login screen shows one button per login provider that is turned on, above the divider. The button takes the user to the provider and back into the app, at the page they were at when their session ended. When the provider's e-mail belongs to no active user, the screen says so, with the e-mail, and tells them to ask an administrator; when the login did not conclude, it says so and lets them try again. Rules: the Framed pane recipe in `DESIGN.md`; the login state read of 07.

**Blocked by:** 05 (The login screen and the gate), 07 (Dex and the login through a provider in the API).

**Status:** ready-for-agent

- [ ] The screen reads the login state and draws one button per provider, with its display name; with no provider on, the divider and the buttons are absent.
- [ ] A button navigates to the start operation of that provider with the current `return_to` when there is one.
- [ ] The notice `no_active_user` shows the e-mail from the login state and tells the person to ask an administrator; the same text for an unknown and a deactivated user.
- [ ] The notice `login_not_concluded` says the login did not conclude and that the person can try again.
- [ ] Screen tests cover one and several providers, none, the two notices, and the return path passed to the start operation.
- [ ] In development, logging in through Dex as the seed's administrator lands in the shell, and as the seed's developer shows the notice with their e-mail.
- [ ] The texts exist in the three supported languages.
