package telemetry_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
)

// served is a handler behind the telemetry of requests, with what it wrote
// to the logs and the spans it ended.
type served struct {
	handler http.Handler
	logs    *bytes.Buffer
	spans   *tracetest.SpanRecorder
}

func serve(routes map[string]http.HandlerFunc) served {
	mux := http.NewServeMux()
	for pattern, route := range routes {
		mux.HandleFunc(pattern, route)
	}
	logs := &bytes.Buffer{}
	spans := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	return served{
		handler: telemetry.Requests(mux, telemetry.NewLogger(logs, slog.LevelDebug), provider),
		logs:    logs,
		spans:   spans,
	}
}

func (s served) answer(request *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	s.handler.ServeHTTP(response, request)
	return response
}

// attributesOf gives the attributes of a span as text, by key.
func attributesOf(span sdktrace.ReadOnlySpan) map[string]string {
	attributes := map[string]string{}
	for _, attribute := range span.Attributes() {
		attributes[string(attribute.Key)] = attribute.Value.String()
	}
	return attributes
}

func TestARequestOpensASpanNamedAfterItsRoute(t *testing.T) {
	var seen string
	api := serve(map[string]http.HandlerFunc{
		"GET /things/{id}": func(w http.ResponseWriter, r *http.Request) {
			seen = telemetry.TraceID(r.Context())
			w.WriteHeader(http.StatusOK)
		},
	})

	api.answer(httptest.NewRequest(http.MethodGet, "/things/42", nil))

	spans := api.spans.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "GET /things/{id}", spans[0].Name())
	assert.Equal(t, trace.SpanKindServer, spans[0].SpanKind())
	assert.Equal(t, map[string]string{
		"http.request.method":       "GET",
		"http.route":                "/things/{id}",
		"http.response.status_code": "200",
	}, attributesOf(spans[0]))
	assert.Equal(t, spans[0].SpanContext().TraceID().String(), seen, "the handler runs inside the span")
}

func TestASpanCarriesNoQueryStringNoHeaderAndNoBody(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"POST /things": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
	})
	request := httptest.NewRequest(http.MethodPost, "/things?token=in-the-query", strings.NewReader("in-the-body"))
	request.Header.Set("Cookie", "session=in-a-header")

	api.answer(request)

	spans := api.spans.Ended()
	require.Len(t, spans, 1)
	recorded := fmt.Sprint(spans[0].Name(), attributesOf(spans[0]), spans[0].Events(), spans[0].Status())
	assert.NotContains(t, recorded, "in-the-query")
	assert.NotContains(t, recorded, "in-a-header")
	assert.NotContains(t, recorded, "in-the-body")
}

func TestAnIncomingTraceparentIsHonoured(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"GET /things": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
	})
	request := httptest.NewRequest(http.MethodGet, "/things", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	api.answer(request)

	spans := api.spans.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", spans[0].SpanContext().TraceID().String())
	assert.Equal(t, "00f067aa0ba902b7", spans[0].Parent().SpanID().String())
}

func TestASuccessfulRequestAndA4xxWriteNoLine(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"GET /things": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
		"POST /things": func(w http.ResponseWriter, r *http.Request) {
			telemetry.RecordFailure(r.Context(), errors.New("the name is required"))
			w.WriteHeader(http.StatusUnprocessableEntity)
		},
	})

	ok := api.answer(httptest.NewRequest(http.MethodGet, "/things", nil))
	refused := api.answer(httptest.NewRequest(http.MethodPost, "/things", nil))
	unknown := api.answer(httptest.NewRequest(http.MethodGet, "/nowhere", nil))

	assert.Equal(t, http.StatusOK, ok.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, refused.Code)
	assert.Equal(t, http.StatusNotFound, unknown.Code)
	assert.Empty(t, api.logs.String())
}

func TestA5xxWritesOneLineWithItsRawErrorAndTheTraceIdentifierOfTheRequest(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"GET /things/{id}": func(w http.ResponseWriter, r *http.Request) {
			telemetry.RecordFailure(r.Context(), errors.New("PostgreSQL does not answer"))
			w.WriteHeader(http.StatusServiceUnavailable)
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/things/42?token=in-the-query", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	response := api.answer(request)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	lines := linesOf(t, api.logs.String())
	require.Len(t, lines, 1)
	assert.Equal(t, "ERROR", lines[0]["level"])
	assert.Equal(t, "request failed", lines[0]["msg"])
	assert.Equal(t, "GET", lines[0]["method"])
	assert.Equal(t, "/things/{id}", lines[0]["route"])
	assert.EqualValues(t, http.StatusServiceUnavailable, lines[0]["status"])
	assert.Equal(t, "PostgreSQL does not answer", lines[0]["error"])
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", lines[0]["trace_id"])
	assert.NotContains(t, api.logs.String(), "in-the-query")

	spans := api.spans.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestA5xxWithNoRecordedFailureStillWritesItsLine(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"GET /things": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) },
	})

	api.answer(httptest.NewRequest(http.MethodGet, "/things", nil))

	lines := linesOf(t, api.logs.String())
	require.Len(t, lines, 1)
	assert.EqualValues(t, http.StatusBadGateway, lines[0]["status"])
	assert.NotContains(t, lines[0], "error")
}

func TestAPanicAnswers500AndWritesOneLineWithItsRawError(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"GET /things": func(http.ResponseWriter, *http.Request) { panic("the index is out of range") },
	})

	response := api.answer(httptest.NewRequest(http.MethodGet, "/things", nil))

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Empty(t, response.Body.String(), "the answer says nothing about the panic")
	lines := linesOf(t, api.logs.String())
	require.Len(t, lines, 1)
	assert.Equal(t, "request panicked", lines[0]["msg"])
	assert.Equal(t, "the index is out of range", lines[0]["error"])
	assert.EqualValues(t, http.StatusInternalServerError, lines[0]["status"])
	assert.Contains(t, lines[0]["stack"], "requests_test.go")
	assert.NotEmpty(t, lines[0]["trace_id"])

	spans := api.spans.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func TestAHandlerThatAbortsItsResponseIsLeftToTheServer(t *testing.T) {
	api := serve(map[string]http.HandlerFunc{
		"GET /things": func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) },
	})

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		api.answer(httptest.NewRequest(http.MethodGet, "/things", nil))
	})
	assert.Empty(t, api.logs.String())
}

func TestRecordingAFailureOutsideARequestDoesNothing(t *testing.T) {
	assert.NotPanics(t, func() {
		telemetry.RecordFailure(t.Context(), errors.New("nobody is listening"))
	})
}
