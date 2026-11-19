package applicationtest

import (
	"context"
	"slices"
	"time"
)

// EndedSessions stands in for the store of the sessions: it keeps each moment
// it was asked to remove the sessions that ended by.
type EndedSessions struct {
	removedAt []time.Time
}

// NewEndedSessions returns a store asked to remove nothing yet.
func NewEndedSessions() *EndedSessions {
	return &EndedSessions{}
}

// Remove keeps the moment.
func (s *EndedSessions) Remove(_ context.Context, at time.Time) error {
	s.removedAt = append(s.removedAt, at)
	return nil
}

// RemovedAt returns each moment the sessions that ended were removed by, in
// order.
func (s *EndedSessions) RemovedAt() []time.Time {
	return slices.Clone(s.removedAt)
}
