package application

import (
	"context"
	"time"
)

// BounceNoticeKeptFor is how long the identifier of a bounce notice is kept:
// longer than its sender delivers a notice again, so a notice delivered again
// still changes nothing. Resend, through Svix, gives up on a notice within
// days.
const BounceNoticeKeptFor = 30 * 24 * time.Hour

// CleanUpBounceNotices removes the identifiers of the bounce notices received
// more than BounceNoticeKeptFor ago. The daily cleanup of the worker runs it
// (Jobs in CODING_STANDARDS.md).
type CleanUpBounceNotices struct {
	notices BounceNotices
	clock   Clock
}

// NewCleanUpBounceNotices builds the use case.
func NewCleanUpBounceNotices(notices BounceNotices, clock Clock) *CleanUpBounceNotices {
	return &CleanUpBounceNotices{notices: notices, clock: clock}
}

// Execute removes the old notices. A run more than it needed removes
// nothing.
func (u *CleanUpBounceNotices) Execute(ctx context.Context) error {
	return u.notices.RemoveReceivedBefore(ctx, u.clock.Now().Add(-BounceNoticeKeptFor))
}
