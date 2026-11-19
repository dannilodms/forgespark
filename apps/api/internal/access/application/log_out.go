package application

import "context"

// LogOut ends the session of the browser: the user is no longer logged in to
// it, and its cookie opens nothing. The audit log records no logout.
type LogOut struct {
	session Session
}

// NewLogOut builds the use case.
func NewLogOut(session Session) *LogOut {
	return &LogOut{session: session}
}

// Execute ends the session, whether a user is logged in to it or not.
func (u *LogOut) Execute(ctx context.Context) error {
	return u.session.End(ctx)
}
