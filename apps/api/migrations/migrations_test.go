package migrations_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

const insufficientPrivilege = "42501"

// refusal runs a statement as the API role and returns the SQLSTATE of the
// error PostgreSQL answers with.
func refusal(t *testing.T, temporary databasetest.Database, statement string) string {
	t.Helper()
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, temporary.APIURL)
	require.NoError(t, err)
	defer connection.Close(ctx)

	_, err = connection.Exec(ctx, statement)

	var failure *pgconn.PgError
	require.ErrorAs(t, err, &failure, "the API role ran: %s", statement)
	return failure.Code
}

func TestTheAPIRoleCannotChangeTheSchema(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	statements := []string{
		"CREATE TABLE intruder (id bigint)",
		"ALTER TABLE goose_db_version ADD COLUMN intruder text",
		"DROP TABLE goose_db_version",
		"CREATE SCHEMA intruder",
	}
	for _, statement := range statements {
		assert.Equal(t, insufficientPrivilege, refusal(t, temporary, statement), statement)
	}
}

func TestTheAPIRoleCannotRewriteWhichMigrationsWereApplied(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	statements := []string{
		"INSERT INTO goose_db_version (version_id, is_applied) VALUES (999, true)",
		"UPDATE goose_db_version SET version_id = 999",
		"DELETE FROM goose_db_version",
	}
	for _, statement := range statements {
		assert.Equal(t, insufficientPrivilege, refusal(t, temporary, statement), statement)
	}
}

func TestTheAPIRoleSeesTheSchemaAtTheVersionTheBinaryExpects(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIURL,
		MaxSize:                  1,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: 10 * time.Second,
	})
	require.NoError(t, err)
	defer pool.Close()

	err = database.CheckSchema(context.Background(), pool, migrations.Files)

	assert.NoError(t, err)
}

func TestTheMigrationsApplyNothingOnASecondRun(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	applied, err := database.Migrate(context.Background(), temporary.MigrationURL, migrations.Files)

	require.NoError(t, err)
	assert.Empty(t, applied)
}

func TestTheAPIRoleAddsAndReadsEntriesOfTheAuditLogAndCannotChangeOrRemoveOne(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, temporary.APIURL)
	require.NoError(t, err)
	defer connection.Close(ctx)

	_, err = connection.Exec(ctx, `
		INSERT INTO audit_log (occurred_at, operation, actor_kind, subject_kind, subject_id)
		VALUES (now(), 'setup_code_issued', 'installer', 'setup_code', 1)`)
	require.NoError(t, err)
	var entries int
	require.NoError(t, connection.QueryRow(ctx, "SELECT count(*) FROM audit_log").Scan(&entries))
	assert.Equal(t, 1, entries)

	statements := []string{
		"UPDATE audit_log SET subject_id = 2",
		"DELETE FROM audit_log",
		"TRUNCATE audit_log",
	}
	for _, statement := range statements {
		assert.Equal(t, insufficientPrivilege, refusal(t, temporary, statement), statement)
	}
	require.NoError(t, connection.QueryRow(ctx, "SELECT count(*) FROM audit_log WHERE subject_id = 1").Scan(&entries))
	assert.Equal(t, 1, entries, "the entry was changed or removed")
}
