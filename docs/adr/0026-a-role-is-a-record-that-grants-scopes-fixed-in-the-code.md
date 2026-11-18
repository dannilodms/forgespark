---
status: accepted
---

# A role is a record that grants scopes fixed in the code

ADR-0008 gave access tokens scopes, and the use case checks the scopes of whoever is acting. ADR-0024 created an administrator and left open what a user who is not one may do. The maintainer wants the consultancy to make its own roles and to say what each may do, without a deploy. The terms user, role, administrator, developer and scope are in `GLOSSARY.md`.

A role is a record with a name and a set of scopes. The list of scopes is fixed in the code, and a person chooses which of them a role grants.

- **One vocabulary.** What a role grants is the scope of ADR-0008: read or write in one area of the system. There is no second list of permissions. A use case checks a scope and never asks which role the actor has.
- **A session acts with the scopes of its user's roles.** They are read on every request, so a change of role shows on the next one. An access token acts with the scopes chosen when it was created, and never with more than the roles of its user grant.
- **A scope arrives with the code that checks it.** A person never defines a scope.
- **Several roles for a user.** A user has at least one role and may do what any of them allows.
- **The administrator role is in the code.** It always grants every scope, one added later included, and it cannot be changed or removed. The setup creates it (ADR-0024). The system refuses to take it from the last administrator who can log in, and refuses to deactivate that user.
- **The developer role is an ordinary role.** The setup creates it with no scope. It gets scopes as the areas of work are built, and it can be changed or removed like a role a person creates.
- **Some things ask for no scope.** Any logged-in user reads the name and the logo of the consultancy and edits their own profile.
- **One thing asks for the role itself.** The River panel opens to administrators only (ADR-0016), because it can run a job again or cancel it.

## Considered Options

- **A fixed set of roles in the code.** Each role and what it may do would be covered by tests, and no screen would be needed. A consultancy that divides its work another way would wait for a deploy.
- **A yes-or-no field that marks an administrator.** The smallest model. It has no way of giving a user part of what an administrator can do.
- **Permissions as a second list, finer than scopes.** A role could allow registering a user and not editing one. The use case would check two things, and what a user may do on a screen could differ from what their access token may do.
- **Scopes defined by a person.** No code could check a scope it does not know. ADR-0022 refused statuses and transitions configured by a person for the same reason.
- **One role for each user.** Nothing to add up, but a person who does two jobs would need a third role that joins them.
- **The developer role in the code as well.** Its scopes would change only with a deploy.

## Consequences

- At the start the scopes are read and write of the data of the consultancy, read and write of users, and read of the system health. Only the administrator role grants them.
- The write scope of users includes giving and taking roles. Whoever holds it can give the administrator role to anyone, themselves included, so granting that scope to a role is granting everything.
- The name of a role a person creates is configuration content, with a version for each supported language (ADR-0005). The screen that creates roles and chooses their scopes waits for the decision on how configuration content is stored. Until then the only roles are the two of the setup, and the frontend translates their names.
- Statuses and transitions stay in the code (ADR-0022). A role is different: what a person configures is which fixed scopes go together, and each scope is still checked by code that a test covers.
- A user who loses a role loses what it allowed in their access tokens as well, with no token revoked.
- Changing the roles of a user, and deactivating or reactivating one, write to the audit log (ADR-0012).
- "Permission" stays out of `GLOSSARY.md`.
