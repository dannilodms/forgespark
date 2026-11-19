package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// inTheBrowser is the context of a request of the browser of the test.
func inTheBrowser() context.Context {
	return applicationtest.InBrowser(context.Background(), theBrowser)
}

// loggedIn logs the user in to the browser of the test with the right code.
func (s system) loggedIn(t *testing.T) domain.UserID {
	t.Helper()
	ana := s.administratorWhoAsked(t)
	require.NoError(t, s.check(codeOfTheFirstRequest))
	return ana
}

func TestTheActorOfASessionIsItsUserWithTheScopesOfTheirRoles(t *testing.T) {
	s := newSystem(t)
	ana := s.loggedIn(t)
	bruno, err := s.users.RegisterUser("bruno.lima@fabrica.example", "en-US")
	require.NoError(t, err)
	brunosBrowser := applicationtest.InBrowser(context.Background(), "the browser of Bruno")
	require.NoError(t, s.sessions.LogIn(brunosBrowser, domain.UserID(bruno)))

	administrator, administratorFound, err := s.findActor.Execute(inTheBrowser())
	require.NoError(t, err)
	developer, developerFound, err := s.findActor.Execute(brunosBrowser)
	require.NoError(t, err)

	require.True(t, administratorFound)
	assert.Equal(t, int64(ana), administrator.User())
	assert.Equal(t, kernel.EveryScope(), administrator.Scopes(), "the administrator role grants every scope")
	require.True(t, developerFound)
	assert.Equal(t, int64(bruno), developer.User())
	assert.Empty(t, developer.Scopes(), "the developer role grants a scope")
}

func TestASessionWithNoUserHasNoActor(t *testing.T) {
	s := newSystem(t)
	require.NoError(t, s.ask("bruno.lima@fabrica.example"))

	_, asked, err := s.findActor.Execute(inTheBrowser())
	require.NoError(t, err)
	_, never, err := s.findActor.Execute(applicationtest.InBrowser(context.Background(), anotherBrowser))
	require.NoError(t, err)

	assert.False(t, asked, "a session before the login has an actor")
	assert.False(t, never, "a browser with no session has an actor")
	_, kept := s.sessions.LoginCodeRequestOf(theBrowser)
	assert.True(t, kept, "the session before the login was ended")
}

// Deactivating a user takes effect at their next request: the session is
// ended, and nobody acts with it.
func TestTheSessionOfADeactivatedUserOrOfAUserThatIsGoneIsEnded(t *testing.T) {
	cases := map[string]func(t *testing.T, s system){
		"a deactivated user": func(t *testing.T, s system) {
			ana := s.loggedIn(t)
			require.NoError(t, s.users.DeactivateUser(consultancyapplication.UserID(ana)))
		},
		"a user that does not exist": func(t *testing.T, s system) {
			require.NoError(t, s.sessions.LogIn(inTheBrowser(), domain.UserID(37)))
		},
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			prepare(t, s)

			_, found, err := s.findActor.Execute(inTheBrowser())

			require.NoError(t, err)
			assert.False(t, found)
			assert.True(t, s.sessions.Ended(theBrowser), "the session was not ended")
			_, loggedIn := s.sessions.UserOf(theBrowser)
			assert.False(t, loggedIn)
		})
	}
}

func TestTheLoggedInUserIsTheirNameEmailPreferredLanguageAndRoles(t *testing.T) {
	s := newSystem(t)
	s.loggedIn(t)
	actor, _, err := s.findActor.Execute(inTheBrowser())
	require.NoError(t, err)

	user, found, err := s.tellLoggedInUser.Execute(inTheBrowser(), actor)

	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "Ana Souza", user.Name)
	assert.Equal(t, "ana.souza@fabrica.example", user.Email.String())
	assert.Equal(t, "pt-BR", user.PreferredLanguage.String())
	assert.Equal(t, []string{"administrator"}, user.Roles)
}

func TestTheLoggedInUserIsNotFoundOnceDeactivated(t *testing.T) {
	s := newSystem(t)
	ana := s.loggedIn(t)
	actor, _, err := s.findActor.Execute(inTheBrowser())
	require.NoError(t, err)
	require.NoError(t, s.users.DeactivateUser(consultancyapplication.UserID(ana)))

	_, found, err := s.tellLoggedInUser.Execute(inTheBrowser(), actor)

	require.NoError(t, err)
	assert.False(t, found)
}

func TestLoggingOutEndsTheSession(t *testing.T) {
	s := newSystem(t)
	s.loggedIn(t)

	err := s.logOut.Execute(inTheBrowser())

	require.NoError(t, err)
	assert.True(t, s.sessions.Ended(theBrowser))
	_, found, err := s.findActor.Execute(inTheBrowser())
	require.NoError(t, err)
	assert.False(t, found, "the user still acts after logging out")
	assert.Len(t, s.auditLog.Entries(), 1, "the logout was recorded")
}

func TestLoggingOutWithNoUserLoggedInEndsTheSessionToo(t *testing.T) {
	s := newSystem(t)
	require.NoError(t, s.ask("bruno.lima@fabrica.example"))

	err := s.logOut.Execute(inTheBrowser())

	require.NoError(t, err)
	assert.True(t, s.sessions.Ended(theBrowser))
	_, kept := s.sessions.LoginCodeRequestOf(theBrowser)
	assert.False(t, kept)
}

// The in-memory sessions stand in for the session of the REST adapter.
var _ application.Session = (*applicationtest.Sessions)(nil)
