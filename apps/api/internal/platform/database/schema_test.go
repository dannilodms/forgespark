package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

func openPool(t *testing.T, connectionURL string) *pgxpool.Pool {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            connectionURL,
		MaxSize:                  2,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: 10 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

func TestTheSchemaCheckPassesWhenEveryMigrationIsApplied(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationURL, twoMigrations())
	require.NoError(t, err)

	err = database.CheckSchema(context.Background(), openPool(t, temporary.MigrationURL), twoMigrations())

	assert.NoError(t, err)
}

func TestTheSchemaCheckFailsWhileAMigrationIsPending(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationURL, onlyTheFirst(twoMigrations()))
	require.NoError(t, err)

	err = database.CheckSchema(context.Background(), openPool(t, temporary.MigrationURL), twoMigrations())

	assert.EqualError(t, err, "the database schema is at version 1 and this binary expects version 2")
}

func TestTheSchemaCheckFailsOnADatabaseNeverMigrated(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)

	err := database.CheckSchema(context.Background(), openPool(t, temporary.MigrationURL), twoMigrations())

	assert.EqualError(t, err, "the database schema is at version 0 and this binary expects version 2")
}

func TestTheSchemaCheckFailsWhenTheDatabaseIsAheadOfTheBinary(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationURL, twoMigrations())
	require.NoError(t, err)

	err = database.CheckSchema(context.Background(), openPool(t, temporary.MigrationURL), onlyTheFirst(twoMigrations()))

	assert.EqualError(t, err, "the database schema is at version 2 and this binary expects version 1")
}

func TestThePoolOpensWhilePostgreSQLIsDownAndTheSchemaCheckFails(t *testing.T) {
	t.Parallel()
	// Port 1 of the machine has no PostgreSQL behind it.
	pool := openPool(t, "postgres://api:api@127.0.0.1:1/software_factory?sslmode=disable")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := database.CheckSchema(ctx, pool, twoMigrations())

	assert.ErrorContains(t, err, "PostgreSQL does not answer")
}
