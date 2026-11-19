package restapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

// actors stands in for the access area: it answers the same actor, nobody,
// or a failure, for every request.
type actors struct {
	actor kernel.Actor
	found bool
	err   error
}

func (a actors) ActorOf(context.Context) (kernel.Actor, bool, error) {
	return a.actor, a.found, a.err
}

type whoActsOutput struct {
	Body struct {
		User   int64          `json:"user"`
		Scopes []kernel.Scope `json:"scopes"`
	}
}

type renameInput struct {
	Body struct {
		Name string `json:"name" minLength:"1"`
	}
}

// servedForUsers is an API with operations that need a logged-in user. The
// handler of the second counts its runs.
func servedForUsers(t *testing.T, found restapi.Actors) (served, *int) {
	t.Helper()
	logs := &lockedBuffer{}
	logger := telemetry.NewLogger(logs, slog.LevelDebug)
	tracing, err := telemetry.StartTracing("api", "", logger)
	require.NoError(t, err)
	t.Cleanup(tracing.Stop)

	mux := http.NewServeMux()
	api := restapi.New(mux, limits)
	restapi.RegisterForUser(api, huma.Operation{
		OperationID: "tell_who_acts",
		Method:      http.MethodGet,
		Path:        "/api/actor",
		Summary:     "Tell who acts",
	}, nil, func(_ context.Context, actor kernel.Actor, _ *struct{}) (*whoActsOutput, error) {
		output := &whoActsOutput{}
		output.Body.User = actor.User()
		output.Body.Scopes = actor.Scopes()
		return output, nil
	})
	runs := 0
	restapi.RegisterForUser(api, huma.Operation{
		OperationID: "rename_actor",
		Method:      http.MethodPost,
		Path:        "/api/actor/rename",
		Summary:     "Rename who acts",
	}, restapi.Failures{
		http.StatusUnprocessableEntity: {thingLocked{}},
	}, func(context.Context, kernel.Actor, *renameInput) (*struct{}, error) {
		runs++
		return nil, restapi.NotLoggedIn
	})
	// Given after the operations are registered, as mountREST may.
	if found != nil {
		api.AuthenticateWith(found)
	}

	server := httptest.NewServer(telemetry.Requests(mux, logger, tracing.Provider))
	t.Cleanup(server.Close)
	return served{url: server.URL, logs: logs, api: api}, &runs
}

func TestAnOperationForAUserReceivesWhoActs(t *testing.T) {
	t.Parallel()
	api, _ := servedForUsers(t, actors{actor: kernel.NewActor(7, []kernel.Scope{"users:read"}), found: true})

	answer := api.send(t, http.MethodGet, "/api/actor", "")

	require.Equal(t, http.StatusOK, answer.status, answer.raw)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(answer.raw), &body))
	assert.Equal(t, map[string]any{"user": float64(7), "scopes": []any{"users:read"}}, body)
}

// Nobody logged in is answered before the body is read, so the input of a
// person who is not logged in is never judged, and the handler never runs.
func TestAnOperationForAUserAnswers401WhenNobodyIsLoggedIn(t *testing.T) {
	t.Parallel()
	api, runs := servedForUsers(t, actors{found: false})

	read := api.send(t, http.MethodGet, "/api/actor", "")
	write := api.send(t, http.MethodPost, "/api/actor/rename", `{"name": ""}`)

	for _, answer := range []answered{read, write} {
		require.Equal(t, http.StatusUnauthorized, answer.status, answer.raw)
		assert.Equal(t, "unauthorized", answer.problem.Code)
		assert.Equal(t, "Unauthorized", answer.problem.Title)
		assert.Equal(t, traceID, answer.problem.TraceID)
		assert.Empty(t, answer.problem.Errors)
	}
	assert.Zero(t, *runs, "the handler ran with nobody logged in")
	assert.Empty(t, api.logs.String(), "a refusal wrote a line")
}

// A handler answers 401 too, when the user it acts for is gone by the time it
// runs.
func TestAnOperationForAUserMayAnswerThatNobodyIsLoggedIn(t *testing.T) {
	t.Parallel()
	api, runs := servedForUsers(t, actors{actor: kernel.NewActor(7, nil), found: true})

	answer := api.send(t, http.MethodPost, "/api/actor/rename", `{"name": "Ana"}`)

	require.Equal(t, http.StatusUnauthorized, answer.status, answer.raw)
	assert.Equal(t, "unauthorized", answer.problem.Code)
	assert.Equal(t, 1, *runs)
}

func TestAFailureToFindWhoActsAnswers500AndLogsTheRawError(t *testing.T) {
	t.Parallel()
	cases := map[string]restapi.Actors{
		"the actors fail":  actors{err: errors.New("pq: the sessions table is gone")},
		"no actors at all": nil,
	}
	for name, found := range cases {
		t.Run(name, func(t *testing.T) {
			api, _ := servedForUsers(t, found)

			answer := api.send(t, http.MethodGet, "/api/actor", "")

			require.Equal(t, http.StatusInternalServerError, answer.status, answer.raw)
			assert.Equal(t, "internal_server_error", answer.problem.Code)
			assert.Empty(t, answer.problem.Detail)
			assert.Contains(t, api.logs.String(), `"level":"ERROR"`)
		})
	}
}

func TestTheSpecOfAnOperationForAUserDeclaresThe401(t *testing.T) {
	t.Parallel()
	api, _ := servedForUsers(t, actors{})
	spec := api.api.OpenAPI()

	for path, operation := range map[string]*huma.Operation{
		"/api/actor":        spec.Paths["/api/actor"].Get,
		"/api/actor/rename": spec.Paths["/api/actor/rename"].Post,
	} {
		unauthorized := operation.Responses["401"]
		require.NotNil(t, unauthorized, "%s declares no 401", path)
		assert.Equal(t, "#/components/schemas/Problem", unauthorized.Content["application/problem+json"].Schema.Ref)
		assert.Contains(t, unauthorized.Description, "`unauthorized`: no user is logged in", path)
	}
	assert.Contains(t, spec.Paths["/api/actor/rename"].Post.Responses, "422", "the failures the operation declares are lost")
}
