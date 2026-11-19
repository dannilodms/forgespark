package applicationtest

import (
	"context"
	"slices"
	"time"
)

// BounceNotices keeps in memory the identifiers of the bounce notices taken.
// It takes part in the in-memory transactions, as the table of the notices
// takes part in the transaction of the use case.
type BounceNotices struct {
	kept []string
}

// NewBounceNotices returns no notice.
func NewBounceNotices() *BounceNotices {
	return &BounceNotices{}
}

// Keep keeps the identifier, and tells false when it was kept before.
func (n *BounceNotices) Keep(_ context.Context, notice string, _ time.Time) (bool, error) {
	if slices.Contains(n.kept, notice) {
		return false, nil
	}
	n.kept = append(n.kept, notice)
	return true, nil
}

// Snapshot lets the notices take part in the in-memory transactions.
func (n *BounceNotices) Snapshot() (restore func()) {
	saved := slices.Clone(n.kept)
	return func() { n.kept = saved }
}

// Kept returns the identifier of each notice taken, in order.
func (n *BounceNotices) Kept() []string {
	return slices.Clone(n.kept)
}
