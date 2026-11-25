package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
)

// linesOf reads what a logger wrote as one JSON object for each line, and
// fails the test on a line that is not one.
func linesOf(t *testing.T, written string) []map[string]any {
	t.Helper()
	lines := []map[string]any{}
	for line := range strings.Lines(written) {
		var object map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &object), "not a JSON object: %s", line)
		lines = append(lines, object)
	}
	return lines
}

// inTrace gives a context that carries a span of the given trace.
func inTrace(t *testing.T, traceID string) context.Context {
	t.Helper()
	parsedTrace, err := trace.TraceIDFromHex(traceID)
	require.NoError(t, err)
	parsedSpan, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	require.NoError(t, err)
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: parsedTrace, SpanID: parsedSpan})
	return trace.ContextWithSpanContext(context.Background(), spanContext)
}

func TestTheLoggerWritesOneJSONObjectForEachLine(t *testing.T) {
	var written bytes.Buffer
	logger := telemetry.NewLogger(&written, slog.LevelInfo)

	logger.Info("the API started", "address", "127.0.0.1:8080")
	logger.Error("request failed", "error", errors.New("PostgreSQL does not answer:\nconnection refused"))

	lines := linesOf(t, written.String())
	require.Len(t, lines, 2)
	assert.Equal(t, "INFO", lines[0]["level"])
	assert.Equal(t, "the API started", lines[0]["msg"])
	assert.Equal(t, "127.0.0.1:8080", lines[0]["address"])
	assert.NotEmpty(t, lines[0]["time"])
	assert.Equal(t, "ERROR", lines[1]["level"])
	assert.Equal(t, "PostgreSQL does not answer:\nconnection refused", lines[1]["error"])
}

func TestALineWrittenDuringARequestCarriesItsTraceIdentifier(t *testing.T) {
	var written bytes.Buffer
	logger := telemetry.NewLogger(&written, slog.LevelInfo)
	request := inTrace(t, "4bf92f3577b34da6a3ce929d0e0e4736")

	logger.ErrorContext(request, "request failed")
	logger.With("route", "GET /readyz").WithGroup("check").ErrorContext(request, "request failed", "part", "postgres")

	lines := linesOf(t, written.String())
	require.Len(t, lines, 2)
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", lines[0]["trace_id"])
	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", lines[1]["trace_id"], "the trace identifier stays at the root of the object")
	assert.Equal(t, "GET /readyz", lines[1]["route"])
}

func TestALineWrittenOutsideARequestCarriesNoTraceIdentifier(t *testing.T) {
	var written bytes.Buffer
	logger := telemetry.NewLogger(&written, slog.LevelInfo)

	logger.Info("the API started")

	lines := linesOf(t, written.String())
	require.Len(t, lines, 1)
	assert.NotContains(t, lines[0], "trace_id")
}

func TestTheLoggerWritesNothingBelowItsLevel(t *testing.T) {
	var written bytes.Buffer
	logger := telemetry.NewLogger(&written, slog.LevelWarn)

	logger.Debug("a detail")
	logger.Info("the API started")
	logger.Warn("something to look at")

	lines := linesOf(t, written.String())
	require.Len(t, lines, 1)
	assert.Equal(t, "something to look at", lines[0]["msg"])
}

// A library that writes with the standard library's logger, such as a server
// of net/http, writes through the logger of the command: a JSON line, never a
// plain one. The test changes the logger of the whole process, so it runs
// alone and puts it back.
func TestTheStandardLoggerWritesThroughTheLoggerOfTheCommand(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(previous)
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
	})
	var written bytes.Buffer

	telemetry.RouteTheStandardLogger(telemetry.NewLogger(&written, slog.LevelInfo))
	log.Print("http: a line of a library")

	lines := linesOf(t, written.String())
	require.Len(t, lines, 1)
	assert.Equal(t, "http: a line of a library", lines[0]["msg"])
	assert.Equal(t, "INFO", lines[0]["level"])
}
