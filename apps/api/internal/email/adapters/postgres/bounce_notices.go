package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// BounceNotices keeps the identifiers of the bounce notices taken, in the
// bounce_notices table.
type BounceNotices struct {
	pool *pgxpool.Pool
}

// NewBounceNotices returns the notices on this pool.
func NewBounceNotices(pool *pgxpool.Pool) *BounceNotices {
	return &BounceNotices{pool: pool}
}

// Keep keeps the identifier of the notice, and tells false when it was kept
// before.
func (n *BounceNotices) Keep(ctx context.Context, notice string, receivedAt time.Time) (bool, error) {
	kept, err := New(database.Within(ctx, n.pool)).KeepBounceNotice(ctx, KeepBounceNoticeParams{ID: notice, ReceivedAt: receivedAt})
	if err != nil {
		return false, fmt.Errorf("keep a bounce notice: %w", err)
	}
	return kept == 1, nil
}
