package main

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
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
// database.
func environmentOf(temporary databasetest.Database) map[string]string {
	return map[string]string{
		"API_LISTEN_ADDRESS":     "127.0.0.1:0",
		"DATABASE_URL":           temporary.APIURL,
		"MIGRATION_DATABASE_URL": temporary.MigrationURL,
	}
}

// startAPI runs the command that serves the API until the test ends, and
// returns the address it listens on.
func startAPI(t *testing.T, environment map[string]string) string {
	t.Helper()
	ctx, stop := context.WithCancel(context.Background())
	var stdout, stderr output
	status := make(chan int, 1)
	go func() { status <- run(ctx, nil, environment, &stdout, &stderr) }()
	t.Cleanup(func() {
		stop()
		assert.Equal(t, 0, <-status, "the API stopped with: %s", stderr.String())
	})

	listening := regexp.MustCompile(`api: listening on (\S+)`)
	var address string
	require.Eventually(t, func() bool {
		found := listening.FindStringSubmatch(stderr.String())
		if found == nil {
			return false
		}
		address = found[1]
		return true
	}, 10*time.Second, 10*time.Millisecond, "the API did not start: %s", stderr.String())
	return "http://" + address
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

	status := run(context.Background(), nil, map[string]string{"API_LISTEN_ADDRESS": "127.0.0.1:0"}, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: DATABASE_URL is required\n", stderr.String())
	assert.Empty(t, stdout.String())
}

func TestTheAPIRefusesToStartWithAnInvalidVariableAndNeverShowsItsValue(t *testing.T) {
	var stdout, stderr output
	environment := map[string]string{
		"API_LISTEN_ADDRESS": "127.0.0.1:0",
		"DATABASE_URL":       "postgres://api:hunter2@localhost:5432/software_factory?sslmode=sometimes",
	}

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: DATABASE_URL is not valid: expected a PostgreSQL connection URL\n", stderr.String())
	assert.NotContains(t, stderr.String()+stdout.String(), "hunter2")
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

	assert.Equal(t, 0, firstStatus, stderr.String())
	assert.Equal(t, "applied 00001_grant_the_api_role_its_rights.sql\n", first.String())
	assert.Equal(t, 0, secondStatus, stderr.String())
	assert.Equal(t, "no migration to apply\n", second.String())
}

func TestMigrateFailsWhenPostgreSQLDoesNotAnswer(t *testing.T) {
	t.Parallel()
	var stdout, stderr output
	environment := map[string]string{
		// Port 1 of the machine has no PostgreSQL behind it.
		"MIGRATION_DATABASE_URL": "postgres://migration:hunter2@127.0.0.1:1/software_factory?sslmode=disable",
	}

	status := run(context.Background(), []string{"migrate"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stderr.String(), "api: apply the migrations")
	assert.NotContains(t, stderr.String(), "hunter2")
}

func TestTheAPIIsAliveAndReadyOnceTheMigrationsAreApplied(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationURL, migrations.Files)
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
	environment := map[string]string{
		"API_LISTEN_ADDRESS": "127.0.0.1:0",
		// Port 1 of the machine has no PostgreSQL behind it.
		"DATABASE_URL": "postgres://api:api@127.0.0.1:1/software_factory?sslmode=disable",
	}

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
