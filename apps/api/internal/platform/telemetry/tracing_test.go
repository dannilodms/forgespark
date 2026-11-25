package telemetry_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry/telemetrytest"
)

func TestSpansAreExportedToTheAddressOfTheConfiguration(t *testing.T) {
	collector := telemetrytest.NewCollector(t)
	var logs bytes.Buffer
	tracing, err := telemetry.StartTracing("api", collector.URL, telemetry.NewLogger(&logs, slog.LevelDebug))
	require.NoError(t, err)

	_, span := tracing.Provider.Tracer("test").Start(context.Background(), "issue a setup code")
	span.End()
	tracing.Stop()

	spans := collector.Spans()
	require.Len(t, spans, 1)
	assert.Equal(t, "issue a setup code", spans[0].Name)
	assert.Equal(t, "api", spans[0].Service)
	assert.Equal(t, []string{"/v1/traces"}, collector.Paths())
	assert.True(t, tracing.Exports())
	assert.Empty(t, logs.String())
}

func TestEveryTraceIsKeptWhateverTheCallerSampled(t *testing.T) {
	collector := telemetrytest.NewCollector(t)
	tracing, err := telemetry.StartTracing("api", collector.URL, telemetry.NewLogger(io.Discard, slog.LevelDebug))
	require.NoError(t, err)
	api := telemetry.Requests(http.NotFoundHandler(), telemetry.NewLogger(io.Discard, slog.LevelDebug), tracing.Provider)
	request := httptest.NewRequest(http.MethodGet, "/things", nil)
	// The last field says the caller did not sample its trace.
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")

	api.ServeHTTP(httptest.NewRecorder(), request)
	tracing.Stop()

	spans := collector.Spans()
	require.Len(t, spans, 1)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", spans[0].TraceID)
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
	collector := telemetrytest.NewCollectorAnswering(t, http.StatusBadRequest)
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
