package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/postgres"
)

// keep keeps the notice in a transaction of its own and tells whether it was
// the first time.
func (s stored) keep(t *testing.T, notices *postgres.BounceNotices, notice string) bool {
	t.Helper()
	var first bool
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		var err error
		first, err = notices.Keep(ctx, notice, noon)
		return err
	})
	require.NoError(t, err)
	return first
}

func TestANoticeIsKeptOnce(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	notices := postgres.NewBounceNotices(temporary.pool)

	first := temporary.keep(t, notices, "msg_2fKq8bX1")
	again := temporary.keep(t, notices, "msg_2fKq8bX1")
	another := temporary.keep(t, notices, "msg_7hZr3cY9")

	assert.True(t, first)
	assert.False(t, again)
	assert.True(t, another)
}

func TestANoticeKeptInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	notices := postgres.NewBounceNotices(temporary.pool)
	failed := errors.New("saving the bounce failed")

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		if _, err := notices.Keep(ctx, "msg_2fKq8bX1", noon); err != nil {
			return err
		}
		return failed
	})

	require.ErrorIs(t, err, failed)
	assert.True(t, temporary.keep(t, notices, "msg_2fKq8bX1"))
}

// A notice delivered twice at once is taken by one of the two: the second
// waits for the first to commit, then finds it kept.
func TestTwoDeliveriesOfANoticeAtOnceKeepItOnce(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	notices := postgres.NewBounceNotices(temporary.pool)
	inside := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan bool, 1)

	go func() {
		var first bool
		err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
			var err error
			first, err = notices.Keep(ctx, "msg_2fKq8bX1", noon)
			close(inside)
			<-release
			return err
		})
		assert.NoError(t, err)
		firstDone <- first
	}()
	<-inside
	time.AfterFunc(200*time.Millisecond, func() { close(release) })

	second := temporary.keep(t, notices, "msg_2fKq8bX1")

	assert.True(t, <-firstDone)
	assert.False(t, second)
}

// The API role keeps a notice, reads it and removes it, and changes none.
func TestTheAPIRoleCannotChangeANotice(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.keep(t, postgres.NewBounceNotices(temporary.pool), "msg_2fKq8bX1")
	const insufficientPrivilege = "42501"

	_, err := temporary.pool.Exec(context.Background(), "UPDATE bounce_notices SET id = 'msg_changed'")

	var failure *pgconn.PgError
	require.ErrorAs(t, err, &failure)
	assert.Equal(t, insufficientPrivilege, failure.Code)
}

// The daily cleanup removes the notices received before a moment, and keeps
// the rest, which a notice delivered again still meets.
func TestTheNoticesReceivedBeforeAMomentAreRemoved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	notices := postgres.NewBounceNotices(temporary.pool)
	for notice, at := range map[string]time.Time{"msg_old": noon.Add(-time.Hour), "msg_now": noon, "msg_new": noon.Add(time.Hour)} {
		_, err := notices.Keep(context.Background(), notice, at)
		require.NoError(t, err)
	}

	err := notices.RemoveReceivedBefore(context.Background(), noon)

	require.NoError(t, err)
	assert.True(t, temporary.keep(t, notices, "msg_old"), "the old notice was not removed")
	assert.False(t, temporary.keep(t, notices, "msg_now"))
	assert.False(t, temporary.keep(t, notices, "msg_new"))
}
