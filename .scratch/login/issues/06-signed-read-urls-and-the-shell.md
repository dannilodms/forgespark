# 06: Signed read URLs and the shell

**What to build:** A logged-in user lands in the shell of the app: the header with the brand mark, the product name and their own menu (name, e-mail, photo or initials, log out); the sidebar with the logo or the initials of the consultancy and its name; a work surface with one sentence saying the areas of work come in the next efforts. The photo and the logo come from the private storage through signed URLs that expire in 15 minutes and are renewed. Nothing that does nothing is shown. Rules: Files in `CODING_STANDARDS.md`, ADR-0017, ADR-0026 (name and logo need no scope); the App shell recipe in `DESIGN.md`.

**Blocked by:** 05 (The login screen and the gate).

**Status:** resolved

- [x] The `files` area has a use case that answers a signed `GET` URL for a confirmed file, for the public host name of the storage, valid for 15 minutes; the `s3` test against Garage fetches an object through it as the browser would.
- [x] The operation that answers the logged-in user carries the signed URL of the photo when there is one; a read of the consultancy answers its name and the signed URL of its logo, for any logged-in user.
- [x] A layout route wraps every screen after the login and draws the App shell recipe cut down to what works: header with the brand, the product name and the user menu; sidebar with the consultancy; work surface with one sentence. No search, notifications, help, favourites or settings.
- [x] Under 960px the sidebar is a sheet opened from a button at the start of the header. The shell holds in both themes.
- [x] The user menu shows the name, the e-mail, the photo or the initials, and one action, log out, which leads to the login with its notice.
- [x] The app reads the user and the consultancy again at an interval shorter than 15 minutes, so an image never breaks on the screen.
- [x] Screen tests cover the menu, the initials when there is no image, the logout, the preferred language and the sheet at a narrow width.
- [x] `DESIGN.md` records that the App shell recipe has code, what this effort leaves out of it and why.

## Comments

### What was built

- `files`:
  - `application`: the use case `SignDownload` (`Execute(ctx, fileID) (url, found, err)`): a URL that reads a confirmed file for 15 minutes, and nothing for a file never confirmed or that does not exist. The interfaces gained `FileRepository.ConfirmedKey` (no lock) and `Storage.SignDownload`; the in-memory storage records each download it signs (`SignedDownload`, `SignedDownloadKeys`), and the in-memory repository saves a confirmed file for the tests of other areas (`SaveConfirmed`).
  - `adapters/s3`: `SignDownload`, on `PresignGetObject` of the client that signs for the public host name. `adapters/postgres`: `ConfirmedObjectKeyOf` (sqlc).
- `consultancy`:
  - `application`: `ActiveUser.PhotoFileID` (the file 0 when there is none), the alias `FileID`, the read-only interface `ConsultancyReader` and the query use case `TellConsultancy` (name and the URL of the logo; a missing consultancy is a failure of the system, since nobody logs in before the setup). In memory: `NewConsultancyReader`, and `RegisterAdministratorWithPhoto` for the tests of `access`.
  - `adapters/postgres`: the two queries of the active user read `photo_file_id`; `ConsultancyReader` reads the name and the logo in one query.
  - `adapters/rest`: `tell_consultancy`, `GET /api/consultancy`, with `restapi.RegisterForUser` (401 `unauthorized`, 500), body `{name, logo_url?}`.
- `access`: `TellLoggedInUser` takes `SignDownload` and answers `PhotoURL`; `GET /api/session` gains `photo_url`, left out when there is no photo.
- `cmd/api`: `mountREST` builds one `SignDownload` for both reads. `cmd/seed`: its storage that stands in for none signs nothing either.
- Spec and client: `openapi.yaml` (`Consultancy`, `photo_url`, `/api/consultancy` in the singular paths of `tools/speclint/ruleset.yaml`) and the web client generated again.
- `apps/web`:
  - `shared/components/app-shell.tsx`, the App shell recipe in code (`AppShell`): the header with the mark, the product name and a slot for the user menu; the sidebar on the frame below the header with a slot for the consultancy, a sheet under 960px opened from a button at the start of the header; the work surface, which scrolls inside itself. `shared/components/image-or-initials.tsx`: the image of a person (a circle) or of the consultancy (a rounded square, the whole logo shown), with the initials while there is none, while it loads or when it fails, and which keeps showing an image until the picture of its new URL has loaded.
  - `access/components/user-menu.tsx` (`UserMenu`): the photo or the initials as the button, then the name, the e-mail and logging out; the menu stays open while the logout is sent and shows its failure. The provisional landing of ticket 05 and its texts are gone.
  - `consultancy`: `data/consultancy.ts` (`consultancyQuery`), `components/consultancy-heading.tsx` (`ConsultancyHeading`), `screens/start-screen.tsx` (`StartScreen`, the sentence on the work surface at the root).
  - Routes: the pathless layout route `routes/_shell.tsx`, whose loader reads the consultancy, and `routes/_shell/index.tsx` at `/`.
  - `loggedInUserQuery` and `consultancyQuery` are read again every ten minutes (`signedURLsRenewEvery` in `shared/api/storage.ts`), on top of the read when the person comes back to the tab.
  - The breakpoint: `useIsMobile` at 960px, and `--breakpoint-shell: 60rem` in `globals.css`, whose `shell:` classes replace `md:` in `ui/components/sidebar.tsx`.
  - Texts in `en-US`, `pt-BR` and `es`: the user menu (`access`), the sidebar and its button (`shared`), the sentence of the work surface (`consultancy`). `testing/api.ts` gained `loggedIn(user, consultancy)` and `theConsultancy`.
- Records: `DESIGN.md` (Recipes: the app shell has code, what it leaves out and why, its images and initials, the breakpoint; Spacing: `breakpoint-shell`), `CODING_STANDARDS.md` (Login: the photo in who is logged in; Users and roles: the read of the consultancy; Files: Between areas and Download; Frontend Layout: the layout route; Data module: the reads renewed every ten minutes), ADR-0017 (the signed `GET` against Garage).

### Decided while building

- **The use case is `SignDownload`**, beside `SignUpload` of the storage, as Download in Files names it; it answers a URL only for a confirmed file, because the object of an upload never confirmed holds whatever its sender sent. It reads the key without a lock: it is a read, and a confirmed file never changes its key.
- **Each read signs its own image.** Who is logged in (`access`) signs the photo, and the consultancy (`consultancy`) its logo, each calling `SignDownload` of `files`. `FindActiveUserByID`, which every request runs for the actor, reads only the identifier of the photo and signs nothing, so finding the actor costs no extra query.
- **`GET /api/consultancy`, `tell_consultancy`**: the consultancy exists once, so its path is singular (ADR-0009, and the list of the lint); it needs a logged-in user and no scope (ADR-0026), and answers only what the shell shows. It answers 500 when no consultancy exists, which cannot happen to a logged-in user.
- **The URLs are optional fields**, left out when there is no image, with `format: uri`; no time of expiry travels with them: the app reads both answers again every ten minutes, five less than the URLs work.
- **No blink at each renewal.** A new URL of the same image would make the avatar of Base UI show the initials while it loads, every ten minutes; `ImageOrInitials` loads the picture of the new URL first and only then shows it, when the browser holds it and the avatar finds it at once.
- **Where the pieces live.** The recipe is in `shared/components` beside `FramedPane`, with slots; the user menu is of `access`; the consultancy and the start screen of the root are of `consultancy` (the shell is the consultancy's, and the next efforts put its areas of work there); the layout route puts them together, so neither area imports the other. Recorded in Layout of the frontend rules.
- **The breakpoint.** DESIGN.md says 960px and shadcn/ui's hook said 768px: the hook now holds 960px, and the Sidebar's `md:` classes became `shell:`, a breakpoint of the theme at 60rem, so the two agree; recorded in DESIGN.md.
- **The images.** A person is a circle with initials, a consultancy a 32px square with 4px corners where the whole logo shows (`object-contain`). The initials take the `muted` fill for now: the coloured circles of the Traits arrive with the first screen that tells people apart. The initials are the first letters of the first and last words.
- **The sidebar lies on the frame without a line**, below the header (`variant="inset"`, as shadcn/ui's sidebar with a header does), and the work surface keeps 12px on its left too while the sidebar is a sheet.
- **Logging out** stays in the open menu while it is sent, with a spinner, and shows its failure there; it then leads to `/login?notice=logged_out` as before.

### What was checked

- Use cases with in-memory implementations: the download of a confirmed file signed for its kept key for 15 minutes; none for a file never confirmed or unknown; a failure of the storage or of the database. The consultancy with its name and the URL of its logo, with its name alone, before the setup (a failure), and with a storage that fails. Who is logged in with the URL of the photo, with none, and with a storage that fails.
- `postgres` in temporary databases: the key of a confirmed file read while another transaction holds the file (no lock), and none for an unconfirmed or unknown file; the consultancy with and without its logo, and none before the setup; the active user found with the file of the photo, by identifier and by e-mail.
- `s3` against Garage: a download signed for the public host name, with `X-Amz-Expires=900`, read by a GET sent as the browser sends it; the URL of one object changed to another key refused with 403; an expired URL refused; the spans of signing a download and an upload carry no key.
- Through the binary: who is logged in carries `photo_url`, and the consultancy its name and `logo_url`, each signed for the public host name for 15 minutes and reading the PNG sent in the setup; with no images both leave the URLs out; the consultancy answers 401 with no session; the spec declares the 401 of both reads. The test of sensitive values gained a scenario that reads both with the images, then fails both with a 5xx, and finds neither the keys of the images, nor the signatures of their URLs, nor the secret key of the storage in the logs or the spans.
- Screen tests (`shared/components/app-shell.test.tsx`, 13): the mark, the product name, the consultancy, the menu and the sentence at a wide width; one button and no search field or link; the initials of the user and of the consultancy; the photo and the logo read from the storage; the menu with the name, the e-mail and one item, the logout to `/login?notice=logged_out`, its failure shown in the open menu; the menu opened and the logout chosen by keyboard; the preferred language `pt-BR` with a browser in `es`; the sheet at 375px and at 959px, and the sidebar beside the work surface at 960px; no page wider than 375px; the photo and the logo read again with new URLs ten minutes later (fails without the interval). The gate and login tests now look for the user menu instead of the landing.
- In Chromium through Caddy (`http://localhost`), with the API and the worker built from this worktree on the development database (migration 25, seeded, no images) and Vite on 5173, driven by a Playwright script outside the repository: logging in through Dex as `seed+admin@fabrica.example` (the button "Log in with Local login provider" of ticket 08) landed at `/` in the shell, in `pt-BR`: the button "Menu de Ana Exemplo" with "AE", the sidebar with "FE" and "Fábrica Exemplo", the sentence on the work surface, one button in the page; the dark theme followed the browser; the menu showed "Ana Exemplo", `seed+admin@fabrica.example` and "Sair"; at 375px no page wider than the window, and "Abrir a barra lateral" opened the sheet "Barra lateral" with the consultancy; `GET /api/session` had no `photo_url` and `GET /api/consultancy` was `{"name":"Fábrica Exemplo"}`; "Sair" led to `/login?notice=logged_out` with "You logged out". The API, the worker and Vite were stopped after; the development database keeps two `logged_in` entries through `local`, and the worker ran the cleanup of the storage when it started, as `task dev` does.
- `go build`, `go vet`, `task api:lint`, `task api:generate:check`, `task api:spec:check`, `task api:spec:lint`, `task api:lists:check`, `task versions:check`, `go test -count=1 -p 2 ./...` of `apps/api`; `task web:generate`, `task web:check` (generate check, `vp check`, imports, knip, missing keys, 155 screen tests), with `OXLINT_TSGOLINT_PATH` set to the worktree's `tsgolint.exe`. Run again after `feat/login` brought tickets 07 and 08.

### What was not checked

- `task check:quick`, `task check` and `task api:test` as wholes (they start the shared stack); each of their parts ran on its own. `task e2e` belongs to ticket 12.
- A photo and a logo in the browser through Caddy and `files.localhost`: the seeded setup has no images, and none was added to the development database. The binary tests read both through the signed URLs from Garage, and the screen tests show both.
- The Chrome DevTools MCP was not connected; the browser check used Playwright's Chromium instead.

### Found, and left for a decision

- **The other areas of the shell.** The search field, the command palette, notifications, help, the items of the sidebar, favourites and settings wait for their features, as DESIGN.md now says; the settings of the next effort will decide whether they reuse `GET /api/consultancy` with more fields, as the spec left open.
- **The coloured circles of people** (Traits of DESIGN.md) are muted initials for now; their colour is chosen with the first screen that tells people apart.
- `docs/machine-tools.yaml` and `docs/providers.md` need no change.

### Settled at the end of the effort

- **The other areas of the shell**: nothing to settle; each arrives with its feature, as DESIGN.md says.
- **The colour of the circles of people**: deferred by the maintainer to the first screen that tells people apart. Recorded under Deferred of the frontend rules in `CODING_STANDARDS.md`.
