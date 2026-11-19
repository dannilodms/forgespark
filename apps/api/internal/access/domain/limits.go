package domain

import "time"

// LoginCodeLimit is a limit on asking for login codes, for each user: at most
// Requests in any Window.
type LoginCodeLimit struct {
	Requests int
	Window   time.Duration
}

// LoginCodeLimits are the limits of asking for a login code, counted on the
// requests of each user: one a minute and five an hour (Login in
// CODING_STANDARDS.md).
func LoginCodeLimits() []LoginCodeLimit {
	return []LoginCodeLimit{
		{Requests: 1, Window: time.Minute},
		{Requests: 5, Window: time.Hour},
	}
}

// Since returns the start of the window that ends at this moment: a request
// asked for after it counts against the limit.
func (l LoginCodeLimit) Since(at time.Time) time.Time {
	return at.Add(-l.Window)
}

// Allows tells whether a user who asked for this many codes in the window may
// ask for one more, and answers LoginCodeLimitReachedError when not.
func (l LoginCodeLimit) Allows(asked int) error {
	if asked >= l.Requests {
		return LoginCodeLimitReachedError{}
	}
	return nil
}

// LoginCodeLimitReachedError is the answer to asking for a login code past a
// limit of the e-mail: the person has to wait before asking for another.
type LoginCodeLimitReachedError struct{}

func (LoginCodeLimitReachedError) Error() string {
	return "too many login codes were asked for this e-mail: one a minute and five an hour are allowed, so wait before asking for another"
}

// Code returns the stable code an inbound adapter answers with.
func (LoginCodeLimitReachedError) Code() string {
	return "login_code_limit_reached"
}
