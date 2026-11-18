//go:build unix

package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// These tests run the API as a process of its own and send it the signals
// of the system, which Go can send to one process on Unix only. Windows has
// no SIGTERM to send: there the same behaviour is tested with the signal
// handed to the code, in shutdown_test.go and in platform/lifecycle, which
// run on every system.

// process is the API running as a process of its own.
type process struct {
	command *exec.Cmd
	url     string
	logs    *output
	// exited receives the exit status of the process.
	exited chan int
}

// runAsTheAPI is the variable that makes the test binary run as the API.
const runAsTheAPI = "RUN_THE_TEST_BINARY_AS_THE_API"

// TestMain lets the test binary stand in for the binary of the API: started
// again with the variable set, it runs main, signals and exit status
// included, instead of the tests.
func TestMain(m *testing.M) {
	if os.Getenv(runAsTheAPI) == "1" {
		main()
		return
	}
	os.Exit(m.Run())
}

// startProcess runs the API as a process of its own, with the environment,
// and waits until it listens.
func startProcess(t *testing.T, environment map[string]string) process {
	t.Helper()
	binary, err := os.Executable()
	require.NoError(t, err)
	command := exec.Command(binary)
	command.Env = []string{runAsTheAPI + "=1"}
	for name, value := range environment {
		command.Env = append(command.Env, name+"="+value)
	}
	logs := &output{}
	command.Stderr = logs
	require.NoError(t, command.Start())
	exited := make(chan int, 1)
	go func() {
		_ = command.Wait()
		exited <- command.ProcessState.ExitCode()
	}()
	t.Cleanup(func() { _ = command.Process.Kill() })

	started := regexp.MustCompile(`"msg":"the API started".*"address":"([^"]+)"`)
	var address string
	require.Eventually(t, func() bool {
		found := started.FindStringSubmatch(logs.String())
		if found == nil {
			return false
		}
		address = found[1]
		return true
	}, 30*time.Second, 10*time.Millisecond, "the process did not start: %s", logs)
	return process{command: command, url: "http://" + address, logs: logs, exited: exited}
}

func (p process) send(t *testing.T, signal os.Signal) {
	t.Helper()
	require.NoError(t, p.command.Process.Signal(signal))
}

// requireExit waits for the process to exit and returns its status.
func (p process) requireExit(t *testing.T, within time.Duration) int {
	t.Helper()
	select {
	case status := <-p.exited:
		return status
	case <-time.After(within):
		require.Fail(t, "the process did not exit", "%s", p.logs)
		return -1
	}
}

func TestTheProcessStopsOnSIGINTWithStatusZero(t *testing.T) {
	t.Parallel()
	api := startProcess(t, environmentOf(databasetest.New(t)))

	api.send(t, syscall.SIGINT)

	assert.Equal(t, 0, api.requireExit(t, 30*time.Second), "%s", api.logs)
	assert.Contains(t, api.logs.String(), `"msg":"the API stopped"`)
}

func TestTheProcessDrainsOnSIGTERMAndASecondSignalExitsItAtOnce(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)
	postgres := stallablePostgres(t, temporary.APIURL)
	environment := environmentOf(temporary)
	environment["DATABASE_URL"] = postgres.url
	environment["HEALTH_CHECK_TIMEOUT"] = "5m"
	environment["SHUTDOWN_DRAIN_DEADLINE"] = "5m"
	api := startProcess(t, environment)
	require.Equal(t, http.StatusOK, (<-newCaller().get(api.url + "/readyz")).status)
	postgres.stall()
	slow := newCaller().get(api.url + "/readyz")
	postgres.requireHolding(t)

	api.send(t, syscall.SIGTERM)

	requireRefused(t, api.url)
	select {
	case status := <-api.exited:
		require.Fail(t, "the process exited with a request in flight", "status %d: %s", status, api.logs)
	case <-time.After(500 * time.Millisecond):
	}

	api.send(t, syscall.SIGTERM)

	assert.NotZero(t, api.requireExit(t, 5*time.Second), "a second signal exits at once, with a status other than zero")
	assert.NotEqual(t, http.StatusOK, (<-slow).status, "the request in flight was dropped")
	assert.NotContains(t, api.logs.String(), `"msg":"the API stopped"`, "the process did not wait for the drain")
}
