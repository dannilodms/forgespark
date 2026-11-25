package migrations_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/forgespark/apps/api/migrations"
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

func TestTheAPIRoleCannotCreateATemporaryTable(t *testing.T) {
	t.Parallel()
	// A database the migrations were applied to, as the development
	// database and the one of a server are.
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)

	assert.Equal(t, insufficientPrivilege, refusal(t, temporary, "CREATE TEMPORARY TABLE intruder (id bigint)"))
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
		ConnectionURL:            temporary.APIConnection(t),
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

	applied, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)

	require.NoError(t, err)
	assert.Empty(t, applied)
}

// River's tables are written by the API, which saves jobs, and by the worker,
// which runs them and elects a leader, both as the API role (Jobs in
// CODING_STANDARDS.md). The reindexer of River rebuilds the indexes of the
// jobs table, which asks for MAINTAIN on it.
func TestTheAPIRoleReadsWritesAndReindexesTheTablesOfRiver(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, temporary.APIURL)
	require.NoError(t, err)
	defer connection.Close(ctx)

	statements := []string{
		`INSERT INTO river_job (kind, args, max_attempts, queue) VALUES ('proof', '{"id": 1}', 5, 'light')`,
		"UPDATE river_job SET state = 'running', attempt = 1, attempted_at = now() WHERE kind = 'proof'",
		"SELECT count(*) FROM river_job WHERE state = 'running'",
		"DELETE FROM river_job WHERE kind = 'proof'",
		"INSERT INTO river_leader (elected_at, expires_at, leader_id) VALUES (now(), now() + interval '10 seconds', 'worker-1')",
		"DELETE FROM river_leader",
		"INSERT INTO river_queue (name) VALUES ('light')",
		"UPDATE river_queue SET paused_at = now() WHERE name = 'light'",
		"DELETE FROM river_queue",
		"INSERT INTO river_notification (payload, topic) VALUES ('{}', 'proof')",
		"DELETE FROM river_notification",
		"REINDEX INDEX CONCURRENTLY river_job_kind",
	}
	for _, statement := range statements {
		_, err := connection.Exec(ctx, statement)
		assert.NoError(t, err, statement)
	}
}

func TestTheAPIRoleCannotChangeTheTablesOfRiver(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	statements := []string{
		"ALTER TABLE river_job ADD COLUMN intruder text",
		"DROP TABLE river_job",
		"TRUNCATE river_job",
		"ALTER TYPE river_job_state ADD VALUE 'intruder'",
	}
	for _, statement := range statements {
		assert.Equal(t, insufficientPrivilege, refusal(t, temporary, statement), statement)
	}
}

func TestTheAPIRoleAddsAndReadsEntriesOfTheAuditLogAndCannotChangeOrRemoveOne(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, temporary.APIURL)
	require.NoError(t, err)
	defer connection.Close(ctx)

	_, err = connection.Exec(ctx, `
		INSERT INTO audit_log (occurred_at, operation, actor_kind, actor_id, subject_kind, subject_id, way_in)
		VALUES (now(), 'setup_code_issued', 'installer', NULL, 'setup_code', 1, NULL),
		       (now(), 'logged_in', 'user', 1, 'user', 1, 'login_code')`)
	require.NoError(t, err)
	var entries int
	require.NoError(t, connection.QueryRow(ctx, "SELECT count(*) FROM audit_log").Scan(&entries))
	assert.Equal(t, 2, entries)

	statements := []string{
		"UPDATE audit_log SET subject_id = 2",
		"UPDATE audit_log SET way_in = 'local' WHERE operation = 'logged_in'",
		"DELETE FROM audit_log",
		"DELETE FROM audit_log WHERE operation = 'logged_in'",
		"TRUNCATE audit_log",
	}
	for _, statement := range statements {
		assert.Equal(t, insufficientPrivilege, refusal(t, temporary, statement), statement)
	}
	require.NoError(t, connection.QueryRow(ctx,
		"SELECT count(*) FROM audit_log WHERE subject_id = 1 AND way_in IS NOT DISTINCT FROM CASE operation WHEN 'logged_in' THEN 'login_code' END").Scan(&entries))
	assert.Equal(t, 2, entries, "an entry was changed or removed")
}
