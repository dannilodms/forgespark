package rest

import (
	"context"
	"fmt"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/session"
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
)

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

// LogIn gives the session a new token, which the answer sets in its cookie,
// keeps the user, and forgets the request and the attempts.
func (s *Session) LogIn(ctx context.Context, user domain.UserID) error {
	if err := s.sessions.RenewToken(ctx); err != nil {
		return fmt.Errorf("renew the token of the session: %w", err)
	}
	s.sessions.Remove(ctx, loginCodeRequestKey)
	s.sessions.Remove(ctx, attemptsWithoutRequestKey)
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
