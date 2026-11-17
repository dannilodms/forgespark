// Package applicationtest holds the in-memory implementations of the
// interfaces the use cases of the audit log declare. A test of a use case, of
// this area or of one that records an entry, builds the use case on them.
package applicationtest

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
)

// AuditLog keeps the entries in memory, in the order they were recorded.
type AuditLog struct {
	entries []application.Entry
}

// NewAuditLog returns an audit log with no entry.
func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

// Append adds an entry.
func (l *AuditLog) Append(_ context.Context, entry application.Entry) error {
	l.entries = append(l.entries, entry)
	return nil
}

// Entries returns what was recorded so far.
func (l *AuditLog) Entries() []application.Entry {
	return append([]application.Entry(nil), l.entries...)
}
