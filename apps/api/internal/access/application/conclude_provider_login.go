package application

import (
	"context"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	consultancyapplication "github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// ProviderLoginDependencies are what concluding a login through a provider
// needs: the providers turned on, the session of the browser, the provider
// links, the users who can log in by e-mail, the step that logs a user in,
// and the clock.
type ProviderLoginDependencies struct {
	Providers      *LoginProviders
	Session        Session
	Links          ProviderLinkRepository
	FindActiveUser *consultancyapplication.FindActiveUser
	LogIn          *LogIn
	Clock          Clock
}

// ConcludeProviderLogin concludes a login through a login provider when the
// provider sends the browser back (Login in CODING_STANDARDS.md, ADR-0011):
// it checks the state against the login the session keeps, has the provider
// identify the person from the code, finds the user by the provider link of
// the account or else by the verified e-mail, linking the account then, and
// logs the user in.
type ConcludeProviderLogin struct {
	with ProviderLoginDependencies
}

// NewConcludeProviderLogin builds the use case.
func NewConcludeProviderLogin(with ProviderLoginDependencies) *ConcludeProviderLogin {
	return &ConcludeProviderLogin{with: with}
}

// ConcludeProviderLoginInput is what the provider sent the browser back
// with: the key of the provider, from the path, the code and the state, and
// whether the provider answered an error, as it does when the person
// cancelled.
type ConcludeProviderLoginInput struct {
	Provider string
	Code     string
	State    string
	Failed   bool
}

// ProviderLoginOutcome is how a login through a provider ended.
type ProviderLoginOutcome int

const (
	// LoginNotConcluded is a login that logged nobody in for any reason but
	// the e-mail: a cancel or an error at the provider, a code it refused, an
	// e-mail it does not say is verified, a callback this browser did not
	// start or with another state.
	LoginNotConcluded ProviderLoginOutcome = iota
	// LoggedInThroughProvider is a login that logged the user in.
	LoggedInThroughProvider
	// NoActiveUser is a verified e-mail of no active user: unknown, or of a
	// deactivated user, refused the same way. The session keeps the e-mail
	// for the login screen.
	NoActiveUser
	// AnotherAccountLinked is an account that is not linked, whose verified
	// e-mail is the one of a user who has another account of the provider
	// linked: a user has one at each provider, so the person logs in with
	// that account, or with a login code.
	AnotherAccountLinked
)

// ConcludedProviderLogin is how the login ended, and the path of the app the
// browser returns to once logged in.
type ConcludedProviderLogin struct {
	Outcome  ProviderLoginOutcome
	ReturnTo domain.ReturnPath
}

// Execute concludes the login. It returns false when no provider with that
// key is turned on. The login the session keeps is used once, whatever the
// outcome. A login refused for an e-mail of no active user keeps that e-mail
// in the session; any other outcome forgets the refusal before it.
func (u *ConcludeProviderLogin) Execute(ctx context.Context, input ConcludeProviderLoginInput) (ConcludedProviderLogin, bool, error) {
	provider, found := u.with.Providers.find(input.Provider)
	if !found {
		return ConcludedProviderLogin{}, false, nil
	}
	login, started, err := u.with.Session.TakeProviderLogin(ctx)
	if err != nil {
		return ConcludedProviderLogin{}, true, err
	}
	if !started {
		return u.notConcluded(ctx, domain.RootPath)
	}
	if input.Failed || input.Code == "" || !login.Accepts(provider.Key(), input.State) {
		return u.notConcluded(ctx, login.ReturnTo())
	}
	identity, err := provider.Identify(ctx, input.Code, login)
	if err != nil {
		// The provider refused the code or could not be reached, and its
		// adapter tells why.
		return u.notConcluded(ctx, login.ReturnTo())
	}
	account, usable := accountOf(provider.Key(), identity)
	if !usable {
		return u.notConcluded(ctx, login.ReturnTo())
	}
	return u.logIn(ctx, account, login.ReturnTo())
}

// account is the account at the provider a login identified, with the
// e-mail the provider says is verified.
type account struct {
	provider domain.LoginProviderKey
	subject  domain.ProviderSubject
	email    kernel.Email
}

// accountOf reads the identity the provider gave, and false when it gives no
// subject or no verified e-mail the system can read.
func accountOf(provider domain.LoginProviderKey, identity ProviderIdentity) (account, bool) {
	subject, err := domain.NewProviderSubject(identity.Subject)
	if err != nil || identity.VerifiedEmail == "" {
		return account{}, false
	}
	email, err := kernel.NewEmail(identity.VerifiedEmail)
	if err != nil {
		return account{}, false
	}
	return account{provider: provider, subject: subject, email: email}, true
}

// logIn logs in the user of the account, through the step every way in ends
// with, and tells the outcome.
func (u *ConcludeProviderLogin) logIn(ctx context.Context, account account, returnTo domain.ReturnPath) (ConcludedProviderLogin, bool, error) {
	var outcome ProviderLoginOutcome
	loggedIn, err := u.with.LogIn.Execute(ctx, account.provider.WayIn(), func(ctx context.Context) (domain.UserID, bool, error) {
		user, userOutcome, err := u.userOf(ctx, account)
		outcome = userOutcome
		return user, userOutcome == LoggedInThroughProvider, err
	})
	if err != nil {
		return ConcludedProviderLogin{}, true, err
	}
	if loggedIn {
		return ConcludedProviderLogin{Outcome: LoggedInThroughProvider, ReturnTo: returnTo}, true, nil
	}
	if outcome != LoggedInThroughProvider && outcome != NoActiveUser {
		return u.refusedWithout(ctx, outcome, returnTo)
	}
	// No active user has the e-mail, or the user found by the link was
	// deactivated since: refused for the e-mail either way.
	if err := u.with.Session.RememberRefusal(ctx, account.email); err != nil {
		return ConcludedProviderLogin{}, true, err
	}
	return ConcludedProviderLogin{Outcome: NoActiveUser, ReturnTo: returnTo}, true, nil
}

// userOf finds the user of the account, inside the transaction of the login:
// by its link, or else the active user with its verified e-mail, linking the
// account to them. A user who has another account of the provider linked is
// known there by that one, and is not found by the e-mail.
func (u *ConcludeProviderLogin) userOf(ctx context.Context, account account) (domain.UserID, ProviderLoginOutcome, error) {
	link, linked, err := u.with.Links.Find(ctx, account.provider, account.subject)
	if err != nil {
		return 0, LoginNotConcluded, err
	}
	if linked {
		return link.User(), LoggedInThroughProvider, nil
	}
	user, found, err := u.with.FindActiveUser.Execute(ctx, account.email)
	if err != nil || !found {
		return 0, NoActiveUser, err
	}
	owner := domain.UserID(user.ID)
	hasLink, err := u.with.Links.HasLink(ctx, account.provider, owner)
	if err != nil || hasLink {
		return 0, AnotherAccountLinked, err
	}
	saved, err := u.with.Links.Save(ctx, domain.LinkProvider(account.provider, account.subject, owner, u.with.Clock.Now()))
	if err != nil {
		return 0, LoginNotConcluded, err
	}
	if !saved {
		return u.linkedMeanwhile(ctx, account)
	}
	return owner, LoggedInThroughProvider, nil
}

// linkedMeanwhile reads the link again when another login at the same moment
// linked the account, or another account of the user at the provider, while
// this one was linking it: the user of the account, or the user known there
// by another account.
func (u *ConcludeProviderLogin) linkedMeanwhile(ctx context.Context, account account) (domain.UserID, ProviderLoginOutcome, error) {
	link, linked, err := u.with.Links.Find(ctx, account.provider, account.subject)
	if err != nil || !linked {
		return 0, AnotherAccountLinked, err
	}
	return link.User(), LoggedInThroughProvider, nil
}

// notConcluded logs nobody in, and forgets the refusal of the login before.
func (u *ConcludeProviderLogin) notConcluded(ctx context.Context, returnTo domain.ReturnPath) (ConcludedProviderLogin, bool, error) {
	return u.refusedWithout(ctx, LoginNotConcluded, returnTo)
}

// refusedWithout logs nobody in for a reason that is not the e-mail, and
// forgets the refusal of the login before.
func (u *ConcludeProviderLogin) refusedWithout(ctx context.Context, outcome ProviderLoginOutcome, returnTo domain.ReturnPath) (ConcludedProviderLogin, bool, error) {
	if err := u.with.Session.ForgetRefusal(ctx); err != nil {
		return ConcludedProviderLogin{}, true, err
	}
	return ConcludedProviderLogin{Outcome: outcome, ReturnTo: returnTo}, true, nil
}
