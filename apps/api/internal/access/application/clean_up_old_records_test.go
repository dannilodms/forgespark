package application_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// requested saves a request of the user asked for at the moment, and ended at
// the other one unless it is zero.
func (s system) requested(t *testing.T, user domain.UserID, askedAt, endedAt time.Time) domain.LoginCodeRequestID {
	t.Helper()
	id, err := s.requests.NextIdentifier(context.Background())
	require.NoError(t, err)
	request, err := domain.AskForLoginCode(id, user, askedAt, bytes.NewReader(make([]byte, 32)))
	require.NoError(t, err)
	if !endedAt.IsZero() {
		request.End(endedAt)
	}
	require.NoError(t, s.requests.Save(context.Background(), request))
	return id
}

func (s system) requestIDs() []domain.LoginCodeRequestID {
	ids := []domain.LoginCodeRequestID{}
	for _, request := range s.requests.All() {
		ids = append(ids, request.ID())
	}
	return ids
}

// The daily cleanup removes what the login keeps only for a while: a login
// code request a day after it ended or expired, an ask for a login code once
// no limit counts it, the sessions that ended, and, in the e-mail area, the
// bounce notices older than 30 days.
func TestTheDailyCleanupRemovesWhatTheLoginKeepsOnlyForAWhile(t *testing.T) {
	s := newSystem(t)
	ana := domain.UserID(s.registered(t, "ana.souza@fabrica.example", "pt-BR"))
	var none time.Time
	expiredTwoDaysAgo := s.requested(t, ana, noon.Add(-48*time.Hour), none)
	endedMoreThanADayAgo := s.requested(t, ana, noon.Add(-25*time.Hour), noon.Add(-25*time.Hour+time.Minute))
	endedLessThanADayAgo := s.requested(t, ana, noon.Add(-23*time.Hour), noon.Add(-23*time.Hour+time.Minute))
	expiredLessThanADayAgo := s.requested(t, ana, noon.Add(-24*time.Hour+5*time.Minute), none)
	open := s.requested(t, ana, noon.Add(-time.Minute), none)
	email, err := kernel.NewEmail("ana.souza@fabrica.example")
	require.NoError(t, err)
	address := theKey(t).AskedAddress(email)
	ctx := context.Background()
	require.NoError(t, s.asks.Add(ctx, address, noon.Add(-61*time.Minute)))
	require.NoError(t, s.asks.Add(ctx, address, noon.Add(-59*time.Minute)))
	_, err = s.notices.Keep(ctx, "msg_old", noon.Add(-31*24*time.Hour))
	require.NoError(t, err)
	_, err = s.notices.Keep(ctx, "msg_recent", noon.Add(-24*time.Hour))
	require.NoError(t, err)

	err = s.cleanUpOldRecords.Execute(ctx)

	require.NoError(t, err)
	assert.ElementsMatch(t, []domain.LoginCodeRequestID{endedLessThanADayAgo, expiredLessThanADayAgo, open}, s.requestIDs())
	assert.NotContains(t, s.requestIDs(), expiredTwoDaysAgo)
	assert.NotContains(t, s.requestIDs(), endedMoreThanADayAgo)
	require.Len(t, s.asks.All(), 1)
	assert.Equal(t, noon.Add(-59*time.Minute), s.asks.All()[0].AskedAt)
	assert.Equal(t, []time.Time{noon}, s.endedSessions.RemovedAt(), "the sessions that ended by now were not removed")
	assert.Equal(t, []string{"msg_recent"}, s.notices.Kept())
}

func TestTheDailyCleanupRunMoreThanItNeededRemovesNothingMore(t *testing.T) {
	s := newSystem(t)
	ana := domain.UserID(s.registered(t, "ana.souza@fabrica.example", "pt-BR"))
	s.requested(t, ana, noon.Add(-time.Hour), noon.Add(-time.Hour+time.Minute))
	require.NoError(t, s.cleanUpOldRecords.Execute(context.Background()))

	err := s.cleanUpOldRecords.Execute(context.Background())

	require.NoError(t, err)
	assert.Len(t, s.requests.All(), 1)
}
