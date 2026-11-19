package postgres_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// addressOf is the address as the limits keep it, under a key of the test.
func addressOf(t *testing.T, typed string) domain.AskedAddress {
	t.Helper()
	key, err := domain.NewLoginCodeKey(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)
	email, err := kernel.NewEmail(typed)
	require.NoError(t, err)
	return key.AskedAddress(email)
}

// asked keeps an ask for the address at this moment, in a transaction of its
// own.
func (s stored) asked(t *testing.T, asks *postgres.AskedAddressRepository, address domain.AskedAddress, at time.Time) {
	t.Helper()
	require.NoError(t, s.transactions.Run(context.Background(), func(ctx context.Context) error {
		return asks.Add(ctx, address, at)
	}))
}

func countSince(t *testing.T, asks *postgres.AskedAddressRepository, address domain.AskedAddress, since time.Time) int {
	t.Helper()
	count, err := asks.CountAskedSince(context.Background(), address, since)
	require.NoError(t, err)
	return count
}

func TestTheAsksForAnAddressSinceAMomentAreCountedForEachAddress(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	asks := postgres.NewAskedAddressRepository(temporary.pool)
	ana, bruno := addressOf(t, "ana@fabrica.example"), addressOf(t, "bruno@fabrica.example")
	for _, minutes := range []int{0, 10, 20} {
		temporary.asked(t, asks, ana, noon.Add(time.Duration(minutes)*time.Minute))
	}
	temporary.asked(t, asks, bruno, noon.Add(15*time.Minute))

	assert.Equal(t, 3, countSince(t, asks, ana, noon.Add(-time.Second)))
	assert.Equal(t, 2, countSince(t, asks, ana, noon), "an ask made at the moment itself is not after it")
	assert.Equal(t, 0, countSince(t, asks, ana, noon.Add(20*time.Minute)))
	assert.Equal(t, 1, countSince(t, asks, bruno, noon))
}

// The table holds the digest of the address and when it was asked for: no
// address.
func TestARowOfAnAskHoldsNoAddress(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	asks := postgres.NewAskedAddressRepository(temporary.pool)
	temporary.asked(t, asks, addressOf(t, "ana.souza@fabrica.example"), noon)

	var row string
	require.NoError(t, temporary.pool.QueryRow(context.Background(),
		"SELECT row_to_json(login_code_asks)::text FROM login_code_asks").Scan(&row))

	assert.NotContains(t, row, "ana.souza")
	assert.NotContains(t, row, "fabrica")
	assert.Contains(t, row, `"address"`)
}

// Two asks for one address at the same moment are counted one after the
// other: the second waits for the transaction of the first, and then counts
// its ask, so it meets the limit of a minute and never breaks the unique
// index of the open requests (ticket 03).
func TestTwoAsksForOneAddressAtTheSameMomentAreCountedOneAfterTheOther(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	asks := postgres.NewAskedAddressRepository(temporary.pool)
	address := addressOf(t, "ana.souza@fabrica.example")
	counted := make(chan int, 1)
	failed := make(chan error, 1)

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		require.NoError(t, asks.Hold(ctx, address))
		require.NoError(t, asks.Add(ctx, address, noon))
		go func() {
			failed <- temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
				if err := asks.Hold(ctx, address); err != nil {
					return err
				}
				count, err := asks.CountAskedSince(ctx, address, noon.Add(-time.Minute))
				counted <- count
				return err
			})
		}()
		temporary.waitForATransactionThatWaits(t)
		return nil
	})

	require.NoError(t, err)
	require.NoError(t, <-failed)
	assert.Equal(t, 1, <-counted, "the second ask did not count the first")
}

// waitForATransactionThatWaits waits until a transaction on the database of
// the test waits for a lock another one holds.
func (s stored) waitForATransactionThatWaits(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting int
		err := s.pool.QueryRow(context.Background(), `
			SELECT count(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`).Scan(&waiting)
		require.NoError(t, err)
		return waiting > 0
	}, 10*time.Second, 20*time.Millisecond, "no transaction waited for the first")
}

func TestHoldingTheAsksForAnAddressHoldsNoOtherAddress(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	asks := postgres.NewAskedAddressRepository(temporary.pool)
	ana, bruno := addressOf(t, "ana@fabrica.example"), addressOf(t, "bruno@fabrica.example")
	const lockNotAvailable = "55P03"

	var sameAddress, anotherAddress error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		require.NoError(t, asks.Hold(ctx, ana))
		sameAddress = temporary.holdWithoutWaiting(t, asks, ana)
		anotherAddress = temporary.holdWithoutWaiting(t, asks, bruno)
		return nil
	})

	require.NoError(t, err)
	var failure *pgconn.PgError
	require.ErrorAs(t, sameAddress, &failure)
	assert.Equal(t, lockNotAvailable, failure.Code)
	assert.NoError(t, anotherAddress)
	assert.NoError(t, temporary.holdWithoutWaiting(t, asks, ana), "the lock outlived its transaction")
}

// holdWithoutWaiting holds the asks for the address in a second transaction
// that gives up at once when the lock is taken.
func (s stored) holdWithoutWaiting(t *testing.T, asks *postgres.AskedAddressRepository, address domain.AskedAddress) error {
	t.Helper()
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := database.Within(ctx, s.pool).Exec(ctx, "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)
		return asks.Hold(ctx, address)
	})
}

// The daily cleanup removes the asks made before a moment, of every address.
func TestTheAsksMadeBeforeAMomentAreRemoved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	asks := postgres.NewAskedAddressRepository(temporary.pool)
	ana, bruno := addressOf(t, "ana@fabrica.example"), addressOf(t, "bruno@fabrica.example")
	temporary.asked(t, asks, ana, noon.Add(-time.Hour))
	temporary.asked(t, asks, bruno, noon.Add(-time.Minute))
	temporary.asked(t, asks, ana, noon)

	err := asks.RemoveAskedBefore(context.Background(), noon)

	require.NoError(t, err)
	assert.Equal(t, 1, countSince(t, asks, ana, noon.Add(-2*time.Hour)))
	assert.Zero(t, countSince(t, asks, bruno, noon.Add(-2*time.Hour)))
}
