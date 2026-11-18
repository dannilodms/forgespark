# 07: Dex and the login through a provider in the API

**What to build:** A person navigates to the operation that starts a login through the local login provider, authenticates at Dex, and comes back logged in with the same session the login code gives. The first login is matched by verified e-mail and leaves a provider link; later logins use the link. A verified e-mail of no active user comes back refused, with the e-mail kept in the session for the screen to show. A read that needs no session tells which providers are on and the last refusal. Rules: Login in `CODING_STANDARDS.md`, ADR-0011. Verifiable by navigating the browser to the start URL and reading who am I afterwards.

**Blocked by:** 04 (Checking the code: the session, who is acting, logging out and the origin).

**Status:** ready-for-agent

- [ ] Dex is a service of the development Compose file only, published on a port of the machine so the browser and the API binary reach the same issuer URL; its configuration file has one static client and two people with fixed passwords, the administrator and the developer of the seed at the default mailbox (`<local part>+admin@` and `<local part>+developer@` of the fictitious default of 09).
- [ ] A login provider is a set of configuration variables: issuer URL, client identifier, client secret, display name and key; the local one is on when they are set and off when empty. When on, a log line at the start of the API names it.
- [ ] The `oidc` adapter does discovery, the authorization code flow with PKCE, `state` and `nonce`, and validates the ID token with `coreos/go-oidc/v3`, taking the e-mail only when the token says it is verified; its test runs against the Dex container, driving Dex's login page over HTTP.
- [ ] The start operation is a `GET` with the provider key in the path and the return path as a query parameter; it stores state, nonce, verifier and return path in the session and redirects to the provider.
- [ ] The callback is a `GET` with the key in the path: it checks the state, exchanges the code, validates the token and the nonce, finds the provider link by provider and subject or else the active user with that verified e-mail and creates the link, logs the user in, writes `logged_in` with the provider key as the way in, and redirects to the return path when it is a path of the app, or to the root.
- [ ] A verified e-mail of no active user (unknown or deactivated) stores the refusal with the e-mail in the session and redirects to the login with `notice=no_active_user`; an unverified e-mail, a cancel or an error at the provider redirects with `notice=login_not_concluded`; a wrong state logs nobody in.
- [ ] A provider link is a row of provider key, subject and user, unique on provider and subject; the `postgres` test proves the uniqueness.
- [ ] A read that needs no session answers the login state: the providers turned on with key and display name, and the e-mail of the last refusal in this browser when there is one.
- [ ] Use case tests cover the match by link, by verified e-mail, the link created once, the refusals, and the audit entry; HTTP tests by the binary cover the redirects and that no token or e-mail appears in a path, query parameter or header of the spec.
- [ ] Both operations are in the contract as operations that answer a redirect, and the lint of the spec passes.
