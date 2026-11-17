package database_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		ConnectionURL:            temporary.APIURL,
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
	for _, span := range spans {
		if span.SpanKind() == trace.SpanKindServer {
			return span
		}
	}
	require.Fail(t, "the request opened no span")
	return nil
}

func spanOfTheQuery(t *testing.T, spans []sdktrace.ReadOnlySpan, sql string) sdktrace.ReadOnlySpan {
	t.Helper()
	for _, span := range spans {
		// The span that prepares the statement carries the SQL text too.
		if textOf(span, "db.query.text") == sql && textOf(span, "pgx.prepare_stmt.name") == "" {
			return span
		}
	}
	require.Fail(t, "no span of a query carries the SQL text", "%s", sql)
	return nil
}

func textOf(span sdktrace.ReadOnlySpan, key string) string {
	for _, attribute := range span.Attributes() {
		if string(attribute.Key) == key {
			return attribute.Value.AsString()
		}
	}
	return ""
}
