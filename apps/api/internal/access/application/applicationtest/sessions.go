package applicationtest

import (
	"context"
	"errors"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
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
	providerLogin          domain.ProviderLogin
	startedProviderLogin   bool
	refusal                kernel.Email
	refused                bool
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

// RememberProviderLogin keeps the login through a provider the browser of
// the context started.
func (s *Sessions) RememberProviderLogin(ctx context.Context, login domain.ProviderLogin) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	kept.providerLogin, kept.startedProviderLogin = login, true
	return nil
}

// TakeProviderLogin returns the login through a provider the session keeps,
// and forgets it.
func (s *Sessions) TakeProviderLogin(ctx context.Context) (domain.ProviderLogin, bool, error) {
	kept, err := s.of(ctx)
	if err != nil {
		return domain.ProviderLogin{}, false, err
	}
	login, started := kept.providerLogin, kept.startedProviderLogin
	kept.providerLogin, kept.startedProviderLogin = domain.ProviderLogin{}, false
	return login, started, nil
}

// RememberRefusal keeps the e-mail the last login through a provider was
// refused for.
func (s *Sessions) RememberRefusal(ctx context.Context, email kernel.Email) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	kept.refusal, kept.refused = email, true
	return nil
}

// ForgetRefusal forgets the refusal of the last login through a provider.
func (s *Sessions) ForgetRefusal(ctx context.Context) error {
	kept, err := s.of(ctx)
	if err != nil {
		return err
	}
	kept.refusal, kept.refused = kernel.Email{}, false
	return nil
}

// Refusal returns the e-mail the last login through a provider was refused
// for.
func (s *Sessions) Refusal(ctx context.Context) (kernel.Email, bool, error) {
	kept, err := s.of(ctx)
	if err != nil {
		return kernel.Email{}, false, err
	}
	return kept.refusal, kept.refused, nil
}

// LogIn renews the token of the session, which RenewalsOf counts, keeps the
// user, and forgets everything else.
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

// ProviderLoginOf returns the login through a provider the session of the
// browser keeps, and false when it keeps none.
func (s *Sessions) ProviderLoginOf(browser string) (domain.ProviderLogin, bool) {
	kept := s.ofBrowser(browser)
	return kept.providerLogin, kept.startedProviderLogin
}

// RefusalOf returns the e-mail the session of the browser keeps as the
// refusal of its last login through a provider, and false when it keeps
// none.
func (s *Sessions) RefusalOf(browser string) (kernel.Email, bool) {
	kept := s.ofBrowser(browser)
	return kept.refusal, kept.refused
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
