package application

import (
	"context"
	"errors"
	"slices"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// LoginProviders are the login providers turned on by configuration, in the
// order the login screen shows them. The command that serves builds them once
// and gives them to the use cases of the login through a provider.
type LoginProviders struct {
	providers []LoginProvider
}

// NewLoginProviders holds the providers turned on, none included. Two
// providers with one key are refused: the key names a provider in its paths,
// its links and the audit log.
func NewLoginProviders(providers ...LoginProvider) (*LoginProviders, error) {
	keys := make([]domain.LoginProviderKey, 0, len(providers))
	for _, provider := range providers {
		keys = append(keys, provider.Key())
	}
	slices.Sort(keys)
	if len(slices.Compact(keys)) != len(providers) {
		return nil, errors.New("two login providers have one key")
	}
	return &LoginProviders{providers: slices.Clone(providers)}, nil
}

// find returns the provider turned on with this key, and false when none is,
// for a text that is no key too.
func (p *LoginProviders) find(text string) (LoginProvider, bool) {
	key, err := domain.NewLoginProviderKey(text)
	if err != nil {
		return nil, false
	}
	position := slices.IndexFunc(p.providers, func(provider LoginProvider) bool { return provider.Key() == key })
	if position < 0 {
		return nil, false
	}
	return p.providers[position], true
}

// TellLoginState tells the login screen what it needs before anyone is
// logged in: the login providers turned on, one button each, and the e-mail
// the last login through a provider in this browser was refused for, which
// the screen shows with its notice. The e-mail reaches the screen in the
// answer, so it travels in no address. It needs no session to exist.
type TellLoginState struct {
	providers *LoginProviders
	session   Session
}

// NewTellLoginState builds the use case.
func NewTellLoginState(providers *LoginProviders, session Session) *TellLoginState {
	return &TellLoginState{providers: providers, session: session}
}

// TurnedOnLoginProvider is a login provider as the login screen shows it.
type TurnedOnLoginProvider struct {
	Key         string
	DisplayName string
}

// LoginState is what the login screen needs before the login.
type LoginState struct {
	// Providers are the login providers turned on, in order.
	Providers []TurnedOnLoginProvider
	// RefusedEmail is the e-mail the last login through a provider in this
	// browser was refused for, when Refused is true.
	RefusedEmail kernel.Email
	Refused      bool
}

// Execute returns the state of the login of the browser. Reading it forgets
// nothing: the next login through a provider replaces the refusal.
func (u *TellLoginState) Execute(ctx context.Context) (LoginState, error) {
	email, refused, err := u.session.Refusal(ctx)
	if err != nil {
		return LoginState{}, err
	}
	return LoginState{Providers: u.providers.turnedOn(), RefusedEmail: email, Refused: refused}, nil
}

// turnedOn lists the providers as the login screen shows them, in order.
func (p *LoginProviders) turnedOn() []TurnedOnLoginProvider {
	turnedOn := make([]TurnedOnLoginProvider, 0, len(p.providers))
	for _, provider := range p.providers {
		turnedOn = append(turnedOn, TurnedOnLoginProvider{Key: provider.Key().String(), DisplayName: provider.DisplayName()})
	}
	return turnedOn
}
