package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/access/application"
	"github.com/dannilodms/forgespark/apps/api/internal/access/application/applicationtest"
	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	consultancyapplication "github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	emailapplication "github.com/dannilodms/forgespark/apps/api/internal/email/application"
)

// codeOfTheFirstRequest is the code of the first request of a test, computed
// apart with Python's hmac module from the key of these tests and the random
// value countingBytes gives it (see
// TestTheDeliveryOfTheEmailGivesTheCodeOfItsRequestAndHowLongItWorks).
const codeOfTheFirstRequest = "774421"

// anotherBrowser is a browser that is not the one of the test.
const anotherBrowser = "a browser of someone else"

// check types the code in the browser of the test.
func (s system) check(code string) error {
	return s.checkFrom(theBrowser, code)
}

func (s system) checkFrom(browser, code string) error {
	ctx := applicationtest.InBrowser(context.Background(), browser)
	return s.checkLoginCode.Execute(ctx, application.CheckLoginCodeInput{Code: code})
}

// administrator registers Ana, an administrator, who asks for a login code
// from the browser of the test, and returns her identifier.
func (s system) administratorWhoAsked(t *testing.T) domain.UserID {
	t.Helper()
	id, err := s.users.RegisterAdministrator("Ana Souza", "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, err)
	require.NoError(t, s.ask("ana.souza@fabrica.example"))
	return domain.UserID(id)
}

func requireInvalid(t *testing.T, err error, attemptsLeft int) {
	t.Helper()
	var invalid domain.LoginCodeInvalidError
	require.ErrorAs(t, err, &invalid)
	assert.Equal(t, attemptsLeft, invalid.AttemptsLeft)
}

func requireEnded(t *testing.T, err error) {
	t.Helper()
	require.ErrorIs(t, err, domain.LoginCodeEndedError{})
}

// requireNobodyLoggedIn checks that no session holds a user and that the audit
// log recorded no login.
func (s system) requireNobodyLoggedIn(t *testing.T, browsers ...string) {
	t.Helper()
	for _, browser := range browsers {
		_, loggedIn := s.sessions.UserOf(browser)
		assert.False(t, loggedIn, "a user is logged in to %s", browser)
		assert.Zero(t, s.sessions.RenewalsOf(browser), "the token of %s was renewed", browser)
	}
	assert.Empty(t, s.auditLog.Entries(), "a login was recorded")
}

// The code of the delivery is the code of the check: both compute it from the
// request, and neither stores it.
func TestTheRightCodeLogsTheUserInToTheBrowserThatAskedForIt(t *testing.T) {
	s := newSystem(t)
	ana := s.administratorWhoAsked(t)
	delivered, err := s.loginCodes.ContentOf(context.Background(), emailapplication.LoginCodeRequest(1))
	require.NoError(t, err)
	require.Equal(t, emailapplication.LoginCodeContent{Code: codeOfTheFirstRequest, Lifetime: 10 * time.Minute}, delivered)
	s.clock.Advance(4 * time.Minute)

	err = s.check(codeOfTheFirstRequest)

	require.NoError(t, err)
	user, loggedIn := s.sessions.UserOf(theBrowser)
	assert.True(t, loggedIn)
	assert.Equal(t, ana, user)
	assert.Equal(t, 1, s.sessions.RenewalsOf(theBrowser), "the token of the session was not renewed at the login")
	_, keeps := s.sessions.LoginCodeRequestOf(theBrowser)
	assert.False(t, keeps, "the session keeps the request it used")
	request := s.requests.All()[0]
	endedAt, ended := request.EndedAt()
	assert.True(t, ended, "the code works more than once")
	assert.Equal(t, noon.Add(4*time.Minute), endedAt)
	entries := s.auditLog.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "logged_in", string(entries[0].Operation()))
	assert.Equal(t, noon.Add(4*time.Minute), entries[0].OccurredAt())
	assert.Equal(t, "user", string(entries[0].Actor().Kind()))
	actor, _ := entries[0].Actor().Identifier()
	assert.Equal(t, int64(ana), actor)
	assert.Equal(t, int64(ana), entries[0].Subject().Identifier())
	wayIn, _ := entries[0].WayIn()
	assert.Equal(t, "login_code", string(wayIn))

	again := s.check(codeOfTheFirstRequest)
	requireInvalid(t, again, 2)
	assert.Len(t, s.auditLog.Entries(), 1, "a used code logged in again")
}

func TestAWrongCodeThreeTimesEndsTheCode(t *testing.T) {
	s := newSystem(t)
	s.administratorWhoAsked(t)

	first := s.check("000000")
	second := s.check("774420")
	third := s.check("123456")
	afterwards := s.check(codeOfTheFirstRequest)

	requireInvalid(t, first, 2)
	requireInvalid(t, second, 1)
	requireEnded(t, third)
	requireEnded(t, afterwards)
	request := s.requests.All()[0]
	assert.Equal(t, 3, request.Attempts())
	_, ended := request.EndedAt()
	assert.True(t, ended)
	s.requireNobodyLoggedIn(t, theBrowser)
}

func TestACodeTenMinutesOldLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.administratorWhoAsked(t)
	s.clock.Advance(10 * time.Minute)

	err := s.check(codeOfTheFirstRequest)

	requireInvalid(t, err, 2)
	s.requireNobodyLoggedIn(t, theBrowser)
}

func TestACodeEndedByANewRequestLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.administratorWhoAsked(t)
	s.clock.Advance(time.Minute)
	require.NoError(t, s.askFrom(anotherBrowser, "ana.souza@fabrica.example"))

	err := s.check(codeOfTheFirstRequest)

	requireInvalid(t, err, 2)
	s.requireNobodyLoggedIn(t, theBrowser, anotherBrowser)
}

// The code works only in the browser that asked for it: the session of
// another browser holds no request, and is answered as a wrong code.
func TestTheRightCodeFromAnotherBrowserLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	s.administratorWhoAsked(t)

	err := s.checkFrom(anotherBrowser, codeOfTheFirstRequest)

	requireInvalid(t, err, 2)
	s.requireNobodyLoggedIn(t, theBrowser, anotherBrowser)
	request := s.requests.All()[0]
	assert.Equal(t, 0, request.Attempts(), "the attempt of another browser counted against the request")
	assert.NoError(t, s.check(codeOfTheFirstRequest), "the browser that asked can no longer log in")
}

// A browser that asked for a code for an e-mail of no active user, or for
// none, is answered the way a request that exists is: whoever types a code
// learns nothing of the e-mail.
func TestABrowserWithNoRequestIsAnsweredAsAWrongCodeThreeTimes(t *testing.T) {
	cases := map[string]func(t *testing.T, s system){
		"an e-mail of no user": func(t *testing.T, s system) { require.NoError(t, s.ask("bruno.lima@fabrica.example")) },
		"no code asked for":    func(*testing.T, system) {},
		"a request that does not exist": func(t *testing.T, s system) {
			ctx := applicationtest.InBrowser(context.Background(), theBrowser)
			require.NoError(t, s.sessions.RememberLoginCodeRequest(ctx, domain.LoginCodeRequestID(37)))
		},
		"an e-mail of a deactivated user": func(t *testing.T, s system) {
			id, err := s.users.RegisterUser("carla.dias@fabrica.example", "es")
			require.NoError(t, err)
			require.NoError(t, s.users.DeactivateUser(id))
			require.NoError(t, s.ask("carla.dias@fabrica.example"))
		},
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			prepare(t, s)

			first := s.check(codeOfTheFirstRequest)
			second := s.check("000000")
			third := s.check(codeOfTheFirstRequest)
			afterwards := s.check(codeOfTheFirstRequest)

			requireInvalid(t, first, 2)
			requireInvalid(t, second, 1)
			requireEnded(t, third)
			requireEnded(t, afterwards)
			s.requireNobodyLoggedIn(t, theBrowser)
			assert.Empty(t, s.requests.All())
		})
	}
}

func TestANewCodeAskedForGivesThreeAttemptsAgain(t *testing.T) {
	s := newSystem(t)
	require.NoError(t, s.ask("bruno.lima@fabrica.example"))
	requireInvalid(t, s.check("000000"), 2)
	requireInvalid(t, s.check("000000"), 1)
	// The limit of one a minute counts the asks for an address of no user too.
	s.clock.Advance(time.Minute)

	require.NoError(t, s.ask("bruno.lima@fabrica.example"))

	requireInvalid(t, s.check("000000"), 2)
}

// A user deactivated after asking for a code does not log in with it: the
// code was used, and the person is told to ask for another, which an e-mail of
// a deactivated user never receives.
func TestTheRightCodeOfAUserDeactivatedSinceTheyAskedLogsNobodyIn(t *testing.T) {
	s := newSystem(t)
	ana := s.administratorWhoAsked(t)
	require.NoError(t, s.users.DeactivateUser(consultancyapplication.UserID(ana)))

	err := s.check(codeOfTheFirstRequest)

	requireEnded(t, err)
	s.requireNobodyLoggedIn(t, theBrowser)
	_, ended := s.requests.All()[0].EndedAt()
	assert.True(t, ended, "the code of a deactivated user still works")
}

// A login is never without its record: when the audit log cannot be written,
// the request stays as it was and nobody logs in.
func TestNobodyLogsInWhenTheLoginCannotBeRecorded(t *testing.T) {
	s := newSystem(t)
	s.administratorWhoAsked(t)
	failure := errors.New("the audit log cannot be written to")
	s.auditLog.FailWith(failure)

	err := s.check(codeOfTheFirstRequest)

	require.ErrorIs(t, err, failure)
	s.requireNobodyLoggedIn(t, theBrowser)
	_, ended := s.requests.All()[0].EndedAt()
	assert.False(t, ended, "the request was ended by a login that did not happen")
	s.auditLog.FailWith(nil)
	assert.NoError(t, s.check(codeOfTheFirstRequest), "the code no longer works")
}
