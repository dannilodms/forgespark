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

// session is what the session of one browser keeps.
type session struct {
	loginCodeRequest       domain.LoginCodeRequestID
	keepsLoginCodeRequest  bool
	attemptsWithoutRequest int
	user                   domain.UserID
	loggedIn               bool
	renewals               int
	ended                  bool
}

// Sessions keeps the session of each browser in memory. Like the session of
// the REST adapter, it is no part of a transaction: what a use case writes in
// it stays.
type Sessions struct {
	browsers map[string]*session
}

// NewSessions returns no session.
func NewSessions() *Sessions {
	return &Sessions{browsers: map[string]*session{}}
}

// errNoBrowser is a request whose context carries no session, which the
// session middleware never lets through.
var errNoBrowser = errors.New("the context carries no session: build it with InBrowser")

// of returns the session of the browser of the context, which exists from
// the first request of the browser on.
func (s *Sessions) of(ctx context.Context) (*session, error) {
	browser, known := ctx.Value(browserKey{}).(string)
	if !known {
		return nil, errNoBrowser
	}
	return s.ofBrowser(browser), nil
}

func (s *Sessions) ofBrowser(browser string) *session {
	kept, exists := s.browsers[browser]
	if !exists {
		kept = &session{}
		s.browsers[browser] = kept
	}
	return kept
}

// RememberLoginCodeRequest keeps the request in the session of the browser of
// the context, and forgets the attempts made with no request.
func (s *Sessions) RememberLoginCodeRequest(ctx context.Context, id domain.LoginCodeRequestID) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	kept.loginCodeRequest = id
	kept.keepsLoginCodeRequest = true
	kept.attemptsWithoutRequest = 0
	return nil
}

// LoginCodeRequest returns the request the session keeps, and
// domain.NoLoginCodeRequest when it keeps none.
func (s *Sessions) LoginCodeRequest(ctx context.Context) (domain.LoginCodeRequestID, error) {
	kept, err := s.of(ctx)
	if err != nil {
		return 0, err
	}
	return kept.loginCodeRequest, nil
}

// AttemptsWithoutRequest returns the codes typed with no request.
func (s *Sessions) AttemptsWithoutRequest(ctx context.Context) (int, error) {
	kept, err := s.of(ctx)
	if err != nil {
		return 0, err
	}
	return kept.attemptsWithoutRequest, nil
}

// RememberAttemptsWithoutRequest keeps the codes typed with no request.
func (s *Sessions) RememberAttemptsWithoutRequest(ctx context.Context, attempts int) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	kept.attemptsWithoutRequest = attempts
	return nil
}

// LogIn renews the token of the session, which RenewalsOf counts, keeps the
// user, and forgets the request and the attempts.
func (s *Sessions) LogIn(ctx context.Context, user domain.UserID) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	*kept = session{user: user, loggedIn: true, renewals: kept.renewals + 1}
	return nil
}

// User returns the user logged in to the session.
func (s *Sessions) User(ctx context.Context) (domain.UserID, bool, error) {
	kept, err := s.of(ctx)
	if err != nil {
		return 0, false, err
	}
	return kept.user, kept.loggedIn, nil
}

// End removes what the session keeps, and marks it ended.
func (s *Sessions) End(ctx context.Context) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	*kept = session{ended: true}
	return nil
}

// LoginCodeRequestOf returns the login code request the session of the
// browser keeps, and false when it keeps none, as when the browser never
// asked for a code.
func (s *Sessions) LoginCodeRequestOf(browser string) (domain.LoginCodeRequestID, bool) {
	kept := s.ofBrowser(browser)
	return kept.loginCodeRequest, kept.keepsLoginCodeRequest
}

// UserOf returns the user logged in to the session of the browser, and false
// when nobody is.
func (s *Sessions) UserOf(browser string) (domain.UserID, bool) {
	kept := s.ofBrowser(browser)
	return kept.user, kept.loggedIn
}

// RenewalsOf counts the times the token of the session of the browser was
// renewed: once at each login.
func (s *Sessions) RenewalsOf(browser string) int {
	return s.ofBrowser(browser).renewals
}

// Ended tells whether the session of the browser was ended and nothing was
// kept in it since.
func (s *Sessions) Ended(browser string) bool {
	return s.ofBrowser(browser).ended
}
