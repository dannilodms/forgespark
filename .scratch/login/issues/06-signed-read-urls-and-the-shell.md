# 06: Signed read URLs and the shell

**What to build:** A logged-in user lands in the shell of the app: the header with the brand mark, the product name and their own menu (name, e-mail, photo or initials, log out); the sidebar with the logo or the initials of the consultancy and its name; a work surface with one sentence saying the areas of work come in the next efforts. The photo and the logo come from the private storage through signed URLs that expire in 15 minutes and are renewed. Nothing that does nothing is shown. Rules: Files in `CODING_STANDARDS.md`, ADR-0017, ADR-0026 (name and logo need no scope); the App shell recipe in `DESIGN.md`.

**Blocked by:** 05 (The login screen and the gate).

**Status:** ready-for-agent

- [ ] The `files` area has a use case that answers a signed `GET` URL for a confirmed file, for the public host name of the storage, valid for 15 minutes; the `s3` test against Garage fetches an object through it as the browser would.
- [ ] The operation that answers the logged-in user carries the signed URL of the photo when there is one; a read of the consultancy answers its name and the signed URL of its logo, for any logged-in user.
- [ ] A layout route wraps every screen after the login and draws the App shell recipe cut down to what works: header with the brand, the product name and the user menu; sidebar with the consultancy; work surface with one sentence. No search, notifications, help, favourites or settings.
- [ ] Under 960px the sidebar is a sheet opened from a button at the start of the header. The shell holds in both themes.
- [ ] The user menu shows the name, the e-mail, the photo or the initials, and one action, log out, which leads to the login with its notice.
- [ ] The app reads the user and the consultancy again at an interval shorter than 15 minutes, so an image never breaks on the screen.
- [ ] Screen tests cover the menu, the initials when there is no image, the logout, the preferred language and the sheet at a narrow width.
- [ ] `DESIGN.md` records that the App shell recipe has code, what this effort leaves out of it and why.
