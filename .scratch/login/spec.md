# Login and the shell of the app

Status: ready-for-agent

The second of three efforts. The first laid the foundation and the setup; the third is the consultancy settings, the users and the system health, specified when this one ends. The decisions below were taken in the grilling of 2026-11-18 and stand on ADR-0011 (login and session), ADR-0016 (jobs), ADR-0017 (files), ADR-0019 (seed), ADR-0024 (setup), ADR-0026 (roles and scopes) and the sections Login, Jobs, E-mail, Files, Seed and Rate limit of `CODING_STANDARDS.md`. The terms are in `GLOSSARY.md`: user, administrator, developer, role, scope, login code, login provider, local login provider, provider link, session, audit log, setup code. Do not reopen them.

## Problem Statement

The setup is concluded and the first administrator exists, and nothing happens next. The login screen is a placeholder that says the login is not built. No session exists, so no request can tell who is acting. The system sends no e-mail and runs no job, so a login code has no way to leave. There is no local login provider, so neither an agent nor an automated test can exercise the login that production will use. A developer who starts the system from scratch has to go through the setup by hand every time, because there is no seed. The images sent in the setup cannot be shown, because nothing signs a URL to read them, and uploads that were never confirmed stay in the bucket for ever.

## Solution

A registered user opens the app and finds the login screen. They enter through a login provider, with the account they already have, or type their e-mail and receive a six-digit login code, which they type on the same screen. Either way they land in the shell of the app: the header with the brand and their own menu, the sidebar with the name and the logo of their consultancy, and a work surface that says the areas of work come in the next efforts. The app speaks their preferred language. From the menu they log out and return to the login screen. A session that ends while they work sends them back to the login with a notice, and brings them back to where they were once they log in again.

Under that, this effort brings the pieces every later feature needs: River and the `worker` command, the e-mail records and their sending, Mailpit and Dex in development, the seed, the periodic job that cleans the file storage, signed URLs to read an image, and the check that a request that changes state came from the app's own origin.

In development and in the tests the only login provider is the local one, backed by Dex. Google, Microsoft and GitHub arrive with the deploy, when the system has an address and an application registered at each.

## User Stories

### Reaching the login

1. As a user, I want every address of the app to lead to the login screen while I am not logged in, so that I never see a half-working screen.
2. As a user who is logged in, I want the address of the login screen to lead to the shell, so that I do not log in twice.
3. As a user, I want the login screen in the language of my browser, with the selector for the other supported languages, so that I can read it before the system knows my preferred language.
4. As a user, I want the login screen to work at a narrow width and in both themes, so that I can log in from the device I have at hand.
5. As a user whose session ended, I want to be taken to the login screen with a notice that says so, so that I understand why I am there.
6. As a user whose session ended, I want to return to the address I was at once I log in again, so that I do not look for it.
7. As the first administrator who just concluded the setup, I want the login screen to tell me the setup is concluded and that I log in below, so that the notices of the setup still make sense now that the login exists.

### Logging in with a login provider

8. As a user, I want a button for each login provider that is turned on, so that I choose the account I already have.
9. As a user, I want the button to take me to the provider, and the provider to bring me back logged in, so that I type no password in the system.
10. As a user, I want a login through a provider accepted only when the provider says my e-mail is verified, so that nobody logs in as me with an account that merely claims my e-mail.
11. As a user, I want my first login through a provider matched by my verified e-mail, so that the administrator who registered me did not have to know my account at the provider.
12. As a user, I want the system to remember my account at the provider after the first login, so that a change of my e-mail at the provider does not log someone else in as me.
13. As a user, I want to log in through more than one provider, so that one provider being down does not lock me out.
14. As a person whose verified e-mail belongs to no active user, I want a notice on the login screen that says no active user has that e-mail and tells me to ask an administrator, so that I know what to do.
15. As the maintainer, I want the same notice whether the e-mail is unknown or belongs to a deactivated user, so that a deactivated person learns nothing about their record from the login.
16. As the maintainer, I want the e-mail of that notice to travel in no address, so that it reaches no browser history, log or span.
17. As a user who cancelled at the provider or whose provider answered an error, I want to be back on the login screen with a notice that the login did not conclude, so that I can try again.
18. As the maintainer, I want the authorization code flow protected against a forged callback and a replayed token, so that a link someone else crafted logs nobody in.

### Logging in with a login code

19. As a user, I want to type my e-mail and ask for a login code, so that I can log in without an account at a login provider.
20. As a user, I want the screen to move to the step of the code once I ask for one, saying that a code was sent if the e-mail belongs to a user, so that I know what to do next without the screen asserting something that may be false.
21. As a user, I want the code to arrive in my mailbox in my preferred language, with how long it works, so that I can use it without guessing.
22. As a user, I want to type the six digits and be logged in, so that the login takes two steps and nothing more.
23. As a user, I want the code to work for 10 minutes and once, so that a code I used or left behind opens nothing.
24. As a user, I want the code to work only in the browser that asked for it, so that a code read from my mailbox elsewhere logs nobody in.
25. As a user, I want a wrong code to tell me so and let me try again, up to three times, so that a typing mistake costs me little.
26. As a user, I want the third wrong attempt to end the code and take me back to the e-mail step with a notice, so that I know I have to ask for a new one.
27. As a user, I want to ask for a new code from the step of the code, 60 seconds after the last one, so that a code that never arrived does not block me.
28. As a user, I want a new code to end the one before, so that only the last code I received works.
29. As a user, I want the e-mail step to be the same whether or not my e-mail is registered, so that nobody learns from the login who has an account.
30. As a user who asked for too many codes, I want a message that says to wait, so that I do not think the e-mail is wrong.
31. As the maintainer, I want one code a minute and five an hour for each e-mail address, so that a mailbox cannot be flooded through the login.
32. As the maintainer, I want the two operations of the login code limited for each IP address in Caddy, with the 429 in the error format of the API, so that one caller cannot hammer them for many addresses.
33. As the maintainer, I want nothing of the code stored, logged or traced, so that a dump of the database or of the logs reveals no code.
34. As the maintainer, I want the code of a deactivated user never sent, with the same answer on the screen, so that a deactivated user cannot log in and learns nothing from trying.
35. As the maintainer, I want the e-mail of the code to leave through a job, so that the response time of asking does not tell whether the e-mail belongs to a user.
36. As the maintainer, I want a failed sending visible in the e-mail record and tried again by the job, so that an e-mail server that was down for a minute costs nobody their login.

### The session

37. As a user, I want to stay logged in across page loads and browser restarts, so that I log in once in a while and not at every visit.
38. As a user, I want my session to end after 7 days without use and, in any case, 30 days after the login, so that a forgotten browser does not stay logged in for ever.
39. As the maintainer, I want the browser to hold only an opaque cookie that JavaScript cannot read, so that a script injected in the page steals no session.
40. As the maintainer, I want a request that changes state refused when it does not come from the app's own origin, so that a page elsewhere cannot act with the cookie of a logged-in user.
41. As the maintainer, I want every request under the contract to carry who is acting, with the scopes of their roles, so that the use cases of the next efforts check scopes without changing how the actor arrives.
42. As the maintainer, I want a deactivated user refused on their next request and their session ended, so that deactivating a user takes effect at once.
43. As a user, I want to log out from my menu and land on the login screen with a notice that I logged out, so that I know the session is over.
44. As the maintainer, I want every login recorded in the audit log, with the user and the way in, so that I can tell later who entered and how.
45. As the maintainer, I want nothing recorded for a logout or a wrong code, so that the audit log holds sensitive operations and not noise.

### The shell

46. As a user, I want to land in the shell of the app after I log in, so that I see where the work will happen.
47. As a user, I want the header to show the brand, the name of the product and my own menu, so that I know where I am and who I am logged in as.
48. As a user, I want my menu to show my name, my e-mail and my photo, with my initials when I have no photo, so that I can tell at a glance I am on the right account.
49. As a user, I want the sidebar to show the name and the logo of the consultancy, with its initials when it has no logo, so that the shell is the consultancy's.
50. As a user, I want the work surface to say that the areas of work come in the next efforts, so that an empty page does not look broken.
51. As a user, I want nothing in the shell that does nothing: no search, no notifications, no help, no settings, until each works, so that I never click a dead control.
52. As a user, I want the shell in my preferred language from the moment I log in, so that the system speaks my language whatever the browser asks for.
53. As a user, I want the shell to work at a narrow width, with the sidebar as a sheet opened from the header, and in both themes, so that the first screen after the login holds the rules of the design system.
54. As a user, I want my photo and the logo shown from the private storage without passing through the API, so that an image loads as fast as the storage serves it.
55. As the maintainer, I want the URL of an image to expire in minutes and to be renewed when the data is read again, so that a URL that leaks stops working soon.

### The local login provider

56. As a coding agent, I want a login provider in development and in the tests that needs no account outside, so that I can log in and exercise the same OAuth code production runs.
57. As the maintainer, I want the local login provider turned on by configuration alone and declared only in the development Compose file, so that production and staging cannot have it.
58. As the maintainer, I want the local login provider to show on the login screen and in a log line at the start of the process, so that its presence is never silent.
59. As a coding agent, I want the people the local provider knows to be the users of the seed, so that logging in as one of them needs no registration.

### E-mail

60. As the maintainer, I want every e-mail the system sends recorded with its recipient, its subject, what it refers to and its status, so that I can tell what was sent and what failed.
61. As the maintainer, I want the e-mail record and its sending job saved in the same transaction, so that no e-mail is recorded and never sent, or sent and never recorded.
62. As the maintainer, I want the sending done by one of two adapters chosen by configuration, SMTP or Resend, so that development, tests and staging use Mailpit and production uses Resend.
63. As the maintainer, I want the Resend adapter to send the identifier of the record as the idempotency key, so that a job that runs twice sends one e-mail.
64. As the maintainer, I want a bounce reported by Resend to mark the record as bounced, after the signature on the raw body is verified and a repeated event is ignored, so that a bounce is recorded once and a forged report changes nothing.
65. As the maintainer, I want the bounce webhook to exist only when its secret is configured, so that an install without Resend exposes no such path.
66. As a user, I want each e-mail with an HTML part and a plain-text part in my preferred language, so that it reads well in any mail client.
67. As a coding agent, I want the build to fail when a supported language lacks a text of an e-mail, so that a missing translation never reaches a mailbox.
68. As the maintainer, I want no e-mail address in a log, a span, a metric label or a job argument, so that the rule of privacy holds for e-mail as for everything else.

### Jobs and the worker

69. As the maintainer, I want the jobs stored in PostgreSQL by River and run by the `worker` command of the API binary, so that a job is saved with the change that asked for it and runs outside the API process.
70. As the maintainer, I want River's tables to arrive as ordinary migrations, exported by River's tooling, so that there is one way to change the schema and one version for readiness to check.
71. As the maintainer, I want the API role to hold on River's tables only what the API and the worker need, so that the principle of the two database roles is kept.
72. As the maintainer, I want a job to call one use case with identifiers as its arguments, so that the River panel and the telemetry show no personal data.
73. As the maintainer, I want a light queue and a heavy queue, each with a limit of concurrent jobs taken from configuration, so that heavy work later does not starve the e-mails.
74. As the maintainer, I want the worker to stop as the API does, taking no new job and giving the running ones a deadline, so that a deploy loses no work.
75. As the maintainer, I want the trace that saved a job linked to the trace that ran it, so that an e-mail can be followed from the request that asked for it.
76. As a coding agent, I want `task dev` to start the worker beside the API and the app, so that a login code arrives in development without a second command.

### The periodic cleanup of the storage

77. As the maintainer, I want a job every hour to remove the uploads started more than 24 hours ago and never confirmed, so that abandoned uploads do not pile up in the bucket.
78. As the maintainer, I want the same job to remove the objects no file names, older than 24 hours, found by listing the bucket, so that a failed removal or a confirmation that did not commit leaves nothing behind for ever.
79. As the maintainer, I want the job to tolerate a missed run and to change nothing on a second run over the same state, so that the rules of periodic and idempotent jobs hold.

### The seed

80. As a coding agent, I want one command that fills an empty database with a consultancy and an administrator, so that I open the system on the login screen and log in as that administrator.
81. As the maintainer, I want the seed to create its data through the use cases of the setup, so that seeded data obeys every rule of the domain.
82. As the maintainer, I want two runs of the seed on an empty database to produce the same data with the same identifiers, and a second run on a seeded database to change nothing, so that the seed is a fixed point.
83. As the maintainer, I want the e-mail addresses of the seed derived from one mailbox taken from configuration, with a fictitious default, so that in staging I can log in as any seeded user from my own mailbox.
84. As the maintainer, I want the seed as its own command in its own image, declared by the development Compose file only, so that production never carries it.
85. As the maintainer, I want nothing real in the seed, so that the public repository names no real person, company or address.

### The development environment and the pipeline

86. As a coding agent, I want Mailpit and Dex started by `task services`, with their addresses in `.env.example`, so that the login works in development as it is.
87. As a coding agent, I want the end-to-end test to log in through the local provider and with a login code read from Mailpit, so that both ways in are proven against the real system.
88. As a coding agent, I want the tests of the Resend adapter to run on a schedule and never on a pull request, and to skip when no key is configured, so that a pull request depends on no outside service.
89. As the maintainer, I want the joints this effort was meant to prove recorded as proven or corrected in `CODING_STANDARDS.md` and the ADRs, so that no rule stays unproven in silence.

## Implementation Decisions

### Domain areas

- **`access`** holds the login: the login code request, the session, the provider link, the login providers, and the use cases that ask for a code, check a code, start and conclude a login through a provider, tell who is logged in, and log out. Access tokens join it later.
- **`email`** holds the e-mail record, the use case that queues an e-mail with its job, the use case that delivers one (run by the job), the use case that marks a bounce, the templates and their translations, and the two sending adapters.
- **`files`** gains the use case that answers a signed URL to read a file, used by other areas, and the periodic cleanup job.
- **`consultancy`** gains the read of the consultancy as the shell shows it: its name and the URL of its logo. ADR-0026 asks no scope for it. Whether the settings of the next effort reuse this operation with more fields is decided then.
- **`audit`** gains the operation `logged_in`, with a new actor kind, the user, and a new column that records the way in: the key of the login provider or `login_code`. The CHECK constraints of the audit log table grow with them.
- The **seed** is a `main` package of its own in the Go module, built into its own image, as Seed in `CODING_STANDARDS.md` says.

### The session

- `alexedwards/scs/v2` with `pgxstore`, its table created by a migration. The cookie is opaque, `HttpOnly`, `Secure` (off in development, which has no certificate), `SameSite=Lax`, so the callback of a provider, a top-level navigation, carries it. Idle life 7 days, absolute life 30 days.
- A session exists before the login too. It holds what the login needs between two requests: the identifier of the login code request this browser asked for, the state, nonce and PKCE verifier of a provider login in progress, the path to return to, and the refusal of the last provider login. Logging in renews the session token and stores the user's identifier.
- The REST adapter reads the session on every request under the contract that needs a user. It loads the user and their roles; a deactivated user, or one that no longer exists, has their session destroyed and gets 401. The actor it passes to the use case carries the user's identifier and the scopes of their roles: every scope for the administrator role, none for the developer role yet. No operation of this effort checks a scope; the mechanism is built so the next effort adds the checks only.
- Several sessions of one user may exist at once. Nothing lists or ends the other sessions; that comes with the actions on the session in a later effort.
- **Who is acting** is read by the app from one operation that answers the logged-in user: name, e-mail, the signed URL of the photo when there is one, preferred language and the keys of the roles. It answers 401 with no session.
- **Logging out** is a named `POST` operation on the session (ADR-0009), which destroys it. The app then opens the login screen with the notice `logged_out`.
- **Cross-site requests.** A middleware of the REST adapter refuses, with 403 in the error format, a `POST` or `PUT` that carries the session cookie when its `Origin` header is present and differs from the app origin of the configuration, or when `Origin` is absent and `Sec-Fetch-Site` says `cross-site`. A request with no cookie, such as the bounce webhook and the access tokens of a later effort, is not checked.

### The login screen and the gate

- The gate of the app now decides between three places: the setup, the login and the shell. It reads the state of the setup first, as today, then the logged-in user; a 401 means the login. A logged-in user who opens the login screen is taken to the shell.
- A read that needs no session answers the **login state**: the login providers turned on, each with its key and its display name, and, when the last provider login in this browser was refused, the e-mail it was refused for. The screen draws one button per provider from it, and the notice with the e-mail from it, so the e-mail travels in no address. Reading clears nothing; the next login attempt replaces the refusal.
- The login screen is one screen in the narrow framed pane: the provider buttons, a divider, then the e-mail step. Its flow is one reducer in the component, whose state is a union of named steps: `email`, `code` (with the e-mail, the time the code was asked for, and the attempts left) and `sending`. A page loaded again starts at `email`; what was typed is not kept.
- The step of the code says that a code was sent if the e-mail belongs to a user. It offers a new code 60 seconds after the last request, counted on the screen. A wrong code shows the error under the field and keeps the step; the third wrong attempt, which the API reports as the code ended, returns to the e-mail step with a notice. Past the limit of codes for the address, the API answers the rule broken and the screen says to wait.
- The notices the login screen shows come from the search parameter `notice`, as today, with new values: `session_ended`, `logged_out`, `login_not_concluded` (the person cancelled at the provider or the provider failed), `no_active_user` (whose e-mail comes from the login state). The notices of the setup stay and their texts change: the administrator logs in below.
- When a request of the app answers 401 outside the login, the data layer sends the app to the login with `notice=session_ended` and `return_to`, the path of the app the person was at. After a login, the app opens `return_to` when it is a path of the app, and the shell otherwise. A provider login receives the path when it starts and gets it back from the session at the callback.
- After a login the app switches to the preferred language of the user. The language selector stays only in the footer of the framed pane, before the login; changing the preferred language comes with the profile, in the next effort. Logging out returns the app to the language of the selector.

### The login code

- Asking for a code is a `POST` that takes the e-mail and answers 202 with no body, the same for every e-mail. For an e-mail that belongs to an active user, the use case ends their earlier open requests, saves a **login code request** (identifier, user, a random value, when it was asked, when it expires, attempts made, when it ended) and the e-mail record with its sending job, in one transaction, and stores the identifier of the request in the session. For an unknown or deactivated e-mail nothing is saved.
- The limits for each address are counted on the requests of that user: one in the last minute and five in the last hour. Past them the operation answers 422 with a rule of its own, which the screen translates as "wait".
- The code is six digits computed with HMAC-SHA256 from a key in the configuration over the identifier of the request and its random value, reduced to six digits. The job that delivers the e-mail computes it to write the e-mail; the check computes it to compare, in constant time. The code is in no table, job argument, log or span, and a test of the logs proves it.
- Checking is a `POST` that takes the six digits. The use case reads the identifier of the request from the session: no request in this browser, a request ended or expired, or a wrong code all answer the same rule broken, "the code is not valid", with how many attempts are left; the third wrong attempt ends the request and answers that the code ended. A right code ends the request, logs the user in, and writes `logged_in` to the audit log with `login_code` as the way in.
- Caddy limits the two operations for each IP address in a zone of their own, `login`, with 10 requests a minute counted together, beside the zone of the setup. The 429 is the one already written in the `Caddyfile`. The numbers go to Rate limit in `CODING_STANDARDS.md`.

### Login providers

- One implementation, generic OIDC, serves every provider that issues an ID token: discovery from the issuer URL, authorization code flow with PKCE, `state` and `nonce` kept in the session, the ID token validated with `coreos/go-oidc/v3`, the e-mail taken from the token only when the token says it is verified. A provider is a set of configuration variables: issuer URL, client identifier, client secret, display name, and a key that names it in paths, in provider links and in the audit log. In this effort the only one configured is the local login provider; Google and Microsoft arrive later as configuration plus the check of how each states a verified e-mail, and GitHub as a second implementation.
- Starting a login is a `GET` that the browser navigates to, with the key of the provider in the path and the path to return to as a query parameter; it stores state, nonce, verifier and return path in the session and redirects to the provider. The callback is a `GET` with the key in the path; it checks the state, exchanges the code, validates the token and the nonce, and then: finds the provider link by provider and subject, or, failing that, the active user with that verified e-mail and creates the link; logs the user in, writes `logged_in` with the key of the provider as the way in, and redirects to the return path or the shell. A verified e-mail of no active user stores the refusal with the e-mail in the session and redirects to the login with `notice=no_active_user`. A cancel or an error at the provider redirects with `notice=login_not_concluded`. Both operations are in the contract as operations that answer a redirect; the generated client does not call them.
- A **provider link** is a row: provider key, subject, user; the pair of provider and subject is unique. It is created at the first login through that provider and never changed by the login.
- The local login provider is **Dex**, a service of the development Compose file only, published on a port of the machine so that the browser and the API binary, which runs on the machine, reach the same issuer URL (`http://localhost:<port>`). Its configuration file holds one static client and a fixed list of people with fixed passwords: the administrator and the developer of the seed, with the default mailbox. The developer is not a user until the next effort; logging in as them shows the notice of an e-mail with no active user, which the screen tests use. When the provider is on, a log line at the start of the API names it; the health report shows it in the next effort.

### The shell

- A layout route wraps every screen after the login and draws the App shell recipe of `DESIGN.md`, cut down to what works: the header with the brand mark, the product name and the user menu; the sidebar with the logo or the initials of the consultancy and its name; the work surface with one sentence. Under 960px the sidebar is a sheet opened from the header. The recipe gets its code with this screen, and `DESIGN.md` records what was left out and why.
- The user menu shows the name, the e-mail and the photo or the initials, and one action: log out.
- The signed URLs of the photo and of the logo are embedded in the answers that carry them, the logged-in user and the consultancy, valid for 15 minutes; the app reads both again at an interval shorter than that, so an image never breaks on the screen. The `files` area signs a `GET` for the public host name of the storage, as it signs the `PUT` of an upload.

### E-mail

- An **e-mail record** has: recipient, subject, what it refers to (a kind and an identifier, here the login code request), its status (`queued`, `sent`, `failed`, `bounced`), the identifier the sender returned, when it was sent or failed, and a stable error code when it failed. The use case that queues an e-mail saves the record and its job in the caller's transaction. The job calls the use case that delivers: it renders the template, sends through the adapter chosen by configuration, and marks the record `sent` or, when the attempts run out, `failed`. A job that runs again on a record already `sent` does nothing.
- The **SMTP adapter** uses `go-mail`, pointed at Mailpit in development, tests and staging. The **Resend adapter** uses Resend's API with the identifier of the record as the idempotency key and as a tag, and stores the identifier Resend returns. The sender address and name come from configuration, with a fictitious default.
- The **bounce webhook** is an operation under `/api`, outside the OpenAPI spec, registered only when its secret is configured. It verifies the signature on the raw body, ignores an event already seen (their identifiers are kept), and calls the use case that marks the record `bounced`. It carries no cookie and is not subject to the origin check.
- Templates live in the backend, one per kind of e-mail, with an HTML part and a plain-text part, and their texts in translation files for every supported language; a test fails when a language lacks a key, as the frontend's does. The first template is the login code: the code, how long it works, and that the e-mail is to be ignored by whoever did not ask for it. The language is the recipient's preferred language. The format of the templates and of their translation files is chosen with this code and written into E-mail in `CODING_STANDARDS.md`.

### Jobs and the worker

- River, with its tables exported by River's tooling into numbered goose migrations, one per River schema version, and the grants the API role needs on them in the same migration. Where that tooling runs from, and whether it is installed on the maintainer's machine, is recorded: in `docs/machine-tools.yaml` if installed there, otherwise as a Go tool of the module, listed in Versions.
- The `worker` command of the API binary runs the jobs, with two queues, light and heavy, each with a limit of concurrent jobs taken from configuration, and a shutdown deadline taken from configuration. The e-mail job and the cleanup job run in the light queue; the cleanup declares a longer timeout. River's OpenTelemetry middleware traces the jobs, and whether it links the saving trace to the running trace is checked and recorded.
- The River panel stays in development only; the next effort serves it behind the session to administrators, with the role check.
- The progress stream of ADR-0016 (`NOTIFY`, Server-Sent Events) is not built: no job of this effort has a screen that waits for it.

### The periodic cleanup of the storage

- A periodic job declared in the worker, every hour. It lists the bucket and, for each object older than 24 hours, removes it when no confirmed file names its key; it marks as abandoned, or removes, the file records of uploads started more than 24 hours ago and never confirmed. The files of the setup confirmed before this effort are untouched, because a confirmed file names its key. The job tolerates a missed run and a second run over the same state.

### The seed

- The seed issues a setup code and concludes the setup through the use cases of the setup, with a fictitious consultancy and its administrator at `<local part>+admin@<domain>`, derived from a mailbox taken from configuration whose default is fictitious and matches Dex. A run on a database whose setup is concluded does nothing. The developer at `<local part>+developer@<domain>` is listed in Dex and created by the seed in the next effort, when registering a user exists.
- `task seed` runs it in development on the development database. Staging runs it on every publish, with the deploy effort.

### Configuration

New variables, each with a value that works in development and listed in `.env.example`: the key of the login code; for the local login provider, its issuer URL, client identifier, client secret and display name (empty turns it off); the sender of e-mail (`smtp` or `resend`), the SMTP address, the sender address and name, the Resend key and the webhook secret; the limits of the two queues and the shutdown deadline of the worker; the mailbox of the seed; the ports of Dex and Mailpit for Compose. The `worker` command requires what it needs and nothing of the API's listening; the `seed` command requires the database and the mailbox.

### Records to keep current

- `CODING_STANDARDS.md`: the numbers of the login code in Rate limit; the format of e-mail templates in E-mail; what Deferred called the second effort moved to what is built, and each joint of this effort marked as proven or corrected; the `seed` and `worker` commands in Commands.
- `DESIGN.md`: the App shell recipe gets code and the record of what this effort leaves out; the login screen as the second use of the narrow framed pane.
- `README.md`: the status line, Local development with the login, `task seed`, Dex and Mailpit.
- `GLOSSARY.md`: done in the grilling (session, provider link, local login provider).
- `docs/providers.md`: nothing changes now. The Resend row stays "decided, not created" and the sender address "not chosen" until an account exists.
- `docs/machine-tools.yaml`: only if River's tooling is installed on the maintainer's machine.

## Testing Decisions

A good test exercises behaviour that someone outside the module can see, and survives a rewrite of the inside. It asserts on what a use case returns and saves, on what an operation answers, on what arrives in a mailbox, or on what a screen shows. It does not assert on which function was called. Assertions use testify.

- **Use cases, with in-memory implementations of their interfaces.** The default seam, where every rule is tested: asking for a code for a known, unknown and deactivated e-mail, with the same answer; the limits of one a minute and five an hour; a new code ending the one before; checking a right code, a wrong one three times, an expired one, one from another browser; the code computed the same by the delivery and by the check and never stored; the provider login matched by link, then by verified e-mail, refused for an unverified e-mail, an unknown one and a deactivated user; the link created once; who is logged in and the logout; the `logged_in` entry with each way in; the e-mail record and its job saved together, delivered once, marked failed when the attempts run out, bounced once for a repeated event; the cleanup over objects named, unnamed, young and old, twice; the seed twice. A fixed clock drives every expiry and limit. Prior art: the use case tests of the setup in `consultancy/application`.
- **The `postgres` adapters, against the development PostgreSQL**, in temporary databases from pgtestdb: the session store, the login code requests and their counts, the unique provider link, the e-mail records, the audit log accepting the new operation and refusing changes as before, and the grants on River's tables. Prior art: the repository tests of the setup code.
- **The `s3` adapter, against Garage**: the signed `GET` for the public host name, the listing with last-modified dates, the removal. Prior art: the existing `s3` tests.
- **The `smtp` adapter, against Mailpit** (new): an e-mail sent arrives with both parts and the right recipient, read through Mailpit's API. **The `resend` adapter, against Resend's test addresses** (new): runs only in the scheduled workflow and skips without a key.
- **The `river` adapter, with River's test helpers in a temporary database** (new): a job worker reads its arguments and calls its use case; the periodic job is scheduled.
- **The `oidc` adapter, against the Dex container** (new): discovery, the redirect built with state and PKCE, the exchange of a code and the validation of the token, through Dex's own login page driven by HTTP.
- **The REST adapter, through HTTP, by the binary**: the cookie set at login and its attributes; 401 with no session and after deactivation; 403 on a cross-origin `POST`; no code, e-mail or token in any path, query parameter or header of the spec; the bounce webhook refusing a wrong signature and absent without a secret; the error format on each failure. Prior art: `cmd/api` tests and `platform/restapi` tests.
- **Caddy, through HTTP, in a container**: the `login` zone answering 429 in the error format after ten requests, and the setup zone unaffected. Prior art: `caddy_test.go`.
- **The screens, whole, in Vitest's Browser Mode with MSW**: the login screen with one and several providers, the two steps, the conditional message, the error under the field, the third wrong attempt, the wait for a new code, the limit reached, each notice; the gate between the setup, the login and the shell, a 401 mid-work with the return after the login, a logged-in user at the login address; the shell with the menu, the initials, the logout, the preferred language and the sheet at a narrow width. Prior art: the setup screen tests.
- **One end-to-end test in Playwright, through Caddy**, continuing the script of the setup on the same database: the setup registers the administrator with the e-mail Dex knows; then a login through Dex, the shell, the logout, a login with a code read from Mailpit, the shell again. A separate check spends the login limit and reads the 429; it runs last. Prior art: `e2e/setup.spec.ts`.
- **`platform`**: the configuration failing at the start and naming each new variable; the test of the logs covering the login code, the e-mail address and the session cookie.

## Out of Scope

- Google, Microsoft and GitHub as configured providers, the check of how each states a verified e-mail, and GitHub's implementation without an ID token. They come with the deploy, with the applications registered and the redirect addresses in `docs/providers.md`.
- Access tokens, MCP, and OAuth for agents.
- Registering, editing and deactivating users, the roles screen, the profile and the change of the preferred language, the consultancy settings: the third effort.
- The health report, with the state of the worker, the e-mail server and the login providers; the River panel behind the session; the scopes checked by use cases.
- A download operation of its own for files; any file other than the two images.
- Search, the command palette, notifications, help, favourites and settings in the shell; listing or ending other sessions.
- Attachments, e-mail other than the login code, receiving e-mail, events other than a bounce.
- The progress stream of jobs and anything that waits on a screen for a job.
- Mailpit in staging behind a password, the seed on every publish, and anything of the deploy.
- The LLM stand-in.

## Further Notes

- The decision to write the Resend adapter and the bounce webhook now, covered only by simulated responses and a scheduled test that cannot run until an account exists, was the maintainer's, against the recommendation of SMTP alone. The pull request that builds them states which paths were never exercised against Resend.
- The code of the login is HMAC over the request; a change of the key ends only the codes not yet used (ADR-0011).
- Dex on a port of the machine, rather than behind Caddy at a host name under `localhost`, is chosen because the API binary runs on the machine and the issuer URL has to be the same for the browser and for the API; a name under `localhost` is resolved by browsers and not reliably by the resolver of every operating system. If the deploy later puts the API in a container, the issuer moves with it.
- Logging in from the setup is still refused: ADR-0024 keeps the setup without a session, and the notices of the login screen are what links the two.
