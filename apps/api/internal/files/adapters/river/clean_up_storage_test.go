package river_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	filesriver "github.com/dannilodms/forgespark/apps/api/internal/files/adapters/river"
	"github.com/dannilodms/forgespark/apps/api/internal/files/application"
	"github.com/dannilodms/forgespark/apps/api/internal/files/application/applicationtest"
	"github.com/dannilodms/forgespark/apps/api/internal/files/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/jobs"
)

var noon = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

// The job of the cleanup names nothing: the River panel and the telemetry
// show the arguments of a job.
func TestTheJobOfTheCleanupCarriesNoArgument(t *testing.T) {
	arguments, err := json.Marshal(filesriver.CleanUpStorageArgs{})

	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(arguments))
}

// The cleanup is light work, and lists the whole bucket, which takes longer
// than River gives a job by default.
func TestTheJobOfTheCleanupRunsInTheLightQueueWithALongerTimeout(t *testing.T) {
	worker := filesriver.NewCleanUpStorageWorker(nil)

	assert.Equal(t, jobs.QueueLight, filesriver.CleanUpStorageArgs{}.InsertOpts().Queue)
	assert.Greater(t, worker.Timeout(&river.Job[filesriver.CleanUpStorageArgs]{}), river.JobTimeoutDefault)
}

func TestTheJobOfTheCleanupCleansUpTheStorage(t *testing.T) {
	t.Parallel()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIConnection(t),
		MaxSize:                  4,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	clock := applicationtest.NewClock(noon)
	files := applicationtest.NewFileRepository()
	objects := applicationtest.NewStorage(clock)
	const old, young domain.ObjectKey = "0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f", "1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e"
	objects.Put(old, []byte("left behind"))
	clock.Advance(25 * time.Hour)
	objects.Put(young, []byte("left behind an hour ago"))
	clock.Advance(time.Hour)
	useCase := application.NewCleanUpStorage(applicationtest.NewTransactions(files), files, objects, clock)
	worker := rivertest.NewWorker(t, riverpgxv5.New(pool), &river.Config{}, filesriver.NewCleanUpStorageWorker(useCase))
	ctx := context.Background()
	transaction, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer transaction.Rollback(ctx)

	result, err := worker.Work(ctx, t, transaction, filesriver.CleanUpStorageArgs{}, nil)

	require.NoError(t, err)
	assert.Equal(t, river.EventKindJobCompleted, result.EventKind)
	assert.Equal(t, []domain.ObjectKey{young}, objects.Keys())
}
