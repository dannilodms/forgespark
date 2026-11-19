// Package jobs holds what every domain area shares to run work that nobody
// waits for on a screen: River, which keeps the jobs in PostgreSQL, the two
// queues, the saving of a job inside the transaction of a use case, and the
// runner that the `worker` command hands the shutdown notice to. Rules: Jobs
// in CODING_STANDARDS.md; the reasons are in ADR-0016.
//
// A domain area meets River in its `adapters/river` package. Outbound, it
// saves a job: the adapter implements the interface its `application`
// declares by calling Saver.Save with the job's arguments, which carry
// identifiers only. Inbound, it runs one: a job worker is a river.Worker
// whose Work reads the identifiers from the job's arguments and calls one use
// case, and the `worker` command adds it to the Kinds the Runner runs, with
// Add, or with AddPeriodic for a job River saves by itself at an interval.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/riverqueue/rivercontrib/otelriver"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

const (
	// QueueLight is the queue of light work, such as sending an e-mail. A job
	// names its queue in its InsertOpts.
	QueueLight = "light"
	// QueueHeavy is the queue of heavy work, such as converting a file, so
	// that it never starves the light queue.
	QueueHeavy = "heavy"

	// maxAttempts is how many times a job runs before River discards it:
	// about six minutes with River's retry policy. A job kind that needs
	// more declares its own in its InsertOpts.
	maxAttempts = 5

	// returnGrace is how long the jobs whose context was cancelled at the
	// shutdown deadline have to return, before the stop gives up on them. A
	// job that respects its context returns at once.
	returnGrace = 15 * time.Second
)

// ErrNoTransaction is returned by Save when the context carries no
// transaction: a job is saved with the business change that asked for it,
// never on its own.
var ErrNoTransaction = errors.New("a job is saved inside the transaction of a use case")

// Kinds is the set of job kinds a worker runs, and the periodic jobs among
// them. The `worker` command builds it and each domain area adds its job
// workers to it with Add, or with AddPeriodic.
type Kinds struct {
	workers  *river.Workers
	periodic []*river.PeriodicJob
}

// NewKinds returns an empty set of job kinds.
func NewKinds() *Kinds {
	return &Kinds{workers: river.NewWorkers(), periodic: nil}
}

// Add registers a job worker for the kind of its arguments. It panics when
// the kind is already registered, as a wiring mistake should.
func Add[T river.JobArgs](kinds *Kinds, worker river.Worker[T]) {
	river.AddWorker(kinds.workers, worker)
}

// AddPeriodic registers a job worker for the kind of its arguments, and has
// River save a job of that kind when the worker starts and then once every
// interval. The arguments of the job are the zero value of their type: a
// periodic job names no record.
//
// River keeps the schedule in the memory of the worker it elects, so a
// restart or a new election starts it over, and a run can be missed. The run
// at the start bounds the wait: a worker restarted more often than the
// interval still runs the job. A periodic job tolerates a missed run and a
// run more than it needed (Jobs in CODING_STANDARDS.md).
func AddPeriodic[T river.JobArgs](kinds *Kinds, worker river.Worker[T], interval time.Duration) {
	Add(kinds, worker)
	kinds.periodic = append(kinds.periodic, river.NewPeriodicJob(
		river.PeriodicInterval(interval),
		func() (river.JobArgs, *river.InsertOpts) {
			var arguments T
			return arguments, nil
		},
		&river.PeriodicJobOpts{ID: "", RunOnStart: true},
	))
}

// Saver saves jobs for the API, inside the transaction of the use case that
// asks for one. The job is visible to the worker once that transaction
// commits, and never when it rolls back.
type Saver struct {
	client *river.Client[pgx.Tx]
}

// NewSaver builds the saver of a process. It runs nothing: the API saves jobs
// and never runs one.
func NewSaver(pool *pgxpool.Pool, tracerProvider trace.TracerProvider, logger *slog.Logger) (*Saver, error) {
	client, err := river.NewClient(riverpgxv5.New(pool), clientConfig(tracerProvider, logger))
	if err != nil {
		return nil, fmt.Errorf("prepare the saving of jobs: %w", err)
	}
	return &Saver{client: client}, nil
}

// Save saves a job of these arguments in the transaction the context carries,
// and fails with ErrNoTransaction when it carries none. The options may be
// nil: the queue, the attempts and the timeout then come from the arguments'
// InsertOpts, or from River's defaults.
func (s *Saver) Save(ctx context.Context, arguments river.JobArgs, options *river.InsertOpts) error {
	transaction, open := database.TransactionOf(ctx)
	if !open {
		return ErrNoTransaction
	}
	if _, err := s.client.InsertTx(ctx, transaction, arguments, options); err != nil {
		return fmt.Errorf("save the job %q: %w", arguments.Kind(), err)
	}
	return nil
}

// RunnerSettings is what the runner takes from configuration and from the
// command that builds it.
type RunnerSettings struct {
	// LightQueueMaxJobs and HeavyQueueMaxJobs are the most jobs each queue
	// runs at once in this process.
	LightQueueMaxJobs int
	HeavyQueueMaxJobs int
	// ShutdownDeadline is how long the running jobs have to finish once the
	// shutdown starts, before their contexts are cancelled.
	ShutdownDeadline time.Duration
	// Kinds are the job kinds this runner runs, and the periodic jobs River
	// schedules by itself, on the one runner it elects as leader.
	Kinds *Kinds
	// TracerProvider traces the saving and the running of a job. With none,
	// nothing is traced.
	TracerProvider trace.TracerProvider
	// Logger receives River's own lines beside the runner's.
	Logger *slog.Logger
}

// Runner runs the jobs of the two queues until the shutdown notice, as the
// `worker` command does.
type Runner struct {
	client   *river.Client[pgx.Tx]
	settings RunnerSettings
}

// NewRunner builds the runner of the `worker` command. River opens no
// connection until Run.
func NewRunner(pool *pgxpool.Pool, settings RunnerSettings) (*Runner, error) {
	config := clientConfig(settings.TracerProvider, settings.Logger)
	config.Queues = map[string]river.QueueConfig{
		QueueLight: {MaxWorkers: settings.LightQueueMaxJobs},
		QueueHeavy: {MaxWorkers: settings.HeavyQueueMaxJobs},
	}
	config.Workers = settings.Kinds.workers
	config.PeriodicJobs = settings.Kinds.periodic
	config.SoftStopTimeout = settings.ShutdownDeadline
	client, err := river.NewClient(riverpgxv5.New(pool), config)
	if err != nil {
		return nil, fmt.Errorf("prepare the running of jobs: %w", err)
	}
	return &Runner{client: client, settings: settings}, nil
}

// Run runs jobs until the shutdown notice ends, then stops as Shutdown in
// CODING_STANDARDS.md says: it takes no new job, gives the running ones the
// shutdown deadline, then cancels their contexts and waits for them to
// return. A job cut this way is made available again by River, with the
// attempt given back, so it runs again on the next worker. Run returns nil
// once River has stopped, and an error when it could not start, as with no
// job kind to run, or a cut job did not return within returnGrace.
func (r *Runner) Run(shutdown context.Context) error {
	// The jobs outlive the notice: River cancels every job at once when the
	// context of Start ends, and the deadline is what gives them time.
	if err := r.client.Start(context.WithoutCancel(shutdown)); err != nil {
		return fmt.Errorf("start the running of jobs: %w", err)
	}
	<-shutdown.Done()

	stop, release := context.WithTimeout(context.Background(), r.settings.ShutdownDeadline+returnGrace)
	defer release()
	if err := r.client.Stop(stop); err != nil {
		return fmt.Errorf("stop the running of jobs: %w", err)
	}
	return nil
}

// clientConfig is what the saver and the runner share: the attempts of a job,
// the logger and the tracing. River's middleware opens the span of each job
// and, on the saving side, writes the trace of the request into the job's
// metadata, so that the span of the run links to the span that saved it.
func clientConfig(tracerProvider trace.TracerProvider, logger *slog.Logger) *river.Config {
	return &river.Config{
		Logger:      logger,
		MaxAttempts: maxAttempts,
		Middleware: []rivertype.Middleware{
			otelriver.NewMiddleware(&otelriver.MiddlewareConfig{
				EnableTracePropagation:      true,
				EnableWorkSpanJobKindSuffix: true,
				TracerProvider:              tracerProvider,
			}),
		},
	}
}
