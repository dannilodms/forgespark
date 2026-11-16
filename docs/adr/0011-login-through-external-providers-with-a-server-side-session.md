---
status: accepted
---

# Login through external providers or a login code, with a server-side session

ADR-0008 left open how a person logs in to the React app. The app and the API share one origin behind Caddy, so the simplest safe arrangement applies: the backend keeps the session and the browser holds only an opaque cookie that JavaScript cannot read.

A person logs in through an external provider or with a login code sent to their e-mail. Both ways end in the same session.

- **Providers.** A person logs in with Google, Microsoft or GitHub. The backend runs the OAuth 2.0 authorization code flow (`golang.org/x/oauth2`) and validates the ID token (`coreos/go-oidc/v3`). GitHub issues no ID token, so its verified e-mail is fetched from its API. The system stores no password.
- **Login code.** A person types their e-mail and receives a login code: six digits, valid for 10 minutes, usable once. The third wrong attempt ends the code, and a new one has to be asked for. The code works only in the browser that asked for it.
- **Nothing of the code is stored.** The code is computed with HMAC from a key in the deploy configuration and a random value kept with the request for the code. The job that sends the e-mail (ADR-0016) and the operation that checks the code compute it again. A dump of the database reveals no code.
- **An unknown e-mail gets the same answer.** The screen says the same thing and nothing is sent, so the login does not tell who is a user. Sending is a job, so the response takes the same time either way.
- **Limits on asking for a code.** For each e-mail address, one code a minute and five an hour, counted in PostgreSQL by the use case; a new code ends the previous one. For each IP address, a limit in Caddy on the two operations of the login code, asking and checking.
- **Session.** `alexedwards/scs/v2` with `pgxstore` keeps session data in PostgreSQL. The cookie is opaque, `HttpOnly`, `Secure` and `SameSite=Lax`. The callback renews the session token and stores the user's identifier. Logging out destroys the session.
- **Who is acting.** The REST adapter reads the user from the session on every request and passes the actor to the use case, the same way an access token does for an agent.
- **Only registered users get in.** A login succeeds only when the e-mail belongs to an existing user: the one a provider reports as verified, or the one a login code was sent to. An unknown login is refused and creates nothing. The first user's e-mail comes from the deploy configuration; other users are registered by an existing user.
- **Recognising the same person.** The first login through a provider is matched by verified e-mail. From then on the system stores the stable identifier that provider gives the account and uses it. A user can have several providers linked.
- **Cross-site requests.** With a cookie carrying the session, a request that changes state must come from the app's own origin. The backend refuses the others.
- **A local provider for development and tests.** An agent or an automated test cannot get past Google's login screen. A fourth provider, generic OIDC, is turned on by configuration alone and pointed at an identity container that runs beside the app, so the same OAuth code as in production is exercised end to end. The identity container is Dex, with a fixed list of users in its configuration file. The production and staging Compose files do not declare it. When it is on, it shows as a button on the login screen, in a log line at the start of the process and in the health report. End-to-end tests also log in with a login code, read from Mailpit.

## Considered Options

- **Authorization Code with PKCE and tokens in the browser.** A short-lived access token in memory and a refresh token in a cookie. It suits a frontend on another domain or an external identity provider issuing API tokens; here it is more parts for the same result, and revoking means waiting for a token to expire.
- **Passwords of our own.** No dependency on outside providers, but it brings password storage, reset by e-mail and protection against guessing.
- **Matching by verified e-mail on every login.** No table of links, but an e-mail can change hands, and whoever gets an account with that e-mail at any provider logs in as the person.
- **A released e-mail domain or open sign-up.** Less registration work, but a wider door, and open sign-up needs an approval screen.
- **One provider per user.** No ambiguity, but no way in when that provider is down or the account is lost.
- **Providers only, with no login code.** The first version of this decision. Nothing could be guessed, so no limit was needed, but a person without an account at one of the three providers had no way in.
- **Telling the person that the e-mail is not registered.** Clearer for someone who mistyped the address, but it lets anyone find out who has an account.
- **The code stored as a hash.** The worker that sends the e-mail could not write the code, because job arguments carry identifiers only (ADR-0016), and the hash of six digits is broken in moments from a dump.
- **The code stored encrypted until the e-mail leaves.** It works, with two states for the same data.
- **Sending the code inside the request.** No job, but the response time would tell whether the e-mail belongs to a user.
- **A limit per e-mail address only.** No new piece, but nothing would slow down one caller asking for codes for many addresses.
- **A login operation that exists in development only, or a command that writes a session.** Simpler than a provider, but the first is a way in that has to be proven off in production, and neither exercises the real login flow.
- **The login code as the only way in for development and tests.** One container less, but the OAuth flow would be covered by adapter tests alone.
- **The binary refusing to start with the local provider in production, or a build tag.** The first depends on another variable being right. The second makes the tested image differ from the published one.

## Consequences

- Each provider needs an application registered with it, and its client identifier and secret in the deploy configuration.
- A person locked out of one provider logs in through another with the same verified e-mail, or with a login code.
- Whoever reads a user's mailbox can log in as that user.
- The login code depends on the e-mail server, an optional part of the system health. With it down, only the providers work.
- Caddy has no rate limit of its own, so its image is built in this repository with a plugin. The plugin and the numbers of the limit are chosen in the walking skeleton.
- The key of the login code is one more secret in the deploy configuration. Changing it only ends the codes not yet used.
- Rate limiting, which ADR-0009 had left out, exists for the login code.
- Staging does not use the local provider. It is public, so it logs in through the real providers, or with a login code read from a Mailpit that sits behind a password (ADR-0019).
- How each provider states that an e-mail is verified has to be checked when the login is built. Microsoft is the known hard case: its e-mail claim is not always verified.
- How long a session lasts, idle and in total, is not decided. It is settled with the modelling of users.
- Agents are unaffected: they keep using access tokens (ADR-0008).
