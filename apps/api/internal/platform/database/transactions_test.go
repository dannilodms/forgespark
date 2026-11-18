package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// poolWithNotes opens a pool on a temporary database that has one table,
// `notes`, to write to.
func poolWithNotes(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := openPool(t, databasetest.New(t).MigrationConnection(t))
	_, err := pool.Exec(context.Background(), "CREATE TABLE notes (text text NOT NULL)")
	require.NoError(t, err)
	return pool
}

// addNote writes the way a `postgres` adapter does: on whatever the context
// carries.
func addNote(ctx context.Context, pool *pgxpool.Pool, text string) error {
	_, err := database.Within(ctx, pool).Exec(ctx, "INSERT INTO notes (text) VALUES ($1)", text)
	return err
}

func notes(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	rows, err := pool.Query(context.Background(), "SELECT text FROM notes ORDER BY text")
	require.NoError(t, err)
	var found []string
	for rows.Next() {
		var text string
		require.NoError(t, rows.Scan(&text))
		found = append(found, text)
	}
	require.NoError(t, rows.Err())
	return found
}

func TestATransactionKeepsEverythingItsOperationWroteWhenTheOperationSucceeds(t *testing.T) {
	t.Parallel()
	pool := poolWithNotes(t)
	transactions := database.NewTransactions(pool)

	err := transactions.Run(context.Background(), func(ctx context.Context) error {
		return errors.Join(addNote(ctx, pool, "first"), addNote(ctx, pool, "second"))
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second"}, notes(t, pool))
}

func TestATransactionKeepsNothingWhenItsOperationFails(t *testing.T) {
	t.Parallel()
	pool := poolWithNotes(t)
	transactions := database.NewTransactions(pool)
	failure := errors.New("the second step failed")

	err := transactions.Run(context.Background(), func(ctx context.Context) error {
		return errors.Join(addNote(ctx, pool, "first"), failure)
	})

	require.ErrorIs(t, err, failure)
	assert.Empty(t, notes(t, pool))
}

func TestWhatIsWrittenInsideATransactionShowsOutsideOnlyAfterItEnds(t *testing.T) {
	t.Parallel()
	pool := poolWithNotes(t)
	transactions := database.NewTransactions(pool)

	var seenFromOutside []string
	err := transactions.Run(context.Background(), func(ctx context.Context) error {
		err := addNote(ctx, pool, "first")
		seenFromOutside = notes(t, pool)
		return err
	})

	require.NoError(t, err)
	assert.Empty(t, seenFromOutside, "a write was visible before its transaction ended")
	assert.Equal(t, []string{"first"}, notes(t, pool))
}

func TestAnOperationThatOpensATransactionInsideAnotherJoinsIt(t *testing.T) {
	t.Parallel()
	pool := poolWithNotes(t)
	transactions := database.NewTransactions(pool)
	failure := errors.New("the outer operation failed")

	err := transactions.Run(context.Background(), func(ctx context.Context) error {
		inner := transactions.Run(ctx, func(ctx context.Context) error {
			return addNote(ctx, pool, "inner")
		})
		return errors.Join(inner, failure)
	})

	require.ErrorIs(t, err, failure)
	assert.Empty(t, notes(t, pool), "the inner operation was saved although the outer one failed")
}

func TestOutsideATransactionAWriteGoesStraightToThePool(t *testing.T) {
	t.Parallel()
	pool := poolWithNotes(t)

	err := addNote(context.Background(), pool, "first")

	require.NoError(t, err)
	assert.Equal(t, []string{"first"}, notes(t, pool))
}
