package jobs_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry/telemetrytest"
)

// These tests stand in for the first `adapters/river` package: a job kind
// whose arguments are identifiers, a job worker that reads them and calls one
// use case, and the saving of the job in the transaction of a use case. The
// use case here records what it was called with.

// lightArgs are the arguments of a job of the light queue.
type lightArgs struct {
	FileID int64 `json:"file_id"`
}

func (lightArgs) Kind() string { return "light_proof" }

func (lightArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{Queue: jobs.QueueLight} }

// heavyArgs are the arguments of a job of the heavy queue.
type heavyArgs struct {
	FileID int64 `json:"file_id"`
}

func (heavyArgs) Kind() string { return "heavy_proof" }

func (heavyArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{Queue: jobs.QueueHeavy} }

// recordingUseCase stands for the one use case a job worker calls.
type recordingUseCase struct {
	calls chan int64
}

func newRecordingUseCase() *recordingUseCase {
	return &recordingUseCase{calls: make(chan int64, 8)}
}

func (u *recordingUseCase) Execute(_ context.Context, fileID int64) error {
	u.calls <- fileID
	return nil
}

// lightWorker is the shape of a job worker: it reads the identifiers from
// the arguments and calls the use case.
type lightWorker struct {
	river.WorkerDefaults[lightArgs]
	useCase *recordingUseCase
}

func (w *lightWorker) Work(ctx context.Context, job *river.Job[lightArgs]) error {
	return w.useCase.Execute(ctx, job.Args.FileID)
}

type heavyWorker struct {
	river.WorkerDefaults[heavyArgs]
	useCase *recordingUseCase
}

func (w *heavyWorker) Work(ctx context.Context, job *river.Job[heavyArgs]) error {
	return w.useCase.Execute(ctx, job.Args.FileID)
}

// slowArgs are the arguments of a job that runs until it is released or cut.
type slowArgs struct{}

func (slowArgs) Kind() string { return "slow_proof" }

func (slowArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{Queue: jobs.QueueHeavy} }

// slowWorker tells when a run starts, then waits for its release or for its
// context to end, and returns the error of the context then, as a job that
// respects its context does.
type slowWorker struct {
	river.WorkerDefaults[slowArgs]
	started chan struct{}
	release chan struct{}
}

func newSlowWorker() *slowWorker {
	return &slowWorker{started: make(chan struct{}, 8), release: make(chan struct{})}
}

func (w *slowWorker) Work(ctx context.Context, _ *river.Job[slowArgs]) error {
	w.started <- struct{}{}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.release:
		return nil
	}
}

// poolOf opens the pool of the API role on a temporary database with every
// migration applied.
func poolOf(t *testing.T, provider trace.TracerProvider) (databasetest.Database, *pgxpool.Pool) {
	t.Helper()
	temporary := databasetest.Migrated(t)
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIConnection(t),
		MaxSize:                  8,
		StatementTimeout:         30 * time.Second,
		IdleInTransactionTimeout: 30 * time.Second,
		TracerProvider:           provider,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return temporary, pool
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelError}))
}

func settingsOf(kinds *jobs.Kinds, deadline time.Duration, provider trace.TracerProvider, logger *slog.Logger) jobs.RunnerSettings {
	return jobs.RunnerSettings{
		LightQueueMaxJobs: 2,
		HeavyQueueMaxJobs: 1,
		ShutdownDeadline:  deadline,
		Kinds:             kinds,
		PeriodicJobs:      nil,
		TracerProvider:    provider,
		Logger:            logger,
	}
}

// runUntilSignalled runs the runner as the `worker` command does, until the
// returned function is called: it sends the signal and returns what Run
// returned, failing the test when Run takes longer than a minute to stop.
func runUntilSignalled(t *testing.T, runner *jobs.Runner) func() error {
	t.Helper()
	shutdown, signal := context.WithCancel(context.Background())
	stopped := make(chan error, 1)
	go func() { stopped <- runner.Run(shutdown) }()
	t.Cleanup(signal)
	return func() error {
		signal()
		select {
		case err := <-stopped:
			return err
		case <-time.After(time.Minute):
			t.Fatal("the runner did not stop within a minute of the signal")
			return nil
		}
	}
}

func receive[T any](t *testing.T, from <-chan T, what string) T {
	t.Helper()
	select {
	case value := <-from:
		return value
	case <-time.After(30 * time.Second):
		t.Fatalf("%s did not happen within 30 seconds", what)
		var zero T
		return zero
	}
}

func TestAJobWorkerReadsTheIdentifiersOfTheJobAndCallsItsUseCase(t *testing.T) {
	t.Parallel()
	_, pool := poolOf(t, nil)
	ctx := context.Background()
	useCase := newRecordingUseCase()
	worker := rivertest.NewWorker(t, riverpgxv5.New(pool), &river.Config{}, &lightWorker{useCase: useCase})
	transaction, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer transaction.Rollback(ctx)

	result, err := worker.Work(ctx, t, transaction, lightArgs{FileID: 42}, nil)

	require.NoError(t, err)
	assert.Equal(t, river.EventKindJobCompleted, result.EventKind)
	assert.Equal(t, int64(42), receive(t, useCase.calls, "the call of the use case"))
}

func TestSaveSavesTheJobInTheTransactionOfTheUseCaseAndNothingWhenItRollsBack(t *testing.T) {
	t.Parallel()
	_, pool := poolOf(t, nil)
	ctx := context.Background()
	transactions := database.NewTransactions(pool)
	saver, err := jobs.NewSaver(pool, nil, quietLogger())
	require.NoError(t, err)

	err = transactions.Run(ctx, func(ctx context.Context) error {
		return saver.Save(ctx, lightArgs{FileID: 7}, nil)
	})
	require.NoError(t, err)
	brokenRule := errors.New("a rule of the use case broke")
	err = transactions.Run(ctx, func(ctx context.Context) error {
		require.NoError(t, saver.Save(ctx, lightArgs{FileID: 8}, nil))
		return brokenRule
	})
	require.ErrorIs(t, err, brokenRule)

	saved := rivertest.RequireInserted(ctx, t, riverpgxv5.New(pool), lightArgs{}, &rivertest.RequireInsertedOpts{
		Queue:       jobs.QueueLight,
		MaxAttempts: 5,
	})
	assert.Equal(t, int64(7), saved.Args.FileID)
	var count int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM river_job WHERE kind = $1", lightArgs{}.Kind()).Scan(&count))
	assert.Equal(t, 1, count, "the job of the transaction that rolled back was saved")
}

func TestSaveRefusesToSaveAJobOutsideATransaction(t *testing.T) {
	t.Parallel()
	_, pool := poolOf(t, nil)
	saver, err := jobs.NewSaver(pool, nil, quietLogger())
	require.NoError(t, err)

	err = saver.Save(context.Background(), lightArgs{FileID: 7}, nil)

	assert.ErrorIs(t, err, jobs.ErrNoTransaction)
}

func TestTheRunnerRunsTheJobsOfBothQueuesAndTheRunLinksToTheTraceThatSavedTheJob(t *testing.T) {
	t.Parallel()
	collector := telemetrytest.NewCollector(t)
	tracing, err := telemetry.StartTracing("api", collector.URL, quietLogger())
	require.NoError(t, err)
	_, pool := poolOf(t, tracing.Provider)
	ctx := context.Background()
	useCase := newRecordingUseCase()
	kinds := jobs.NewKinds()
	jobs.Add(kinds, &lightWorker{useCase: useCase})
	jobs.Add(kinds, &heavyWorker{useCase: useCase})
	runner, err := jobs.NewRunner(pool, settingsOf(kinds, 10*time.Second, tracing.Provider, quietLogger()))
	require.NoError(t, err)
	saver, err := jobs.NewSaver(pool, tracing.Provider, quietLogger())
	require.NoError(t, err)

	// The jobs are saved inside the span of a request, as a use case does.
	requestCtx, request := tracing.Provider.Tracer("test").Start(ctx, "POST /api/proofs")
	err = database.NewTransactions(pool).Run(requestCtx, func(ctx context.Context) error {
		if err := saver.Save(ctx, lightArgs{FileID: 1}, nil); err != nil {
			return err
		}
		return saver.Save(ctx, heavyArgs{FileID: 2}, nil)
	})
	require.NoError(t, err)
	request.End()
	stop := runUntilSignalled(t, runner)
	called := []int64{receive(t, useCase.calls, "the first job"), receive(t, useCase.calls, "the second job")}
	require.NoError(t, stop())
	tracing.Stop()

	assert.ElementsMatch(t, []int64{1, 2}, called)
	requestTrace := request.SpanContext().TraceID().String()
	runs := slices.DeleteFunc(collector.Spans(), func(span telemetrytest.Span) bool {
		return !strings.HasPrefix(span.Name, "river.work/")
	})
	require.Len(t, runs, 2, "every job opened its span: %v", collector.Spans())
	for _, run := range runs {
		assert.NotEqual(t, requestTrace, run.TraceID, "the run of %s is its own trace, not a child of the request", run.Name)
		assert.Equal(t, "ok", run.Attributes["status"])
		// The link points at the span that saved the job, which River's
		// middleware opens under the span of the request.
		require.Len(t, run.Links, 1, "the run of %s links to one span", run.Name)
		assert.Equal(t, requestTrace, run.Links[0].TraceID, "the run of %s links to the trace of the request that saved it", run.Name)
		saving := spanWithID(collector.Spans(), run.Links[0].SpanID)
		require.NotNil(t, saving, "the span the run of %s links to was exported", run.Name)
		assert.Equal(t, request.SpanContext().SpanID().String(), saving.ParentSpanID, "the saving span %s is under the request", saving.Name)
	}
}

func spanWithID(spans []telemetrytest.Span, spanID string) *telemetrytest.Span {
	index := slices.IndexFunc(spans, func(span telemetrytest.Span) bool { return span.SpanID == spanID })
	if index < 0 {
		return nil
	}
	return &spans[index]
}

func TestTheShutdownTakesNoNewJobCutsTheRunningOneAtTheDeadlineAndBothRunAgainLater(t *testing.T) {
	t.Parallel()
	_, pool := poolOf(t, nil)
	ctx := context.Background()
	slow := newSlowWorker()
	kinds := jobs.NewKinds()
	jobs.Add(kinds, slow)
	runner, err := jobs.NewRunner(pool, settingsOf(kinds, 300*time.Millisecond, nil, quietLogger()))
	require.NoError(t, err)
	saver, err := jobs.NewSaver(pool, nil, quietLogger())
	require.NoError(t, err)
	// Two jobs of a queue that runs one at a time: the first runs, the second
	// waits.
	err = database.NewTransactions(pool).Run(ctx, func(ctx context.Context) error {
		if err := saver.Save(ctx, slowArgs{}, nil); err != nil {
			return err
		}
		return saver.Save(ctx, slowArgs{}, nil)
	})
	require.NoError(t, err)
	stop := runUntilSignalled(t, runner)
	receive(t, slow.started, "the first job")

	began := time.Now()
	err = stop()

	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(began), 300*time.Millisecond, "the running job had the deadline")
	assert.Less(t, time.Since(began), 10*time.Second, "the running job was cut at the deadline")
	assert.Empty(t, slow.started, "the waiting job was taken after the signal")
	rows, err := pool.Query(ctx, "SELECT state::text, attempt FROM river_job WHERE kind = $1 ORDER BY id", slowArgs{}.Kind())
	require.NoError(t, err)
	type job struct {
		state   string
		attempt int
	}
	var left []job
	for rows.Next() {
		var read job
		require.NoError(t, rows.Scan(&read.state, &read.attempt))
		left = append(left, read)
	}
	require.NoError(t, rows.Err())
	// River gives the cut job back its attempt and leaves both available.
	assert.Equal(t, []job{{state: "available", attempt: 0}, {state: "available", attempt: 0}}, left)

	// The next worker runs both.
	close(slow.release)
	next, err := jobs.NewRunner(pool, settingsOf(kinds, 10*time.Second, nil, quietLogger()))
	require.NoError(t, err)
	stopNext := runUntilSignalled(t, next)
	receive(t, slow.started, "the first job again")
	receive(t, slow.started, "the second job")
	require.Eventually(t, func() bool {
		var completed int
		require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM river_job WHERE kind = $1 AND state = 'completed'", slowArgs{}.Kind()).Scan(&completed))
		return completed == 2
	}, 30*time.Second, 50*time.Millisecond, "the two jobs completed on the next worker")
	require.NoError(t, stopNext())
}

// lockedBuffer collects the logs of a runner, which writes from its own
// goroutine while the test reads.
type lockedBuffer struct {
	mu   sync.Mutex
	text bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.String()
}

func TestTheRunnerStandsByUntilTheSignalWhenNoJobKindIsRegistered(t *testing.T) {
	t.Parallel()
	_, pool := poolOf(t, nil)
	logs := &lockedBuffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	runner, err := jobs.NewRunner(pool, settingsOf(jobs.NewKinds(), time.Second, nil, logger))
	require.NoError(t, err)
	stop := runUntilSignalled(t, runner)

	require.Eventually(t, func() bool { return strings.Contains(logs.String(), "no job kind") }, 10*time.Second, 10*time.Millisecond)
	assert.NoError(t, stop())
}
