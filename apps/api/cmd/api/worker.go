package main

import (
	"context"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

// workJobs runs the jobs until the shutdown notice, then stops without
// cutting a job short of its deadline, as the API does with its requests
// (Jobs and Shutdown in CODING_STANDARDS.md). It needs the database and
// nothing of the listening address of the API. Its logs are JSON lines on the
// error output, River's own included.
func workJobs(shutdown context.Context, environment map[string]string, _, stderr io.Writer) error {
	settings, err := config.Load(config.Worker, environment)
	if err != nil {
		return err
	}

	logger := telemetry.NewLogger(stderr, settings.LogLevel)
	if err := work(shutdown, settings, logger); err != nil {
		logger.Error("the worker stopped", slog.String("error", err.Error()))
		return errLogged
	}
	logger.Info("the worker stopped")
	return nil
}

// work wires the worker and runs it. It refuses to start on a schema that is
// not the one this binary expects: unlike the API, it has no readiness to
// tell, so a pending migration stops it at once.
//
// When it returns, the pool is closed and then the last spans are sent, in
// that order.
func work(shutdown context.Context, settings config.Config, logger *slog.Logger) error {
	tracing, pool, stop, err := tracedDatabase(settings, logger)
	if err != nil {
		return err
	}
	defer stop()
	if err := checkSchema(shutdown, pool, settings); err != nil {
		return err
	}

	runner, err := jobs.NewRunner(pool, jobs.RunnerSettings{
		LightQueueMaxJobs: settings.WorkerLightQueueMaxJobs,
		HeavyQueueMaxJobs: settings.WorkerHeavyQueueMaxJobs,
		ShutdownDeadline:  settings.WorkerShutdownDeadline,
		Kinds:             mountJobs(pool),
		PeriodicJobs:      nil,
		TracerProvider:    tracing.Provider,
		Logger:            logger,
	})
	if err != nil {
		return err
	}
	logger.Info("the worker started",
		slog.Int("light_queue_max_jobs", settings.WorkerLightQueueMaxJobs),
		slog.Int("heavy_queue_max_jobs", settings.WorkerHeavyQueueMaxJobs))
	return runner.Run(shutdown)
}

// mountJobs builds the job kinds the worker runs: it wires the job worker of
// each domain area to its use case and its adapters, by hand, as mountREST
// does for the REST operations. No domain area has a job yet: the e-mails and
// the cleanup of the storage bring the first ones.
func mountJobs(_ *pgxpool.Pool) *jobs.Kinds {
	return jobs.NewKinds()
}
