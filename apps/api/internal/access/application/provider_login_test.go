package application_test

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/access/application"
	"github.com/dannilodms/forgespark/apps/api/internal/access/application/applicationtest"
	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	consultancyapplication "github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
)

// The account of Ana at the local provider, with the e-mail she is
// registered with, verified.
var anaAtTheProvider = applicationtest.Account{
	Subject:       "ana-at-the-local-provider",
	Email:         "Ana.Souza@Fabrica.example",
	EmailVerified: true,
}

// startFrom starts a login through the provider from the browser and returns
// where the browser is sent.
func (s system) startFrom(t *testing.T, browser, provider, returnTo string) application.StartedProviderLogin {
	t.Helper()
	ctx := applicationtest.InBrowser(context.Background(), browser)
	started, found, err := s.startProviderLogin.Execute(ctx, application.StartProviderLoginInput{Provider: provider, ReturnTo: returnTo})
	require.NoError(t, err)
	require.True(t, found, "the provider %s is not turned on", provider)
	return started
}

// stateOf reads the state the browser was sent to the provider with.
func stateOf(t *testing.T, started application.StartedProviderLogin) string {
	t.Helper()
	address, err := url.Parse(started.AuthorizationURL)
	require.NoError(t, err)
	return address.Query().Get("state")
}

// throughTheProvider starts a login from the browser, logs the person in at
// the local provider with the account, and brings the browser back to the
// callback with the code and the state.
func (s system) throughTheProvider(t *testing.T, browser string, account applicationtest.Account) application.ConcludedProviderLogin {
	t.Helper()
	state := stateOf(t, s.startFrom(t, browser, "local", "/projects"))
	code, err := s.localProvider.LogInAs(state, account)
	require.NoError(t, err)
	return s.callback(t, browser, application.ConcludeProviderLoginInput{Provider: "local", Code: code, State: state})
}

func (s system) callback(t *testing.T, browser string, input application.ConcludeProviderLoginInput) application.ConcludedProviderLogin {
	t.Helper()
	ctx := applicationtest.InBrowser(context.Background(), browser)
	concluded, found, err := s.concludeProviderLogin.Execute(ctx, input)
	require.NoError(t, err)
	require.True(t, found, "the provider %s is not turned on", input.Provider)
	return concluded
}

// registeredAna registers Ana, an administrator, with the e-mail of her
// account at the local provider, and returns her identifier.
func (s system) registeredAna(t *testing.T) domain.UserID {
	t.Helper()
	id, err := s.users.RegisterAdministrator("Ana Souza", "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, err)
	return domain.UserID(id)
}

func (s system) deactivate(t *testing.T, user domain.UserID) {
	t.Helper()
	require.NoError(t, s.users.DeactivateUser(consultancyapplication.UserID(user)))
}

// requireNotConcluded checks the browser is sent back to the login with the
// notice that the login did not conclude, with nobody logged in, nothing
// linked and nothing recorded.
func (s system) requireNotConcluded(t *testing.T, concluded application.ConcludedProviderLogin) {
	t.Helper()
	assert.Equal(t, application.LoginNotConcluded, concluded.Outcome)
	s.requireNobodyLoggedIn(t, theBrowser)
	assert.Empty(t, s.links.All(), "an account was linked")
	_, refused := s.sessions.RefusalOf(theBrowser)
	assert.False(t, refused, "the session keeps a refusal")
}

func TestStartingALoginThroughAProviderKeepsItInTheSessionAndSendsTheBrowserToTheProvider(t *testing.T) {
	s := newSystem(t)

	started := s.startFrom(t, theBrowser, "local", "/projects?tab=files")

	login, kept := s.sessions.ProviderLoginOf(theBrowser)
	require.True(t, kept, "the session keeps no login")
	assert.Equal(t, "local", login.Provider().String())
	assert.Equal(t, "/projects?tab=files", login.ReturnTo().String())
	address, err := url.Parse(started.AuthorizationURL)
	require.NoError(t, err)
	assert.Equal(t, "local.provider.example", address.Host)
	assert.Equal(t, login.State(), address.Query().Get("state"))
	assert.Equal(t, login.Nonce(), address.Query().Get("nonce"))
	assert.NotEmpty(t, address.Query().Get("code_challenge"))
	assert.NotContains(t, started.AuthorizationURL, login.Verifier(), "the verifier left the system")
}

func TestTwoLoginsStartedGetTheirOwnStateNonceAndVerifier(t *testing.T) {
	s := newSystem(t)

	first := s.startFrom(t, theBrowser, "local", "/")
	firstLogin, _ := s.sessions.ProviderLoginOf(theBrowser)
	second := s.startFrom(t, theBrowser, "local", "/")
	secondLogin, _ := s.sessions.ProviderLoginOf(theBrowser)

	assert.NotEqual(t, stateOf(t, first), stateOf(t, second))
	assert.NotEqual(t, firstLogin.Nonce(), secondLogin.Nonce())
	assert.NotEqual(t, firstLogin.Verifier(), secondLogin.Verifier())
}

func TestAPathToReturnToThatIsNotOfTheAppReturnsToTheRoot(t *testing.T) {
	s := newSystem(t)

	s.startFrom(t, theBrowser, "local", "https://elsewhere.example/")

	login, _ := s.sessions.ProviderLoginOf(theBrowser)
	assert.Equal(t, domain.RootPath, login.ReturnTo())
}

func TestAProviderThatIsNotTurnedOnIsNotFound(t *testing.T) {
	s := newSystem(t)
	ctx := applicationtest.InBrowser(context.Background(), theBrowser)

	_, startFound, startErr := s.startProviderLogin.Execute(ctx, application.StartProviderLoginInput{Provider: "google", ReturnTo: "/"})
	_, callbackFound, callbackErr := s.concludeProviderLogin.Execute(ctx, application.ConcludeProviderLoginInput{Provider: "Not A Key"})

	require.NoError(t, startErr)
	require.NoError(t, callbackErr)
	assert.False(t, startFound)
	assert.False(t, callbackFound)
	_, kept := s.sessions.ProviderLoginOf(theBrowser)
	assert.False(t, kept)
}

// A provider that does not answer sends the browser back to the login, and
// the session keeps no login to come back to.
func TestAProviderThatCannotBeReachedStartsNoLogin(t *testing.T) {
	s := newSystem(t)
	s.localProvider.Unreachable(errors.New("dial tcp: connection refused"))

	started := s.startFrom(t, theBrowser, "local", "/projects")

	assert.Empty(t, started.AuthorizationURL)
	_, kept := s.sessions.ProviderLoginOf(theBrowser)
	assert.False(t, kept)
}

func TestTheFirstLoginThroughAProviderFindsTheUserByTheVerifiedEmailAndLinksTheAccount(t *testing.T) {
	s := newSystem(t)
	ana := s.registeredAna(t)
	s.clock.Advance(3 * time.Minute)

	concluded := s.throughTheProvider(t, theBrowser, anaAtTheProvider)

	assert.Equal(t, application.LoggedInThroughProvider, concluded.Outcome)
	assert.Equal(t, "/projects", concluded.ReturnTo.String())
	user, loggedIn := s.sessions.UserOf(theBrowser)
	assert.True(t, loggedIn)
	assert.Equal(t, ana, user)
	assert.Equal(t, 1, s.sessions.RenewalsOf(theBrowser), "the token of the session was not renewed at the login")
	_, kept := s.sessions.ProviderLoginOf(theBrowser)
	assert.False(t, kept, "the session keeps the login it used")
	links := s.links.All()
	require.Len(t, links, 1)
	assert.Equal(t, "local", links[0].Provider().String())
	assert.Equal(t, "ana-at-the-local-provider", links[0].Subject().String())
	assert.Equal(t, ana, links[0].User())
	assert.Equal(t, noon.Add(3*time.Minute), links[0].LinkedAt())
	entries := s.auditLog.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "logged_in", string(entries[0].Operation()))
	assert.Equal(t, noon.Add(3*time.Minute), entries[0].OccurredAt())
	actor, _ := entries[0].Actor().Identifier()
	assert.Equal(t, int64(ana), actor)
	assert.Equal(t, int64(ana), entries[0].Subject().Identifier())
	wayIn, _ := entries[0].WayIn()
	assert.Equal(t, "local", string(wayIn))
}

// Once linked, the account logs its user in whatever e-mail it has: an
// e-mail that changes hands at the provider does not log someone else in.
func TestALaterLoginFindsTheUserByTheLinkAndCreatesNoOtherLink(t *testing.T) {
	s := newSystem(t)
	ana := s.registeredAna(t)
	s.throughTheProvider(t, theBrowser, anaAtTheProvider)
	_, err := s.users.RegisterUser("bruno.lima@fabrica.example", "pt-BR")
	require.NoError(t, err)
	renamed := anaAtTheProvider
	renamed.Email = "bruno.lima@fabrica.example"

	concluded := s.throughTheProvider(t, anotherBrowser, renamed)

	assert.Equal(t, application.LoggedInThroughProvider, concluded.Outcome)
	user, loggedIn := s.sessions.UserOf(anotherBrowser)
	assert.True(t, loggedIn)
	assert.Equal(t, ana, user)
	assert.Len(t, s.links.All(), 1, "the account was linked again")
	assert.Len(t, s.auditLog.Entries(), 2)
}

// The e-mail of an account that someone else got at the provider, once the
// user has an account of that provider linked, logs nobody in: the user is
// known there by the account they linked. The login screen tells them so,
// since trying again with that account never works (story 17).
func TestAnotherAccountWithTheEmailOfAUserLinkedAtTheProviderLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	s.throughTheProvider(t, anotherBrowser, anaAtTheProvider)
	someoneElse := anaAtTheProvider
	someoneElse.Subject = "someone-else-with-the-email-of-ana"
	entriesBefore := len(s.auditLog.Entries())

	concluded := s.throughTheProvider(t, theBrowser, someoneElse)

	assert.Equal(t, application.AnotherAccountLinked, concluded.Outcome)
	assert.Equal(t, "/projects", concluded.ReturnTo.String())
	_, loggedIn := s.sessions.UserOf(theBrowser)
	assert.False(t, loggedIn)
	assert.Len(t, s.links.All(), 1)
	assert.Len(t, s.auditLog.Entries(), entriesBefore)
	_, refused := s.sessions.RefusalOf(theBrowser)
	assert.False(t, refused, "the session keeps a refusal for an e-mail")
}

// Two first logins of one account at the same moment both find it unlinked;
// the second to link it finds the link the first saved meanwhile, and logs
// the same user in (ticket 07).
func TestTwoFirstLoginsOfOneAccountAtTheSameMomentLogTheUserIn(t *testing.T) {
	s := newSystem(t)
	ana := s.registeredAna(t)
	s.links.LinkMeanwhile(domain.LinkProvider("local", "ana-at-the-local-provider", ana, noon))

	concluded := s.throughTheProvider(t, theBrowser, anaAtTheProvider)

	assert.Equal(t, application.LoggedInThroughProvider, concluded.Outcome)
	user, loggedIn := s.sessions.UserOf(theBrowser)
	assert.True(t, loggedIn)
	assert.Equal(t, ana, user)
	assert.Len(t, s.links.All(), 1)
	assert.Len(t, s.auditLog.Entries(), 1)
}

// A login that would link an account when another account of the provider
// was linked to the user meanwhile logs nobody in, as when it was linked
// before.
func TestAFirstLoginWhileAnotherAccountOfTheUserIsLinkedMeanwhileLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	ana := s.registeredAna(t)
	s.links.LinkMeanwhile(domain.LinkProvider("local", "another-account-of-ana", ana, noon))

	concluded := s.throughTheProvider(t, theBrowser, anaAtTheProvider)

	assert.Equal(t, application.AnotherAccountLinked, concluded.Outcome)
	s.requireNobodyLoggedIn(t, theBrowser)
	assert.Len(t, s.links.All(), 1)
	assert.Empty(t, s.auditLog.Entries())
}

func TestAVerifiedEmailOfNoActiveUserIsRefusedAndTheSessionKeepsIt(t *testing.T) {
	cases := map[string]func(t *testing.T, s system){
		"an e-mail of no user": func(*testing.T, system) {},
		"an e-mail of a deactivated user": func(t *testing.T, s system) {
			s.deactivate(t, s.registeredAna(t))
		},
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			prepare(t, s)

			concluded := s.throughTheProvider(t, theBrowser, anaAtTheProvider)

			assert.Equal(t, application.NoActiveUser, concluded.Outcome)
			email, refused := s.sessions.RefusalOf(theBrowser)
			assert.True(t, refused)
			assert.Equal(t, "ana.souza@fabrica.example", email.String())
			s.requireNobodyLoggedIn(t, theBrowser)
			assert.Empty(t, s.links.All(), "an account of no active user was linked")
		})
	}
}

// A user deactivated since their account was linked is refused as an e-mail
// of no active user, with the e-mail the provider gave.
func TestALinkedAccountOfADeactivatedUserIsRefusedForItsEmail(t *testing.T) {
	s := newSystem(t)
	ana := s.registeredAna(t)
	s.throughTheProvider(t, anotherBrowser, anaAtTheProvider)
	s.deactivate(t, ana)
	entriesBefore := len(s.auditLog.Entries())

	concluded := s.throughTheProvider(t, theBrowser, anaAtTheProvider)

	assert.Equal(t, application.NoActiveUser, concluded.Outcome)
	email, refused := s.sessions.RefusalOf(theBrowser)
	assert.True(t, refused)
	assert.Equal(t, "ana.souza@fabrica.example", email.String())
	_, loggedIn := s.sessions.UserOf(theBrowser)
	assert.False(t, loggedIn)
	assert.Len(t, s.auditLog.Entries(), entriesBefore)
}

func TestAnEmailTheProviderDoesNotSayIsVerifiedLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	unverified := anaAtTheProvider
	unverified.EmailVerified = false

	concluded := s.throughTheProvider(t, theBrowser, unverified)

	s.requireNotConcluded(t, concluded)
}

func TestAnAccountWithNoUsableEmailOrSubjectLogsNobodyIn(t *testing.T) {
	for name, account := range map[string]applicationtest.Account{
		"an e-mail that is not an address": {Subject: "ana-at-the-local-provider", Email: "ana souza", EmailVerified: true},
		"no subject":                       {Subject: "", Email: anaAtTheProvider.Email, EmailVerified: true},
	} {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			s.registeredAna(t)

			concluded := s.throughTheProvider(t, theBrowser, account)

			s.requireNotConcluded(t, concluded)
		})
	}
}

func TestACancelOrAnErrorAtTheProviderLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	state := stateOf(t, s.startFrom(t, theBrowser, "local", "/projects"))

	concluded := s.callback(t, theBrowser, application.ConcludeProviderLoginInput{Provider: "local", State: state, Failed: true})

	s.requireNotConcluded(t, concluded)
	assert.Equal(t, "/projects", concluded.ReturnTo.String(), "the path to return to was lost")
}

func TestACodeTheProviderRefusesLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	s.localProvider.RefuseCodesWith(errors.New(`the provider refused the code: 400 invalid_grant`))

	concluded := s.throughTheProvider(t, theBrowser, anaAtTheProvider)

	s.requireNotConcluded(t, concluded)
}

// A callback that brings another state was not sent by the provider for this
// browser: a link crafted elsewhere. It logs nobody in, and uses up the login
// it pretended to conclude.
func TestACallbackWithAnotherStateLogsNobodyInAndEndsTheLogin(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	state := stateOf(t, s.startFrom(t, theBrowser, "local", "/projects"))
	code, err := s.localProvider.LogInAs(state, anaAtTheProvider)
	require.NoError(t, err)

	forged := s.callback(t, theBrowser, application.ConcludeProviderLoginInput{Provider: "local", Code: code, State: "a-state-made-up-elsewhere"})
	afterwards := s.callback(t, theBrowser, application.ConcludeProviderLoginInput{Provider: "local", Code: code, State: state})

	s.requireNotConcluded(t, forged)
	s.requireNotConcluded(t, afterwards)
}

// The code of a person who logged in at the provider from another browser
// does not log this browser in: the state is in the session of the browser
// that started.
func TestACallbackInABrowserThatStartedNoLoginLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	state := stateOf(t, s.startFrom(t, anotherBrowser, "local", "/projects"))
	code, err := s.localProvider.LogInAs(state, anaAtTheProvider)
	require.NoError(t, err)

	concluded := s.callback(t, theBrowser, application.ConcludeProviderLoginInput{Provider: "local", Code: code, State: state})

	s.requireNotConcluded(t, concluded)
	assert.Equal(t, domain.RootPath, concluded.ReturnTo)
}

func TestACallbackUsedOnceLogsNobodyInTheSecondTime(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	state := stateOf(t, s.startFrom(t, theBrowser, "local", "/projects"))
	code, err := s.localProvider.LogInAs(state, anaAtTheProvider)
	require.NoError(t, err)
	input := application.ConcludeProviderLoginInput{Provider: "local", Code: code, State: state}
	first := s.callback(t, theBrowser, input)
	require.Equal(t, application.LoggedInThroughProvider, first.Outcome)

	again := s.callback(t, anotherBrowser, input)

	assert.Equal(t, application.LoginNotConcluded, again.Outcome)
	_, loggedIn := s.sessions.UserOf(anotherBrowser)
	assert.False(t, loggedIn)
}

func TestACallbackOfAnotherProviderThanTheOneStartedLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	state := stateOf(t, s.startFrom(t, theBrowser, "local", "/projects"))
	code, err := s.localProvider.LogInAs(state, anaAtTheProvider)
	require.NoError(t, err)

	concluded := s.callback(t, theBrowser, application.ConcludeProviderLoginInput{Provider: "another", Code: code, State: state})

	s.requireNotConcluded(t, concluded)
}

// A login that cannot be recorded logs nobody in, and links nothing: the
// link and the entry are in one transaction.
func TestALoginThroughAProviderThatCannotBeRecordedLinksNothingAndLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.registeredAna(t)
	s.auditLog.FailWith(errors.New("the audit log cannot be written"))
	state := stateOf(t, s.startFrom(t, theBrowser, "local", "/projects"))
	code, err := s.localProvider.LogInAs(state, anaAtTheProvider)
	require.NoError(t, err)
	ctx := applicationtest.InBrowser(context.Background(), theBrowser)

	_, _, err = s.concludeProviderLogin.Execute(ctx, application.ConcludeProviderLoginInput{Provider: "local", Code: code, State: state})

	require.Error(t, err)
	assert.Empty(t, s.links.All(), "the link was kept")
	_, loggedIn := s.sessions.UserOf(theBrowser)
	assert.False(t, loggedIn)
}

// The next login through a provider replaces the refusal of the one before:
// a login that is not refused for an e-mail forgets it, and a login forgets
// it too, whatever its way in.
func TestTheNextLoginReplacesTheRefusal(t *testing.T) {
	s := newSystem(t)
	s.throughTheProvider(t, theBrowser, anaAtTheProvider)
	_, refused := s.sessions.RefusalOf(theBrowser)
	require.True(t, refused)
	unverified := anaAtTheProvider
	unverified.EmailVerified = false

	s.throughTheProvider(t, theBrowser, unverified)

	_, refused = s.sessions.RefusalOf(theBrowser)
	assert.False(t, refused, "a login that did not conclude kept the refusal before it")
	s.throughTheProvider(t, theBrowser, anaAtTheProvider)
	s.registeredAna(t)
	s.throughTheProvider(t, theBrowser, anaAtTheProvider)
	_, refused = s.sessions.RefusalOf(theBrowser)
	assert.False(t, refused, "the login kept the refusal before it")
}

func TestTheLoginStateTellsTheProvidersTurnedOnAndTheRefusalOfThisBrowser(t *testing.T) {
	s := newSystem(t)
	s.throughTheProvider(t, theBrowser, anaAtTheProvider)
	ctx := applicationtest.InBrowser(context.Background(), theBrowser)

	state, err := s.tellLoginState.Execute(ctx)
	again, againErr := s.tellLoginState.Execute(ctx)
	elsewhere, elsewhereErr := s.tellLoginState.Execute(applicationtest.InBrowser(context.Background(), anotherBrowser))

	require.NoError(t, err)
	require.NoError(t, againErr)
	require.NoError(t, elsewhereErr)
	assert.Equal(t, []application.TurnedOnLoginProvider{
		{Key: "local", DisplayName: "Local login provider"},
		{Key: "another", DisplayName: "Another provider"},
	}, state.Providers)
	assert.True(t, state.Refused)
	assert.Equal(t, "ana.souza@fabrica.example", state.RefusedEmail.String())
	assert.Equal(t, state, again, "reading the state forgot the refusal")
	assert.False(t, elsewhere.Refused, "another browser reads the refusal of this one")
}

func TestTwoLoginProvidersWithOneKeyAreRefused(t *testing.T) {
	_, err := application.NewLoginProviders(
		applicationtest.NewLoginProvider("local", "One"),
		applicationtest.NewLoginProvider("local", "Two"),
	)

	assert.Error(t, err)
}
