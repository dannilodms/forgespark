---
status: accepted
---

# Login through external providers, with a server-side session

ADR-0008 left open how a person logs in to the React app. The app and the API share one origin behind Caddy, so the simplest safe arrangement applies: the backend keeps the session and the browser holds only an opaque cookie that JavaScript cannot read.

- **Providers.** A person logs in with Google, Microsoft or GitHub. The backend runs the OAuth 2.0 authorization code flow (`golang.org/x/oauth2`) and validates the ID token (`coreos/go-oidc/v3`). GitHub issues no ID token, so its verified e-mail is fetched from its API. The system stores no password.
- **Session.** `alexedwards/scs/v2` with `pgxstore` keeps session data in PostgreSQL. The cookie is opaque, `HttpOnly`, `Secure` and `SameSite=Lax`. The callback renews the session token and stores the user's identifier. Logging out destroys the session.
- **Who is acting.** The REST adapter reads the user from the session on every request and passes the actor to the use case, the same way an access token does for an agent.
- **Only registered users get in.** A login succeeds only when the e-mail the provider reports as verified belongs to an existing user. An unknown login is refused and creates nothing. The first user's e-mail comes from the deploy configuration; other users are registered by an existing user.
- **Recognising the same person.** The first login through a provider is matched by verified e-mail. From then on the system stores the stable identifier that provider gives the account and uses it. A user can have several providers linked.
- **Cross-site requests.** With a cookie carrying the session, a request that changes state must come from the app's own origin. The backend refuses the others.

## Considered Options

- **Authorization Code with PKCE and tokens in the browser.** A short-lived access token in memory and a refresh token in a cookie. It suits a frontend on another domain or an external identity provider issuing API tokens; here it is more parts for the same result, and revoking means waiting for a token to expire.
- **Passwords of our own.** No dependency on outside providers, but it brings password storage, reset by e-mail and protection against guessing.
- **Matching by verified e-mail on every login.** No table of links, but an e-mail can change hands, and whoever gets an account with that e-mail at any provider logs in as the person.
- **A released e-mail domain or open sign-up.** Less registration work, but a wider door, and open sign-up needs an approval screen.
- **One provider per user.** No ambiguity, but no way in when that provider is down or the account is lost.

## Consequences

- Each provider needs an application registered with it, and its client identifier and secret in the deploy configuration.
- A person locked out of one provider logs in through another with the same verified e-mail.
- How each provider states that an e-mail is verified has to be checked when the login is built. Microsoft is the known hard case: its e-mail claim is not always verified.
- How long a session lasts, idle and in total, is not decided. It is settled with the modelling of users.
- Agents are unaffected: they keep using access tokens (ADR-0008).
