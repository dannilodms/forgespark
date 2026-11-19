// Package applicationtest holds the in-memory implementations of the
// interfaces the use cases of the e-mail area declare. A test builds a use
// case on them, so every rule of an e-mail is tested with no database, no job
// and no e-mail server. A test of another area that queues an e-mail builds
// QueueEmail on the same implementations.
package applicationtest

import "time"

// Clock is a clock that stands still until a test moves it.
type Clock struct {
	now time.Time
}

// NewClock returns a clock stopped at this time.
func NewClock(now time.Time) *Clock {
	return &Clock{now: now}
}

// Now returns the time the clock is stopped at.
func (c *Clock) Now() time.Time {
	return c.now
}

// Advance moves the clock forward.
func (c *Clock) Advance(by time.Duration) {
	c.now = c.now.Add(by)
}
