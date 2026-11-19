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

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
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

func (s stored) countSince(t *testing.T, user domain.UserID, since time.Time) int {
	t.Helper()
	count, err := s.requests.CountAskedSince(context.Background(), user, since)
	require.NoError(t, err)
	return count
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

func TestTheRequestsAskedForSinceAMomentAreCountedForEachUser(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	bruno := temporary.aUser(t, "bruno@fabrica.example")
	for _, minutes := range []int{0, 10, 20} {
		request := temporary.ask(t, ana, noon.Add(time.Duration(minutes)*time.Minute))
		request.End(noon.Add(time.Duration(minutes+1) * time.Minute))
		require.NoError(t, temporary.requests.Save(context.Background(), request))
	}
	temporary.ask(t, bruno, noon.Add(15*time.Minute))

	assert.Equal(t, 3, temporary.countSince(t, ana, noon.Add(-time.Second)), "an ended request still counts")
	assert.Equal(t, 2, temporary.countSince(t, ana, noon), "a request asked for at the moment itself is not after it")
	assert.Equal(t, 0, temporary.countSince(t, ana, noon.Add(20*time.Minute)))
	assert.Equal(t, 1, temporary.countSince(t, bruno, noon))
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
