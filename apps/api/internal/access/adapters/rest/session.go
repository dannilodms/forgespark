package rest

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/session"
)

// The keys of what the login keeps in the session of a browser.
const (
	// loginCodeRequestKey is the identifier of the login code request the
	// browser asked for last, domain.NoLoginCodeRequest included.
	loginCodeRequestKey = "login_code_request"
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
// earlier one. The session is saved with the answer, which sets its cookie.
func (s *Session) RememberLoginCodeRequest(ctx context.Context, id domain.LoginCodeRequestID) error {
	s.sessions.Put(ctx, loginCodeRequestKey, int64(id))
	return nil
}
