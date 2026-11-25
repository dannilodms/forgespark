// Package applicationtest holds the in-memory implementations of the
// interfaces the use cases of the audit log declare. A test of a use case, of
// this area or of one that records an entry, builds the use case on them.
package applicationtest

import (
	"context"
	"slices"

	"github.com/dannilodms/forgespark/apps/api/internal/audit/application"
)

// AuditLog keeps the entries in memory, in the order they were recorded.
type AuditLog struct {
	entries []application.Entry
	failure error
}

// NewAuditLog returns an audit log with no entry.
func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

// FailWith makes every later Append fail with this error, as an audit log
// that cannot be written to.
func (l *AuditLog) FailWith(failure error) {
	l.failure = failure
}

// Append adds an entry.
func (l *AuditLog) Append(_ context.Context, entry application.Entry) error {
	if l.failure != nil {
		return l.failure
	}
	l.entries = append(l.entries, entry)
	return nil
}

// Entries returns what was recorded so far.
func (l *AuditLog) Entries() []application.Entry {
	return slices.Clone(l.entries)
}

// Snapshot lets the audit log take part in the in-memory transactions of the
// area whose use case records an entry: it gives a function that puts the
// audit log back as it is now.
func (l *AuditLog) Snapshot() (restore func()) {
	saved := slices.Clone(l.entries)
	return func() { l.entries = saved }
}
