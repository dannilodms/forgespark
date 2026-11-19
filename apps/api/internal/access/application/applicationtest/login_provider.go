package applicationtest

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

// Account is an account at a stand-in login provider: its subject, its
// e-mail, and whether the provider says the e-mail is verified.
type Account struct {
	Subject       string
	Email         string
	EmailVerified bool
}

// authorization is a login the stand-in was sent to: what the browser
// brought it, and the account the person logged in with.
type authorization struct {
	nonce     string
	challenge string
	account   Account
	used      bool
}

// LoginProvider stands in for a login provider: it remembers each login it
// is sent to, lets a test log a person in to one with an account, and gives
// the account back for its code, once, to the login with the verifier of
// the challenge and the nonce it was sent, as a provider and the validation
// of its ID token do.
type LoginProvider struct {
	key            domain.LoginProviderKey
	displayName    string
	authorizations map[string]*authorization
	codes          map[string]string
	unreachable    error
	failure        error
}

var _ application.LoginProvider = (*LoginProvider)(nil)

// NewLoginProvider returns a provider with this key and display name, sent
// to no login yet.
func NewLoginProvider(key domain.LoginProviderKey, displayName string) *LoginProvider {
	return &LoginProvider{
		key:            key,
		displayName:    displayName,
		authorizations: map[string]*authorization{},
		codes:          map[string]string{},
	}
}

// Key returns the key of the provider.
func (p *LoginProvider) Key() domain.LoginProviderKey {
	return p.key
}

// DisplayName returns the name the login screen shows.
func (p *LoginProvider) DisplayName() string {
	return p.displayName
}

// Unreachable makes every later call fail with this error, as a provider
// that does not answer.
func (p *LoginProvider) Unreachable(failure error) {
	p.unreachable = failure
}

// RefuseCodesWith makes every later exchange of a code fail with this error,
// as a provider that refuses the code or answers a token that is not valid.
func (p *LoginProvider) RefuseCodesWith(failure error) {
	p.failure = failure
}

// AuthorizationURL remembers the login and returns an address of the
// stand-in with its state, its nonce and the challenge of its verifier.
func (p *LoginProvider) AuthorizationURL(_ context.Context, login domain.ProviderLogin) (string, error) {
	if p.unreachable != nil {
		return "", p.unreachable
	}
	challenge := challengeOf(login.Verifier())
	p.authorizations[login.State()] = &authorization{nonce: login.Nonce(), challenge: challenge}
	query := url.Values{"state": {login.State()}, "nonce": {login.Nonce()}, "code_challenge": {challenge}}
	return "https://" + p.key.String() + ".provider.example/authorize?" + query.Encode(), nil
}

// LogInAs logs the person in at the provider, for the login sent there with
// this state, with the account, and returns the code the provider sends the
// browser back with.
func (p *LoginProvider) LogInAs(state string, account Account) (string, error) {
	sent, known := p.authorizations[state]
	if !known {
		return "", errors.New("the stand-in provider was sent no login with this state")
	}
	sent.account = account
	code := fmt.Sprintf("code-%d", len(p.codes)+1)
	p.codes[code] = state
	return code, nil
}

// Identify gives the account of the code, once, to the login with the
// verifier of the challenge it was sent and the nonce of its ID token.
func (p *LoginProvider) Identify(_ context.Context, code string, login domain.ProviderLogin) (application.ProviderIdentity, error) {
	if p.unreachable != nil {
		return application.ProviderIdentity{}, p.unreachable
	}
	if p.failure != nil {
		return application.ProviderIdentity{}, p.failure
	}
	sent, err := p.authorizationOf(code)
	if err != nil {
		return application.ProviderIdentity{}, err
	}
	if sent.challenge != challengeOf(login.Verifier()) {
		return application.ProviderIdentity{}, errors.New("the verifier is not the one of the challenge")
	}
	if sent.nonce != login.Nonce() {
		return application.ProviderIdentity{}, errors.New("the ID token carries another nonce")
	}
	sent.used = true
	identity := application.ProviderIdentity{Subject: sent.account.Subject}
	if sent.account.EmailVerified {
		identity.VerifiedEmail = sent.account.Email
	}
	return identity, nil
}

func (p *LoginProvider) authorizationOf(code string) (*authorization, error) {
	state, known := p.codes[code]
	if !known {
		return nil, errors.New("the stand-in provider gave no such code")
	}
	sent := p.authorizations[state]
	if sent.used {
		return nil, errors.New("the code was used")
	}
	return sent, nil
}

// challengeOf is the S256 challenge of a verifier (RFC 7636).
func challengeOf(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
