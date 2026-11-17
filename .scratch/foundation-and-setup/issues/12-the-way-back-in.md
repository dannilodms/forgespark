# 12 - The way back in: a new administrator registered with a setup code

Status: ready-for-agent
Blocked by: 11

Read `../spec.md`, ADR-0024 and, in `CODING_STANDARDS.md`, Setup and Users and roles.

## What to build

When no administrator can log in, whoever installs the system issues a setup code, opens the address of the setup and registers a new administrator. Nothing else changes.

- `issue-setup-code` says what the code opens from the real state: the first setup while no consultancy exists, and the registration of a new administrator afterwards. In the second case it also says that the users that exist are not changed.
- Opening the setup with such a code answers that it opens the registration of a new administrator.
- The use case that registers an administrator with a setup code. In one transaction it locks the setup code, checks that it is valid and that the setup is concluded, refuses an e-mail that belongs to a user, creates the user with the administrator role, ends the setup code and writes to the audit log. The actor of the entry is the setup code used, and its subject is the user created.
- `POST /api/setup/administrators`, with the setup code and the data of the administrator in the body, answering 201. An e-mail in use answers 409 with `email_in_use`; an install whose setup is not concluded answers 409 with `setup_not_concluded`. The photo is accepted as in the first setup. It falls under the limit of ticket 07.
- Concluding the first setup with a code issued after the setup was concluded still answers `setup_already_concluded`.
- On the screen, a code that opens the registration of a new administrator leaves out the step of the consultancy: the code, the administrator and the review. The review says that a new administrator will be registered and that the users that exist are not changed.
- An e-mail in use shows on its field, on the step of the administrator.
- Concluding leads to the login screen with a notice that the administrator was registered.
- The address of the setup shows the step of the code only while a setup code is waiting, and leads to the login screen once the code is used or expires.
- Every new text in the three supported languages.

## Acceptance criteria

- [ ] On an install whose setup is concluded, the command says the code opens the registration of a new administrator, and the address of the setup shows the step of the code.
- [ ] A person registers a new administrator from the code to the login screen, without seeing the data of the consultancy.
- [ ] The new user holds the administrator role, with the preferred language the screen was in.
- [ ] The consultancy, the users that existed and their roles are unchanged afterwards.
- [ ] The e-mail of an existing user is refused on its field, whether that user is deactivated or not, and nothing is saved.
- [ ] The setup code is ended afterwards, and the address of the setup leads to the login screen again.
- [ ] Registering an administrator with a setup code leaves one entry in the audit log.
- [ ] On an install whose setup is not concluded, the operation answers `setup_not_concluded`.
- [ ] The business rules are tested through the use case, and screen tests cover the shorter path and the e-mail in use.
- [ ] The spec, the generated code, the generated client and the lint are up to date.
