package databasetest_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

func connect(t *testing.T, connectionURL string) *pgx.Conn {
	t.Helper()
	connection, err := pgx.Connect(context.Background(), connectionURL)
	require.NoError(t, err)
	t.Cleanup(func() { connection.Close(context.Background()) })
	return connection
}

func TestAMigratedDatabaseIsAtTheVersionTheBinaryExpects(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	applied, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)

	require.NoError(t, err)
	assert.Empty(t, applied, "the template left a migration to apply")
}

func TestAMigratedDatabaseIsReachedByTheTwoRoles(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	var migrationRole, apiRole string
	err := connect(t, temporary.MigrationURL).QueryRow(context.Background(), "SELECT current_user").Scan(&migrationRole)
	require.NoError(t, err)
	err = connect(t, temporary.APIURL).QueryRow(context.Background(), "SELECT current_user").Scan(&apiRole)
	require.NoError(t, err)

	assert.Equal(t, "migration", migrationRole)
	assert.Equal(t, "api", apiRole)
}

func TestEachMigratedDatabaseIsItsOwn(t *testing.T) {
	t.Parallel()
	first := databasetest.Migrated(t)
	second := databasetest.Migrated(t)
	ctx := context.Background()

	_, err := connect(t, first.MigrationURL).Exec(ctx, "CREATE TABLE only_in_the_first (id bigint)")
	require.NoError(t, err)

	var found bool
	err = connect(t, second.MigrationURL).QueryRow(ctx, "SELECT to_regclass('only_in_the_first') IS NOT NULL").Scan(&found)
	require.NoError(t, err)
	assert.False(t, found, "a table created in one database shows in another")
}
