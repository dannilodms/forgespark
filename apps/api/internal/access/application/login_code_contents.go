package application

import (
	"context"
	"fmt"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
)

// LoginCodeContents gives the content of the e-mail of a login code when the
// e-mail is delivered: the code, computed from its request, and how long it
// works. The code is in no table and no job argument, so this is where it
// comes to exist, for the e-mail alone. The worker wires it as the content
// source of the template of the login code.
type LoginCodeContents struct {
	requests LoginCodeRequestRepository
	key      domain.LoginCodeKey
}

// NewLoginCodeContents builds the source. The key comes from configuration,
// and has at least domain.LoginCodeKeyMinSize bytes.
func NewLoginCodeContents(requests LoginCodeRequestRepository, key []byte) (*LoginCodeContents, error) {
	held, err := domain.NewLoginCodeKey(key)
	if err != nil {
		return nil, err
	}
	return &LoginCodeContents{requests: requests, key: held}, nil
}

// ContentOf gives the content of the e-mail that refers to a login code
// request. The code is given whether or not the request still works: the
// e-mail leaves within seconds of the request, and a code that no longer
// works is refused when it is checked.
func (s *LoginCodeContents) ContentOf(ctx context.Context, reference emailapplication.Reference) (emailapplication.Content, error) {
	// A reference of another kind differs from the login code request of
	// the same identifier.
	if reference != emailapplication.LoginCodeRequest(reference.Identifier()) {
		return nil, fmt.Errorf("the e-mail of a login code refers to a %s, not to a login code request", reference.Kind())
	}
	request, found, err := s.requests.Find(ctx, domain.LoginCodeRequestID(reference.Identifier()))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no login code request %d exists", reference.Identifier())
	}
	return emailapplication.LoginCodeContent{
		Code:     request.Code(s.key).Reveal(),
		Lifetime: domain.LoginCodeLifetime,
	}, nil
}
