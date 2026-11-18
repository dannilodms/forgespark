# Foundation and setup

Status: resolved

The first of three efforts. The second is the login and the shell of the app, and the third is the consultancy settings, the users and the system health. Each of those is specified when the one before it ends.

## Problem Statement

Whoever installs the system gets nothing they can use. `apps/api` is empty, `apps/web` shows a provisional page, and no command starts a database, an API or a file storage. A new install has no consultancy and no user, and nothing can create them: ADR-0011 lets in registered users only, and nobody is registered.

The maintainer and the coding agents have the same problem from the other side. The ADRs and `CODING_STANDARDS.md` describe a backend, a contract, a pipeline and a development environment that were read in documentation and never run, so every rule in them is still unproven.

## Solution

Whoever installs the system runs one command in the API container. It prints a setup code. They open the app, which shows only the setup screen, type the code, fill in the consultancy and its first administrator, review what they typed and conclude. The consultancy, the administrator and the two first roles are saved together, the setup code stops working, and the app leads to the login screen. In this effort that screen is provisional: it says the setup is concluded and that the login is not built yet.

When no administrator can log in, the same command issues a new setup code, which opens the setup at the step of the administrator and registers a new one.

Under that, this effort lays the foundation that every later feature stands on: the Go module with its layout, the database with migrations, the REST contract generated from Go and the client generated from it, the file storage, the audit log, the development environment in Compose behind Caddy, and the checks of a pull request.

The decisions are in ADR-0024 (setup), ADR-0026 (roles and scopes), ADR-0017 (files), ADR-0012 (audit log) and in the sections Setup and Users and roles of `CODING_STANDARDS.md`. The terms are in `GLOSSARY.md`. Do not reopen them.

## User Stories

### Issuing a setup code

1. As whoever installs the system, I want one command of the API binary to print a setup code, so that I can open the setup without a user or a password existing.
2. As whoever installs the system, I want the command to say what the code opens, the first setup or the registration of a new administrator, so that I know what I am about to do.
3. As whoever installs the system, I want the command to print when the code stops working and the address to open, so that I do not have to look either up.
4. As whoever installs the system, I want the code in four short groups with no character that reads like another, so that I can type it without mistakes.
5. As whoever installs the system, I want a new setup code to end the one before, so that a code I lost sight of opens nothing.
6. As whoever installs the system, I want the command to fail with a clear message when the database is unreachable or the migrations were not applied, so that I know what to fix.
7. As the maintainer, I want the setup code stored only as a hash and absent from every log and span, so that a dump of the database or of the logs opens no setup.
8. As the maintainer, I want each setup code issued recorded in the audit log, so that I can tell later that the setup was opened.

### Reaching the setup

9. As whoever installs the system, I want every address of a new install to lead to the setup screen, so that I cannot land on a screen that does not work yet.
10. As a visitor to an install whose setup is concluded, I want the address of the setup to lead to the login screen, so that the setup cannot be seen again.
11. As whoever installs the system, I want the address of the setup to show the field for the code while a setup code is issued and still valid, so that I can register a new administrator when none can log in.
12. As whoever installs the system, I want the setup screen in the language of my browser, with a selector for the other supported languages, so that I can read it before any user has a preferred language.
13. As whoever installs the system, I want the setup screen to work at a narrow width and in both themes, so that I can do the setup from the device I have at hand.

### Typing the setup code

14. As whoever installs the system, I want to type the code in upper or lower case, with or without the hyphens, so that the way I type it does not matter.
15. As whoever installs the system, I want a typed O, I or L read as the digit it resembles, so that a misread character still works.
16. As whoever installs the system, I want one message for a code that is wrong, expired or already used, telling me to issue a new one, so that I know what to do in each case.
17. As whoever installs the system, I want a wrong code to leave my setup code working, so that a typing mistake costs me nothing.
18. As the maintainer, I want the checking of a setup code limited for each IP address, so that the operation cannot be hammered.
19. As a person who is over that limit, I want a message that says to wait, so that I do not think the code is wrong.
20. As the maintainer, I want the code kept only in the memory of the screen and sent in the body of each request, so that it never reaches the browser history, an address or a header.
21. As whoever installs the system, I want to be asked for the code again after I load the page again, so that a code is never left stored in the browser.

### The consultancy

22. As whoever installs the system, I want to register the name the consultancy is known by and its legal name, so that the system shows the first and documents carry the second.
23. As whoever installs the system, I want to choose the country of the consultancy from a list written in my language, so that I do not type a code.
24. As whoever installs the system, I want to type the tax ID as my country writes it, so that no rule of another country refuses it.
25. As whoever installs the system, I want the time zone suggested from my browser and the currency suggested from the country, so that I only confirm them.
26. As whoever installs the system, I want to leave out the e-mail, the telephone, the website and the address, so that I can conclude the setup with what I know now.
27. As whoever installs the system, I want an address in lines, city, region and postal code, so that an address of any country fits.
28. As whoever installs the system, I want to choose a logo and see it on the screen, so that I know it is the right image.
29. As whoever installs the system, I want an image of the wrong type or over 2 MB refused when I choose it, with the reason, so that I pick another.
30. As whoever installs the system, I want to remove or replace the logo I chose, so that a wrong choice is not final.

### The administrator

31. As whoever installs the system, I want to register the name, the e-mail, the country and the national ID of the first administrator, so that the consultancy has a person who can log in and register the others.
32. As whoever installs the system, I want to leave out the telephone and the photo of the administrator, so that I can conclude without them.
33. As the first administrator, I want my preferred language to be the language the setup screen was in when it was concluded, so that the system speaks my language from the first login.
34. As the first administrator, I want my e-mail stored in lowercase, so that the way it was typed does not keep me from logging in.

### Reviewing and concluding

35. As whoever installs the system, I want to move back and forth between the steps without losing what I typed, so that I can correct an earlier step.
36. As whoever installs the system, I want a review of everything before concluding, the e-mail of the administrator in plain sight, so that I catch a typing mistake while I am still at the server.
37. As whoever installs the system, I want an error on the field it is about, in my language, so that I know what to correct.
38. As whoever installs the system, I want an error to show an identifier I can pass on, so that the maintainer finds it in the logs.
39. As whoever installs the system, I want the consultancy, the administrator and the first roles saved together or not at all, so that a browser closed halfway leaves no consultancy without an administrator.
40. As whoever installs the system, I want the setup code to stop working when I conclude, so that nobody opens the setup after me.
41. As whoever installs the system, I want to be asked for a new code when mine expires while I fill in the setup, with what I typed still on the screen, so that an hour is not a deadline for typing.
42. As whoever installs the system, I want an image I already sent to stay valid when I continue with a new code, so that I do not choose it again.
43. As whoever installs the system, I want concluding to lead to the login screen with a notice that the setup is concluded, so that I know it worked and what comes next.
44. As a second person who opened the setup with the same code, I want to be told the setup is already concluded when I try to conclude, so that I do not think my data was saved.
45. As the maintainer, I want the concluding of a setup recorded in the audit log, so that I can tell when the consultancy was registered and which administrator was created.

### The first roles

46. As the first administrator, I want to hold the administrator role from the moment the setup is concluded, so that I may do everything once the login exists.
47. As the maintainer, I want the developer role created in the setup with no scope, so that the first developers registered later have a role to receive.

### The way back in

48. As whoever installs the system, I want to issue a setup code after the setup is concluded, so that a consultancy whose administrators cannot log in gets a new one.
49. As whoever installs the system, I want that code to open the setup at the step of the administrator, so that I do not see or change the data of the consultancy.
50. As whoever installs the system, I want the e-mail of an existing user refused, with the reason, so that the setup changes no user it did not create.
51. As an existing user, I want my data, my roles and my sessions left as they are when a new administrator is registered, so that the way back in affects nobody else.
52. As the maintainer, I want the registration of an administrator by a setup code recorded in the audit log, so that the use of the way back in leaves a trace.

### The foundation

53. As a coding agent, I want one command that starts the services of development and one that starts the API and the app, so that I can run the system without reading a script.
54. As a coding agent, I want one command that runs every check of a pull request, so that what fails in the pipeline fails the same way on the machine.
55. As a coding agent, I want the build to fail when an import breaks the dependency rule, in the backend or in the frontend, so that the layout is kept by tooling.
56. As a coding agent, I want the build to fail when the OpenAPI spec, the code generated by sqlc or the generated client is stale, so that the contract and the code never drift apart.
57. As a coding agent, I want the build to fail on a version that is not exact, so that a range or a floating tag never enters.
58. As a coding agent, I want the build to fail on a file, an export or a dependency that nothing uses, so that dead code does not pile up.
59. As the maintainer, I want every REST failure in one format, unknown routes, wrong methods and panics included, so that the app translates any error by its code.
60. As the maintainer, I want the API to refuse to start when a variable is missing or invalid, naming the variable and never its value, so that a wrong configuration shows at once.
61. As the maintainer, I want the API to stop without cutting a request short, so that a restart loses no work.
62. As the maintainer, I want liveness and readiness answered outside the contract, and `/up` through Caddy, so that Compose and a monitor can tell whether the system is up.
63. As the maintainer, I want a test that proves the logs carry nothing sensitive, so that the rule is kept by more than attention.
64. As the maintainer, I want the three conventions of the first backend code written into `CODING_STANDARDS.md` when the code first uses them, so that every later session finds them.
65. As the maintainer, I want the pull request to run the tests of use cases, of adapters, of screens and the end-to-end test, so that a merge into `main` was exercised whole.

## Implementation Decisions

### Shape of the backend

- One Go module in `apps/api`, laid out as `CODING_STANDARDS.md` describes. It builds one binary with three commands in this effort: the one that serves the API, `migrate` and `issue-setup-code`. The worker command arrives with River in the second effort. Commands are parsed with the standard library.
- Dependencies are wired by hand in the `main` of each command, with no injection library. A use case receives a clock when it is built and takes the current time from it. Both are recorded in `CODING_STANDARDS.md` with the first code that uses them.
- Configuration follows the Configuration rules: environment variables read once into one typed struct, a failure at the start that names the variable, and a committed example file.
- The domain areas of this effort are `consultancy`, `files` and `audit`. `platform` holds the configuration, the pool, the migrations, the identifier encoder, the process lifecycle, liveness and readiness, the logger and the tracer. `kernel` gets the value objects that more than one area uses, starting with the supported language and the country.

### Database

- PostgreSQL 18, one database and one schema, reached through pgx and sqlc. Migrations are numbered SQL files applied by goose through the `migrate` command. The API never changes the schema.
- Two database roles are created now: the migration role, which owns the tables, and the API role, which reads and writes rows. On the audit log table the API role can only insert and select. The monitoring role arrives with the telemetry collector.
- The first migration creates `pg_stat_statements`, which proves that the migration role can do it without being a superuser. (It did not hold: see Joints this effort proves.)
- Tables of this effort, in prose:
  - The consultancy: its name, legal name, country, tax ID, time zone, currency, the optional e-mail, telephone and website, the optional address (two lines, city, region, postal code) and the optional logo, held as the identifier of a file. The database refuses a second row.
  - Users: name, e-mail in lowercase and unique, country, national ID, the optional telephone and photo, the preferred language, and whether the user is deactivated.
  - Roles, and which user has which role. A role created by the setup carries a fixed key, `administrator` or `developer`, which the frontend translates. A name typed by a person comes with the screen that creates roles.
  - Setup codes: the hash, when it was issued, when it expires and when it was ended. Whether a code is valid follows from those and is never stored.
  - Files: the object key, the declared type and size, what the file is for, and when its upload was confirmed.
  - The audit log: when, which operation, who acted and on what.
- No scope exists yet, because no use case of this effort checks one. The table that links a role to its scopes arrives with the first scope, in the third effort. The administrator role stores no scope: it grants all of them by rule.

### The setup code

- Twenty characters of Crockford's Base32 in four groups of five, about 100 bits, from the cryptographic random source. The letters I, L, O and U are never printed. Reading a typed code ignores case, hyphens and spaces, and reads O as 0 and I or L as 1.
- Only the SHA-256 hash of the normalised code is stored, and a code is found by its hash. It lasts one hour. Issuing a code ends every earlier one.
- A wrong code changes nothing. The only limit is in Caddy.
- What a code opens is not stored: it opens the first setup while no consultancy exists, and the registration of a new administrator afterwards.

### Use cases of the `consultancy` area

- **Issue a setup code.** Run by the command only, with no REST operation (ADR-0008). It ends earlier codes, stores the hash, writes to the audit log and returns the code, its expiry and what it opens. The command prints them and the address of the setup to its own output, which is not the output of the container.
- **Tell the state of the setup.** A query with no session. It says whether the setup is concluded and whether a setup code is waiting.
- **Open the setup.** It checks a setup code and answers what it opens and when it expires. It saves nothing.
- **Start an upload of the setup.** It checks the setup code and asks the `files` area to start an upload for the logo of the consultancy or for the photo of the administrator.
- **Conclude the setup.** It validates everything first, then confirms the uploads it was given, then, in one transaction: locks the setup code, checks that it is valid and that no consultancy exists, creates the administrator role and the developer role, the consultancy and the administrator with the administrator role, ends the setup code and writes to the audit log.
- **Register an administrator with a setup code.** The way back in. In one transaction it locks the setup code, checks that it is valid and that the setup is concluded, refuses an e-mail that belongs to a user, creates the user with the administrator role, ends the setup code and writes to the audit log.

### Domain rules

- The consultancy requires its name, legal name, country, tax ID, time zone and currency. A user requires a name, an e-mail, a country, a national ID and a preferred language.
- A country is an ISO 3166-1 two-letter code, a currency an ISO 4217 code and a time zone an IANA name. The backend checks each against a list it carries.
- A tax ID and a national ID are free text within a length limit, stored as typed. No country has a rule of its own.
- An e-mail is lowered before it is stored. A telephone is in international form, a plus sign and digits, checked for its form only and stored without spaces. A website is an `http` or `https` address.
- The preferred language is one of the supported languages.

### The `files` area

- An S3 adapter on the AWS SDK for Go v2, against Garage, following the Files rules: one private bucket, a random object key, a signed URL that expires in minutes, and a confirmation that checks the object.
- Two use cases in this effort: start an upload and confirm an upload. Both are called by the use cases of the setup, which check the setup code. The area has no REST operation of its own yet, and no download: nothing shows a stored image until the shell of the app, in the second effort.
- An image is a PNG, a JPEG or a WebP of at most 2 MB. Starting an upload refuses another declared type or a larger declared size. Confirming reads the size of the object and its first bytes, and removes an object that is over the limit or is not one of the three types. Nothing is resized or cropped.
- An image this small goes in one request, so no upload in parts is built.
- The browser reaches Garage at a host name of its own in the `Caddyfile`, and the bucket accepts cross-origin requests from the origin of the app only.

### The `audit` area

- One use case records an entry. An entry holds the time, the operation, the kind of actor with its identifier, and what was acted on. It carries identifiers only, never an e-mail or a name.
- Operations of this effort: a setup code issued, a setup concluded, an administrator registered with a setup code. The actor of the first is whoever installs the system, with no identifier. The actor of the other two is the setup code used, and the subject is the user created.
- An entry is written in the transaction of the operation it records, so an operation never succeeds without its record. This is the one effect on another area that does not wait for the commit, and `CODING_STANDARDS.md` gains that sentence with this code.
- No operation reads the audit log in this effort.

### REST contract

Every operation is under `/api`, follows the API conventions and is described in the committed OpenAPI spec. None of them needs a session.

| Operation | What it does | Answers |
| --- | --- | --- |
| `GET /api/setup` | Tells the state of the setup | 200 with whether it is concluded and whether a setup code is waiting |
| `POST /api/setup/open` | Opens the setup with a setup code | 200 with what the code opens, `first_setup` or `new_administrator`, and when it expires |
| `POST /api/setup/uploads` | Starts the upload of an image | 201 with the identifier of the file, the signed URL and when it expires |
| `POST /api/setup/conclude` | Concludes the first setup | 204 |
| `POST /api/setup/administrators` | Registers an administrator, the way back in | 201 |

- The setup code travels in the body of each `POST`, never in a path, a query parameter or a header.
- Stable error codes: `setup_code_invalid` (422) for a code that is wrong, expired or ended, with one answer for the three; `setup_already_concluded` (409) and `setup_not_concluded` (409) for a code used on the wrong operation; `email_in_use` (409); `too_many_requests` (429), answered by Caddy. A failed validation is a 422 that names the field.
- The identifier of a file leaves the backend with its prefix and its Sqids encoding (ADR-0010). It is the only identifier exposed in this effort.
- The plural-path lint lists `/api/setup` among the singular paths.
- The foundation of the REST adapter comes with the first operation: the error format on every failure, strict input, the size limit and the read timeout on bodies. The check of the origin arrives with the session, in the second effort.

### Process, health and telemetry

- Liveness, readiness and the shutdown are built as Health and Shutdown describe, with the shutdown test. The streaming part of that test waits for the first stream.
- Logs are JSON lines on the output, each with the trace identifier. A request opens a span and pgx is traced, so the `trace_id` of an error body is real. The exporter is wired and sends nothing without an address. Metrics and the wrapper around the metric exporter come with the health report, in the third effort.

### Frontend

- `apps/web` gets its first two areas: `consultancy`, with the setup screen, and `access`, with the provisional login screen that the second effort replaces.
- The client, the Zod schemas and the simulated responses are generated by orval from the committed spec. Only the data module of an area uses them. The setup has no collection: its one read goes through TanStack Query.
- Routes: every address leads to the setup until it is concluded. Afterwards the address of the setup leads to the login screen, except while a setup code is waiting, and every other address leads to the login screen too, because nothing else exists yet.
- The setup screen is one route whose state is a union of named steps in one reducer: the code, the consultancy, the administrator and the review. When the code opens the registration of a new administrator, the step of the consultancy is left out.
- A request of the setup that answers `setup_code_invalid` sends the screen back to the step of the code and keeps everything typed. A valid code returns the person to the step they were on.
- The form is TanStack Form on the generated Zod schemas. It checks only what the contract declares, and a 422 shows on the field it names. An error is translated by its code and shows its `trace_id`.
- An image is checked for type and size when chosen, sent to the storage right then, and shown from the browser's own copy. The review shows it the same way.
- The country, the currency and the time zone are searchable lists whose names the browser writes in the language of the screen. The time zone starts at the one of the browser, and the currency at the one of the chosen country.
- The screen is framed as the login screen of the removed prototype was: a title bar with the product name, a status bar that holds the language selector and nothing else, and one bordered pane in the centre of a page with the page framing of `DESIGN.md`. At a narrow width the pane takes the width of the page.
- The language selector changes the language of the screen at once. The language it shows when the setup is concluded is sent as the preferred language of the administrator.
- Concluding leads to the login screen with a notice that the setup is concluded.
- Every text is in the three supported languages, in the translation files of its area. English is the source.
- A component the screens need and `apps/web/src/ui` lacks enters through the shadcn CLI, and `DESIGN.md` gains it in the same change.

### Development environment and delivery

- `compose.yaml` at the root is development. It extends a base file in `deploy/` that holds the tag of every image. It runs PostgreSQL, Garage and Caddy, and a step that creates the bucket, its key and its cross-origin rule. Log files rotate by size.
- The API and Vite run on the machine. The browser opens the app through Caddy on `localhost`, which forwards `/api` to the API, the rest to Vite, and the host name of the files to Garage.
- The Caddy image is built in this repository with a rate-limit plugin, chosen in this effort. Caddy limits, for each IP address, the `POST` operations of the setup, and answers 429 in the error format of the API. The starting number is ten requests a minute; the plugin and the number are recorded in `CODING_STANDARDS.md`.
- Caddy sets `Cache-Control: no-store` under `/api` and answers `/up` with the status code of readiness.
- Compression, the security headers, the Brotli copies and the cache headers of the built app wait for the deploy.
- Task targets: starting the services, starting the API and the app, applying the migrations, issuing a setup code, generating the spec and the code, and the checks. `task check` runs everything a pull request runs.
- `VERSION` is created at 0.0.0 (ADR-0023). How the builds read it is decided with the health report.

### Pipeline

- `ci.yml` gains the jobs of the backend, of the app and of the end-to-end test, each added to the `needs` of the `Checks` job. Each calls Task targets. A pull request that changes only documents skips them.
- Checks that guard the code from the first line: the Go linter and formatter (golangci-lint with gofumpt), the rules on imports between layers and between areas, on `else` and on nesting, generated code up to date, the lint of the spec (vacuum), the breaking-change report (oasdiff, which does not fail), exact versions, Knip, and the import rule of the frontend. Where the linter has no rule, a small analyzer in the repository checks it.

### Documents changed with the code

- `CODING_STANDARDS.md`: the three conventions, the tools behind Enforcement, the rate-limit plugin with its number, and the sentence on the audit log and the transaction.
- `README.md` and `README.pt-BR.md`: the prerequisites, the configuration and how to run the system in development, including how to issue a setup code.
- `DESIGN.md` for every component added, and `docs/machine-tools.yaml` for every tool installed on the maintainer's machine.

## Testing Decisions

A good test exercises behaviour that someone outside the module can see, and survives a rewrite of the inside. It asserts on what a use case returns and saves, on what an operation answers, or on what a screen shows. It does not assert on which function was called.

- **Use cases, with in-memory implementations of their interfaces.** The default seam, and where every business rule is tested: issuing a code and ending the earlier ones; opening with a code that is valid, wrong, expired or ended; reading a code typed in each accepted form; concluding with valid data, with each required field missing, with a consultancy already there, and twice with the same code; the two roles and who holds them; the way back in, its refusal of an e-mail in use and of an install not yet set up; each audit entry; the rules of an image when an upload starts and when it is confirmed. A fixed clock drives the expiry. Assertions use testify.
- **The `postgres` adapters, against the development PostgreSQL.** Each test gets a temporary database cloned by pgtestdb from a template the migrations built. They cover what only the database shows: the lock on the setup code under two concluding transactions, the single consultancy, the unique e-mail, and the audit log table refusing an update and a delete by the API role.
- **The `s3` adapter, against Garage.** A signed URL for the public host name accepts an upload from another origin, the confirmation reads the size and the first bytes, and an object over the limit is removed.
- **The REST adapter, through HTTP.** Only what the adapter alone does: the error format on a validation error, an unknown route, a wrong method and a panic; strict input; the size limit of a body; the prefix of an identifier; and that no setup code appears in a path, a query parameter or a header of the spec.
- **`platform`.** The shutdown test, the configuration that fails at the start and names the variable, and the test that proves the logs carry no setup code and no personal data.
- **The setup screen, whole, in Vitest's Browser Mode with MSW.** A test acts as a person does. It covers the first setup from the code to the review, the way back in without the step of the consultancy, an error on its field, a code that expires halfway with the form kept, an image refused when chosen, the limit reached, the change of language, and each route before and after the setup is concluded.
- **One end-to-end test in Playwright, through Caddy.** It issues a setup code with the command, concludes the setup with both images and lands on the login screen with its notice. It then issues a second code and registers a new administrator. A separate check sends requests past the limit and reads the 429 in the error format.

Prior art is thin, because this is the first backend code and the first screen. The one existing test is the missing-key check of the translation files in `apps/web`, which shows the style: plain data in, plain assertions out. The setup screen sets the pattern for every later screen test, and the use case tests of the setup set it for the backend.

## Out of Scope

- The login with a login code, the local login provider, the session, River, the worker command, the e-mails, Mailpit, the seed and the shell of the app: the second effort. That includes the periodic job that removes uploads never confirmed, the download of a file, and the check of the origin of a request.
- The consultancy settings, the profile, the users, the scopes and their checks, the health report with its screen and the River panel: the third effort.
- The three real login providers, the screen that creates roles and chooses their scopes, access tokens, MCP, LLM providers and the telemetry containers.
- Staging, production, the publish workflow, the secrets, the backup and the first release.
- `task outdated` with its weekly workflow, the limit on the size of the JavaScript, the report of load times, the Brotli copies and the closed Content-Security-Policy, which `CODING_STANDARDS.md` now ties to the deploy.
- A rule for the tax ID or the national ID of any country, a search of an address by postal code, and the cropping or resizing of an image.

## Further Notes

- Until the cleanup job exists, an image sent in a setup that was abandoned stays in the bucket. That is accepted for development.
- The login screen of the second effort is built from the flow, the scenarios and the texts of the removed prototype, which remain in the history of the repository (`.scratch/login-screen-prototype/spec.md` at commit `cc48eb1`).
- Docker reaches the API and Vite on the machine through the host gateway. The address differs between Docker Desktop and the runner of the pipeline, and the Compose file has to serve both.

### Issues

In `issues/`, in dependency order. Each one says which issues block it.

1. `01-backend-skeleton.md`
2. `02-backend-guards.md`
3. `03-shutdown-logs-and-trace.md`
4. `04-issue-a-setup-code.md`
5. `05-rest-contract-and-opening-the-setup.md`
6. `06-conclude-the-setup-in-the-backend.md`
7. `07-development-behind-caddy.md`
8. `08-setup-gate-and-code-step.md`
9. `09-setup-form.md`
10. `10-setup-images-in-the-backend.md`
11. `11-setup-images-on-the-screen.md`
12. `12-the-way-back-in.md`
13. `13-end-to-end-test.md`

### Joints this effort proves

Each was recorded in an ADR as read in documentation and not tried. A joint that does not hold changes the decision, in its ADR, before the work goes on.

Checked when the effort closed (issue 13): ten joints are proven, and one did not hold. Where a joint needed something its ADR did not expect, the ADR says what.

- The vacuum lint on the spec Huma generates: `snake_case`, the rule for plural paths with its list of singular ones, and the `$schema` property Huma adds to bodies (ADR-0009). **Proven** (issue 05). Plural paths needed a function of vacuum written in this repository, and the API turns the `$schema` links off; the lint accepts them all the same. ADR-0009 says both.
- The error format on unknown routes, wrong methods and panics, which bypass Huma's error hook (ADR-0009). **Proven** (issue 05), as ADR-0009 wrote it.
- pgtestdb with the version of goose in use (ADR-0015). **Proven** (issue 04), with goose 3.30.1. The test helper builds the template with the function the `migrate` command calls, not with the migrator of pgtestdb, as ADR-0015 says.
- The migration role creating `pg_stat_statements` without being a superuser (ADR-0015). **Did not hold** (issue 01): PostgreSQL does not mark the extension as trusted, so only a superuser creates it. The superuser creates it with the roles, on the first start of the container, and no migration does. ADR-0015 says so.
- The AWS SDK for Go v2 against Garage: path-style addresses, checksum settings, and a URL signed for the public host name (ADR-0017). **Proven** (issue 10), with what ADR-0017 now lists: path-style addresses, the checksum defaults of the SDK, and a second client that signs for the public host name.
- orval on the spec Huma generates: the client, the Zod schemas and the simulated responses (ADR-0013). **Proven** (issue 08), with the configuration ADR-0013 records.
- The generated Zod schemas in a TanStack Form (ADR-0013). **Proven** (issues 08, 09 and 12), with the adjustments ADR-0013 records.
- A screen test in Vitest's Browser Mode with MSW serving the generated responses (ADR-0014). **Proven** (issue 08), with what ADR-0014 now records.
- The Caddy image built with a rate-limit plugin, and its 429 in the error format of the API (ADR-0011). **Proven** (issue 07), with `mholt/caddy-ratelimit`, which ADR-0011 names. The end-to-end test of issue 13 reads the 429 through the Caddy of development.
- Compose files that extend one base, with the tag of every image in the base only (ADR-0019). **Proven** (issues 01, 07 and 10), with what ADR-0019 now records: the volumes and configs are declared in the file that extends, and the image built in the repository keeps its versions in its Dockerfile.
- `/up` through Caddy, answering with the status code of readiness and no body (ADR-0012). **Proven** (issue 07), as ADR-0012 wrote it. The end-to-end test waits on it for the API.

### Joints still not tried

This list replaces the walking skeleton. Each joint is proven by the effort named beside it.

- **Second effort.** sqlc reading the SQL that River exports, goose applying it, and the permissions the API role needs on River's tables. Whether River's OpenTelemetry middleware links the trace that saved a job to the one that ran it. The login with a login code read from Mailpit, and through Dex. The seed creating its data twice with the same result.
- **Third effort.** A TanStack DB collection loaded by following the cursor of the API to the end. The check of the job worker and of the e-mail server for the health report.
- **The deploy.** Backblaze B2 with the AWS SDK for Go v2, and pgBackRest against it. The Content-Security-Policy with Base UI and the fonts, and a Vite plugin that writes Brotli copies under Vite+. The server created from the cloud-init file, the life of a staging server, its reserved address and its certificates. The publish and its check of the variables. The restore test. A pull request from a fork. `task outdated` and its weekly workflow. The first release and the ruleset that protects its tag. What ADR-0012 lists about Grafana and the collector waits for the telemetry containers.
- **Their own efforts.** The MCP handler on the same router as Huma (ADR-0008). The LLM stand-in and a call to a provider that spends no tokens (ADR-0004, ADR-0019). How each login provider states that an e-mail is verified (ADR-0011).
