package postgres_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// stored is a migrated temporary database, reached as the API role, with the
// repository under test on it.
type stored struct {
	pool          *pgxpool.Pool
	transactions  *database.Transactions
	setupCodes    *postgres.SetupCodeRepository
	consultancies *postgres.ConsultancyRepository
	users         *postgres.UserRepository
	roles         *postgres.RoleRepository
	setupState    *postgres.SetupStateReader
	activeUsers   *postgres.ActiveUserReader
}

func newStored(t *testing.T) stored {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:    databasetest.Migrated(t).APIConnection(t),
		MaxSize:          4,
		StatementTimeout: 10 * time.Second,
		// A test holds one transaction open while a second one runs. On a
		// loaded machine the second takes its time to connect, and the first
		// must not be ended for waiting.
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return stored{
		pool:          pool,
		transactions:  database.NewTransactions(pool),
		setupCodes:    postgres.NewSetupCodeRepository(pool),
		consultancies: postgres.NewConsultancyRepository(pool),
		users:         postgres.NewUserRepository(pool),
		roles:         postgres.NewRoleRepository(pool),
		setupState:    postgres.NewSetupStateReader(pool),
		activeUsers:   postgres.NewActiveUserReader(pool),
	}
}

// issue saves a new setup code, in a transaction of its own, and returns it.
// Each one is drawn from different bytes, so no two have the same hash.
func (s stored) issue(t *testing.T, issuedAt time.Time) *domain.SetupCode {
	t.Helper()
	var setupCode *domain.SetupCode
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := s.setupCodes.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		random := bytes.NewReader(bytes.Repeat([]byte{byte(id)}, 20))
		setupCode, _, err = domain.IssueSetupCode(id, issuedAt, random)
		if err != nil {
			return err
		}
		return s.setupCodes.Save(ctx, setupCode)
	})
	require.NoError(t, err)
	return setupCode
}

func (s stored) notEnded(t *testing.T) []*domain.SetupCode {
	t.Helper()
	var loaded []*domain.SetupCode
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		loaded, err = s.setupCodes.LoadNotEnded(ctx)
		return err
	})
	require.NoError(t, err)
	return loaded
}

func sqlState(t *testing.T, err error) string {
	t.Helper()
	var failure *pgconn.PgError
	require.ErrorAs(t, err, &failure)
	return failure.Code
}

func TestASetupCodeIsLoadedAsItWasSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.issue(t, noon)

	loaded := temporary.notEnded(t)

	require.Len(t, loaded, 1)
	assert.Equal(t, saved.ID(), loaded[0].ID())
	assert.Equal(t, saved.Hash(), loaded[0].Hash())
	assert.Equal(t, noon, loaded[0].IssuedAt())
	assert.Equal(t, noon.Add(time.Hour), loaded[0].ExpiresAt())
	_, ended := loaded[0].EndedAt()
	assert.False(t, ended)
}

func TestASetupCodeSavedAgainAfterItIsEndedIsNoLongerAmongTheOnesNotEnded(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.issue(t, noon)
	endedAt := noon.Add(10 * time.Minute)

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		loaded, err := temporary.setupCodes.LoadNotEnded(ctx)
		if err != nil {
			return err
		}
		loaded[0].End(endedAt)
		return temporary.setupCodes.Save(ctx, loaded[0])
	})

	require.NoError(t, err)
	assert.Empty(t, temporary.notEnded(t))
	var storedEndedAt time.Time
	err = temporary.pool.QueryRow(context.Background(), "SELECT ended_at FROM setup_codes").Scan(&storedEndedAt)
	require.NoError(t, err)
	assert.True(t, endedAt.Equal(storedEndedAt), "ended at %s", storedEndedAt)
}

func TestEachIdentifierReservedIsNew(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ctx := context.Background()

	first, err := temporary.setupCodes.NextIdentifier(ctx)
	require.NoError(t, err)
	second, err := temporary.setupCodes.NextIdentifier(ctx)
	require.NoError(t, err)

	assert.Positive(t, int64(first))
	assert.Greater(t, int64(second), int64(first))
}

func TestLoadingASetupCodeLocksItUntilTheTransactionEnds(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.issue(t, noon)
	const lockNotAvailable = "55P03"

	var whileLocked, afterwards error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := temporary.setupCodes.LoadNotEnded(ctx)
		whileLocked = temporary.loadWithoutWaiting(t)
		return err
	})
	afterwards = temporary.loadWithoutWaiting(t)

	require.NoError(t, err)
	assert.Equal(t, lockNotAvailable, sqlState(t, whileLocked))
	assert.NoError(t, afterwards)
}

// loadWithoutWaiting loads the setup codes in a second transaction that gives
// up at once when a row is locked, where PostgreSQL would wait.
func (s stored) loadWithoutWaiting(t *testing.T) error {
	t.Helper()
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := database.Within(ctx, s.pool).Exec(ctx, "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)
		_, err = s.setupCodes.LoadNotEnded(ctx)
		return err
	})
}

func TestTheDatabaseRefusesASecondSetupCodeThatWasNotEnded(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.issue(t, noon)
	const uniqueViolation = "23505"

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := temporary.setupCodes.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		second, _, err := domain.IssueSetupCode(id, noon, bytes.NewReader(bytes.Repeat([]byte{31}, 20)))
		if err != nil {
			return err
		}
		return temporary.setupCodes.Save(ctx, second)
	})

	assert.Equal(t, uniqueViolation, sqlState(t, err))
	assert.Len(t, temporary.notEnded(t), 1)
}

func TestASetupCodeIsFoundByItsHashWhetherItWasEndedOrNot(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ctx := context.Background()
	first := temporary.issue(t, noon)
	endedAt := noon.Add(10 * time.Minute)
	first.End(endedAt)
	require.NoError(t, temporary.setupCodes.Save(ctx, first))
	second := temporary.issue(t, noon.Add(10*time.Minute))

	foundFirst, firstExists, err := temporary.setupCodes.FindByHash(ctx, first.Hash())
	require.NoError(t, err)
	foundSecond, secondExists, err := temporary.setupCodes.FindByHash(ctx, second.Hash())
	require.NoError(t, err)

	require.True(t, firstExists)
	assert.Equal(t, first.ID(), foundFirst.ID())
	assert.Equal(t, noon.Add(time.Hour), foundFirst.ExpiresAt())
	foundEndedAt, ended := foundFirst.EndedAt()
	assert.True(t, ended)
	assert.Equal(t, endedAt, foundEndedAt)
	require.True(t, secondExists)
	assert.Equal(t, second.ID(), foundSecond.ID())
	assert.Equal(t, noon.Add(10*time.Minute), foundSecond.IssuedAt())
	_, ended = foundSecond.EndedAt()
	assert.False(t, ended)
}

func TestNoSetupCodeIsFoundByAHashThatWasNeverStored(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.issue(t, noon)
	_, never, err := domain.IssueSetupCode(1, noon, bytes.NewReader(bytes.Repeat([]byte{31}, 20)))
	require.NoError(t, err)

	found, exists, err := temporary.setupCodes.FindByHash(context.Background(), never.Hash())

	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, found)
}

func TestLoadingASetupCodeByItsHashLocksItUntilTheTransactionEnds(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	issued := temporary.issue(t, noon)
	const lockNotAvailable = "55P03"

	var loaded *domain.SetupCode
	var found bool
	var whileLocked, afterwards error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		loaded, found, err = temporary.setupCodes.LoadByHash(ctx, issued.Hash())
		whileLocked = temporary.loadWithoutWaiting(t)
		return err
	})
	afterwards = temporary.loadWithoutWaiting(t)

	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, issued.ID(), loaded.ID())
	assert.Equal(t, lockNotAvailable, sqlState(t, whileLocked))
	assert.NoError(t, afterwards)
}

func TestFindingASetupCodeByItsHashLocksNothing(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	issued := temporary.issue(t, noon)

	var whileFound error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, _, err := temporary.setupCodes.FindByHash(ctx, issued.Hash())
		whileFound = temporary.loadWithoutWaiting(t)
		return err
	})

	require.NoError(t, err)
	assert.NoError(t, whileFound)
}
