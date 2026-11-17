package telemetry_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	collectorpb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

// collector stands in for the telemetry collector: it keeps what arrives over
// OTLP/HTTP and answers with the status it was given.
type collector struct {
	*httptest.Server
	mu       sync.Mutex
	paths    []string
	received []*collectorpb.ExportTraceServiceRequest
}

func newCollector(t *testing.T, status int) *collector {
	t.Helper()
	fake := &collector{}
	fake.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}
		var export collectorpb.ExportTraceServiceRequest
		if !assert.NoError(t, proto.Unmarshal(body, &export)) {
			return
		}
		fake.mu.Lock()
		defer fake.mu.Unlock()
		fake.paths = append(fake.paths, r.URL.Path)
		fake.received = append(fake.received, &export)
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(status)
	}))
	t.Cleanup(fake.Close)
	return fake
}

// exported is one span that reached the collector, with the service that
// sent it.
type exported struct {
	service string
	span    *tracepb.Span
}

func (c *collector) spans() []exported {
	c.mu.Lock()
	defer c.mu.Unlock()
	var spans []exported
	for _, export := range c.received {
		for _, resource := range export.GetResourceSpans() {
			spans = append(spans, spansOf(resource)...)
		}
	}
	return spans
}

func spansOf(resource *tracepb.ResourceSpans) []exported {
	var service string
	for _, attribute := range resource.GetResource().GetAttributes() {
		if attribute.GetKey() == "service.name" {
			service = attribute.GetValue().GetStringValue()
		}
	}
	var spans []exported
	for _, scope := range resource.GetScopeSpans() {
		for _, span := range scope.GetSpans() {
			spans = append(spans, exported{service: service, span: span})
		}
	}
	return spans
}

func TestSpansAreExportedToTheAddressOfTheConfiguration(t *testing.T) {
	collector := newCollector(t, http.StatusOK)
	var logs bytes.Buffer
	tracing, err := telemetry.StartTracing("api", collector.URL, telemetry.NewLogger(&logs, slog.LevelDebug))
	require.NoError(t, err)

	_, span := tracing.Provider.Tracer("test").Start(context.Background(), "issue a setup code")
	span.End()
	tracing.Stop()

	spans := collector.spans()
	require.Len(t, spans, 1)
	assert.Equal(t, "issue a setup code", spans[0].span.GetName())
	assert.Equal(t, "api", spans[0].service)
	assert.Equal(t, []string{"/v1/traces"}, collector.paths)
	assert.Empty(t, logs.String())
}

func TestEveryTraceIsKeptWhateverTheCallerSampled(t *testing.T) {
	collector := newCollector(t, http.StatusOK)
	tracing, err := telemetry.StartTracing("api", collector.URL, telemetry.NewLogger(io.Discard, slog.LevelDebug))
	require.NoError(t, err)
	api := telemetry.Requests(http.NotFoundHandler(), telemetry.NewLogger(io.Discard, slog.LevelDebug), tracing.Provider)
	request := httptest.NewRequest(http.MethodGet, "/things", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")

	api.ServeHTTP(httptest.NewRecorder(), request)
	tracing.Stop()

	spans := collector.spans()
	require.Len(t, spans, 1)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", hex.EncodeToString(spans[0].span.GetTraceId()))
}

func TestWithNoAddressASpanStillHasATraceIdentifierAndNothingIsExported(t *testing.T) {
	var logs bytes.Buffer
	tracing, err := telemetry.StartTracing("api", "", telemetry.NewLogger(&logs, slog.LevelDebug))
	require.NoError(t, err)

	ctx, span := tracing.Provider.Tracer("test").Start(context.Background(), "issue a setup code")
	span.End()
	tracing.Stop()

	assert.Len(t, telemetry.TraceID(ctx), 32)
	assert.False(t, tracing.Exports(), "no address, so no exporter")
	assert.Empty(t, logs.String(), "an export that was tried and failed would leave a line")
}

func TestAFailedExportIsLoggedAndTheCommandGoesOn(t *testing.T) {
	collector := newCollector(t, http.StatusBadRequest)
	var logs bytes.Buffer
	tracing, err := telemetry.StartTracing("api", collector.URL, telemetry.NewLogger(&logs, slog.LevelDebug))
	require.NoError(t, err)

	_, span := tracing.Provider.Tracer("test").Start(context.Background(), "issue a setup code")
	span.End()
	tracing.Stop()

	lines := linesOf(t, logs.String())
	require.Len(t, lines, 1)
	assert.Equal(t, "ERROR", lines[0]["level"])
	assert.Equal(t, "export of traces failed", lines[0]["msg"])
	assert.NotEmpty(t, lines[0]["error"])
}
