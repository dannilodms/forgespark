package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	auditapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application/applicationtest"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	consultancyapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application/applicationtest"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	emailapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application/applicationtest"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	filesapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// countingBytes stands in for the random source: it gives 100, 101, 102 and
// so on, so the random value of the first request is 100 to 131 and the one
// of the second 132 to 163.
type countingBytes struct {
	next byte
}

func (c *countingBytes) Read(buffer []byte) (int, error) {
	for i := range buffer {
		buffer[i] = c.next
		c.next++
	}
	return len(buffer), nil
}

// theKey is the key of the login code of these tests: the bytes 0 to 31.
func theKey(t *testing.T) domain.LoginCodeKey {
	t.Helper()
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}
	key, err := domain.NewLoginCodeKey(raw)
	require.NoError(t, err)
	return key
}

// theBrowser is the browser every request of a test comes from, unless the
// test says otherwise.
const theBrowser = "the browser of Ana"

// system is the login built on the in-memory implementations of this area,
// of the consultancy area, of the e-mail area, of the files area and of the
// audit log, with the clock stopped at noon.
type system struct {
	clock            *applicationtest.Clock
	requests         *applicationtest.LoginCodeRequestRepository
	asks             *applicationtest.AskedAddressRepository
	sessions         *applicationtest.Sessions
	users            *consultancyapplicationtest.UserRepository
	emails           *emailapplicationtest.EmailRepository
	deliveries       *emailapplicationtest.DeliveryJobs
	notices          *emailapplicationtest.BounceNotices
	endedSessions    *applicationtest.EndedSessions
	auditLog         *auditapplicationtest.AuditLog
	askForLoginCode  *application.AskForLoginCode
	loginCodes       *application.LoginCodeContents
	checkLoginCode   *application.CheckLoginCode
	findActor        *application.FindActor
	tellLoggedInUser *application.TellLoggedInUser
	logOut           *application.LogOut
	// The files the photos of the users are, and the storage of their bytes.
	files   *filesapplicationtest.FileRepository
	objects *filesapplicationtest.Storage
	// The login through a provider: the provider links, and two providers
	// turned on, the local one first.
	links                 *applicationtest.ProviderLinkRepository
	localProvider         *applicationtest.LoginProvider
	startProviderLogin    *application.StartProviderLogin
	concludeProviderLogin *application.ConcludeProviderLogin
	tellLoginState        *application.TellLoginState
	// The daily cleanup, which removes what the login and the e-mail keep
	// only for a while.
	cleanUpOldRecords *application.CleanUpOldRecords
}

func newSystem(t *testing.T) system {
	t.Helper()
	clock := applicationtest.NewClock(noon)
	requests := applicationtest.NewLoginCodeRequestRepository()
	asks := applicationtest.NewAskedAddressRepository()
	sessions := applicationtest.NewSessions()
	users := consultancyapplicationtest.NewUserRepository()
	emails := emailapplicationtest.NewEmailRepository()
	deliveries := emailapplicationtest.NewDeliveryJobs()
	notices := emailapplicationtest.NewBounceNotices()
	endedSessions := applicationtest.NewEndedSessions()
	auditLog := auditapplicationtest.NewAuditLog()
	links := applicationtest.NewProviderLinkRepository()
	files := filesapplicationtest.NewFileRepository()
	objects := filesapplicationtest.NewStorage(clock)
	transactions := applicationtest.NewTransactions(requests, asks, emails, deliveries, auditLog, links)
	login := application.LoginCodeDependencies{
		Transactions: transactions,
		Requests:     requests,
		Session:      sessions,
		Key:          theKey(t),
		Clock:        clock,
	}
	activeUsers := consultancyapplicationtest.NewActiveUserReader(users)
	findActiveUserByID := consultancyapplication.NewFindActiveUserByID(activeUsers)
	logIn := application.NewLogIn(application.LogInDependencies{
		Transactions:       transactions,
		Session:            sessions,
		FindActiveUserByID: findActiveUserByID,
		RecordEntry:        auditapplication.NewRecordEntry(auditLog),
		Clock:              clock,
	})
	localProvider := applicationtest.NewLoginProvider("local", "Local login provider")
	providers, err := application.NewLoginProviders(localProvider, applicationtest.NewLoginProvider("another", "Another provider"))
	require.NoError(t, err)
	return system{
		clock:         clock,
		requests:      requests,
		asks:          asks,
		sessions:      sessions,
		users:         users,
		emails:        emails,
		deliveries:    deliveries,
		notices:       notices,
		endedSessions: endedSessions,
		auditLog:      auditLog,
		askForLoginCode: application.NewAskForLoginCode(
			login,
			asks,
			consultancyapplication.NewFindActiveUser(activeUsers),
			emailapplication.NewQueueEmail(transactions, emails, deliveries, clock),
			&countingBytes{next: 100},
		),
		loginCodes:       application.NewLoginCodeContents(requests, theKey(t)),
		checkLoginCode:   application.NewCheckLoginCode(login, logIn),
		findActor:        application.NewFindActor(sessions, findActiveUserByID),
		tellLoggedInUser: application.NewTellLoggedInUser(findActiveUserByID, filesapplication.NewSignDownload(files, objects)),
		logOut:           application.NewLogOut(sessions),
		files:            files,
		objects:          objects,
		links:            links,
		localProvider:    localProvider,
		// The random source counts from 0, apart from the one of the login
		// code, so the values of a login differ from one login to the next.
		startProviderLogin: application.NewStartProviderLogin(providers, sessions, &countingBytes{next: 0}),
		concludeProviderLogin: application.NewConcludeProviderLogin(application.ProviderLoginDependencies{
			Providers:      providers,
			Session:        sessions,
			Links:          links,
			FindActiveUser: consultancyapplication.NewFindActiveUser(activeUsers),
			LogIn:          logIn,
			Clock:          clock,
		}),
		tellLoginState: application.NewTellLoginState(providers, sessions),
		cleanUpOldRecords: application.NewCleanUpOldRecords(application.OldRecords{
			Requests:             requests,
			Asks:                 asks,
			EndedSessions:        endedSessions,
			CleanUpBounceNotices: emailapplication.NewCleanUpBounceNotices(notices, clock),
			Clock:                clock,
		}),
	}
}

// registered stores an active user with this e-mail and preferred language.
func (s system) registered(t *testing.T, email, language string) int64 {
	t.Helper()
	id, err := s.users.RegisterUser(email, language)
	require.NoError(t, err)
	return int64(id)
}

// ask asks for a login code from the browser of the test.
func (s system) ask(email string) error {
	return s.askFrom(theBrowser, email)
}

func (s system) askFrom(browser, email string) error {
	ctx := applicationtest.InBrowser(context.Background(), browser)
	return s.askForLoginCode.Execute(ctx, application.AskForLoginCodeInput{Email: email})
}

// requireNothingSaved checks that no request, e-mail or job was saved.
func (s system) requireNothingSaved(t *testing.T) {
	t.Helper()
	require.Empty(t, s.requests.All(), "a login code request was saved")
	require.Empty(t, s.emails.All(), "an e-mail was saved")
	require.Empty(t, s.deliveries.Saved(), "a job was saved")
}

func TestAKnownAnUnknownAndADeactivatedEmailGetTheSameAnswer(t *testing.T) {
	cases := map[string]struct {
		prepare func(t *testing.T, s system)
		saves   bool
	}{
		"an e-mail of an active user": {
			prepare: func(t *testing.T, s system) { s.registered(t, "ana.souza@fabrica.example", "pt-BR") },
			saves:   true,
		},
		"an e-mail of no user": {
			prepare: func(t *testing.T, s system) { s.registered(t, "bruno.lima@fabrica.example", "pt-BR") },
		},
		"an e-mail of a deactivated user": {
			prepare: func(t *testing.T, s system) {
				id, err := s.users.RegisterUser("ana.souza@fabrica.example", "pt-BR")
				require.NoError(t, err)
				require.NoError(t, s.users.DeactivateUser(id))
			},
		},
	}
	for name, example := range cases {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			example.prepare(t, s)

			err := s.ask("ana.souza@fabrica.example")

			require.NoError(t, err, "the answer differs")
			_, remembered := s.sessions.LoginCodeRequestOf(theBrowser)
			assert.True(t, remembered, "the session of the browser was not written")
			if example.saves {
				return
			}
			s.requireNothingSaved(t)
			kept, _ := s.sessions.LoginCodeRequestOf(theBrowser)
			assert.Equal(t, domain.NoLoginCodeRequest, kept)
		})
	}
}

func TestAskingSavesARequestThatWorksForTenMinutesAndQueuesTheEmailInThePreferredLanguage(t *testing.T) {
	s := newSystem(t)
	user := s.registered(t, "ana.souza@fabrica.example", "pt-BR")

	err := s.ask("  Ana.Souza@Fabrica.EXAMPLE ")

	require.NoError(t, err)
	require.Len(t, s.requests.All(), 1)
	request := s.requests.All()[0]
	assert.Equal(t, domain.UserID(user), request.User())
	assert.Equal(t, noon, request.AskedAt())
	assert.Equal(t, noon.Add(10*time.Minute), request.ExpiresAt())
	assert.Equal(t, 0, request.Attempts())
	_, ended := request.EndedAt()
	assert.False(t, ended)
	require.Len(t, s.emails.All(), 1)
	email := s.emails.All()[0]
	assert.Equal(t, "ana.souza@fabrica.example", email.Recipient().String(), "the e-mail goes to the address of the user")
	assert.Equal(t, emailapplication.LoginCode, email.Template())
	assert.Equal(t, "pt-BR", email.Language().String())
	assert.Equal(t, emailapplication.LoginCodeRequest(int64(request.ID())), email.Reference())
	assert.Equal(t, []emailapplication.EmailID{email.ID()}, s.deliveries.Saved())
	kept, _ := s.sessions.LoginCodeRequestOf(theBrowser)
	assert.Equal(t, request.ID(), kept, "the session keeps the request the browser asked for")
}

func TestAnEmailThatIsNotAnAddressIsAFieldThatBreaksARule(t *testing.T) {
	cases := map[string]struct {
		typed string
		code  string
	}{
		"empty":             {typed: "  ", code: "required"},
		"not an address":    {typed: "ana.souza", code: "email_invalid"},
		"with a name":       {typed: "Ana <ana.souza@fabrica.example>", code: "email_invalid"},
		"over its limit":    {typed: strings.Repeat("a", kernel.EmailMaxLength) + "@fabrica.example", code: "too_long"},
		"two addresses":     {typed: "ana@fabrica.example, bruno@fabrica.example", code: "email_invalid"},
		"an address of one": {typed: "@fabrica.example", code: "email_invalid"},
	}
	for name, example := range cases {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			s.registered(t, "ana.souza@fabrica.example", "pt-BR")

			err := s.ask(example.typed)

			var invalid kernel.InvalidFieldsError
			require.ErrorAs(t, err, &invalid)
			require.Len(t, invalid.Fields, 1)
			assert.Equal(t, "email", invalid.Fields[0].Field)
			assert.Equal(t, example.code, invalid.Fields[0].Rule.Code())
			s.requireNothingSaved(t)
			assert.Empty(t, s.asks.All(), "a text that is no address was counted")
		})
	}
}

func TestASecondCodeWithinAMinuteIsRefusedAndChangesNothing(t *testing.T) {
	s := newSystem(t)
	s.registered(t, "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, s.ask("ana.souza@fabrica.example"))
	first, _ := s.sessions.LoginCodeRequestOf(theBrowser)

	s.clock.Advance(59 * time.Second)
	err := s.ask("ana.souza@fabrica.example")

	require.ErrorIs(t, err, domain.LoginCodeLimitReachedError{})
	require.Len(t, s.requests.All(), 1)
	_, ended := s.requests.All()[0].EndedAt()
	assert.False(t, ended, "the refused request ended the one before")
	assert.Len(t, s.emails.All(), 1)
	assert.Len(t, s.asks.All(), 1, "the refused ask was counted")
	kept, _ := s.sessions.LoginCodeRequestOf(theBrowser)
	assert.Equal(t, first, kept, "the refused request changed the session")
}

// The limits count the asks for the typed address, whether or not it belongs
// to a user, so an address of no user answers the same 422 as the address of
// a user, and the limit tells nothing of who is a user (ADR-0011).
func TestTheLimitsAnswerTheSameForEveryAddress(t *testing.T) {
	cases := map[string]func(t *testing.T, s system){
		"an e-mail of an active user": func(t *testing.T, s system) { s.registered(t, "ana.souza@fabrica.example", "pt-BR") },
		"an e-mail of no user":        func(*testing.T, system) {},
		"an e-mail of a deactivated user": func(t *testing.T, s system) {
			id, err := s.users.RegisterUser("ana.souza@fabrica.example", "pt-BR")
			require.NoError(t, err)
			require.NoError(t, s.users.DeactivateUser(id))
		},
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			prepare(t, s)
			require.NoError(t, s.ask("ana.souza@fabrica.example"))

			s.clock.Advance(59 * time.Second)
			withinTheMinute := s.ask("Ana.Souza@Fabrica.example")
			s.clock.Advance(time.Second)
			aMinuteAfter := s.ask("ana.souza@fabrica.example")

			require.ErrorIs(t, withinTheMinute, domain.LoginCodeLimitReachedError{}, "the address typed in another case is the same address")
			require.NoError(t, aMinuteAfter)
			assert.Len(t, s.asks.All(), 2, "the refused ask was counted")
		})
	}
}

// The limits keep the keyed digest of the address and when it was asked for,
// never the address.
func TestTheLimitsKeepNoAddress(t *testing.T) {
	s := newSystem(t)
	email, err := kernel.NewEmail("bruno.lima@fabrica.example")
	require.NoError(t, err)

	require.NoError(t, s.ask("  Bruno.Lima@Fabrica.example"))

	require.Len(t, s.asks.All(), 1)
	ask := s.asks.All()[0]
	assert.Equal(t, theKey(t).AskedAddress(email).Digest(), ask.Digest)
	assert.NotContains(t, string(ask.Digest), "bruno")
	assert.Equal(t, noon, ask.AskedAt)
}

func TestACodeIsAllowedAgainOneMinuteAfterTheLast(t *testing.T) {
	s := newSystem(t)
	s.registered(t, "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, s.ask("ana.souza@fabrica.example"))

	s.clock.Advance(time.Minute)
	err := s.ask("ana.souza@fabrica.example")

	require.NoError(t, err)
	assert.Len(t, s.requests.All(), 2)
}

func TestASixthCodeWithinAnHourIsRefusedAndAllowedOnceTheFirstIsAnHourOld(t *testing.T) {
	for name, user := range map[string]bool{"an e-mail of an active user": true, "an e-mail of no user": false} {
		t.Run(name, func(t *testing.T) {
			s := newSystem(t)
			if user {
				s.registered(t, "ana.souza@fabrica.example", "pt-BR")
			}
			for range 5 {
				require.NoError(t, s.ask("ana.souza@fabrica.example"))
				s.clock.Advance(10 * time.Minute)
			}
			// The five were asked at noon, 12:10, 12:20, 12:30 and 12:40; it is
			// 12:50.

			sixth := s.ask("ana.souza@fabrica.example")
			s.clock.Advance(10 * time.Minute)
			anHourAfterTheFirst := s.ask("ana.souza@fabrica.example")

			require.ErrorIs(t, sixth, domain.LoginCodeLimitReachedError{})
			require.NoError(t, anHourAfterTheFirst)
			assert.Len(t, s.asks.All(), 6)
		})
	}
}

func TestTheLimitsCountTheAsksOfEachAddressApart(t *testing.T) {
	s := newSystem(t)
	s.registered(t, "ana.souza@fabrica.example", "pt-BR")
	s.registered(t, "bruno.lima@fabrica.example", "en-US")
	require.NoError(t, s.ask("ana.souza@fabrica.example"))
	require.NoError(t, s.ask("carla.dias@fabrica.example"))

	err := s.askFrom("the browser of Bruno", "bruno.lima@fabrica.example")

	require.NoError(t, err)
	assert.Len(t, s.requests.All(), 2)
}

func TestANewRequestEndsTheOneBefore(t *testing.T) {
	s := newSystem(t)
	s.registered(t, "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, s.ask("ana.souza@fabrica.example"))
	s.clock.Advance(2 * time.Minute)

	err := s.askFrom("another browser of Ana", "ana.souza@fabrica.example")

	require.NoError(t, err)
	requests := s.requests.All()
	require.Len(t, requests, 2)
	endedAt, ended := requests[0].EndedAt()
	assert.True(t, ended, "the earlier request still works")
	assert.Equal(t, noon.Add(2*time.Minute), endedAt)
	_, ended = requests[1].EndedAt()
	assert.False(t, ended)
	kept, _ := s.sessions.LoginCodeRequestOf("another browser of Ana")
	assert.Equal(t, requests[1].ID(), kept)
}

func TestNothingIsSavedAndTheSessionIsLeftAsItWasWhenTheEmailCannotBeQueued(t *testing.T) {
	s := newSystem(t)
	s.registered(t, "ana.souza@fabrica.example", "pt-BR")
	require.NoError(t, s.ask("ana.souza@fabrica.example"))
	first, _ := s.sessions.LoginCodeRequestOf(theBrowser)
	s.clock.Advance(time.Minute)
	failure := errors.New("the database refused the job")
	s.deliveries.FailWith(failure)

	err := s.ask("ana.souza@fabrica.example")

	require.ErrorIs(t, err, failure)
	require.Len(t, s.requests.All(), 1)
	_, ended := s.requests.All()[0].EndedAt()
	assert.False(t, ended, "the earlier request was ended by a request that was not saved")
	assert.Len(t, s.emails.All(), 1)
	assert.Len(t, s.asks.All(), 1, "an ask that saved nothing was counted")
	kept, _ := s.sessions.LoginCodeRequestOf(theBrowser)
	assert.Equal(t, first, kept)
}
