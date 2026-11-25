package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
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

// RemoveReceivedBefore removes the notices received before the moment.
func (n *BounceNotices) RemoveReceivedBefore(ctx context.Context, before time.Time) error {
	if err := New(database.Within(ctx, n.pool)).RemoveBounceNoticesReceivedBefore(ctx, before); err != nil {
		return fmt.Errorf("remove the old bounce notices: %w", err)
	}
	return nil
}
