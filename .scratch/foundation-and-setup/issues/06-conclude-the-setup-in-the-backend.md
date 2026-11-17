# 06 - Conclude the setup in the backend

Status: ready-for-agent
Blocked by: 05

Read `../spec.md`, ADR-0024, ADR-0026 and, in `CODING_STANDARDS.md`, Setup, Users and roles, Domain modelling and API conventions. The terms consultancy, user, role, administrator, developer, tax ID and national ID are in `GLOSSARY.md`.

## What to build

A caller with a valid setup code registers the consultancy and its first administrator in one operation, and anyone can ask whether the setup is concluded. Images come in ticket 10.

- The consultancy, the user and the role as aggregates of the `consultancy` area, with the tables "Database" in `../spec.md` describes. The database refuses a second consultancy and a second user with the same e-mail.
- The domain rules of `../spec.md`: the required fields of each; the country, the currency and the time zone checked against lists the backend carries; the tax ID and the national ID as free text within a length limit; the e-mail lowered; the telephone in international form; the website as an `http` or `https` address; the preferred language as a supported language. The value objects that more than one area will use go to `kernel`.
- The use case that concludes the setup. It validates everything first. Then, in one transaction, it locks the setup code, checks that it is valid and that no consultancy exists, creates the administrator role and the developer role, the consultancy and the administrator with the administrator role, ends the setup code and writes to the audit log. The actor of the entry is the setup code used, and its subject is the user created.
- `POST /api/setup/conclude`, with the setup code and the data in the body, answering 204. A code that is not valid answers `setup_code_invalid`; a consultancy that already exists answers 409 with `setup_already_concluded`; a failed validation answers 422 and names the field.
- The query that tells the state of the setup, and `GET /api/setup`: whether the setup is concluded and whether a setup code is waiting. It needs no session.
- The use case that opens the setup now answers from the real state.

No scope exists yet, so no table links a role to its scopes. The administrator role stores none.

## Acceptance criteria

- [ ] With valid data and a valid code, the consultancy, the administrator, the two roles and the audit entry exist afterwards, the administrator holds the administrator role, and the setup code is ended.
- [ ] The developer role exists and no user holds it.
- [ ] `GET /api/setup` says not concluded before and concluded after, and says a setup code is waiting only while one is issued and still valid.
- [ ] Each required field missing, a country, a currency, a time zone or a language outside its list, a malformed e-mail, telephone or website, and an identifier over the length limit each answer 422 naming the field, and save nothing.
- [ ] An e-mail typed in upper case is stored in lowercase.
- [ ] Concluding twice with the same code saves once: the second attempt answers `setup_code_invalid`.
- [ ] Two concluding requests at the same moment save one consultancy. A test against PostgreSQL proves it.
- [ ] A failure at any point of the transaction leaves no consultancy, no user, no role and a setup code that still works.
- [ ] No e-mail, name or identifier of a person appears in a log line, a span or the audit entry. The test of sensitive values guards them.
- [ ] The business rules are tested through the use case with in-memory implementations, and the spec, the generated code and the lint are up to date.
