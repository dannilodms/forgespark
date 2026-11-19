package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// stored is a migrated temporary database, reached as the API role, with the
// repository under test on it.
type stored struct {
	pool         *pgxpool.Pool
	transactions *database.Transactions
	emails       *postgres.EmailRepository
}

func newStored(t *testing.T) stored {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIConnection(t),
		MaxSize:                  4,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return stored{pool: pool, transactions: database.NewTransactions(pool), emails: postgres.NewEmailRepository(pool)}
}

// queue saves a new e-mail in pt-BR, in a transaction of its own.
func (s stored) queue(t *testing.T) *domain.Email {
	t.Helper()
	recipient, err := domain.NewRecipient("ana.souza@fabrica.example")
	require.NoError(t, err)
	language, err := kernel.NewLanguage("pt-BR")
	require.NoError(t, err)
	var email *domain.Email
	err = s.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := s.emails.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		email, err = domain.Queue(id, recipient, domain.LoginCode, language, domain.LoginCodeRequest(42), noon)
		if err != nil {
			return err
		}
		return s.emails.Save(ctx, email)
	})
	require.NoError(t, err)
	return email
}

// change loads the e-mail, changes it and saves it, in a transaction of its
// own.
func (s stored) change(t *testing.T, id domain.EmailID, change func(email *domain.Email) error) {
	t.Helper()
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		email, _, err := s.emails.Load(ctx, id)
		if err != nil {
			return err
		}
		if err := change(email); err != nil {
			return err
		}
		return s.emails.Save(ctx, email)
	})
	require.NoError(t, err)
}

func (s stored) load(t *testing.T, id domain.EmailID) (*domain.Email, bool) {
	t.Helper()
	var email *domain.Email
	var found bool
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		email, found, err = s.emails.Load(ctx, id)
		return err
	})
	require.NoError(t, err)
	return email, found
}

func TestAQueuedEmailIsLoadedAsItWasSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.queue(t)

	loaded, found := temporary.load(t, saved.ID())

	require.True(t, found)
	assert.Equal(t, saved, loaded)
	assert.Equal(t, "Seu código de login", loaded.Subject())
	assert.Equal(t, domain.Queued, loaded.Status())
}

func TestASentEmailIsLoadedWithTheIdentifierTheSenderGaveIt(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.queue(t)
	sentAt := noon.Add(2 * time.Second)

	temporary.change(t, saved.ID(), func(email *domain.Email) error {
		return email.MarkSent("abc123@worker.fabrica.example", sentAt)
	})

	loaded, _ := temporary.load(t, saved.ID())
	assert.Equal(t, domain.Sent, loaded.Status())
	messageID, sent := loaded.MessageID()
	assert.True(t, sent)
	assert.Equal(t, "abc123@worker.fabrica.example", messageID)
	loadedAt, _ := loaded.SentAt()
	assert.Equal(t, sentAt, loadedAt)
	_, _, failed := loaded.Failure()
	assert.False(t, failed)
}

func TestTheTableTakesEveryCodeOfAFailedEmail(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	failedAt := noon.Add(6 * time.Minute)

	for _, code := range []domain.FailureCode{domain.ContentUnavailable, domain.SendingFailed, domain.DeliveryFailed} {
		saved := temporary.queue(t)

		temporary.change(t, saved.ID(), func(email *domain.Email) error { return email.MarkFailed(code, failedAt) })

		loaded, _ := temporary.load(t, saved.ID())
		assert.Equal(t, domain.Failed, loaded.Status())
		loadedCode, loadedAt, failed := loaded.Failure()
		assert.True(t, failed)
		assert.Equal(t, code, loadedCode)
		assert.Equal(t, failedAt, loadedAt)
		_, sent := loaded.MessageID()
		assert.False(t, sent)
	}
}

func TestNoEmailIsLoadedForAnIdentifierThatWasNeverSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)

	loaded, found := temporary.load(t, domain.EmailID(987654))

	assert.False(t, found)
	assert.Nil(t, loaded)
}

func TestLoadingAnEmailLocksItUntilTheTransactionEnds(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.queue(t)
	const lockNotAvailable = "55P03"

	var whileLocked, afterwards error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, _, err := temporary.emails.Load(ctx, saved.ID())
		whileLocked = temporary.loadWithoutWaiting(t, saved.ID())
		return err
	})
	afterwards = temporary.loadWithoutWaiting(t, saved.ID())

	require.NoError(t, err)
	var failure *pgconn.PgError
	require.ErrorAs(t, whileLocked, &failure)
	assert.Equal(t, lockNotAvailable, failure.Code)
	assert.NoError(t, afterwards)
}

// loadWithoutWaiting loads the e-mail in a second transaction that gives up
// at once when its row is locked, where PostgreSQL would wait.
func (s stored) loadWithoutWaiting(t *testing.T, id domain.EmailID) error {
	t.Helper()
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := database.Within(ctx, s.pool).Exec(ctx, "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)
		_, _, err = s.emails.Load(ctx, id)
		return err
	})
}

func TestAnEmailSavedInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	recipient, err := domain.NewRecipient("ana.souza@fabrica.example")
	require.NoError(t, err)
	language, err := kernel.NewLanguage("en-US")
	require.NoError(t, err)
	failed := errors.New("a later step failed")
	var id domain.EmailID

	err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		id, err = temporary.emails.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		email, err := domain.Queue(id, recipient, domain.LoginCode, language, domain.LoginCodeRequest(1), noon)
		if err != nil {
			return err
		}
		if err := temporary.emails.Save(ctx, email); err != nil {
			return err
		}
		return failed
	})

	require.ErrorIs(t, err, failed)
	_, found := temporary.load(t, id)
	assert.False(t, found)
}

// The statuses, the templates and the kinds of reference are fixed lists of
// the domain, which the table checks too.
func TestTheTableRefusesAStatusATemplateOrAReferenceTheDomainDoesNotHave(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.queue(t)
	const checkViolation = "23514"

	for _, statement := range []string{
		"UPDATE emails SET status = 'opened' WHERE id = $1",
		"UPDATE emails SET template = 'newsletter' WHERE id = $1",
		"UPDATE emails SET reference_kind = 'invoice' WHERE id = $1",
		"UPDATE emails SET failure_code = 'tired' WHERE id = $1",
	} {
		_, err := temporary.pool.Exec(context.Background(), statement, int64(saved.ID()))

		var failure *pgconn.PgError
		require.ErrorAs(t, err, &failure, statement)
		assert.Equal(t, checkViolation, failure.Code, statement)
	}
	_, err := temporary.pool.Exec(context.Background(), "UPDATE emails SET status = 'bounced' WHERE id = $1", int64(saved.ID()))
	assert.NoError(t, err, "a bounced e-mail is a status of the table")
}
