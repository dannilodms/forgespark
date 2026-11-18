// Package application holds the use cases of the audit log. They are what
// another domain area calls: nothing else of this area is imported from
// outside it.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/domain"
)

// AuditLog is where entries are kept. Nothing reads it back yet, and nothing
// ever changes or removes an entry.
type AuditLog interface {
	// Append adds an entry. Inside the transaction of a use case it is saved
	// or lost together with what that use case saves.
	Append(ctx context.Context, entry domain.Entry) error
}

// RecordEntry records one sensitive operation in the audit log.
//
// The use case of the operation calls it inside its own transaction, so the
// operation never succeeds without its record. It is one of the two effects
// on another area that do not wait for the commit; the other is the
// confirmation of the images of the setup (Writes in CODING_STANDARDS.md).
type RecordEntry struct {
	auditLog AuditLog
}

// NewRecordEntry builds the use case on the audit log it writes to.
func NewRecordEntry(auditLog AuditLog) *RecordEntry {
	return &RecordEntry{auditLog: auditLog}
}

// RecordEntryInput describes the operation to record. OccurredAt is the time
// the use case of the operation took from its clock.
type RecordEntryInput struct {
	OccurredAt time.Time
	Operation  Operation
	Actor      Actor
	Subject    Subject
}

// Execute records the entry.
func (u *RecordEntry) Execute(ctx context.Context, input RecordEntryInput) error {
	entry, err := domain.NewEntry(input.OccurredAt, input.Operation, input.Actor, input.Subject)
	if err != nil {
		return err
	}
	return u.auditLog.Append(ctx, entry)
}
