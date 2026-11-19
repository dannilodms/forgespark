package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/audit/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// apiPool opens a pool, as the API role, on a migrated temporary database.
func apiPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIConnection(t),
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
	entry, err := domain.NewEntry(noon, domain.SetupCodeIssued, domain.Installer(), domain.SetupCode(7), domain.NoWayIn)
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

// The table refuses an operation its CHECK constraint does not list, so each
// operation of the domain is appended once.
func TestTheTableTakesEveryOperationTheAuditLogRecords(t *testing.T) {
	t.Parallel()
	pool := apiPool(t)
	auditLog := postgres.NewAuditLog(pool)
	entries := map[domain.Operation]func() (domain.Entry, error){
		domain.SetupCodeIssued: func() (domain.Entry, error) {
			return domain.NewEntry(noon, domain.SetupCodeIssued, domain.Installer(), domain.SetupCode(7), domain.NoWayIn)
		},
		domain.SetupConcluded: func() (domain.Entry, error) {
			return domain.NewEntry(noon, domain.SetupConcluded, domain.SetupCodeUsed(7), domain.User(3), domain.NoWayIn)
		},
		domain.AdministratorRegistered: func() (domain.Entry, error) {
			return domain.NewEntry(noon, domain.AdministratorRegistered, domain.SetupCodeUsed(8), domain.User(4), domain.NoWayIn)
		},
		domain.LoggedIn: func() (domain.Entry, error) {
			return domain.NewEntry(noon, domain.LoggedIn, domain.ActingUser(4), domain.User(4), kernel.LoginCodeWayIn)
		},
	}

	for operation, entryOf := range entries {
		entry, err := entryOf()
		require.NoError(t, err)
		assert.NoError(t, auditLog.Append(context.Background(), entry), "the operation %s", operation)
	}
	assert.Equal(t, len(entries), entriesIn(t, pool))
}

// A login keeps the user who logged in and the way in: the login code, or the
// key of a login provider. Any other operation keeps no way in.
func TestALoginIsARowWithTheUserAndTheWayIn(t *testing.T) {
	t.Parallel()
	pool := apiPool(t)
	auditLog := postgres.NewAuditLog(pool)
	for _, way := range []kernel.WayIn{kernel.LoginCodeWayIn, "local"} {
		entry, err := domain.NewEntry(noon, domain.LoggedIn, domain.ActingUser(4), domain.User(4), way)
		require.NoError(t, err)
		require.NoError(t, auditLog.Append(context.Background(), entry))
	}
	require.NoError(t, auditLog.Append(context.Background(), setupCodeIssued(t)))

	rows, err := pool.Query(context.Background(), `
		SELECT concat_ws(' ', operation, actor_kind, coalesce(actor_id::text, 'none'), subject_kind, subject_id,
		                 coalesce(way_in, '(none)'))
		FROM audit_log ORDER BY id`)
	require.NoError(t, err)
	found, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	assert.Equal(t, []string{
		"logged_in user 4 user 4 login_code",
		"logged_in user 4 user 4 local",
		"setup_code_issued installer none setup_code 7 (none)",
	}, found)
}

// The table holds a way in for every login and for nothing else, whoever
// writes the row.
func TestTheTableRefusesALoginWithNoWayInAndAWayInForAnotherOperation(t *testing.T) {
	t.Parallel()
	pool := apiPool(t)

	statements := []string{
		`INSERT INTO audit_log (occurred_at, operation, actor_kind, actor_id, subject_kind, subject_id)
		 VALUES (now(), 'logged_in', 'user', 4, 'user', 4)`,
		`INSERT INTO audit_log (occurred_at, operation, actor_kind, subject_kind, subject_id, way_in)
		 VALUES (now(), 'setup_code_issued', 'installer', 'setup_code', 7, 'login_code')`,
	}
	for _, statement := range statements {
		_, err := pool.Exec(context.Background(), statement)
		var failure *pgconn.PgError
		require.ErrorAs(t, err, &failure, statement)
		assert.Equal(t, "23514", failure.Code, "a CHECK constraint: %s", statement)
	}
	assert.Zero(t, entriesIn(t, pool))
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
