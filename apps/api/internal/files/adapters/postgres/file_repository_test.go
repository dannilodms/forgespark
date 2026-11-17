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

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// stored is a migrated temporary database, reached as the API role, with the
// repository under test on it.
type stored struct {
	pool         *pgxpool.Pool
	transactions *database.Transactions
	files        *postgres.FileRepository
}

func newStored(t *testing.T) stored {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIURL,
		MaxSize:                  4,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return stored{pool: pool, transactions: database.NewTransactions(pool), files: postgres.NewFileRepository(pool)}
}

// start saves the record of a new file, in a transaction of its own. Each
// one gets a key of its own.
func (s stored) start(t *testing.T, purpose domain.Purpose, mediaType string, size int64) *domain.File {
	t.Helper()
	declared, err := domain.DeclareImage(mediaType, size)
	require.NoError(t, err)
	var file *domain.File
	err = s.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := s.files.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		key, err := domain.NewObjectKey(bytes.NewReader(bytes.Repeat([]byte{byte(id)}, 16)))
		if err != nil {
			return err
		}
		file, err = domain.StartUpload(id, key, purpose, declared, noon)
		if err != nil {
			return err
		}
		return s.files.Save(ctx, file)
	})
	require.NoError(t, err)
	return file
}

func (s stored) load(t *testing.T, id domain.FileID) (*domain.File, bool) {
	t.Helper()
	var file *domain.File
	var found bool
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		file, found, err = s.files.Load(ctx, id)
		return err
	})
	require.NoError(t, err)
	return file, found
}

func TestAFileIsLoadedAsItWasSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.start(t, domain.UserPhoto, "image/webp", 2048)

	loaded, found := temporary.load(t, saved.ID())

	require.True(t, found)
	assert.Equal(t, saved.ID(), loaded.ID())
	assert.Equal(t, saved.ObjectKey(), loaded.ObjectKey())
	assert.Equal(t, domain.UserPhoto, loaded.Purpose())
	assert.Equal(t, domain.WebP, loaded.Declared().MediaType())
	assert.Equal(t, int64(2048), loaded.Declared().Size())
	assert.Equal(t, noon, loaded.StartedAt())
	_, confirmed := loaded.ConfirmedAt()
	assert.False(t, confirmed)
}

func TestAFileConfirmedIsLoadedWithTheTimeOfItsConfirmation(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.start(t, domain.ConsultancyLogo, "image/png", 1024)
	confirmedAt := noon.Add(3 * time.Minute)

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		file, _, err := temporary.files.Load(ctx, saved.ID())
		if err != nil {
			return err
		}
		if err := file.Confirm(domain.StoredObject{Size: 1024, FirstBytes: []byte("\x89PNG\r\n\x1a\n")}, confirmedAt); err != nil {
			return err
		}
		return temporary.files.Save(ctx, file)
	})

	require.NoError(t, err)
	loaded, _ := temporary.load(t, saved.ID())
	loadedAt, confirmed := loaded.ConfirmedAt()
	assert.True(t, confirmed)
	assert.Equal(t, confirmedAt, loadedAt)
}

func TestNoFileIsLoadedForAnIdentifierThatWasNeverSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.start(t, domain.ConsultancyLogo, "image/png", 1024)

	loaded, found := temporary.load(t, domain.FileID(987654))

	assert.False(t, found)
	assert.Nil(t, loaded)
}

func TestLoadingAFileLocksItUntilTheTransactionEnds(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	saved := temporary.start(t, domain.ConsultancyLogo, "image/png", 1024)
	const lockNotAvailable = "55P03"

	var whileLocked, afterwards error
	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, _, err := temporary.files.Load(ctx, saved.ID())
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

// loadWithoutWaiting loads the file in a second transaction that gives up at
// once when its row is locked, where PostgreSQL would wait.
func (s stored) loadWithoutWaiting(t *testing.T, id domain.FileID) error {
	t.Helper()
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		_, err := database.Within(ctx, s.pool).Exec(ctx, "SET LOCAL lock_timeout = '100ms'")
		require.NoError(t, err)
		_, _, err = s.files.Load(ctx, id)
		return err
	})
}

// A file saved in a transaction that fails is not kept: the repository writes
// in the transaction of the context, never past it.
func TestAFileSavedInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	declared, err := domain.DeclareImage("image/png", 1024)
	require.NoError(t, err)
	failed := errors.New("a later step failed")
	var id domain.FileID

	err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		id, err = temporary.files.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		file, err := domain.StartUpload(id, domain.ObjectKey("a-key-of-a-file-not-kept"), domain.ConsultancyLogo, declared, noon)
		if err != nil {
			return err
		}
		if err := temporary.files.Save(ctx, file); err != nil {
			return err
		}
		return failed
	})

	require.ErrorIs(t, err, failed)
	_, found := temporary.load(t, id)
	assert.False(t, found)
}

func TestTwoFilesNeverShareAnObjectKey(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	first := temporary.start(t, domain.ConsultancyLogo, "image/png", 1024)
	declared, err := domain.DeclareImage("image/png", 1024)
	require.NoError(t, err)
	const uniqueViolation = "23505"

	err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := temporary.files.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		second, err := domain.StartUpload(id, first.ObjectKey(), domain.UserPhoto, declared, noon)
		if err != nil {
			return err
		}
		return temporary.files.Save(ctx, second)
	})

	var failure *pgconn.PgError
	require.ErrorAs(t, err, &failure)
	assert.Equal(t, uniqueViolation, failure.Code)
}
