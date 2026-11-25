package postgres_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/access/adapters/postgres"
	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// stored is a migrated temporary database, reached as the API role, with the
// repository under test on it.
type stored struct {
	pool         *pgxpool.Pool
	transactions *database.Transactions
	requests     *postgres.LoginCodeRequestRepository
}

func newStored(t *testing.T) stored {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:    databasetest.Migrated(t).APIConnection(t),
		MaxSize:          4,
		StatementTimeout: 10 * time.Second,
		// A test holds one transaction open while a second one runs.
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return stored{pool: pool, transactions: database.NewTransactions(pool), requests: postgres.NewLoginCodeRequestRepository(pool)}
}

// aUser writes a user row past the consultancy area, which owns users: a
// request refers to one. It returns its identifier.
func (s stored) aUser(t *testing.T, email string) domain.UserID {
	t.Helper()
	var id int64
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO users (id, name, email, country, national_id, preferred_language, deactivated)
		VALUES (nextval('users_id_seq'), 'Someone', $1, 'BR', '1', 'pt-BR', false)
		RETURNING id`, email).Scan(&id)
	require.NoError(t, err)
	return domain.UserID(id)
}

// ask saves a new request of the user at this moment, in a transaction of its
// own. Each one is drawn from a random value of its own.
func (s stored) ask(t *testing.T, user domain.UserID, at time.Time) *domain.LoginCodeRequest {
	t.Helper()
	var request *domain.LoginCodeRequest
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := s.requests.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		request, err = domain.AskForLoginCode(id, user, at, bytes.NewReader(bytes.Repeat([]byte{byte(id)}, 32)))
		if err != nil {
			return err
		}
		return s.requests.Save(ctx, request)
	})
	require.NoError(t, err)
	return request
}

func (s stored) find(t *testing.T, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool) {
	t.Helper()
	found, ok, err := s.requests.Find(context.Background(), id)
	require.NoError(t, err)
	return found, ok
}

func TestARequestIsFoundAsItWasSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.ask(t, temporary.aUser(t, "ana@fabrica.example"), noon)

	found, ok := temporary.find(t, saved.ID())

	require.True(t, ok)
	assert.Equal(t, saved, found)
}

func TestAnEndedRequestIsFoundWithWhenItWasEnded(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.ask(t, temporary.aUser(t, "ana@fabrica.example"), noon)

	saved.End(noon.Add(time.Minute))
	require.NoError(t, temporary.requests.Save(context.Background(), saved))

	found, _ := temporary.find(t, saved.ID())
	endedAt, ended := found.EndedAt()
	assert.True(t, ended)
	assert.Equal(t, noon.Add(time.Minute), endedAt)
}

func TestNoRequestIsFoundForAnIdentifierThatWasNeverSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)

	found, ok := temporary.find(t, domain.LoginCodeRequestID(987654))

	assert.False(t, ok)
	assert.Nil(t, found)
}

func TestTheRequestsOfAUserThatWereNotEndedAreLoaded(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	earlier := temporary.ask(t, ana, noon)
	earlier.End(noon.Add(time.Minute))
	require.NoError(t, temporary.requests.Save(context.Background(), earlier))
	open := temporary.ask(t, ana, noon.Add(time.Minute))
	temporary.ask(t, temporary.aUser(t, "bruno@fabrica.example"), noon)

	var loaded []*domain.LoginCodeRequest
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		loaded, err = temporary.requests.LoadNotEndedOf(ctx, ana)
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, []*domain.LoginCodeRequest{open}, loaded)
}

func TestLoadingTheRequestsOfAUserLocksThemUntilTheTransactionEnds(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	temporary.ask(t, ana, noon)
	const lockNotAvailable = "55P03"

	var whileLocked, afterwards error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := temporary.requests.LoadNotEndedOf(ctx, ana)
		whileLocked = temporary.loadWithoutWaiting(t, ana)
		return err
	})
	afterwards = temporary.loadWithoutWaiting(t, ana)

	require.NoError(t, err)
	var failure *pgconn.PgError
	require.ErrorAs(t, whileLocked, &failure)
	assert.Equal(t, lockNotAvailable, failure.Code)
	assert.NoError(t, afterwards)
}

// loadWithoutWaiting loads the requests of the user in a second transaction
// that gives up at once when a row is locked, where PostgreSQL would wait.
func (s stored) loadWithoutWaiting(t *testing.T, user domain.UserID) error {
	t.Helper()
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := database.Within(ctx, s.pool).Exec(ctx, "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)
		_, err = s.requests.LoadNotEndedOf(ctx, user)
		return err
	})
}

// A check loads its request with a lock, so two checks of one code count
// their attempts one after the other. The attempts are saved with the request.
func TestARequestLoadedForItsCheckIsLockedAndKeepsItsAttempts(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.ask(t, temporary.aUser(t, "ana@fabrica.example"), noon)
	key, err := domain.NewLoginCodeKey(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)
	const lockNotAvailable = "55P03"

	var whileLocked error
	err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		loaded, found, err := temporary.requests.Load(ctx, saved.ID())
		require.True(t, found)
		require.Equal(t, saved, loaded)
		whileLocked = temporary.loadOneWithoutWaiting(t, saved.ID())
		_ = loaded.Check("not the code", key, noon.Add(time.Minute))
		return errors.Join(err, temporary.requests.Save(ctx, loaded))
	})

	require.NoError(t, err)
	var failure *pgconn.PgError
	require.ErrorAs(t, whileLocked, &failure)
	assert.Equal(t, lockNotAvailable, failure.Code)
	assert.NoError(t, temporary.loadOneWithoutWaiting(t, saved.ID()))
	found, _ := temporary.find(t, saved.ID())
	assert.Equal(t, 1, found.Attempts())
}

func TestNoRequestIsLoadedForAnIdentifierThatWasNeverSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)

	for _, id := range []domain.LoginCodeRequestID{domain.NoLoginCodeRequest, 987654} {
		err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
			loaded, found, err := temporary.requests.Load(ctx, id)
			assert.False(t, found, "the request %d", id)
			assert.Nil(t, loaded)
			return err
		})
		require.NoError(t, err)
	}
}

// loadOneWithoutWaiting loads the request in a second transaction that gives
// up at once when its row is locked.
func (s stored) loadOneWithoutWaiting(t *testing.T, id domain.LoginCodeRequestID) error {
	t.Helper()
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := database.Within(ctx, s.pool).Exec(ctx, "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)
		_, _, err = s.requests.Load(ctx, id)
		return err
	})
}

func TestTheDatabaseRefusesASecondRequestOfAUserThatWasNotEnded(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	temporary.ask(t, ana, noon)
	const uniqueViolation = "23505"

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := temporary.requests.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		request, err := domain.AskForLoginCode(id, ana, noon.Add(time.Minute), bytes.NewReader(make([]byte, 32)))
		if err != nil {
			return err
		}
		return temporary.requests.Save(ctx, request)
	})

	var failure *pgconn.PgError
	require.ErrorAs(t, err, &failure)
	assert.Equal(t, uniqueViolation, failure.Code)
}

func TestARequestSavedInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	failed := errors.New("a later step failed")
	var id domain.LoginCodeRequestID

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		id, err = temporary.requests.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		request, err := domain.AskForLoginCode(id, ana, noon, bytes.NewReader(make([]byte, 32)))
		if err != nil {
			return err
		}
		if err := temporary.requests.Save(ctx, request); err != nil {
			return err
		}
		return failed
	})

	require.ErrorIs(t, err, failed)
	_, found := temporary.find(t, id)
	assert.False(t, found)
}

// The code is computed from the identifier and the random value, with a key
// the database never holds: no column of a request carries it.
func TestARowOfARequestHoldsNoCode(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.ask(t, temporary.aUser(t, "ana@fabrica.example"), noon)
	key, err := domain.NewLoginCodeKey(bytes.Repeat([]byte{7}, 32))
	require.NoError(t, err)
	code := saved.Code(key).Reveal()

	var row string
	require.NoError(t, temporary.pool.QueryRow(context.Background(),
		"SELECT row_to_json(login_code_requests)::text FROM login_code_requests WHERE id = $1", int64(saved.ID())).Scan(&row))

	assert.NotContains(t, row, code)
	assert.Contains(t, row, `"random_value"`)
}

// The daily cleanup removes the requests that ended, or expired, before a
// moment, and keeps the others.
func TestTheRequestsThatEndedOrExpiredBeforeAMomentAreRemoved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	expired := temporary.ask(t, ana, noon.Add(-time.Hour))
	expired.End(noon.Add(time.Hour))
	require.NoError(t, temporary.requests.Save(context.Background(), expired))
	ended := temporary.ask(t, ana, noon.Add(-5*time.Minute))
	ended.End(noon.Add(-4 * time.Minute))
	require.NoError(t, temporary.requests.Save(context.Background(), ended))
	endedLater := temporary.ask(t, ana, noon.Add(-time.Minute))
	endedLater.End(noon.Add(time.Minute))
	require.NoError(t, temporary.requests.Save(context.Background(), endedLater))
	open := temporary.ask(t, ana, noon)

	err := temporary.requests.RemoveEndedOrExpiredBefore(context.Background(), noon)

	require.NoError(t, err)
	_, keptExpired := temporary.find(t, expired.ID())
	_, keptEnded := temporary.find(t, ended.ID())
	_, keptEndedLater := temporary.find(t, endedLater.ID())
	_, keptOpen := temporary.find(t, open.ID())
	assert.False(t, keptExpired, "a request that expired, then ended, was kept")
	assert.False(t, keptEnded, "a request that ended before the moment was kept")
	assert.True(t, keptEndedLater)
	assert.True(t, keptOpen)
}
