package restapi_test

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/restapi"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
)

// A handler of its own is one Huma does not serve, under the prefix and
// outside the spec, such as the bounce notices of Resend. It reads its body
// and answers its failures with the functions of the REST adapter, so the
// limits of a body and the error format hold for it too.

// ownHandler takes a body, refuses one that says "refuse" and fails on one
// that says "fail".
func ownHandler(w http.ResponseWriter, r *http.Request) {
	body, read := restapi.ReadBody(w, r, limits)
	if !read {
		return
	}
	switch string(body) {
	case "refuse":
		restapi.WriteRefusal(w, r, http.StatusUnauthorized, "the signature is not valid")
	case "fail":
		restapi.WriteFailure(w, r, errors.New("dial tcp 127.0.0.1:5432: connection refused"))
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func serveOwnHandler(t *testing.T) served {
	t.Helper()
	logs := &lockedBuffer{}
	logger := telemetry.NewLogger(logs, slog.LevelDebug)
	tracing, err := telemetry.StartTracing("api", "", logger)
	require.NoError(t, err)
	t.Cleanup(tracing.Stop)
	mux := http.NewServeMux()
	api := restapi.New(mux, limits)
	mux.HandleFunc("POST /api/notices", ownHandler)
	server := httptest.NewServer(telemetry.Requests(mux, logger, tracing.Provider))
	t.Cleanup(server.Close)
	return served{url: server.URL, logs: logs, api: api}
}

func TestAHandlerOfItsOwnRefusesInTheErrorFormat(t *testing.T) {
	t.Parallel()
	api := serveOwnHandler(t)

	answer := api.send(t, http.MethodPost, "/api/notices", "refuse")

	require.Equal(t, http.StatusUnauthorized, answer.status)
	assert.Equal(t, problem{
		Title:   "Unauthorized",
		Status:  http.StatusUnauthorized,
		Detail:  "the signature is not valid",
		Code:    "unauthorized",
		Params:  map[string]any{},
		TraceID: traceID,
	}, answer.problem)
	assert.Empty(t, api.logs.String(), "a 4xx writes no line")
}

func TestAHandlerOfItsOwnFailsWithA500ThatSaysNothingOfTheErrorAndLogsIt(t *testing.T) {
	t.Parallel()
	api := serveOwnHandler(t)

	answer := api.send(t, http.MethodPost, "/api/notices", "fail")

	require.Equal(t, http.StatusInternalServerError, answer.status)
	assert.Equal(t, problem{
		Title:   "Internal Server Error",
		Status:  http.StatusInternalServerError,
		Code:    "internal_server_error",
		Params:  map[string]any{},
		TraceID: traceID,
	}, answer.problem)
	assert.NotContains(t, answer.raw, "5432")
	assert.Contains(t, api.logs.String(), `"route":"/api/notices"`)
	assert.Contains(t, api.logs.String(), "dial tcp 127.0.0.1:5432: connection refused")
}

func TestAHandlerOfItsOwnRefusesABodyOverTheSizeLimit(t *testing.T) {
	t.Parallel()
	api := serveOwnHandler(t)

	answer := api.send(t, http.MethodPost, "/api/notices", strings.Repeat("a", int(limits.BodyBytes)+1))

	require.Equal(t, http.StatusRequestEntityTooLarge, answer.status, answer.raw)
	assert.Equal(t, "request_entity_too_large", answer.problem.Code)
	assert.Equal(t, traceID, answer.problem.TraceID)
}

func TestAHandlerOfItsOwnTakesABodyAtTheSizeLimit(t *testing.T) {
	t.Parallel()
	api := serveOwnHandler(t)

	answer := api.send(t, http.MethodPost, "/api/notices", strings.Repeat("a", int(limits.BodyBytes)))

	assert.Equal(t, http.StatusNoContent, answer.status, answer.raw)
}

func TestAHandlerOfItsOwnAnswersACallerThatStallsOnItsBodyWhenTheReadTimeoutEnds(t *testing.T) {
	t.Parallel()
	api := serveOwnHandler(t)
	connection, err := net.Dial("tcp", strings.TrimPrefix(api.url, "http://"))
	require.NoError(t, err)
	t.Cleanup(func() { connection.Close() })

	// The body says it has 100 bytes and sends 3.
	_, err = io.WriteString(connection, "POST /api/notices HTTP/1.1\r\nHost: api\r\nContent-Length: 100\r\n\r\nabc")
	require.NoError(t, err)
	require.NoError(t, connection.SetReadDeadline(time.Now().Add(5*time.Second)))
	response, err := http.ReadResponse(bufio.NewReader(connection), nil)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusRequestTimeout, response.StatusCode)
	assert.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"code":"request_timeout"`)
}
