package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transactions opens and closes the transaction of a use case on the pool.
// It is what a command passes to a use case that declares a transactions
// interface.
type Transactions struct {
	pool *pgxpool.Pool
}

// NewTransactions returns the transactions of this pool.
func NewTransactions(pool *pgxpool.Pool) *Transactions {
	return &Transactions{pool: pool}
}

// Run calls the operation inside one transaction, which travels in the
// context the operation receives. The transaction is committed when the
// operation returns nil and rolled back when it returns an error, which Run
// returns as it is.
//
// An operation that calls Run again, as a use case of another area may, joins
// the transaction that is already open: nothing is saved until the outermost
// operation succeeds.
func (t *Transactions) Run(ctx context.Context, operation func(ctx context.Context) error) error {
	if _, open := ctx.Value(transactionKey{}).(pgx.Tx); open {
		return operation(ctx)
	}
	return pgx.BeginFunc(ctx, t.pool, func(transaction pgx.Tx) error {
		return operation(context.WithValue(ctx, transactionKey{}, transaction))
	})
}

type transactionKey struct{}

// Querier is what a `postgres` adapter runs its SQL on. The code sqlc
// generates takes one.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// Within returns where an adapter runs its SQL for this context: the
// transaction a use case opened with Run, or the pool when there is none, as
// for a query use case. An adapter calls it in every method and keeps no
// connection of its own, so everything a use case writes, through any
// adapter of any area, is in its one transaction.
func Within(ctx context.Context, pool *pgxpool.Pool) Querier {
	if transaction, open := ctx.Value(transactionKey{}).(pgx.Tx); open {
		return transaction
	}
	return pool
}
