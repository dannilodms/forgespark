package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/email/application"
)

// A bounce notice is kept for 30 days, longer than its sender delivers it
// again, and then removed by the daily cleanup.
func TestTheCleanupRemovesTheBounceNoticesReceivedMoreThan30DaysAgo(t *testing.T) {
	room := newMailroom()
	ctx := context.Background()
	for notice, daysAgo := range map[string]int{"msg_old": 31, "msg_month": 30, "msg_recent": 2} {
		_, err := room.notices.Keep(ctx, notice, noon.Add(-time.Duration(daysAgo)*24*time.Hour))
		require.NoError(t, err)
	}

	err := application.NewCleanUpBounceNotices(room.notices, room.clock).Execute(ctx)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"msg_month", "msg_recent"}, room.notices.Kept())
}
