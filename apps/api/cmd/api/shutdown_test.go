package main

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// The API has no slow operation yet, so these tests make one: readiness asks
// PostgreSQL, and a proxy holds the answer back for as long as the test
// needs. The readiness check is given a minute, so that only the shutdown
// ends the request.

// apiBehindAStallablePostgres starts the API on a migrated database that the
// test can stall.
func apiBehindAStallablePostgres(t *testing.T, drainDeadline string) (servedAPI, *stallingPostgres) {
	t.Helper()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationURL, migrations.Files)
	require.NoError(t, err)
	postgres := stallablePostgres(t, temporary.APIURL)
	environment := environmentOf(temporary)
	environment["DATABASE_URL"] = postgres.url
	environment["HEALTH_CHECK_TIMEOUT"] = "1m"
	environment["SHUTDOWN_DRAIN_DEADLINE"] = drainDeadline
	return launchAPI(t, environment), postgres
}

func TestTheSignalLetsASlowRequestFinishAndTheAPIExitsWithStatusZero(t *testing.T) {
	t.Parallel()
	api, postgres := apiBehindAStallablePostgres(t, "1m")
	monitor := newCaller()
	require.Equal(t, http.StatusOK, (<-monitor.get(api.url + "/readyz")).status, "the API is ready before the signal")
	postgres.stall()
	slow := newCaller().get(api.url + "/readyz")
	postgres.requireHolding(t)

	api.signal()

	requireRefused(t, api.url)
	// The monitor has no way in but the connection it opened before.
	assert.Equal(t, answer{status: http.StatusOK}, <-monitor.get(api.url+"/livez"), "liveness keeps answering")
	assert.Equal(t, answer{status: http.StatusServiceUnavailable}, <-monitor.get(api.url+"/readyz"), "readiness fails from the signal on")
	select {
	case status := <-api.status:
		require.Fail(t, "the API exited with a request in flight", "status %d", status)
	default:
	}
	postgres.resume()
	assert.Equal(t, answer{status: http.StatusOK}, <-slow, "the slow request finished with its answer")
	assert.Equal(t, 0, <-api.status, "%s", api.logs)
	assert.Equal(t, []string{"the API started", "request failed", "the API stopped"}, api.messages(t))
	stopped := api.lines(t)[2]
	assert.Equal(t, "INFO", stopped["level"])
}

func TestAtTheDrainDeadlineTheRequestLeftIsCutAndTheAPIExitsWithAnError(t *testing.T) {
	t.Parallel()
	api, postgres := apiBehindAStallablePostgres(t, "200ms")
	require.Equal(t, http.StatusOK, (<-newCaller().get(api.url + "/readyz")).status)
	postgres.stall()
	slow := newCaller().get(api.url + "/readyz")
	postgres.requireHolding(t)

	api.signal()

	require.Eventually(t, func() bool {
		return strings.Contains(api.logs.String(), "requests cut by the shutdown")
	}, 10*time.Second, 10*time.Millisecond, "the drain deadline cut nothing: %s", api.logs)
	// The request was cut while PostgreSQL held its answer. Closing the pool
	// waits for PostgreSQL, so the test lets it answer again.
	postgres.resume()
	assert.Equal(t, 1, <-api.status, "%s", api.logs)
	assert.NotEqual(t, http.StatusOK, (<-slow).status, "the request was cancelled")
	lines := api.lines(t)
	cut := lineWith(lines, "requests cut by the shutdown")
	require.NotNil(t, cut, "%s", api.logs)
	assert.Equal(t, "ERROR", cut["level"])
	assert.EqualValues(t, 1, cut["count"])
	stopped := lineWith(lines, "the API stopped")
	require.NotNil(t, stopped, "%s", api.logs)
	assert.Equal(t, "ERROR", stopped["level"])
	assert.Equal(t, "requests cut by the shutdown: 1", stopped["error"])
}

// lineWith finds the line of the logs that carries the message.
func lineWith(lines []map[string]any, message string) map[string]any {
	index := slices.IndexFunc(lines, func(line map[string]any) bool { return line["msg"] == message })
	if index < 0 {
		return nil
	}
	return lines[index]
}
