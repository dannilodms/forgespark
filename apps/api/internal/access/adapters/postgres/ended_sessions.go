package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
)

// EndedSessions removes from the sessions table, which pgxstore keeps for
// platform/session, the sessions that ended. The store of the API runs no
// cleanup of its own: the daily cleanup of the worker does it.
type EndedSessions struct {
	pool *pgxpool.Pool
}

// NewEndedSessions returns the sessions on this pool.
func NewEndedSessions(pool *pgxpool.Pool) *EndedSessions {
	return &EndedSessions{pool: pool}
}

// Remove removes the sessions whose expiry passed by the moment: idle too
// long, or past their lifetime.
func (s *EndedSessions) Remove(ctx context.Context, at time.Time) error {
	if err := New(database.Within(ctx, s.pool)).RemoveEndedSessions(ctx, at); err != nil {
		return fmt.Errorf("remove the sessions that ended: %w", err)
	}
	return nil
}
