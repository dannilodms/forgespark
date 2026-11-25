package database

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
)

// PoolSettings is what the pool takes from configuration.
type PoolSettings struct {
	// ConnectionURL is the connection of the role the pool logs in as. It
	// prints as a placeholder, and only the driver receives the URL itself.
	ConnectionURL config.PostgresURL
	// MaxSize is the most connections the pool opens.
	MaxSize int32
	// StatementTimeout is how long one statement may run before PostgreSQL
	// cancels it.
	StatementTimeout time.Duration
	// IdleInTransactionTimeout is how long a transaction may sit idle before
	// PostgreSQL ends its session.
	IdleInTransactionTimeout time.Duration
	// TracerProvider opens a span for each query, under the span of the
	// request that made it. With none, the pool is not traced.
	TracerProvider trace.TracerProvider
}

// NewPool builds the one pool of the process. It opens no connection yet, so
// the process starts while PostgreSQL is down and readiness tells.
func NewPool(settings PoolSettings) (*pgxpool.Pool, error) {
	parsed, err := pgxpool.ParseConfig(settings.ConnectionURL.ConnectionString())
	if err != nil {
		// The error of the driver quotes the connection URL.
		return nil, errors.New("the connection of the pool is not valid")
	}
	parsed.MaxConns = settings.MaxSize
	parsed.ConnConfig.RuntimeParams["statement_timeout"] = milliseconds(settings.StatementTimeout)
	parsed.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = milliseconds(settings.IdleInTransactionTimeout)
	if settings.TracerProvider != nil {
		// A span records the SQL text. The option that adds the parameters
		// stays off: they hold what a person typed.
		parsed.ConnConfig.Tracer = otelpgx.NewTracer(otelpgx.WithTracerProvider(settings.TracerProvider))
	}

	return pgxpool.NewWithConfig(context.Background(), parsed)
}

func milliseconds(duration time.Duration) string {
	return strconv.FormatInt(duration.Milliseconds(), 10)
}
