package restapi_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

type openDoorInput struct {
	Door string `path:"door"`
}

// servedDoors is an API with an operation whose path names a door, and only
// the front door exists.
func servedDoors(t *testing.T) served {
	t.Helper()
	logs := &lockedBuffer{}
	logger := telemetry.NewLogger(logs, slog.LevelDebug)
	tracing, err := telemetry.StartTracing("api", "", logger)
	require.NoError(t, err)
	t.Cleanup(tracing.Stop)

	mux := http.NewServeMux()
	api := restapi.New(mux, limits)
	restapi.Register(api, huma.Operation{
		OperationID:   "open_door",
		Method:        http.MethodPost,
		Path:          "/api/doors/{door}/open",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Open a door",
	}, restapi.Failures{
		http.StatusNotFound: {restapi.NotFound},
	}, func(_ context.Context, input *openDoorInput) (*struct{}, error) {
		if input.Door != "front" {
			return nil, restapi.NotFound
		}
		return &struct{}{}, nil
	})

	server := httptest.NewServer(telemetry.Requests(mux, logger, tracing.Provider))
	t.Cleanup(server.Close)
	return served{url: server.URL, logs: logs, api: api}
}

// A path that names nothing that exists is answered 404 `not_found`, in the
// error format, with no line in the logs.
func TestAnOperationAnswersNotFoundForAPathThatNamesNothing(t *testing.T) {
	t.Parallel()
	api := servedDoors(t)

	answer := api.send(t, http.MethodPost, "/api/doors/back/open", "")
	opened := api.send(t, http.MethodPost, "/api/doors/front/open", "")

	require.Equal(t, http.StatusNotFound, answer.status, answer.raw)
	assert.Equal(t, "not_found", answer.problem.Code)
	assert.Equal(t, "Not Found", answer.problem.Title)
	assert.Equal(t, traceID, answer.problem.TraceID)
	assert.Equal(t, http.StatusNoContent, opened.status)
	assert.Empty(t, api.logs.String())
	described := api.api.OpenAPI().Paths["/api/doors/{door}/open"].Post.Responses["404"].Description
	assert.Contains(t, described, "`not_found`")
}
