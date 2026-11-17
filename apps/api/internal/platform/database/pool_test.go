package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

func sqlState(t *testing.T, err error) string {
	t.Helper()
	var failure *pgconn.PgError
	require.ErrorAs(t, err, &failure)
	return failure.Code
}

func TestThePoolCancelsAStatementThatOutlastsItsTimeout(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIURL,
		MaxSize:                  1,
		StatementTimeout:         100 * time.Millisecond,
		IdleInTransactionTimeout: 10 * time.Second,
	})
	require.NoError(t, err)
	defer pool.Close()

	_, err = pool.Exec(context.Background(), "SELECT pg_sleep(5)")

	const queryCanceled = "57014"
	assert.Equal(t, queryCanceled, sqlState(t, err))
}

func TestThePoolEndsATransactionLeftIdle(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIURL,
		MaxSize:                  1,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: 100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer pool.Close()
	transaction, err := pool.Begin(context.Background())
	require.NoError(t, err)
	defer transaction.Rollback(context.Background())

	time.Sleep(time.Second)
	_, err = transaction.Exec(context.Background(), "SELECT 1")

	assert.Error(t, err)
}

func TestThePoolOpensNoMoreConnectionsThanItsMaximumSize(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIURL,
		MaxSize:                  2,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: 10 * time.Second,
	})
	require.NoError(t, err)
	defer pool.Close()
	first, err := pool.Acquire(context.Background())
	require.NoError(t, err)
	defer first.Release()
	second, err := pool.Acquire(context.Background())
	require.NoError(t, err)
	defer second.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	third, err := pool.Acquire(ctx)
	if err == nil {
		third.Release()
	}

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
