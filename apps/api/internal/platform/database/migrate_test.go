package database_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
)

// twoMigrations stands for the migrations embedded in a binary.
func twoMigrations() fstest.MapFS {
	return fstest.MapFS{
		"00001_create_clients.sql":  {Data: []byte("-- +goose Up\nCREATE TABLE clients (id bigint PRIMARY KEY);\n")},
		"00002_create_projects.sql": {Data: []byte("-- +goose Up\nCREATE TABLE projects (id bigint PRIMARY KEY);\n")},
	}
}

func onlyTheFirst(migrations fstest.MapFS) fstest.MapFS {
	return fstest.MapFS{"00001_create_clients.sql": migrations["00001_create_clients.sql"]}
}

func tables(t *testing.T, connectionURL string) []string {
	t.Helper()
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, connectionURL)
	require.NoError(t, err)
	defer connection.Close(ctx)
	rows, err := connection.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename <> 'goose_db_version' ORDER BY tablename")
	require.NoError(t, err)
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	return names
}

func TestMigrateAppliesThePendingMigrationsInOrder(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)

	applied, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), twoMigrations())

	require.NoError(t, err)
	assert.Equal(t, []string{"00001_create_clients.sql", "00002_create_projects.sql"}, applied)
	assert.Equal(t, []string{"clients", "projects"}, tables(t, temporary.MigrationURL))
}

func TestMigrateRunTwiceChangesNothingTheSecondTime(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), twoMigrations())
	require.NoError(t, err)

	applied, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), twoMigrations())

	require.NoError(t, err)
	assert.Empty(t, applied)
	assert.Equal(t, []string{"clients", "projects"}, tables(t, temporary.MigrationURL))
}

func TestMigrateAppliesOnlyWhatTheDatabaseLacks(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), onlyTheFirst(twoMigrations()))
	require.NoError(t, err)

	applied, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), twoMigrations())

	require.NoError(t, err)
	assert.Equal(t, []string{"00002_create_projects.sql"}, applied)
}
