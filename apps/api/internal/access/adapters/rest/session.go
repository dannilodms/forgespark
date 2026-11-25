package rest

import (
	"context"
	"fmt"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/session"
)

// The keys of what the login keeps in the session of a browser.
const (
	// loginCodeRequestKey is the identifier of the login code request the
	// browser asked for last, domain.NoLoginCodeRequest included.
	loginCodeRequestKey = "login_code_request"
	// attemptsWithoutRequestKey is how many codes the browser typed with no
	// request to check them against.
	attemptsWithoutRequestKey = "login_code_attempts"
	// userKey is the identifier of the user logged in.
	userKey = "user"
	// The login through a provider the browser started: the key of the
	// provider, the state, the nonce, the verifier and the path to return
	// to.
	providerLoginProviderKey = "provider_login_provider"
	providerLoginStateKey    = "provider_login_state"
	providerLoginNonceKey    = "provider_login_nonce"
	providerLoginVerifierKey = "provider_login_verifier"
	providerLoginReturnToKey = "provider_login_return_to"
	// refusalKey is the e-mail the last login through a provider was
	// refused for.
	refusalKey = "provider_login_refused_email"
)

// providerLoginKeys are the keys of a login through a provider.
var providerLoginKeys = []string{
	providerLoginProviderKey, providerLoginStateKey, providerLoginNonceKey, providerLoginVerifierKey, providerLoginReturnToKey,
}

// Session is the session of the browser of a request, read and written
// through the sessions every request under the API loads: it implements the
// Session the use cases of the area declare. The context of the request
// carries the session.
type Session struct {
	sessions *session.Sessions
}

// NewSession returns the session of each request on these sessions.
func NewSession(sessions *session.Sessions) *Session {
	return &Session{sessions: sessions}
}

// RememberLoginCodeRequest keeps the request in the session, in place of any
// earlier one, and forgets the attempts made with no request. The session is
// saved with the answer, which sets its cookie.
func (s *Session) RememberLoginCodeRequest(ctx context.Context, id domain.LoginCodeRequestID) error {
	s.sessions.Put(ctx, loginCodeRequestKey, int64(id))
	s.sessions.Remove(ctx, attemptsWithoutRequestKey)
	return nil
}

// LoginCodeRequest returns the request the session keeps, and
// domain.NoLoginCodeRequest when it keeps none.
func (s *Session) LoginCodeRequest(ctx context.Context) (domain.LoginCodeRequestID, error) {
	return domain.LoginCodeRequestID(s.sessions.GetInt64(ctx, loginCodeRequestKey)), nil
}

// AttemptsWithoutRequest returns the codes typed with no request.
func (s *Session) AttemptsWithoutRequest(ctx context.Context) (int, error) {
	return s.sessions.GetInt(ctx, attemptsWithoutRequestKey), nil
}

// RememberAttemptsWithoutRequest keeps the codes typed with no request.
func (s *Session) RememberAttemptsWithoutRequest(ctx context.Context, attempts int) error {
	s.sessions.Put(ctx, attemptsWithoutRequestKey, attempts)
	return nil
}

// RememberProviderLogin keeps the login through a provider in the session,
// in place of any earlier one. The session is saved with the answer, which
// sets its cookie, before the browser leaves for the provider.
func (s *Session) RememberProviderLogin(ctx context.Context, login domain.ProviderLogin) error {
	s.sessions.Put(ctx, providerLoginProviderKey, login.Provider().String())
	s.sessions.Put(ctx, providerLoginStateKey, login.State())
	s.sessions.Put(ctx, providerLoginNonceKey, login.Nonce())
	s.sessions.Put(ctx, providerLoginVerifierKey, login.Verifier())
	s.sessions.Put(ctx, providerLoginReturnToKey, login.ReturnTo().String())
	return nil
}

// TakeProviderLogin returns the login through a provider the session keeps,
// and removes it from the session.
func (s *Session) TakeProviderLogin(ctx context.Context) (domain.ProviderLogin, bool, error) {
	stored := domain.StoredProviderLogin{
		Provider: s.sessions.PopString(ctx, providerLoginProviderKey),
		State:    s.sessions.PopString(ctx, providerLoginStateKey),
		Nonce:    s.sessions.PopString(ctx, providerLoginNonceKey),
		Verifier: s.sessions.PopString(ctx, providerLoginVerifierKey),
		ReturnTo: s.sessions.PopString(ctx, providerLoginReturnToKey),
	}
	if stored.Provider == "" {
		return domain.ProviderLogin{}, false, nil
	}
	login, err := domain.RestoreProviderLogin(stored)
	if err != nil {
		return domain.ProviderLogin{}, false, fmt.Errorf("read the login through a provider of the session: %w", err)
	}
	return login, true, nil
}

// RememberRefusal keeps the e-mail the last login through a provider was
// refused for.
func (s *Session) RememberRefusal(ctx context.Context, email kernel.Email) error {
	s.sessions.Put(ctx, refusalKey, email.String())
	return nil
}

// ForgetRefusal forgets the refusal of the last login through a provider.
func (s *Session) ForgetRefusal(ctx context.Context) error {
	s.sessions.Remove(ctx, refusalKey)
	return nil
}

// Refusal returns the e-mail the last login through a provider was refused
// for, and false when the session keeps none. A session that was never
// written is not created by the read.
func (s *Session) Refusal(ctx context.Context) (kernel.Email, bool, error) {
	kept := s.sessions.GetString(ctx, refusalKey)
	if kept == "" {
		return kernel.Email{}, false, nil
	}
	email, err := kernel.NewEmail(kept)
	if err != nil {
		return kernel.Email{}, false, fmt.Errorf("read the refusal of the session: %w", err)
	}
	return email, true, nil
}

// LogIn gives the session a new token, which the answer sets in its cookie,
// keeps the user, and forgets everything the login kept: the request, the
// attempts, the login through a provider and its refusal.
func (s *Session) LogIn(ctx context.Context, user domain.UserID) error {
	if err := s.sessions.RenewToken(ctx); err != nil {
		return fmt.Errorf("renew the token of the session: %w", err)
	}
	s.sessions.Remove(ctx, loginCodeRequestKey)
	s.sessions.Remove(ctx, attemptsWithoutRequestKey)
	for _, key := range providerLoginKeys {
		s.sessions.Remove(ctx, key)
	}
	s.sessions.Remove(ctx, refusalKey)
	s.sessions.Put(ctx, userKey, int64(user))
	return nil
}

// User returns the user logged in to the session.
func (s *Session) User(ctx context.Context) (domain.UserID, bool, error) {
	user := s.sessions.GetInt64(ctx, userKey)
	return domain.UserID(user), user > 0, nil
}

// End removes the session from the store, and the answer expires its cookie.
func (s *Session) End(ctx context.Context) error {
	if err := s.sessions.Destroy(ctx); err != nil {
		return fmt.Errorf("end the session: %w", err)
	}
	return nil
}
