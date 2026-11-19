package applicationtest

import (
	"context"
	"slices"
	"time"
)

// BounceNotices keeps in memory the identifiers of the bounce notices taken,
// with when each was received. It takes part in the in-memory transactions,
// as the table of the notices takes part in the transaction of the use case.
type BounceNotices struct {
	kept []keptNotice
}

type keptNotice struct {
	id         string
	receivedAt time.Time
}

// NewBounceNotices returns no notice.
func NewBounceNotices() *BounceNotices {
	return &BounceNotices{}
}

// Keep keeps the identifier, and tells false when it was kept before.
func (n *BounceNotices) Keep(_ context.Context, notice string, receivedAt time.Time) (bool, error) {
	if slices.ContainsFunc(n.kept, func(kept keptNotice) bool { return kept.id == notice }) {
		return false, nil
	}
	n.kept = append(n.kept, keptNotice{id: notice, receivedAt: receivedAt})
	return true, nil
}

// RemoveReceivedBefore removes the notices received before the moment.
func (n *BounceNotices) RemoveReceivedBefore(_ context.Context, before time.Time) error {
	n.kept = slices.DeleteFunc(n.kept, func(kept keptNotice) bool { return kept.receivedAt.Before(before) })
	return nil
}

// Snapshot lets the notices take part in the in-memory transactions.
func (n *BounceNotices) Snapshot() (restore func()) {
	saved := slices.Clone(n.kept)
	return func() { n.kept = saved }
}

// Kept returns the identifier of each notice taken, in order.
func (n *BounceNotices) Kept() []string {
	identifiers := make([]string, 0, len(n.kept))
	for _, kept := range n.kept {
		identifiers = append(identifiers, kept.id)
	}
	return identifiers
}
