# 04 - Issue a setup code

Status: ready-for-agent
Blocked by: 01

Read `../spec.md`, ADR-0024 and, in `CODING_STANDARDS.md`, Setup, Use cases, Domain modelling, Database and Tests. The terms setup code and audit log are in `GLOSSARY.md`.

## What to build

Whoever installs the system runs `issue-setup-code` in the API container and gets a setup code on the screen. This is the first use case of the repository, so it also sets how a use case, its repository and its tests are written.

- The `consultancy` area, with the setup code as its first aggregate, and the use case that issues one. It ends every earlier code, stores the hash of the new one and returns the code, its expiry and what it opens. No REST operation exists for it (ADR-0008).
- The code as "The setup code" in `../spec.md` describes it: twenty characters of Crockford's Base32 in four groups of five, from the cryptographic random source, lasting one hour, stored only as the SHA-256 hash of its normalised form.
- What the code opens is the first setup for now. Ticket 12 makes it depend on whether the setup is concluded.
- The `issue-setup-code` command of the binary. It prints the code, when it stops working, what it opens and the address of the setup, taken from configuration, to its own output. The code reaches no log and no span.
- A Task target that runs the command in development.
- The `audit` area, with the use case that records an entry and its table, on which the API role can only insert and select. Issuing a setup code records an entry whose actor is whoever installs the system, with no identifier. The entry is written in the transaction of the operation it records: add that sentence to `CODING_STANDARDS.md`.
- The transaction opened and closed through an interface, and the repository that locks the row it loads, as Use cases and Domain modelling describe.
- sqlc for the queries of both areas, with its generated code committed and a check that fails when it is stale, in `task check` and in the pipeline.
- A use case takes the current time from a clock it receives when it is built, and the tests assert with testify. These are the other two conventions of the first backend code: record both in `CODING_STANDARDS.md`.
- `postgres` adapter tests on a temporary database cloned by pgtestdb from a template the migrations built.
- `README.md` and `README.pt-BR.md` say how to issue a setup code.

## Acceptance criteria

- [ ] The command prints a code in four groups of five characters, with none of the letters I, L, O and U, together with its expiry, what it opens and the address of the setup.
- [ ] Two codes issued in a row differ, and the first one is ended.
- [ ] The table of setup codes holds a hash and never the code.
- [ ] The code appears in no log line. The test of sensitive values of ticket 03 guards it.
- [ ] Each code issued leaves one entry in the audit log, and a failure to write the entry leaves no code stored.
- [ ] The API role cannot update or delete a row of the audit log. A test proves it against the database.
- [ ] The command fails with a clear message, and a status other than zero, when the database is unreachable or a migration is pending.
- [ ] The use case is tested with in-memory implementations of its interfaces and a fixed clock. The adapters are tested against PostgreSQL, each test on its own temporary database.
- [ ] pgtestdb works with the version of goose in use. If it does not, ADR-0015 is updated before the work goes on.
- [ ] Stale sqlc code fails the check.
- [ ] `CODING_STANDARDS.md` records the clock, the assertion library and the sentence on the audit log and the transaction.
