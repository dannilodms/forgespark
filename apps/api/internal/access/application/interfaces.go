// Package application holds the use cases of the access area and the
// interfaces they need: asking for a login code and the content of its
// e-mail, which the delivery of the e-mail area asks for; checking the code,
// which logs the user in; starting and concluding a login through a login
// provider, and the state of the login the login screen reads; who acts in a
// request, who is logged in, and logging out. A use case is a type built once, with what it needs, and run
// with Execute.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// Transactions opens and closes the transaction of a use case that writes.
type Transactions interface {
	// Run calls the operation inside one transaction, and passes it the
	// context every repository must receive to take part. The transaction is
	// committed when the operation returns nil and rolled back, with nothing
	// saved, when it returns an error. The use cases of other areas it calls
	// join it.
	Run(ctx context.Context, operation func(ctx context.Context) error) error
}

// Clock tells the current time. A use case never reads the time of the
// machine by itself, so a test decides what time it is.
type Clock interface {
	Now() time.Time
}

// LoginCodeRequestRepository keeps the login code requests.
type LoginCodeRequestRepository interface {
	// NextIdentifier reserves the identifier of a request that does not
	// exist yet.
	NextIdentifier(ctx context.Context) (domain.LoginCodeRequestID, error)
	// LoadNotEndedOf loads every request of the user that was not ended,
	// expired or not, and locks each one until the transaction ends. A user
	// has one at most.
	LoadNotEndedOf(ctx context.Context, user domain.UserID) ([]*domain.LoginCodeRequest, error)
	// CountAskedSince counts the requests of the user asked for after this
	// moment, ended or not. It locks nothing.
	CountAskedSince(ctx context.Context, user domain.UserID, since time.Time) (int, error)
	// Find returns the request with this identifier, ended or not, and false
	// when there is none. It locks nothing: it serves the delivery of the
	// e-mail of the code, which saves nothing of the request.
	Find(ctx context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error)
	// Load loads the request with this identifier, ended or not, and false
	// when there is none, and locks it until the transaction ends, so two
	// checks of its code count their attempts one after the other.
	Load(ctx context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error)
	// Save stores a request whole, whether it is new or changed.
	Save(ctx context.Context, request *domain.LoginCodeRequest) error
}

// ProviderLinkRepository keeps the provider links. A link is created at the
// first login through a provider and never changed, so nothing loads one to
// change it and nothing locks.
type ProviderLinkRepository interface {
	// Find returns the link of the account at the provider, and false when
	// the account was never linked.
	Find(ctx context.Context, provider domain.LoginProviderKey, subject domain.ProviderSubject) (*domain.ProviderLink, bool, error)
	// HasLink tells whether the user has an account of the provider linked
	// already: a user has one at each provider at most.
	HasLink(ctx context.Context, provider domain.LoginProviderKey, user domain.UserID) (bool, error)
	// Save stores a new link. A second link of the same account, or a second
	// link of the user at the same provider, fails.
	Save(ctx context.Context, link *domain.ProviderLink) error
}

// LoginProvider is a login provider turned on (ADR-0011): where the browser
// goes to log in through it, and who it says the person is when it sends them
// back. The `oidc` adapter is the one implementation.
type LoginProvider interface {
	// Key returns the key that names the provider.
	Key() domain.LoginProviderKey
	// DisplayName returns the name the login screen shows on its button.
	DisplayName() string
	// AuthorizationURL returns the address of the provider the browser is
	// sent to for this login, with its state, its nonce and the challenge of
	// its verifier. It fails when the provider cannot be reached.
	AuthorizationURL(ctx context.Context, login domain.ProviderLogin) (string, error)
	// Identify exchanges the code the provider sent the browser back with,
	// with the verifier of the login, validates the ID token it answers and
	// its nonce, and returns who the token says the person is. It fails when
	// the provider refuses the code, cannot be reached, or answers a token
	// that is not valid for this login.
	Identify(ctx context.Context, code string, login domain.ProviderLogin) (ProviderIdentity, error)
}

// ProviderIdentity is who a login provider says the person is: the account,
// and its e-mail when the provider says the e-mail is verified. An e-mail the
// provider does not say is verified is never given.
type ProviderIdentity struct {
	// Subject is the identifier the provider gives the account.
	Subject string
	// VerifiedEmail is the e-mail of the account, as the provider wrote it,
	// and empty when the provider does not say it is verified.
	VerifiedEmail string
}

// Session is the session of the browser that sent the request, which exists
// before the login: what the login keeps there between two requests of that
// browser, and, once it logs in, its user. The context of the request carries
// it. It takes no part in a transaction: what a use case writes in it stays.
type Session interface {
	// RememberLoginCodeRequest keeps the login code request this browser
	// asked for, in place of any earlier one, and forgets the attempts made
	// with no request. domain.NoLoginCodeRequest is kept for an e-mail of no
	// active user, the same way.
	RememberLoginCodeRequest(ctx context.Context, id domain.LoginCodeRequestID) error
	// LoginCodeRequest returns the login code request this browser asked for
	// last, and domain.NoLoginCodeRequest when it asked for none, or asked for
	// an e-mail of no active user.
	LoginCodeRequest(ctx context.Context) (domain.LoginCodeRequestID, error)
	// AttemptsWithoutRequest returns how many codes this browser typed with no
	// request to check them against since it last asked for one.
	AttemptsWithoutRequest(ctx context.Context) (int, error)
	// RememberAttemptsWithoutRequest keeps how many codes this browser typed
	// with no request.
	RememberAttemptsWithoutRequest(ctx context.Context, attempts int) error
	// RememberProviderLogin keeps the login through a provider this browser
	// started, in place of any earlier one.
	RememberProviderLogin(ctx context.Context, login domain.ProviderLogin) error
	// TakeProviderLogin returns the login through a provider this browser
	// started last, and false when it started none, and forgets it: a
	// callback uses it once.
	TakeProviderLogin(ctx context.Context) (domain.ProviderLogin, bool, error)
	// RememberRefusal keeps the verified e-mail the last login through a
	// provider in this browser was refused for, because it belongs to no
	// active user, in place of any earlier one.
	RememberRefusal(ctx context.Context, email kernel.Email) error
	// ForgetRefusal forgets the refusal of the last login through a
	// provider: the login that came after it was not refused for an e-mail.
	ForgetRefusal(ctx context.Context) error
	// Refusal returns the e-mail the last login through a provider in this
	// browser was refused for, and false when it was not refused. Reading it
	// forgets nothing.
	Refusal(ctx context.Context) (kernel.Email, bool, error)
	// LogIn renews the token of the session, so a token known before the
	// login opens nothing, keeps the user in it, and forgets what the login
	// kept: the login code request, the attempts, the login through a
	// provider and its refusal.
	LogIn(ctx context.Context, user domain.UserID) error
	// User returns the user logged in to the session, and false when nobody
	// is.
	User(ctx context.Context) (domain.UserID, bool, error)
	// End ends the session: what it kept is removed, and its cookie opens
	// nothing.
	End(ctx context.Context) error
}
