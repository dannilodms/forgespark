package main

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
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

// standInStorage stands in for the file storage of the worker, which cleans
// it up as soon as it starts. The bucket of development is shared by the
// database of development and by every temporary one: a cleanup there with a
// temporary database would remove objects the database of development names.
// The stand-in lists the objects it was given, and removes each one asked
// for, or cuts the connection of each removal, as a storage that goes away
// does: the error of the request then carries its URL, with the key in it.
type standInStorage struct {
	mu       sync.Mutex
	objects  map[string]time.Time
	cut      bool
	removals chan string
}

func newStandInStorage(objects map[string]time.Time, cut bool) *standInStorage {
	return &standInStorage{objects: objects, cut: cut, removals: make(chan string, 16)}
}

// environment gives the variables of a file storage that is this stand-in.
func (s *standInStorage) environment(t *testing.T, environment map[string]string) map[string]string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(s.answer))
	t.Cleanup(server.Close)
	environment["FILE_STORAGE_URL"] = server.URL
	environment["FILE_STORAGE_PUBLIC_URL"] = server.URL
	environment["FILE_STORAGE_REGION"] = "garage"
	environment["FILE_STORAGE_BUCKET"] = "files"
	environment["FILE_STORAGE_ACCESS_KEY_ID"] = "GK0123"
	environment["FILE_STORAGE_SECRET_ACCESS_KEY"] = "a-secret-key-of-the-stand-in"
	return environment
}

func (s *standInStorage) answer(answer http.ResponseWriter, request *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	answer.Header().Set("Content-Type", "application/xml")
	if request.Method == http.MethodDelete {
		s.remove(answer, strings.TrimPrefix(request.URL.Path, "/files/"))
		return
	}
	contents := ""
	for _, key := range slices.Sorted(maps.Keys(s.objects)) {
		contents += fmt.Sprintf("<Contents><Key>%s</Key><LastModified>%s</LastModified><Size>64</Size></Contents>",
			key, s.objects[key].UTC().Format("2006-01-02T15:04:05.000Z"))
	}
	_, _ = fmt.Fprintf(answer, `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>files</Name><KeyCount>%d</KeyCount>
<MaxKeys>1000</MaxKeys><IsTruncated>false</IsTruncated>%s</ListBucketResult>`, len(s.objects), contents)
}

func (s *standInStorage) remove(answer http.ResponseWriter, key string) {
	// The SDK and River try again: the test reads the first removal only.
	select {
	case s.removals <- key:
	default:
	}
	if s.cut {
		cutConnection(answer)
		return
	}
	delete(s.objects, key)
	answer.WriteHeader(http.StatusNoContent)
}

// cutConnection closes the connection of the request without an answer.
func cutConnection(answer http.ResponseWriter) {
	connection, _, err := http.NewResponseController(answer).Hijack()
	if err != nil {
		return
	}
	_ = connection.Close()
}

// workerEnvironmentOf gives the worker the variables it needs: the database of
// the API role, the sender of e-mail and a stand-in of the file storage with
// nothing in it, and nothing of the listening address.
func workerEnvironmentOf(t *testing.T, temporary databasetest.Database) map[string]string {
	t.Helper()
	return newStandInStorage(map[string]time.Time{}, false).environment(t, withEmailSender(map[string]string{
		"DATABASE_URL":             temporary.APIURL,
		"WORKER_SHUTDOWN_DEADLINE": "2s",
	}))
}

// withEmailSender adds what the delivery of e-mail needs: the sender, the
// Mailpit of development at the address the tests receive in
// EMAIL_SMTP_ADDRESS, and theLoginCodeKey, which the content of the e-mail of
// a login code is computed with.
func withEmailSender(environment map[string]string) map[string]string {
	environment["EMAIL_SENDER"] = "smtp"
	environment["EMAIL_SMTP_ADDRESS"] = os.Getenv("EMAIL_SMTP_ADDRESS")
	environment["EMAIL_FROM_ADDRESS"] = "no-reply@fabrica.example"
	environment["EMAIL_FROM_NAME"] = "Fábrica Exemplo"
	environment["LOGIN_CODE_KEY"] = theLoginCodeKey
	return environment
}

// theLoginCodeKey is the key of the login code of the tests: the bytes 0 to
// 31, in hexadecimal.
const theLoginCodeKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func TestTheWorkerRunsOnTheDatabaseTheSenderOfEmailAndTheFileStorageAndStopsOnTheSignalWithStatusZero(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)
	worker := launchWorker(t, workerEnvironmentOf(t, temporary))

	worker.signal()

	assert.Equal(t, 0, <-worker.status, "%s", worker.logs)
	lines := worker.logs.String()
	assert.Contains(t, lines, `"msg":"the worker stopped"`)
	for line := range strings.Lines(lines) {
		assert.True(t, strings.HasPrefix(line, "{"), "a line of the logs is not JSON: %s", line)
	}
}

// The cleanup of the storage runs when the worker starts, and then every
// hour. It removes the object more than a day old that no file names, and
// keeps the younger one.
func TestTheWorkerCleansUpTheStorageWhenItStarts(t *testing.T) {
	t.Parallel()
	const old, young = "0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f", "1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e"
	storage := newStandInStorage(map[string]time.Time{
		old:   time.Now().Add(-25 * time.Hour),
		young: time.Now().Add(-time.Hour),
	}, false)
	worker := launchWorker(t, storage.environment(t, withEmailSender(map[string]string{"DATABASE_URL": databasetest.Migrated(t).APIURL})))

	removed := receiveWithin(t, storage.removals, 30*time.Second, "the removal of the old object")
	worker.signal()

	assert.Equal(t, 0, <-worker.status, "%s", worker.logs)
	assert.Equal(t, old, removed)
	assert.Empty(t, storage.removals, "the young object was removed")
}

func receiveWithin[T any](t *testing.T, from <-chan T, wait time.Duration, what string) T {
	t.Helper()
	select {
	case value := <-from:
		return value
	case <-time.After(wait):
		t.Fatalf("%s did not happen within %s", what, wait)
		var zero T
		return zero
	}
}

func TestTheWorkerRefusesToStartWithoutTheDatabaseTheSenderOfEmailOrTheFileStorage(t *testing.T) {
	var stdout, stderr output

	status := run(context.Background(), []string{"worker"}, map[string]string{}, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: DATABASE_URL is required; "+
		"FILE_STORAGE_URL is required; FILE_STORAGE_PUBLIC_URL is required; FILE_STORAGE_REGION is required; "+
		"FILE_STORAGE_BUCKET is required; FILE_STORAGE_ACCESS_KEY_ID is required; FILE_STORAGE_SECRET_ACCESS_KEY is required; "+
		"LOGIN_CODE_KEY is required; EMAIL_SENDER is required; EMAIL_FROM_ADDRESS is required; EMAIL_FROM_NAME is required\n",
		stderr.String())
}

// withResend makes the worker send through Resend, with a key that no
// account has: the worker calls Resend only when it delivers an e-mail.
func withResend(environment map[string]string) map[string]string {
	environment["EMAIL_SENDER"] = "resend"
	delete(environment, "EMAIL_SMTP_ADDRESS")
	environment["RESEND_API_KEY"] = "re_no_account_has_this_key"
	return environment
}

func TestTheWorkerStartsWithResendAsTheSenderAndNoSMTPServer(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)
	worker := launchWorker(t, withResend(workerEnvironmentOf(t, temporary)))

	worker.signal()

	assert.Equal(t, 0, <-worker.status, "%s", worker.logs)
	assert.NotContains(t, worker.logs.String(), "re_no_account_has_this_key")
}

func TestTheWorkerRefusesToSendThroughResendWithoutItsKey(t *testing.T) {
	var stdout, stderr output
	environment := newStandInStorage(map[string]time.Time{}, false).environment(t, withResend(withEmailSender(map[string]string{
		"DATABASE_URL": "postgres://api:api@localhost:5432/software_factory",
	})))
	delete(environment, "RESEND_API_KEY")

	status := run(context.Background(), []string{"worker"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: RESEND_API_KEY is required\n", stderr.String())
}

func TestTheWorkerRefusesToStartWithAnInvalidQueueLimit(t *testing.T) {
	var stdout, stderr output
	environment := newStandInStorage(map[string]time.Time{}, false).environment(t, withEmailSender(map[string]string{
		"DATABASE_URL":                "postgres://api:api@localhost:5432/software_factory",
		"WORKER_HEAVY_QUEUE_MAX_JOBS": "0",
	}))

	status := run(context.Background(), []string{"worker"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: WORKER_HEAVY_QUEUE_MAX_JOBS is not valid: expected at least 1\n", stderr.String())
}

func TestTheWorkerStopsAtOnceWhileAMigrationIsPending(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	var stdout, logs output

	status := run(context.Background(), []string{"worker"}, workerEnvironmentOf(t, temporary), &stdout, &logs)

	assert.Equal(t, 1, status)
	assert.Contains(t, logs.String(), `"msg":"the worker stopped"`)
	assert.Contains(t, logs.String(), "this binary expects version")
	assert.Contains(t, logs.String(), "Apply the migrations with the migrate command first")
}

func TestTheWorkerStopsAtOnceWhilePostgreSQLIsDown(t *testing.T) {
	t.Parallel()
	var stdout, logs output
	environment := newStandInStorage(map[string]time.Time{}, false).environment(t, withEmailSender(map[string]string{
		// Port 1 of the machine has no PostgreSQL behind it.
		"DATABASE_URL": "postgres://api:hunter2@127.0.0.1:1/software_factory?sslmode=disable",
	}))

	status := run(context.Background(), []string{"worker"}, environment, &stdout, &logs)

	assert.Equal(t, 1, status)
	assert.Contains(t, logs.String(), `"msg":"the worker stopped"`)
	assert.Contains(t, logs.String(), "PostgreSQL does not answer")
	assert.NotContains(t, logs.String(), "hunter2")
}
