// Package postgres keeps the audit log in PostgreSQL. Its SQL is in the .sql
// file beside this one, and sqlc generates the functions that run it: the
// files named sqlc_*.go and *.sql.go, which are never edited.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/forgespark/apps/api/internal/audit/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
)

// AuditLog keeps the entries in the audit_log table, where the API role can
// insert and select and nothing else.
type AuditLog struct {
	pool *pgxpool.Pool
}

// NewAuditLog returns the audit log on this pool.
func NewAuditLog(pool *pgxpool.Pool) *AuditLog {
	return &AuditLog{pool: pool}
}

// Append inserts the entry. It runs in the transaction the context carries,
// which is the one of the operation the entry records: the two are saved
// together or not at all.
func (l *AuditLog) Append(ctx context.Context, entry domain.Entry) error {
	err := New(database.Within(ctx, l.pool)).AppendAuditLogEntry(ctx, AppendAuditLogEntryParams{
		OccurredAt:  entry.OccurredAt(),
		Operation:   string(entry.Operation()),
		ActorKind:   string(entry.Actor().Kind()),
		ActorID:     actorIDOf(entry.Actor()),
		SubjectKind: string(entry.Subject().Kind()),
		SubjectID:   entry.Subject().Identifier(),
		WayIn:       wayInOf(entry),
	})
	if err != nil {
		return fmt.Errorf("append the entry to the audit log: %w", err)
	}
	return nil
}

// wayInOf is the way in of a login, and nil for any other operation.
func wayInOf(entry domain.Entry) *string {
	wayIn, has := entry.WayIn()
	if !has {
		return nil
	}
	text := string(wayIn)
	return &text
}

func actorIDOf(actor domain.Actor) *int64 {
	identifier, identified := actor.Identifier()
	if !identified {
		return nil
	}
	return &identifier
}
