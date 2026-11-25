package lifecycle_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/lifecycle"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
)

// The streaming part of the shutdown test (a stream that receives its final
// event) waits for the first stream: see Shutdown in CODING_STANDARDS.md.

// logs collects what the server writes, and can be read while it runs.
type logs struct {
	mu      sync.Mutex
	written bytes.Buffer
}

func (l *logs) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.written.Write(p)
}

// lines reads what was written as one JSON object for each line.
func (l *logs) lines(t *testing.T) []map[string]any {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	lines := []map[string]any{}
	for line := range strings.Lines(l.written.String()) {
		var object map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &object), "not a JSON object: %s", line)
		lines = append(lines, object)
	}
	return lines
}

// server is a handler served by the lifecycle, as the API is.
type server struct {
	url string
	// signal starts the shutdown, as the first SIGTERM does.
	signal context.CancelFunc
	// stopped receives what Serve returned.
	stopped chan error
	logs    *logs
}

func start(t *testing.T, handler http.Handler, drainDeadline time.Duration) server {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	shutdown, signal := context.WithCancel(context.Background())
	written := &logs{}
	stopped := make(chan error, 1)
	go func() {
		stopped <- lifecycle.Serve(shutdown, listener, handler, drainDeadline, telemetry.NewLogger(written, slog.LevelDebug))
	}()
	t.Cleanup(signal)
	return server{url: "http://" + listener.Addr().String(), signal: signal, stopped: stopped, logs: written}
}

// answer is what a client received.
type answer struct {
	status int
	body   string
	err    error
}

// get sends one request on a connection of its own and gives the answer when
// it arrives.
func get(url string) <-chan answer {
	return getWith(&http.Client{Transport: &http.Transport{}}, url)
}

func getWith(client *http.Client, url string) <-chan answer {
	answered := make(chan answer, 1)
	go func() {
		response, err := client.Get(url)
		if err != nil {
			answered <- answer{err: err}
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		answered <- answer{status: response.StatusCode, body: string(body), err: err}
	}()
	return answered
}

// requireRefused waits until the server takes no new connection.
func requireRefused(t *testing.T, url string) {
	t.Helper()
	address := strings.TrimPrefix(url, "http://")
	require.Eventually(t, func() bool {
		connection, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			return true
		}
		_ = connection.Close()
		return false
	}, 20*time.Second, 10*time.Millisecond, "the server still takes a new connection")
}

// slow is a request that stays in flight until the test lets it finish.
type slow struct {
	entered chan struct{}
	release chan struct{}
}

func newSlow() slow {
	return slow{entered: make(chan struct{}, 8), release: make(chan struct{})}
}

func (s slow) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.entered <- struct{}{}
	select {
	case <-s.release:
		_, _ = io.WriteString(w, "finished")
	case <-r.Context().Done():
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func TestTheSignalLetsARequestInFlightFinishAndRefusesANewConnection(t *testing.T) {
	t.Parallel()
	request := newSlow()
	api := start(t, request, 30*time.Second)
	answered := get(api.url)
	<-request.entered

	api.signal()

	requireRefused(t, api.url)
	select {
	case err := <-api.stopped:
		require.Fail(t, "the server stopped with a request in flight", "%v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(request.release)
	assert.Equal(t, answer{status: http.StatusOK, body: "finished"}, <-answered)
	assert.NoError(t, <-api.stopped, "every request finished")
	assert.Empty(t, api.logs.lines(t), "no request was cut")
}

func TestAConnectionAlreadyOpenIsStillAnsweredWhileTheDrainLasts(t *testing.T) {
	t.Parallel()
	request := newSlow()
	mux := http.NewServeMux()
	mux.Handle("GET /slow", request)
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	api := start(t, mux, 30*time.Second)
	monitor := &http.Client{Transport: &http.Transport{}}
	require.Equal(t, http.StatusOK, (<-getWith(monitor, api.url+"/livez")).status)
	answered := get(api.url + "/slow")
	<-request.entered

	api.signal()

	requireRefused(t, api.url)
	// The monitor has no way in but the connection it opened before.
	assert.Equal(t, answer{status: http.StatusOK}, <-getWith(monitor, api.url+"/livez"))
	close(request.release)
	assert.Equal(t, http.StatusOK, (<-answered).status)
	assert.NoError(t, <-api.stopped)
}

func TestAReachedDrainDeadlineCancelsWhatIsLeftAndLogsHowMany(t *testing.T) {
	t.Parallel()
	request := newSlow()
	api := start(t, request, 200*time.Millisecond)
	first, second := get(api.url), get(api.url)
	<-request.entered
	<-request.entered

	api.signal()

	assert.Error(t, <-api.stopped, "a request was cut")
	assert.NotEqual(t, http.StatusOK, (<-first).status, "the context of the request was cancelled")
	assert.NotEqual(t, http.StatusOK, (<-second).status, "the context of the request was cancelled")
	lines := api.logs.lines(t)
	require.Len(t, lines, 1)
	assert.Equal(t, "ERROR", lines[0]["level"])
	assert.Equal(t, "requests cut by the shutdown", lines[0]["msg"])
	assert.EqualValues(t, 2, lines[0]["count"])
}

func TestWithNothingInFlightTheServerStopsAtOnce(t *testing.T) {
	t.Parallel()
	api := start(t, http.NotFoundHandler(), 30*time.Second)
	require.Equal(t, http.StatusNotFound, (<-get(api.url)).status)

	started := time.Now()
	api.signal()

	assert.NoError(t, <-api.stopped)
	assert.Less(t, time.Since(started), 5*time.Second)
	assert.Empty(t, api.logs.lines(t))
}

func TestServeFailsWhenItCannotTakeConnections(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	err = lifecycle.Serve(context.Background(), listener, http.NotFoundHandler(), time.Second, telemetry.NewLogger(io.Discard, slog.LevelDebug))

	assert.Error(t, err)
}

func TestTheFirstSignalStartsTheShutdownAndASecondOneExitsAtOnce(t *testing.T) {
	t.Parallel()
	signals := make(chan os.Signal, 2)
	exited := make(chan int, 1)

	shutdown := lifecycle.ShutdownNotice(signals, func(status int) { exited <- status })

	require.NoError(t, shutdown.Err(), "no signal yet")
	signals <- syscall.SIGTERM
	<-shutdown.Done()
	select {
	case status := <-exited:
		require.Fail(t, "the first signal exited the process", "status %d", status)
	case <-time.After(100 * time.Millisecond):
	}
	signals <- os.Interrupt
	assert.NotZero(t, <-exited, "the second signal exits with a status other than zero")
}
