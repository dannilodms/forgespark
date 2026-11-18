package main

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry/telemetrytest"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

const (
	// traceparent is the header a caller sends to put a request in its trace.
	traceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	// traceID is the trace identifier that header carries.
	traceID = "4bf92f3577b34da6a3ce929d0e0e4736"
)

func TestASuccessfulRequestAndA4xxWriteNoLine(t *testing.T) {
	t.Parallel()
	api := launchAPI(t, environmentOf(databasetest.New(t)))
	caller := newCaller()

	alive := <-caller.get(api.url + "/livez")
	unknown := <-caller.get(api.url + "/nowhere")
	wrongMethod := caller.send(http.MethodPost, api.url+"/livez", nil, "")

	assert.Equal(t, http.StatusOK, alive.status)
	assert.Equal(t, http.StatusNotFound, unknown.status)
	assert.Equal(t, http.StatusMethodNotAllowed, wrongMethod.status)
	assert.Equal(t, []string{"the API started"}, api.messages(t))
}

func TestA5xxWritesOneJSONLineWithItsRawErrorAndTheTraceIdentifierOfItsRequest(t *testing.T) {
	t.Parallel()
	// No migration was applied, so the API is not ready.
	api := launchAPI(t, environmentOf(databasetest.New(t)))

	notReady := <-newCaller().get(api.url+"/readyz?verbose=in-the-query", "traceparent", traceparent)

	assert.Equal(t, http.StatusServiceUnavailable, notReady.status)
	lines := api.lines(t)
	require.Len(t, lines, 2, "%s", api.logs)
	assert.Equal(t, "ERROR", lines[1]["level"])
	assert.Equal(t, "request failed", lines[1]["msg"])
	assert.Equal(t, "GET", lines[1]["method"])
	assert.Equal(t, "/readyz", lines[1]["route"])
	assert.EqualValues(t, http.StatusServiceUnavailable, lines[1]["status"])
	// The version the binary expects grows with each migration, so the test
	// does not say which it is.
	assert.Regexp(t, `^the database schema is at version 0 and this binary expects version \d+$`, lines[1]["error"])
	assert.Equal(t, traceID, lines[1]["trace_id"])
	assert.NotContains(t, api.logs.String(), "in-the-query")
}

func TestTheStartAndTheStopOfTheProcessAreLogged(t *testing.T) {
	t.Parallel()
	api := launchAPI(t, environmentOf(databasetest.New(t)))

	api.signal()

	assert.Equal(t, 0, <-api.status)
	lines := api.lines(t)
	require.Len(t, lines, 2, "%s", api.logs)
	assert.Equal(t, "the API started", lines[0]["msg"])
	assert.Equal(t, "INFO", lines[0]["level"])
	assert.Equal(t, api.url, "http://"+lines[0]["address"].(string))
	assert.Equal(t, "the API stopped", lines[1]["msg"])
	assert.Equal(t, "INFO", lines[1]["level"])
}

func TestTheLogsGoToTheErrorOutputFromTheLevelOfTheConfigurationUp(t *testing.T) {
	t.Parallel()
	environment := environmentOf(databasetest.New(t))
	environment["LOG_LEVEL"] = "error"
	var stdout, logs output
	// The shutdown notice is already given, so the API starts and stops.
	shutdown, signal := context.WithCancel(context.Background())
	signal()

	status := run(shutdown, nil, environment, &stdout, &logs)

	assert.Equal(t, 0, status)
	assert.Empty(t, logs.String(), "the start and the stop are below the level")
	assert.Empty(t, stdout.String())
}

func TestTheAPIExportsTheSpanOfARequestAndOfItsQueriesToTheAddressOfTheConfiguration(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)
	collector := telemetrytest.NewCollector(t)
	environment := environmentOf(temporary)
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL
	api := launchAPI(t, environment)

	setup := <-newCaller().get(api.url+"/api/setup", "traceparent", traceparent)
	api.signal()

	require.Equal(t, http.StatusOK, setup.status)
	require.Equal(t, 0, <-api.status, "%s", api.logs)
	spans := collector.Spans()
	request := spanNamed(spans, "GET /api/setup")
	require.NotNil(t, request, "%v", spans)
	assert.Equal(t, "api", request.Service)
	assert.Equal(t, traceID, request.TraceID, "the incoming traceparent is honoured")
	assert.Equal(t, "00f067aa0ba902b7", request.ParentSpanID)
	query := spanOfTheQuery(spans, "-- name: ReadSetupState ")
	require.NotNil(t, query, "%v", spans)
	assert.Equal(t, request.SpanID, query.ParentSpanID, "the query is a span under the span of the request")
	assert.Equal(t, traceID, query.TraceID)
	assert.Equal(t, []string{"the API started", "the API stopped"}, api.messages(t))
}

func TestLivenessAndReadinessOpenNoSpanAndARequestUnderAPIStillDoes(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	_, err := database.Migrate(context.Background(), temporary.MigrationConnection(t), migrations.Files)
	require.NoError(t, err)
	collector := telemetrytest.NewCollector(t)
	environment := environmentOf(temporary)
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL
	api := launchAPI(t, environment)
	caller := newCaller()

	alive := <-caller.get(api.url + "/livez")
	ready := <-caller.get(api.url + "/readyz")
	followed := <-caller.get(api.url+"/readyz", "traceparent", traceparent)
	setup := <-caller.get(api.url + "/api/setup")
	api.signal()

	require.Equal(t, http.StatusOK, alive.status)
	require.Equal(t, http.StatusOK, ready.status)
	require.Equal(t, http.StatusOK, followed.status)
	require.Equal(t, http.StatusOK, setup.status)
	require.Equal(t, 0, <-api.status, "%s", api.logs)
	spans := collector.Spans()
	assert.Nil(t, spanNamed(spans, "GET /livez"), "liveness opened a span")
	assert.Nil(t, spanNamed(spans, "GET /readyz"), "readiness opened a span")
	assert.Nil(t, spanOfTheQuery(spans, "goose_db_version"), "the query of readiness opened a span of its own")
	request := spanNamed(spans, "GET /api/setup")
	require.NotNil(t, request, "%v", spans)
	assert.NotNil(t, spanOfTheQuery(spans, "-- name: ReadSetupState "), "%v", spans)
	for _, span := range spans {
		assert.Equal(t, request.TraceID, span.TraceID, "a span outside the request under /api: %s", span.Name)
	}
}

func TestWithNoExportAddressTheAPIStartsAndExportsNothing(t *testing.T) {
	t.Parallel()
	environment := environmentOf(databasetest.New(t))
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = ""
	api := launchAPI(t, environment)

	alive := <-newCaller().get(api.url + "/livez")
	api.signal()

	assert.Equal(t, http.StatusOK, alive.status)
	assert.Equal(t, 0, <-api.status)
	// An export that was tried would fail, with nothing listening, and a
	// failed export leaves a line.
	assert.Equal(t, []string{"the API started", "the API stopped"}, api.messages(t))
}

// spanOfTheQuery finds the span of the query whose SQL text holds the
// fragment, such as the name sqlc writes at its top. The span that prepares
// its statement carries the same text and is left out.
func spanOfTheQuery(spans []telemetrytest.Span, fragment string) *telemetrytest.Span {
	index := slices.IndexFunc(spans, func(span telemetrytest.Span) bool {
		return strings.Contains(span.Attributes["db.query.text"], fragment) && span.Attributes["pgx.prepare_stmt.name"] == ""
	})
	if index < 0 {
		return nil
	}
	return &spans[index]
}

// spanNamed finds the first span with the name.
func spanNamed(spans []telemetrytest.Span, name string) *telemetrytest.Span {
	index := slices.IndexFunc(spans, func(span telemetrytest.Span) bool { return span.Name == name })
	if index < 0 {
		return nil
	}
	return &spans[index]
}
