package database_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
)

func TestAQueryMadeDuringARequestIsASpanUnderItWithItsSQLTextAndNoParameter(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	spans := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIConnection(t),
		MaxSize:                  1,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: 10 * time.Second,
		TracerProvider:           provider,
	})
	require.NoError(t, err)
	defer pool.Close()
	var answered string
	api := telemetry.Requests(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		err = pool.QueryRow(r.Context(), "SELECT upper($1::text)", "ana@example.com").Scan(&answered)
	}), telemetry.NewLogger(io.Discard, slog.LevelError), provider)

	api.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	require.NoError(t, err)
	require.Equal(t, "ANA@EXAMPLE.COM", answered)
	request := spanOfTheRequest(t, spans.Ended())
	query := spanOfTheQuery(t, spans.Ended(), "SELECT upper($1::text)")
	assert.Equal(t, request.SpanContext().SpanID(), query.Parent().SpanID(), "the query is a span under the span of the request")
	assert.Equal(t, request.SpanContext().TraceID(), query.SpanContext().TraceID())
	for _, span := range spans.Ended() {
		recorded := fmt.Sprint(span.Name(), span.Attributes(), span.Events(), span.Status())
		assert.NotContains(t, recorded, "ana@example.com", "the span %q carries a parameter", span.Name())
		assert.NotContains(t, recorded, "ANA@EXAMPLE.COM", "the span %q carries a result", span.Name())
	}
}

func spanOfTheRequest(t *testing.T, spans []sdktrace.ReadOnlySpan) sdktrace.ReadOnlySpan {
	t.Helper()
	index := slices.IndexFunc(spans, func(span sdktrace.ReadOnlySpan) bool {
		return span.SpanKind() == trace.SpanKindServer
	})
	require.GreaterOrEqual(t, index, 0, "the request opened no span")
	return spans[index]
}

func spanOfTheQuery(t *testing.T, spans []sdktrace.ReadOnlySpan, sql string) sdktrace.ReadOnlySpan {
	t.Helper()
	index := slices.IndexFunc(spans, func(span sdktrace.ReadOnlySpan) bool {
		// The span that prepares the statement carries the SQL text too.
		return textOf(span, "db.query.text") == sql && textOf(span, "pgx.prepare_stmt.name") == ""
	})
	require.GreaterOrEqual(t, index, 0, "no span of a query carries the SQL text: %s", sql)
	return spans[index]
}

func textOf(span sdktrace.ReadOnlySpan, key string) string {
	attributes := span.Attributes()
	index := slices.IndexFunc(attributes, func(each attribute.KeyValue) bool {
		return string(each.Key) == key
	})
	if index < 0 {
		return ""
	}
	return attributes[index].Value.AsString()
}
