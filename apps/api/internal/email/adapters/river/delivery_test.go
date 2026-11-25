package river_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	riverqueue "github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/email/adapters/river"
	"github.com/dannilodms/forgespark/apps/api/internal/email/application"
	"github.com/dannilodms/forgespark/apps/api/internal/email/application/applicationtest"
	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/jobs"
)

const recipient = "ana.souza@fabrica.example"

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// poolOf opens the pool of the API role on a temporary database with every
// migration applied, where River keeps its jobs.
func poolOf(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIConnection(t),
		MaxSize:                  4,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

func TestTheJobOfAnEmailIsSavedInTheTransactionInTheLightQueueWithItsIdentifierAlone(t *testing.T) {
	t.Parallel()
	pool := poolOf(t)
	saver, err := jobs.NewSaver(pool, nil, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	require.NoError(t, err)
	deliveries := river.NewDeliveryJobs(saver)

	err = database.NewTransactions(pool).Run(context.Background(), func(ctx context.Context) error {
		return deliveries.Save(ctx, 37)
	})

	require.NoError(t, err)
	// The arguments name the kind of the job the helper looks for.
	job := rivertest.RequireInserted(context.Background(), t, riverpgxv5.New(pool), river.DeliveryArgs{EmailID: 0}, &rivertest.RequireInsertedOpts{
		Queue:       jobs.QueueLight,
		MaxAttempts: 5,
	})
	assert.Equal(t, int64(37), job.Args.EmailID)
	var arguments string
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT args::text FROM river_job WHERE kind = 'deliver_email'").Scan(&arguments))
	assert.JSONEq(t, `{"email_id": `+strconv.Itoa(37)+`}`, arguments)
}

// office is the e-mail area on its in-memory implementations, with its job
// worker run by River's test helper in a temporary database.
type office struct {
	pool           *pgxpool.Pool
	emails         *applicationtest.EmailRepository
	sender         *applicationtest.Sender
	deliveryWorker *river.DeliveryWorker
	worker         *rivertest.Worker[river.DeliveryArgs, pgx.Tx]
	id             domain.EmailID
}

// newOffice queues the e-mail of the login code request 42 to Ana, whose code
// the source of login codes gives.
func newOffice(t *testing.T) office {
	t.Helper()
	clock := applicationtest.NewClock(noon)
	emails := applicationtest.NewEmailRepository()
	transactions := applicationtest.NewTransactions(emails)
	sender := applicationtest.NewSender()
	loginCodes := applicationtest.NewContentSource()
	loginCodes.Give(domain.LoginCodeRequest(42), domain.LoginCodeContent{Code: "482915", Lifetime: 10 * time.Minute})
	language, err := kernel.NewLanguage("en-US")
	require.NoError(t, err)
	queue := application.NewQueueEmail(transactions, emails, applicationtest.NewDeliveryJobs(), clock)
	require.NoError(t, queue.Execute(context.Background(), application.QueueEmailInput{
		Recipient: recipient,
		Template:  application.LoginCode,
		Language:  language,
		Reference: application.LoginCodeRequest(42),
	}))
	deliver := application.NewDeliverEmail(transactions, emails, application.ContentSources{application.LoginCode: loginCodes}, sender, clock)
	pool := poolOf(t)
	worker := river.NewDeliveryWorker(deliver)
	return office{
		pool:           pool,
		emails:         emails,
		sender:         sender,
		deliveryWorker: worker,
		worker:         rivertest.NewWorker(t, riverpgxv5.New(pool), &riverqueue.Config{}, worker),
		id:             emails.All()[0].ID(),
	}
}

func (o office) status(t *testing.T) domain.Status {
	t.Helper()
	email, found := o.emails.Find(o.id)
	require.True(t, found)
	return email.Status()
}

// work runs the job of the e-mail once, as River would, with this many
// attempts in all.
func (o office) work(t *testing.T, maxAttempts int) (*rivertest.WorkResult, error) {
	t.Helper()
	transaction, err := o.pool.Begin(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { transaction.Rollback(context.Background()) })
	return o.worker.Work(context.Background(), t, transaction, river.DeliveryArgs{EmailID: int64(o.id)}, &riverqueue.InsertOpts{MaxAttempts: maxAttempts})
}

func TestTheJobWorkerDeliversTheEmailOfItsJob(t *testing.T) {
	t.Parallel()
	post := newOffice(t)

	result, err := post.work(t, 5)

	require.NoError(t, err)
	assert.Equal(t, riverqueue.EventKindJobCompleted, result.EventKind)
	sent := post.sender.Sent()
	require.Len(t, sent, 1)
	assert.Equal(t, recipient, sent[0].To)
	assert.Contains(t, sent[0].Text, "482915")
	assert.Equal(t, domain.Sent, post.status(t))
}

func TestAJobThatFailsBeforeItsLastAttemptLeavesTheEmailQueued(t *testing.T) {
	t.Parallel()
	post := newOffice(t)
	unreachable := errors.New("dial tcp 127.0.0.1:1025: connection refused")
	post.sender.FailWith(unreachable)

	result, err := post.work(t, 5)

	require.ErrorIs(t, err, unreachable)
	assert.Equal(t, riverqueue.EventKindJobFailed, result.EventKind)
	assert.NotEqual(t, rivertype.JobStateDiscarded, result.Job.State, "River runs the job again")
	assert.Equal(t, domain.Queued, post.status(t))
}

// River's timeout of a job ends its context with the deadline, which is not
// the cut of a shutdown: the last attempt that runs out of its time marks the
// e-mail failed, so it is not left queued with no job.
func TestTheLastAttemptOfAJobThatRunsOutOfItsTimeMarksTheEmailFailed(t *testing.T) {
	t.Parallel()
	post := newOffice(t)
	post.sender.HangUntilTheContextEnds()
	post.worker = rivertest.NewWorker(t, riverpgxv5.New(post.pool), &riverqueue.Config{JobTimeout: 50 * time.Millisecond}, post.deliveryWorker)

	result, err := post.work(t, 1)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, rivertype.JobStateDiscarded, result.Job.State, "River runs the job no more")
	email, _ := post.emails.Find(post.id)
	code, _, failed := email.Failure()
	assert.True(t, failed, "the e-mail is left queued with no job")
	assert.Equal(t, domain.SendingFailed, code)
}

func TestTheLastAttemptOfAJobThatFailsMarksTheEmailFailed(t *testing.T) {
	t.Parallel()
	post := newOffice(t)
	unreachable := errors.New("dial tcp 127.0.0.1:1025: connection refused")
	post.sender.FailWith(unreachable)

	result, err := post.work(t, 1)

	require.ErrorIs(t, err, unreachable)
	assert.Equal(t, riverqueue.EventKindJobFailed, result.EventKind)
	assert.Equal(t, rivertype.JobStateDiscarded, result.Job.State, "River runs the job no more")
	email, _ := post.emails.Find(post.id)
	code, _, failed := email.Failure()
	assert.True(t, failed)
	assert.Equal(t, domain.SendingFailed, code)
}
