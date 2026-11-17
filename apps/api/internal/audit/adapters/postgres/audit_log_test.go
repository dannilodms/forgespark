package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// apiPool opens a pool, as the API role, on a migrated temporary database.
func apiPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIURL,
		MaxSize:                  2,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: 10 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

func setupCodeIssued(t *testing.T) domain.Entry {
	t.Helper()
	entry, err := domain.NewEntry(noon, domain.SetupCodeIssued, domain.Installer(), domain.SetupCode(7))
	require.NoError(t, err)
	return entry
}

func entriesIn(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var entries int
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT count(*) FROM audit_log").Scan(&entries))
	return entries
}

func TestAnEntryAppendedIsARowOfTheAuditLogTable(t *testing.T) {
	t.Parallel()
	pool := apiPool(t)
	auditLog := postgres.NewAuditLog(pool)

	err := auditLog.Append(context.Background(), setupCodeIssued(t))

	require.NoError(t, err)
	var (
		occurredAt           time.Time
		operation, actorKind string
		actorID              *int64
		subjectKind          string
		subjectID            int64
	)
	err = pool.QueryRow(context.Background(), `
		SELECT occurred_at, operation, actor_kind, actor_id, subject_kind, subject_id
		FROM audit_log`).Scan(&occurredAt, &operation, &actorKind, &actorID, &subjectKind, &subjectID)
	require.NoError(t, err)
	assert.True(t, noon.Equal(occurredAt), "occurred at %s", occurredAt)
	assert.Equal(t, "setup_code_issued", operation)
	assert.Equal(t, "installer", actorKind)
	assert.Nil(t, actorID, "whoever installs the system has no identifier")
	assert.Equal(t, "setup_code", subjectKind)
	assert.Equal(t, int64(7), subjectID)
}

func TestAnEntryAppendedInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	pool := apiPool(t)
	auditLog := postgres.NewAuditLog(pool)
	failure := errors.New("the operation failed after its entry was written")

	err := database.NewTransactions(pool).Run(context.Background(), func(ctx context.Context) error {
		return errors.Join(auditLog.Append(ctx, setupCodeIssued(t)), failure)
	})

	require.ErrorIs(t, err, failure)
	assert.Zero(t, entriesIn(t, pool))
}

func TestAnEntryAppendedInATransactionThatSucceedsIsKept(t *testing.T) {
	t.Parallel()
	pool := apiPool(t)
	auditLog := postgres.NewAuditLog(pool)

	err := database.NewTransactions(pool).Run(context.Background(), func(ctx context.Context) error {
		return auditLog.Append(ctx, setupCodeIssued(t))
	})

	require.NoError(t, err)
	assert.Equal(t, 1, entriesIn(t, pool))
}
