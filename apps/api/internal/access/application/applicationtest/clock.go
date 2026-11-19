// Package applicationtest holds the in-memory implementations of the
// interfaces the use cases of the access area declare. A test builds a use
// case on them, and on the in-memory implementations of the consultancy and
// e-mail areas, so every rule of the login is tested with no database, no
// session store and no e-mail server.
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
