---
status: accepted
---

# A record with a lifecycle is a state machine in the domain

Several records move through named stages: a project and a request from first contact to maintenance, an e-mail from queued to sent, an upload from asked for to available, a document while it is processed. ADR-0007 says that state changes through a method named after the business operation, and ADR-0009 that the server tells the client which actions are allowed. With those two rules alone, each method checks the current status by itself, and the allowed actions are written a second time in the query. The two drift apart.

An aggregate that has a lifecycle therefore has a status, and one transition table that decides what can happen in each status.

- **The statuses are a fixed set in code.** A value object in `domain` lists them. The column is `text` with a `CHECK` constraint, as Database in `CODING_STANDARDS.md` already says for a fixed set of values.
- **One transition table for each aggregate**, in `domain`. For each operation it lists the statuses the operation can start from and the status it leads to. The method named after the operation asks the table. An operation the table does not allow is a typed error with a stable code.
- **Allowed actions come from the same table.** The query use case passes the status it read to a function in `domain` and gets the operations that status allows. It then keeps the ones the actor may perform. Neither the screen nor an agent derives them.
- **Work in progress is a status.** The use case that starts the work moves the record to a status that says so and saves the job in the same transaction (ADR-0016). The job's use case moves it on, or to failed when the attempts run out. Asking for a new attempt is a transition out of failed.
- **A record processed in steps has one status for each step.** Each step is a job that moves the record to the next one, so a failure is retried from the step that broke. The document processing pipeline keeps its state this way.
- **The name is configuration content; the set is not.** The name shown for each status keeps one version per supported language (ADR-0005). Adding or removing a status, or changing a transition, is a code change.
- **Plain Go.** A type, a table and a function. No library and no engine.
- **Not everything is a state machine.** A record with two conditions keeps a field: an access token that is revoked or not, a version that is an automatic draft or not. A condition that follows from time and counters, such as a login code that expired, is computed and never stored. System health is computed when it is asked for.
- **On a screen**, a flow with several steps is one reducer whose state is a union of named steps. An upload in parts and the login code are the first two.

## Considered Options

- **A check of the status written in each method.** Less to write for an aggregate with two operations. The allowed actions are then a second copy of the same rules, and nothing fails when the two disagree.
- **Statuses and transitions configured by a person.** A consultancy that changes its process would need no deploy. The rules would live in data: no test covers them, a stable error code cannot name one, an MCP tool cannot describe what it does, and it needs screens of its own.
- **A state machine library (looplab/fsm, qmuntal/stateless).** Events as strings and callbacks inside the aggregate, for a table of a dozen lines.
- **A workflow engine.** Rejected in ADR-0016: a server of its own with a database of its own.
- **Event sourcing.** The history of transitions comes for free, with everything else it brings. Deferred in ADR-0007.
- **No status, derived from dates.** A project with a delivery date is delivered. Nothing to keep in step, but the allowed actions spread over conditions on several columns, and combinations that mean nothing can be stored.
- **XState on the screens.** A library and a notation of its own for flows that a reducer covers.

## Consequences

- The statuses of a project and of a request are not decided. They come with the modelling of those domain areas.
- A new status needs a migration that changes the `CHECK` constraint, and a deploy.
- Whether a status can be divided further by configuration content, such as a stage a person names inside "in progress", is decided with the modelling of projects.
- How the name of a status is stored follows what ADR-0005 left open for configuration content. Each status needs a name before a person writes one.
- A rule that needs more than the status, such as a project that cannot be delivered with a request open, stays in the method. Whether the allowed actions reflect such a rule is settled with the first transition that has one.
- A transition records nothing by itself. Who moved a record, when, and from which status belongs to the history of work that screens show, which is not modelled.
- The steps of the document processing pipeline are named when the processors are implemented. Where files are converted stays open (ADR-0016).
- The asynchronous task of ADR-0004, when an agent needs it, is a record of this kind: created, running, then completed, failed or cancelled.
