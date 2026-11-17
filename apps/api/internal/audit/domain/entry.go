// Package domain holds what an entry of the audit log is: the permanent
// record of one sensitive operation. An entry carries identifiers only, never
// a name or an e-mail, and nothing changes it once it is recorded.
package domain

import (
	"fmt"
	"time"
)

// Operation is a sensitive operation the audit log records.
type Operation string

// SetupCodeIssued is recorded when whoever installs the system issues a
// setup code.
const SetupCodeIssued Operation = "setup_code_issued"

// ActorKind tells what kind of thing acted.
type ActorKind string

// InstallerKind is whoever installs the system: a person on the server, who
// is not a user and has no identifier.
const InstallerKind ActorKind = "installer"

// Actor is who performed an operation.
type Actor struct {
	kind       ActorKind
	identifier int64
	identified bool
}

// Installer is whoever installs the system.
func Installer() Actor {
	return Actor{kind: InstallerKind}
}

// Kind tells what kind of thing acted.
func (a Actor) Kind() ActorKind {
	return a.kind
}

// Identifier returns the identifier of the actor, and false when its kind
// has none.
func (a Actor) Identifier() (int64, bool) {
	return a.identifier, a.identified
}

// SubjectKind tells what kind of thing was acted on.
type SubjectKind string

// SetupCodeKind is a setup code.
const SetupCodeKind SubjectKind = "setup_code"

// Subject is what an operation acted on.
type Subject struct {
	kind       SubjectKind
	identifier int64
}

// SetupCode is the setup code with this identifier.
func SetupCode(identifier int64) Subject {
	return Subject{kind: SetupCodeKind, identifier: identifier}
}

// Kind tells what kind of thing was acted on.
func (s Subject) Kind() SubjectKind {
	return s.kind
}

// Identifier returns the identifier of what was acted on.
func (s Subject) Identifier() int64 {
	return s.identifier
}

// Entry is one record of the audit log.
type Entry struct {
	occurredAt time.Time
	operation  Operation
	actor      Actor
	subject    Subject
}

// recorded lists, for each operation the audit log records, the kind of its
// actor and the kind of its subject. A new sensitive operation is a new line
// here, and a new value in the CHECK constraints of the audit log table.
var recorded = map[Operation]struct {
	actor   ActorKind
	subject SubjectKind
}{
	SetupCodeIssued: {actor: InstallerKind, subject: SetupCodeKind},
}

// NewEntry builds the entry of an operation. It refuses an operation the
// audit log does not record, an actor or a subject of another kind than the
// operation has, a subject with no identifier and a missing time.
func NewEntry(occurredAt time.Time, operation Operation, actor Actor, subject Subject) (Entry, error) {
	expected, known := recorded[operation]
	if !known {
		return Entry{}, fmt.Errorf("the audit log does not record the operation %q", operation)
	}
	if actor.kind != expected.actor {
		return Entry{}, fmt.Errorf("the actor of %q is %q, not %q", operation, expected.actor, actor.kind)
	}
	if subject.kind != expected.subject {
		return Entry{}, fmt.Errorf("the subject of %q is %q, not %q", operation, expected.subject, subject.kind)
	}
	if subject.identifier <= 0 {
		return Entry{}, fmt.Errorf("the subject of %q has no identifier", operation)
	}
	if occurredAt.IsZero() {
		return Entry{}, fmt.Errorf("the entry of %q has no time", operation)
	}
	return Entry{occurredAt: occurredAt, operation: operation, actor: actor, subject: subject}, nil
}

// OccurredAt returns when the operation happened.
func (e Entry) OccurredAt() time.Time {
	return e.occurredAt
}

// Operation returns the operation recorded.
func (e Entry) Operation() Operation {
	return e.operation
}

// Actor returns who performed the operation.
func (e Entry) Actor() Actor {
	return e.actor
}

// Subject returns what the operation acted on.
func (e Entry) Subject() Subject {
	return e.subject
}
