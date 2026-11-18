# 04: Checking the code: the session, who is acting, logging out and the origin

**What to build:** The six digits typed in the browser that asked for them log the user in. From then on every request under the contract knows who is acting, with the scopes of their roles; the user can ask who they are and log out; a deactivated user is refused and their session ends; a request that changes state from another origin is refused; and the login is written to the audit log. Rules: Login and Users and roles in `CODING_STANDARDS.md`, ADR-0011, ADR-0012, ADR-0026. Verifiable with `curl`: code, cookie, who am I, log out.

**Blocked by:** 03 (Asking for a login code).

**Status:** ready-for-agent

- [ ] A `POST` takes the six digits and reads the request's identifier from the session. No request in this browser, a request ended or expired, or a wrong code answer the same rule, "the code is not valid", with the attempts left; the comparison is in constant time.
- [ ] The third wrong attempt ends the request and answers that the code ended; a right code ends the request, renews the session token and stores the user's identifier.
- [ ] The REST adapter reads the session on every operation that needs a user, loads the user and their roles, and passes an actor with the user's identifier and the scopes of their roles: every scope for the administrator role, none for the developer role. A missing session answers 401; a deactivated or missing user gets 401 and their session destroyed.
- [ ] An operation answers the logged-in user: name, e-mail, preferred language and the keys of the roles (the photo URL arrives in 06).
- [ ] A named `POST` operation on the session logs out and destroys the session.
- [ ] A `POST` or `PUT` with the session cookie is refused with 403 in the error format when `Origin` is present and differs from the app origin, or when `Origin` is absent and `Sec-Fetch-Site` is `cross-site`; a request with no cookie is not checked.
- [ ] The audit log gains the actor kind `user` and the operation `logged_in`, with a new column for the way in (`login_code` here, the provider key later); the CHECK constraints grow with them, and the `postgres` test shows the table still refuses updates and deletes.
- [ ] Use case tests cover a right code, a wrong one three times, an expired one, one from another browser, the audit entry, who is logged in and the logout; a fixed clock drives expiry.
- [ ] HTTP tests by the binary cover the cookie and its attributes, 401 without a session and after deactivation, the 403 on a cross-origin `POST`, and that no code, e-mail or token appears in any path, query parameter or header of the spec.
- [ ] The test of the logs covers the session cookie.
