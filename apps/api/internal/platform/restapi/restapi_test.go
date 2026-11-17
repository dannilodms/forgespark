package restapi_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

// The trace an incoming `traceparent` names, so that a test knows which
// `trace_id` an error body must carry.
const (
	traceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	traceID     = "4bf92f3577b34da6a3ce929d0e0e4736"
)

// The limits of the API under test, small enough for a test to reach.
var limits = restapi.Limits{BodyBytes: 512, BodyReadTimeout: 200 * time.Millisecond}

// thingNameTaken is a broken rule, as a use case returns one.
type thingNameTaken struct {
	name string
}

func (thingNameTaken) Error() string { return "another thing has this name" }

func (thingNameTaken) Code() string { return "thing_name_taken" }

func (e thingNameTaken) Params() map[string]any { return map[string]any{"name": e.name} }

// thingArchived is a broken rule the operation under test does not declare.
type thingArchived struct{}

func (thingArchived) Error() string { return "the thing is archived" }

func (thingArchived) Code() string { return "thing_archived" }

type renameThingInput struct {
	ThingID string `path:"thing_id"`
	Notify  bool   `query:"notify"`
	Body    RenameThingRequest
}

type RenameThingRequest struct {
	Name string `json:"name" minLength:"3" maxLength:"400"`
}

type renameThingOutput struct {
	Body RenamedThing
}

type RenamedThing struct {
	Name string `json:"name"`
}

// renameThing stands in for an operation of a domain area. What the caller
// names the thing decides what the use case behind it does.
func renameThing(_ context.Context, input *renameThingInput) (*renameThingOutput, error) {
	switch input.Body.Name {
	case "taken":
		return nil, thingNameTaken{name: "taken"}
	case "archived":
		return nil, thingArchived{}
	case "broken":
		return nil, errors.New(`pq: relation "things" does not exist`)
	case "panic":
		panic("index out of range in renameThing")
	}
	return &renameThingOutput{Body: RenamedThing{Name: input.Body.Name}}, nil
}

// served is an API with one operation, behind the telemetry of a request, as
// the command that serves mounts it.
type served struct {
	url  string
	logs *lockedBuffer
}

func serve(t *testing.T) served {
	t.Helper()
	logs := &lockedBuffer{}
	logger := telemetry.NewLogger(logs, slog.LevelDebug)
	tracing, err := telemetry.StartTracing("api", "", logger)
	require.NoError(t, err)
	t.Cleanup(tracing.Stop)

	mux := http.NewServeMux()
	api := restapi.New(mux, limits)
	restapi.Register(api, huma.Operation{
		OperationID: "rename_thing",
		Method:      http.MethodPost,
		Path:        "/api/things/{thing_id}/rename",
		Summary:     "Rename a thing",
	}, restapi.Failures{http.StatusConflict: {thingNameTaken{}}}, renameThing)

	server := httptest.NewServer(telemetry.Requests(mux, logger, tracing.Provider))
	t.Cleanup(server.Close)
	return served{url: server.URL, logs: logs}
}

// answered is what the API answered: its status, its headers and its body
// read as the error format.
type answered struct {
	status  int
	headers http.Header
	raw     string
	problem problem
}

// problem is the error format, as a client reads it.
type problem struct {
	Title   string         `json:"title"`
	Status  int            `json:"status"`
	Detail  string         `json:"detail"`
	Code    string         `json:"code"`
	Params  map[string]any `json:"params"`
	TraceID string         `json:"trace_id"`
	Errors  []struct {
		Location string         `json:"location"`
		Code     string         `json:"code"`
		Params   map[string]any `json:"params"`
		Message  string         `json:"message"`
	} `json:"errors"`
}

func (s served) send(t *testing.T, method, path, body string) answered {
	t.Helper()
	request, err := http.NewRequest(method, s.url+path, strings.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("traceparent", traceparent)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	answer := answered{status: response.StatusCode, headers: response.Header, raw: string(raw)}
	if response.StatusCode >= http.StatusBadRequest {
		require.Equal(t, "application/problem+json", response.Header.Get("Content-Type"), answer.raw)
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		require.NoError(t, decoder.Decode(&answer.problem), answer.raw)
	}
	return answer
}

// lockedBuffer collects the logs, which the server writes from its own
// goroutines.
type lockedBuffer struct {
	mu      sync.Mutex
	written bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.written.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.written.String()
}

func TestAnOperationAnswersWhatItsHandlerReturns(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"ledger"}`)

	assert.Equal(t, http.StatusOK, answer.status)
	assert.JSONEq(t, `{"name":"ledger"}`, answer.raw)
	assert.Equal(t, "application/json", answer.headers.Get("Content-Type"))
}

func TestAFailedValidationAnswersInTheErrorFormatAndNamesTheField(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"ab"}`)

	assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
	assert.Equal(t, "Unprocessable Entity", answer.problem.Title)
	assert.Equal(t, http.StatusUnprocessableEntity, answer.problem.Status)
	assert.Equal(t, "validation_failed", answer.problem.Code)
	assert.Equal(t, map[string]any{}, answer.problem.Params)
	assert.Equal(t, traceID, answer.problem.TraceID)
	require.Len(t, answer.problem.Errors, 1)
	assert.Equal(t, "body.name", answer.problem.Errors[0].Location)
	assert.Equal(t, "invalid", answer.problem.Errors[0].Code)
	assert.Equal(t, map[string]any{}, answer.problem.Errors[0].Params)
	assert.Equal(t, "expected length >= 3", answer.problem.Errors[0].Message)
	assert.NotContains(t, answer.raw, `"ab"`, "the answer repeats the value that was refused")
}

func TestAnUnknownPropertyOfTheBodyIsRejected(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"ledger","colour":"red"}`)

	assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
	assert.Equal(t, "validation_failed", answer.problem.Code)
	require.Len(t, answer.problem.Errors, 1)
	assert.Equal(t, "body.colour", answer.problem.Errors[0].Location)
	assert.Equal(t, "unexpected property", answer.problem.Errors[0].Message)
}

func TestAnUnknownQueryParameterIsRejected(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename?notify=true&colour=red", `{"name":"ledger"}`)

	assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
	assert.Equal(t, "validation_failed", answer.problem.Code)
	assert.Equal(t, traceID, answer.problem.TraceID)
	require.Len(t, answer.problem.Errors, 1)
	assert.Equal(t, "query.colour", answer.problem.Errors[0].Location)
}

func TestABodyThatIsNotJSONAnswersInTheErrorFormatWithoutRepeatingIt(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"K7M2X-9QW4T-B8N6R-C3V5Z"`)

	assert.Equal(t, http.StatusBadRequest, answer.status)
	assert.Equal(t, "bad_request", answer.problem.Code)
	assert.Equal(t, traceID, answer.problem.TraceID)
	assert.NotContains(t, answer.raw, "K7M2X", "the answer repeats the body that was refused")
}

func TestABodyOverTheSizeLimitIsRejected(t *testing.T) {
	t.Parallel()
	api := serve(t)
	// The limit of this API is 512 bytes. The name alone is within what the
	// contract allows, so only the size of the body is wrong.
	body := `{"name":"` + strings.Repeat("a", 390) + `"` + strings.Repeat(" ", 200) + `}`

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", body)

	assert.Equal(t, http.StatusRequestEntityTooLarge, answer.status)
	assert.Equal(t, "request_entity_too_large", answer.problem.Code)
	assert.Equal(t, traceID, answer.problem.TraceID)
}

func TestABodyAtTheSizeLimitIsAccepted(t *testing.T) {
	t.Parallel()
	api := serve(t)
	body := `{"name":"` + strings.Repeat("a", 390) + `"`
	body += strings.Repeat(" ", 512-len(body)-1) + `}`
	require.Len(t, body, 512)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", body)

	assert.Equal(t, http.StatusOK, answer.status, answer.raw)
}

func TestABrokenRuleTheOperationDeclaresAnswersWithItsStatusCodeAndParameters(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"taken"}`)

	assert.Equal(t, http.StatusConflict, answer.status)
	assert.Equal(t, "Conflict", answer.problem.Title)
	assert.Equal(t, http.StatusConflict, answer.problem.Status)
	assert.Equal(t, "thing_name_taken", answer.problem.Code)
	assert.Equal(t, map[string]any{"name": "taken"}, answer.problem.Params)
	assert.Equal(t, "another thing has this name", answer.problem.Detail)
	assert.Equal(t, traceID, answer.problem.TraceID)
	assert.Empty(t, answer.problem.Errors)
	assert.Empty(t, api.logs.String(), "a 4xx wrote a line to the logs")
}

func TestAnErrorThatIsNoDeclaredBrokenRuleAnswers500AndSaysNothingOfIt(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		name   string
		logged string
	}{
		"an error of the infrastructure":        {name: "broken", logged: `relation \"things\" does not exist`},
		"a broken rule that was never declared": {name: "archived", logged: "the operation does not declare the code thing_archived"},
	}
	for name, failing := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			api := serve(t)

			answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"`+failing.name+`"}`)

			assert.Equal(t, http.StatusInternalServerError, answer.status)
			assert.JSONEq(t, `{"title":"Internal Server Error","status":500,"code":"internal_server_error","params":{},"trace_id":"`+traceID+`"}`, answer.raw)
			logs := api.logs.String()
			assert.Equal(t, 1, strings.Count(logs, "\n"), "one line for a 5xx: %s", logs)
			assert.Contains(t, logs, `"msg":"request failed"`)
			assert.Contains(t, logs, `"route":"/api/things/{thing_id}/rename"`)
			assert.Contains(t, logs, `"trace_id":"`+traceID+`"`)
			assert.Contains(t, logs, failing.logged)
		})
	}
}

func TestAPanicAnswersInTheErrorFormatAndSaysNothingOfIt(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"panic"}`)

	assert.Equal(t, http.StatusInternalServerError, answer.status)
	assert.JSONEq(t, `{"title":"Internal Server Error","status":500,"code":"internal_server_error","params":{},"trace_id":"`+traceID+`"}`, answer.raw)
	logs := api.logs.String()
	assert.Equal(t, 1, strings.Count(logs, "\n"), "one line for a panic: %s", logs)
	assert.Contains(t, logs, "index out of range in renameThing")
	assert.Contains(t, logs, "restapi_test.go", "the line of a panic carries no stack")
	assert.Contains(t, logs, `"trace_id":"`+traceID+`"`)
}

func TestAnUnknownRouteAnswersInTheErrorFormat(t *testing.T) {
	t.Parallel()
	api := serve(t)
	for _, path := range []string{"/api/things/7/remove", "/api/nothing", "/api/", "/api"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			answer := api.send(t, http.MethodGet, path, "")

			assert.Equal(t, http.StatusNotFound, answer.status)
			assert.JSONEq(t, `{"title":"Not Found","status":404,"code":"not_found","params":{},"trace_id":"`+traceID+`"}`, answer.raw)
		})
	}
}

func TestAWrongMethodAnswersInTheErrorFormatAndNamesTheOnesAllowed(t *testing.T) {
	t.Parallel()
	api := serve(t)
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, "PURGE"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			answer := api.send(t, method, "/api/things/7/rename", "")

			assert.Equal(t, http.StatusMethodNotAllowed, answer.status)
			assert.JSONEq(t, `{"title":"Method Not Allowed","status":405,"code":"method_not_allowed","params":{},"trace_id":"`+traceID+`"}`, answer.raw)
			assert.Equal(t, "POST", answer.headers.Get("Allow"))
		})
	}
}

func TestACallerThatStallsOnItsBodyIsAnsweredWhenTheReadTimeoutEnds(t *testing.T) {
	t.Parallel()
	api := serve(t)
	connection, err := net.Dial("tcp", strings.TrimPrefix(api.url, "http://"))
	require.NoError(t, err)
	defer connection.Close()
	// The caller announces 64 bytes and sends 9. The read timeout of this
	// API is 200 milliseconds.
	_, err = connection.Write([]byte("POST /api/things/7/rename HTTP/1.1\r\nHost: api\r\nContent-Type: application/json\r\n" +
		"traceparent: " + traceparent + "\r\nContent-Length: 64\r\n\r\n" + `{"name":"`))
	require.NoError(t, err)
	require.NoError(t, connection.SetReadDeadline(time.Now().Add(10*time.Second)))

	response, err := http.ReadResponse(bufio.NewReader(connection), nil)

	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusRequestTimeout, response.StatusCode)
	assert.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"title":"Request Timeout","status":408,"detail":"request body read timeout","code":"request_timeout","params":{},"trace_id":"`+traceID+`"}`, string(body))
}

func TestABodyOfAnotherMediaTypeIsRejectedInTheErrorFormat(t *testing.T) {
	t.Parallel()
	api := serve(t)
	request, err := http.NewRequest(http.MethodPost, api.url+"/api/things/7/rename", strings.NewReader("name=ledger"))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)

	require.NoError(t, err)
	defer response.Body.Close()
	var refused problem
	require.NoError(t, json.NewDecoder(response.Body).Decode(&refused))
	assert.Equal(t, http.StatusUnsupportedMediaType, response.StatusCode)
	assert.Equal(t, "application/problem+json", response.Header.Get("Content-Type"))
	assert.Equal(t, "unsupported_media_type", refused.Code)
	assert.Len(t, refused.TraceID, 32)
}

func TestTheAPIServesNoDocumentationPageNoSpecAndNoSchema(t *testing.T) {
	t.Parallel()
	api := serve(t)
	paths := []string{
		"/docs", "/openapi.json", "/openapi.yaml", "/schemas/RenamedThing.json",
		"/api/docs", "/api/openapi.json", "/api/openapi.yaml", "/api/schemas/RenamedThing.json",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			response, err := http.Get(api.url + path)

			require.NoError(t, err)
			defer response.Body.Close()
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
		})
	}
}

func TestAnAnswerCarriesNoLinkToASchema(t *testing.T) {
	t.Parallel()
	api := serve(t)

	renamed := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"ledger"}`)
	refused := api.send(t, http.MethodPost, "/api/things/7/rename", `{"name":"ab"}`)

	assert.NotContains(t, renamed.raw, "$schema")
	assert.Empty(t, renamed.headers.Get("Link"))
	assert.NotContains(t, refused.raw, "$schema")
	assert.Empty(t, refused.headers.Get("Link"))
}

func TestARequiredPropertyThatIsMissingIsNamedAsTheFieldItIs(t *testing.T) {
	t.Parallel()
	api := serve(t)

	answer := api.send(t, http.MethodPost, "/api/things/7/rename", `{}`)

	assert.Equal(t, http.StatusUnprocessableEntity, answer.status)
	assert.Equal(t, "validation_failed", answer.problem.Code)
	require.Len(t, answer.problem.Errors, 1)
	assert.Equal(t, "body.name", answer.problem.Errors[0].Location)
	assert.Equal(t, "required", answer.problem.Errors[0].Code)
	assert.Equal(t, "expected required property name to be present", answer.problem.Errors[0].Message)
}
