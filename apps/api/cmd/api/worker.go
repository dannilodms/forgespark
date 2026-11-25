package main

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	accesspostgres "github.com/dannilodms/forgespark/apps/api/internal/access/adapters/postgres"
	accessriver "github.com/dannilodms/forgespark/apps/api/internal/access/adapters/river"
	accessapplication "github.com/dannilodms/forgespark/apps/api/internal/access/application"
	emailpostgres "github.com/dannilodms/forgespark/apps/api/internal/email/adapters/postgres"
	emailresend "github.com/dannilodms/forgespark/apps/api/internal/email/adapters/resend"
	emailriver "github.com/dannilodms/forgespark/apps/api/internal/email/adapters/river"
	emailsmtp "github.com/dannilodms/forgespark/apps/api/internal/email/adapters/smtp"
	emailapplication "github.com/dannilodms/forgespark/apps/api/internal/email/application"
	filespostgres "github.com/dannilodms/forgespark/apps/api/internal/files/adapters/postgres"
	filesriver "github.com/dannilodms/forgespark/apps/api/internal/files/adapters/river"
	filesapplication "github.com/dannilodms/forgespark/apps/api/internal/files/application"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/clock"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/jobs"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
)

// cleanUpInterval is how often the worker cleans up the file storage. It also
// does when it starts.
const cleanUpInterval = time.Hour

// dailyCleanUpInterval is how often the worker removes the records the login
// and the e-mail keep only for a while: once a day, so the database is
// reached for it no more often (the maintainer may move to a database billed
// by use). It also does when it starts.
const dailyCleanUpInterval = 24 * time.Hour

// workJobs runs the jobs until the shutdown notice, then stops without
// cutting a job short of its deadline, as the API does with its requests
// (Jobs and Shutdown in CODING_STANDARDS.md). It needs the database and the
// file storage, and nothing of the listening address of the API. Its logs are
// JSON lines on the error output, River's own included.
func workJobs(shutdown context.Context, environment map[string]string, _, stderr io.Writer) error {
	settings, err := config.Load(config.Worker, environment)
	if err != nil {
		return err
	}

	logger := telemetry.NewLogger(stderr, settings.LogLevel)
	telemetry.RouteTheStandardLogger(logger)
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

	kinds, err := mountJobs(pool, settings, tracing.Provider)
	if err != nil {
		return err
	}
	runner, err := jobs.NewRunner(pool, jobs.RunnerSettings{
		LightQueueMaxJobs: settings.WorkerLightQueueMaxJobs,
		HeavyQueueMaxJobs: settings.WorkerHeavyQueueMaxJobs,
		ShutdownDeadline:  settings.WorkerShutdownDeadline,
		Kinds:             kinds,
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

// mountJobs builds the job kinds the worker runs, the periodic ones among
// them: it wires the job worker of each domain area to its use case and its
// adapters, by hand, as mountREST does for the REST operations. A new kind is
// one more line.
func mountJobs(pool *pgxpool.Pool, settings config.Config, tracerProvider trace.TracerProvider) (*jobs.Kinds, error) {
	// The area that asks for the e-mails of a template gives their content:
	// the access area computes the code of a login code.
	loginCodeKey, err := loginCodeKeyOf(settings)
	if err != nil {
		return nil, err
	}
	loginCodes := accessapplication.NewLoginCodeContents(accesspostgres.NewLoginCodeRequestRepository(pool), loginCodeKey)
	contents := emailapplication.ContentSources{emailapplication.LoginCode: loginCodes}

	kinds := jobs.NewKinds()
	jobs.Add(kinds, emailriver.NewDeliveryWorker(deliverEmail(pool, emailSender(settings, tracerProvider), contents)))
	jobs.AddPeriodic(kinds, filesriver.NewCleanUpStorageWorker(cleanUpStorage(pool, settings, tracerProvider)), cleanUpInterval)
	jobs.AddPeriodic(kinds, accessriver.NewCleanUpOldRecordsWorker(cleanUpOldRecords(pool)), dailyCleanUpInterval)
	return kinds, nil
}

// cleanUpOldRecords wires the daily cleanup: the records of the login and the
// sessions in PostgreSQL, and the bounce notices, which the e-mail area
// removes.
func cleanUpOldRecords(pool *pgxpool.Pool) *accessapplication.CleanUpOldRecords {
	return accessapplication.NewCleanUpOldRecords(accessapplication.OldRecords{
		Requests:             accesspostgres.NewLoginCodeRequestRepository(pool),
		Asks:                 accesspostgres.NewAskedAddressRepository(pool),
		EndedSessions:        accesspostgres.NewEndedSessions(pool),
		CleanUpBounceNotices: emailapplication.NewCleanUpBounceNotices(emailpostgres.NewBounceNotices(pool), clock.System{}),
		Clock:                clock.System{},
	})
}

// cleanUpStorage wires the cleanup of the file storage: the files in
// PostgreSQL and the storage the configuration names.
func cleanUpStorage(pool *pgxpool.Pool, settings config.Config, tracerProvider trace.TracerProvider) *filesapplication.CleanUpStorage {
	return filesapplication.NewCleanUpStorage(
		database.NewTransactions(pool),
		filespostgres.NewFileRepository(pool),
		fileStorageOf(settings, tracerProvider),
		clock.System{},
	)
}

// deliverEmail wires the delivery of an e-mail: its record in PostgreSQL, the
// sender the configuration chose and the content of each template.
func deliverEmail(pool *pgxpool.Pool, sender emailapplication.Sender, contents emailapplication.ContentSources) *emailapplication.DeliverEmail {
	return emailapplication.NewDeliverEmail(
		database.NewTransactions(pool),
		emailpostgres.NewEmailRepository(pool),
		contents,
		sender,
		clock.System{},
	)
}

// emailSender is the sender EMAIL_SENDER names: the API of Resend, or an
// SMTP server.
func emailSender(settings config.Config, tracerProvider trace.TracerProvider) emailapplication.Sender {
	if settings.EmailSender == config.ResendSender {
		return emailresend.NewSender(emailresend.Settings{
			URL:            emailresend.URL,
			APIKey:         settings.ResendAPIKey,
			FromAddress:    settings.EmailFromAddress.String(),
			FromName:       settings.EmailFromName,
			TracerProvider: tracerProvider,
		})
	}
	return emailsmtp.NewSender(emailsmtp.Settings{
		Host:           settings.EmailSMTPAddress.Host(),
		Port:           settings.EmailSMTPAddress.Port(),
		FromAddress:    settings.EmailFromAddress.String(),
		FromName:       settings.EmailFromName,
		TracerProvider: tracerProvider,
	})
}
