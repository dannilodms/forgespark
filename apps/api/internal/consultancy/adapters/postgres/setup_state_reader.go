package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
)

// SetupStateReader reads the state of the setup in one query. It builds no
// aggregate and locks nothing.
type SetupStateReader struct {
	pool *pgxpool.Pool
}

// NewSetupStateReader returns the reader on this pool.
func NewSetupStateReader(pool *pgxpool.Pool) *SetupStateReader {
	return &SetupStateReader{pool: pool}
}

// Read tells whether the consultancy exists and whether a setup code works at
// this moment.
func (r *SetupStateReader) Read(ctx context.Context, at time.Time) (application.SetupState, error) {
	row, err := New(database.Within(ctx, r.pool)).ReadSetupState(ctx, at)
	if err != nil {
		return application.SetupState{}, fmt.Errorf("read the state of the setup: %w", err)
	}
	return application.SetupState{Concluded: row.Concluded, SetupCodeWaiting: row.SetupCodeWaiting}, nil
}
