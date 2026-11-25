package main

import (
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/forgespark/apps/api/migrations"
)

// output collects what a command writes, and can be read while it runs.
type output struct {
	mu   sync.Mutex
	text bytes.Buffer
}

func (o *output) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text.Write(p)
}

func (o *output) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text.String()
}

// environmentOf gives the variables a command reads, pointed at a temporary
// database and at the file storage of development.
func environmentOf(temporary databasetest.Database) map[string]string {
	return withFileStorage(map[string]string{
		"API_LISTEN_ADDRESS":     "127.0.0.1:0",
		"APP_ORIGIN":             appOrigin(),
		"LOGIN_CODE_KEY":         theLoginCodeKey,
		"DATABASE_URL":           temporary.APIURL,
		"MIGRATION_DATABASE_URL": temporary.MigrationURL,
	})
}

// appOrigin is the origin of the app in development, as `task api:test`
// reads it from .env.example: http, so the cookie of the session travels
// over the plain HTTP of the tests.
func appOrigin() string {
	return os.Getenv("APP_ORIGIN")
}

// startAPI runs the command that serves the API until the test ends, and
// returns the address it listens on. The API must stop with status 0.
func startAPI(t *testing.T, environment map[string]string) string {
	t.Helper()
	api := launchAPI(t, environment)
	t.Cleanup(func() {
		api.signal()
		assert.Equal(t, 0, <-api.status, "the API stopped with: %s", api.logs)
	})
	return api.url
}

func statusOf(t *testing.T, url string) int {
	t.Helper()
	response, err := http.Get(url)
	require.NoError(t, err)
	defer response.Body.Close()
	return response.StatusCode
}

func TestTheAPIRefusesToStartWithARequiredVariableMissing(t *testing.T) {
	var stdout, stderr output

	environment := withFileStorage(map[string]string{
		"API_LISTEN_ADDRESS": "127.0.0.1:0",
		"APP_ORIGIN":         appOrigin(),
		"LOGIN_CODE_KEY":     theLoginCodeKey,
	})

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: DATABASE_URL is required\n", stderr.String())
	assert.Empty(t, stdout.String())
}

func TestTheAPIRefusesToStartWithAnInvalidVariableAndNeverShowsItsValue(t *testing.T) {
	var stdout, stderr output
	environment := withFileStorage(map[string]string{
		"API_LISTEN_ADDRESS": "127.0.0.1:0",
		"APP_ORIGIN":         appOrigin(),
		"LOGIN_CODE_KEY":     theLoginCodeKey,
		"DATABASE_URL":       "postgres://api:hunter2@localhost:5432/forgespark?sslmode=sometimes",
	})

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: DATABASE_URL is not valid: expected a PostgreSQL connection URL\n", stderr.String())
	assert.NotContains(t, stderr.String()+stdout.String(), "hunter2")
}

func TestTheAPIRefusesToStartWithoutTheFileStorage(t *testing.T) {
	var stdout, stderr output
	environment := map[string]string{
		"API_LISTEN_ADDRESS": "127.0.0.1:0",
		"APP_ORIGIN":         appOrigin(),
		"LOGIN_CODE_KEY":     theLoginCodeKey,
		"DATABASE_URL":       "postgres://api:api@localhost:5432/forgespark",
	}

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: FILE_STORAGE_URL is required; FILE_STORAGE_PUBLIC_URL is required; "+
		"FILE_STORAGE_REGION is required; FILE_STORAGE_BUCKET is required; FILE_STORAGE_ACCESS_KEY_ID is required; "+
		"FILE_STORAGE_SECRET_ACCESS_KEY is required\n", stderr.String())
}

func TestAnUnknownCommandPrintsTheCommandsAndFails(t *testing.T) {
	var stdout, stderr output

	status := run(context.Background(), []string{"reticulate"}, map[string]string{}, &stdout, &stderr)

	assert.Equal(t, 2, status)
	assert.Contains(t, stderr.String(), `api: unknown command "reticulate"`)
	assert.Contains(t, stderr.String(), "migrate")
}

func TestMigrateAppliesTheMigrationsAndASecondRunAppliesNothing(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	var first, second, stderr output

	firstStatus := run(context.Background(), []string{"migrate"}, environmentOf(temporary), &first, &stderr)
	secondStatus := run(context.Background(), []string{"migrate"}, environmentOf(temporary), &second, &stderr)

	// One line for each migration the binary carries, in the order of their
	// numbers.
	files, err := fs.Glob(migrations.Files, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	assert.Equal(t, "00001_grant_the_api_role_its_rights.sql", files[0])
	assert.Equal(t, 0, firstStatus, stderr.String())
	assert.Equal(t, "applied "+strings.Join(files, "\napplied ")+"\n", first.String())
	assert.Equal(t, 0, secondStatus, stderr.String())
	assert.Equal(t, "no migration to apply\n", second.String())
}

func TestMigrateFailsWhenPostgreSQLDoesNotAnswer(t *testing.T) {
	t.Parallel()
	var stdout, stderr output
	environment := map[string]string{
		// Port 1 of the machine has no PostgreSQL behind it.
		"MIGRATION_DATABASE_URL": "postgres://migration:hunter2@127.0.0.1:1/forgespark?sslmode=disable",
	}

	status := run(context.Background(), []string{"migrate"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stderr.String(), "api: apply the migrations")
	assert.NotContains(t, stderr.String(), "hunter2")
}

func TestTheAPIIsAliveAndReadyOnceTheMigrationsAreApplied(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)

	api := startAPI(t, environmentOf(temporary))

	assert.Equal(t, http.StatusOK, statusOf(t, api+"/livez"))
	assert.Equal(t, http.StatusOK, statusOf(t, api+"/readyz"))
}

func TestTheAPIIsAliveAndNotReadyWhileAMigrationIsPending(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)

	api := startAPI(t, environmentOf(temporary))

	assert.Equal(t, http.StatusOK, statusOf(t, api+"/livez"))
	assert.Equal(t, http.StatusServiceUnavailable, statusOf(t, api+"/readyz"))
}

func TestTheAPIIsAliveAndNotReadyWhilePostgreSQLIsDown(t *testing.T) {
	t.Parallel()
	environment := withFileStorage(map[string]string{
		"API_LISTEN_ADDRESS": "127.0.0.1:0",
		"APP_ORIGIN":         appOrigin(),
		"LOGIN_CODE_KEY":     theLoginCodeKey,
		// Port 1 of the machine has no PostgreSQL behind it.
		"DATABASE_URL": "postgres://api:api@127.0.0.1:1/forgespark?sslmode=disable",
	})

	api := startAPI(t, environment)

	assert.Equal(t, http.StatusOK, statusOf(t, api+"/livez"))
	assert.Equal(t, http.StatusServiceUnavailable, statusOf(t, api+"/readyz"))
}

func TestLivenessAndReadinessAreOutsideTheAPIPrefix(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)

	api := startAPI(t, environmentOf(temporary))

	assert.Equal(t, http.StatusNotFound, statusOf(t, api+"/api/livez"))
	assert.Equal(t, http.StatusNotFound, statusOf(t, api+"/api/readyz"))
}

func TestAskingForHelpPrintsTheCommandsAndSucceeds(t *testing.T) {
	var stdout, stderr output

	status := run(context.Background(), []string{"-h"}, map[string]string{}, &stdout, &stderr)

	assert.Equal(t, 0, status)
	assert.Contains(t, stderr.String(), "migrate")
}
