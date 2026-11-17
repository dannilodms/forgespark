package database

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolSettings is what the pool takes from configuration.
type PoolSettings struct {
	// ConnectionURL is the connection of the role the pool logs in as.
	ConnectionURL string
	// MaxSize is the most connections the pool opens.
	MaxSize int32
	// StatementTimeout is how long one statement may run before PostgreSQL
	// cancels it.
	StatementTimeout time.Duration
	// IdleInTransactionTimeout is how long a transaction may sit idle before
	// PostgreSQL ends its session.
	IdleInTransactionTimeout time.Duration
}

// NewPool builds the one pool of the process. It opens no connection yet, so
// the process starts while PostgreSQL is down and readiness tells.
func NewPool(settings PoolSettings) (*pgxpool.Pool, error) {
	parsed, err := pgxpool.ParseConfig(settings.ConnectionURL)
	if err != nil {
		// The error of the driver quotes the connection URL.
		return nil, errors.New("the connection of the pool is not valid")
	}
	parsed.MaxConns = settings.MaxSize
	parsed.ConnConfig.RuntimeParams["statement_timeout"] = milliseconds(settings.StatementTimeout)
	parsed.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = milliseconds(settings.IdleInTransactionTimeout)

	return pgxpool.NewWithConfig(context.Background(), parsed)
}

func milliseconds(duration time.Duration) string {
	return strconv.FormatInt(duration.Milliseconds(), 10)
}
