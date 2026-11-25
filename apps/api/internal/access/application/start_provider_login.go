package application

import (
	"context"
	"io"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
)

// StartProviderLogin starts a login through a login provider (Login in
// CODING_STANDARDS.md, ADR-0011): the browser navigates to it, and is sent to
// the provider, with a state, a nonce and the challenge of a PKCE verifier,
// which the session of the browser keeps with the path to return to.
type StartProviderLogin struct {
	providers *LoginProviders
	session   Session
	random    io.Reader
}

// NewStartProviderLogin builds the use case. The state, the nonce and the
// verifier are drawn from the random source.
func NewStartProviderLogin(providers *LoginProviders, session Session, random io.Reader) *StartProviderLogin {
	return &StartProviderLogin{providers: providers, session: session, random: random}
}

// StartProviderLoginInput names the provider, by its key, and the path of
// the app to return to once logged in, as the browser gave them.
type StartProviderLoginInput struct {
	Provider string
	ReturnTo string
}

// StartedProviderLogin is where the browser goes next: the address of the
// provider, or nowhere when the provider could not be reached, and the
// browser goes back to the login.
type StartedProviderLogin struct {
	AuthorizationURL string
	ReturnTo         domain.ReturnPath
}

// Execute starts the login. It returns false when no provider with that key
// is turned on. A path to return to that is not a path of the app returns to
// its root. A provider that cannot be reached starts nothing: the session
// keeps no login, and the answer has no address.
func (u *StartProviderLogin) Execute(ctx context.Context, input StartProviderLoginInput) (StartedProviderLogin, bool, error) {
	provider, found := u.providers.find(input.Provider)
	if !found {
		return StartedProviderLogin{}, false, nil
	}
	returnTo := domain.NewReturnPath(input.ReturnTo)
	login, err := domain.StartProviderLogin(provider.Key(), returnTo, u.random)
	if err != nil {
		return StartedProviderLogin{}, true, err
	}
	address, err := provider.AuthorizationURL(ctx, login)
	if err != nil {
		return StartedProviderLogin{ReturnTo: returnTo}, true, nil
	}
	if err := u.session.RememberProviderLogin(ctx, login); err != nil {
		return StartedProviderLogin{}, true, err
	}
	return StartedProviderLogin{AuthorizationURL: address, ReturnTo: returnTo}, true, nil
}
