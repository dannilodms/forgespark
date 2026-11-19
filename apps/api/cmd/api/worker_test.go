package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// runningWorker is the worker command, running as the process would, until
// the test tells it to stop.
type runningWorker struct {
	signal context.CancelFunc
	status <-chan int
	logs   *output
}

// launchWorker runs the worker command and waits until it says it started.
func launchWorker(t *testing.T, environment map[string]string) runningWorker {
	t.Helper()
	shutdown, signal := context.WithCancel(context.Background())
	var stdout output
	logs := &output{}
	status := make(chan int, 1)
	go func() { status <- run(shutdown, []string{"worker"}, environment, &stdout, logs) }()
	t.Cleanup(signal)

	require.Eventually(t, func() bool {
		return strings.Contains(logs.String(), `"msg":"the worker started"`)
	}, 20*time.Second, 10*time.Millisecond, "the worker did not start: %s", logs)
	return runningWorker{signal: signal, status: status, logs: logs}
}

// workerEnvironmentOf gives the worker the variables it needs: the database of
// the API role and the sender of e-mail, and nothing of the listening address
// or the file storage.
func workerEnvironmentOf(temporary databasetest.Database) map[string]string {
	return withEmailSender(map[string]string{
		"DATABASE_URL":             temporary.APIURL,
		"WORKER_SHUTDOWN_DEADLINE": "2s",
	})
}

// withEmailSender adds the sender of e-mail: the Mailpit of development, at
// the address the tests receive in EMAIL_SMTP_ADDRESS.
func withEmailSender(environment map[string]string) map[string]string {
	environment["EMAIL_SENDER"] = "smtp"
	environment["EMAIL_SMTP_ADDRESS"] = os.Getenv("EMAIL_SMTP_ADDRESS")
	environment["EMAIL_FROM_ADDRESS"] = "no-reply@fabrica.example"
	environment["EMAIL_FROM_NAME"] = "Fábrica Exemplo"
	return environment
}

func TestTheWorkerRunsOnTheDatabaseAloneAndStopsOnTheSignalWithStatusZero(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)
	worker := launchWorker(t, workerEnvironmentOf(temporary))

	worker.signal()

	assert.Equal(t, 0, <-worker.status, "%s", worker.logs)
	lines := worker.logs.String()
	assert.Contains(t, lines, `"msg":"the worker stopped"`)
	for line := range strings.Lines(lines) {
		assert.True(t, strings.HasPrefix(line, "{"), "a line of the logs is not JSON: %s", line)
	}
}

func TestTheWorkerRefusesToStartWithoutTheDatabaseOrTheSenderOfEmail(t *testing.T) {
	var stdout, stderr output

	status := run(context.Background(), []string{"worker"}, map[string]string{}, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: DATABASE_URL is required; EMAIL_SENDER is required; EMAIL_SMTP_ADDRESS is required; "+
		"EMAIL_FROM_ADDRESS is required; EMAIL_FROM_NAME is required\n", stderr.String())
}

func TestTheWorkerRefusesToStartWithAnInvalidQueueLimit(t *testing.T) {
	var stdout, stderr output
	environment := withEmailSender(map[string]string{
		"DATABASE_URL":                "postgres://api:api@localhost:5432/software_factory",
		"WORKER_HEAVY_QUEUE_MAX_JOBS": "0",
	})

	status := run(context.Background(), []string{"worker"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: WORKER_HEAVY_QUEUE_MAX_JOBS is not valid: expected at least 1\n", stderr.String())
}

func TestTheWorkerStopsAtOnceWhileAMigrationIsPending(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	var stdout, logs output

	status := run(context.Background(), []string{"worker"}, workerEnvironmentOf(temporary), &stdout, &logs)

	assert.Equal(t, 1, status)
	assert.Contains(t, logs.String(), `"msg":"the worker stopped"`)
	assert.Contains(t, logs.String(), "this binary expects version")
	assert.Contains(t, logs.String(), "Apply the migrations with the migrate command first")
}

func TestTheWorkerStopsAtOnceWhilePostgreSQLIsDown(t *testing.T) {
	t.Parallel()
	var stdout, logs output
	environment := withEmailSender(map[string]string{
		// Port 1 of the machine has no PostgreSQL behind it.
		"DATABASE_URL": "postgres://api:hunter2@127.0.0.1:1/software_factory?sslmode=disable",
	})

	status := run(context.Background(), []string{"worker"}, environment, &stdout, &logs)

	assert.Equal(t, 1, status)
	assert.Contains(t, logs.String(), `"msg":"the worker stopped"`)
	assert.Contains(t, logs.String(), "PostgreSQL does not answer")
	assert.NotContains(t, logs.String(), "hunter2")
}
