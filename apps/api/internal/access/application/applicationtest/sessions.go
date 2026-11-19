package applicationtest

import (
	"context"
	"errors"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

// browserKey is the key of the browser a context of a test carries.
type browserKey struct{}

// InBrowser returns the context of a request sent by this browser, as the
// session middleware gives a request the session of its cookie.
func InBrowser(ctx context.Context, browser string) context.Context {
	return context.WithValue(ctx, browserKey{}, browser)
}

// Sessions keeps the session of each browser in memory. Like the session of
// the REST adapter, it is no part of a transaction: what a use case writes in
// it stays.
type Sessions struct {
	loginCodeRequests map[string]domain.LoginCodeRequestID
}

// NewSessions returns no session.
func NewSessions() *Sessions {
	return &Sessions{loginCodeRequests: map[string]domain.LoginCodeRequestID{}}
}

// errNoBrowser is a request whose context carries no session, which the
// session middleware never lets through.
var errNoBrowser = errors.New("the context carries no session: build it with InBrowser")

// RememberLoginCodeRequest keeps the request in the session of the browser of
// the context.
func (s *Sessions) RememberLoginCodeRequest(ctx context.Context, id domain.LoginCodeRequestID) error {
	browser, known := ctx.Value(browserKey{}).(string)
	if !known {
		return errNoBrowser
	}
	s.loginCodeRequests[browser] = id
	return nil
}

// LoginCodeRequestOf returns the login code request the session of the
// browser keeps, and false when it keeps none, as when the browser never
// asked for a code.
func (s *Sessions) LoginCodeRequestOf(browser string) (domain.LoginCodeRequestID, bool) {
	id, kept := s.loginCodeRequests[browser]
	return id, kept
}
