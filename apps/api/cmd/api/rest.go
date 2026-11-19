package main

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	accesspostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/postgres"
	accessrest "github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/rest"
	accessapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	auditpostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/adapters/postgres"
	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	consultancypostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/adapters/postgres"
	consultancyrest "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/adapters/rest"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	emailpostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/postgres"
	emailriver "github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/river"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	filespostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/files/adapters/postgres"
	filess3 "github.com/dannilodms/software-factory-manager/apps/api/internal/files/adapters/s3"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/clock"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/session"
)

// bodyLimits bound the body of every REST operation: a megabyte, sent within
// five seconds. A file never passes through the API, so a body is JSON.
var bodyLimits = restapi.Limits{BodyBytes: 1 << 20, BodyReadTimeout: 5 * time.Second}

// served are what the operations reach outside the process: the database,
// the file storage, the saving of jobs, the sessions of the browsers and the
// login providers turned on; and the key the login codes are computed with.
// Generating the spec passes none of the first five, and a key of its own:
// the use cases are built and never run.
type served struct {
	pool           *pgxpool.Pool
	storage        *filess3.Storage
	saver          *jobs.Saver
	sessions       *session.Sessions
	loginProviders []accessapplication.LoginProvider
	loginCodeKey   []byte
}

// mountREST builds the REST adapter on the mux and registers every operation
// of every domain area: it wires each use case to its adapters, by hand. It
// fails on a key of the login code too short, and on two login providers
// with one key.
//
// The OpenAPI spec is generated from what this function registers, so an
// operation that is served is in the spec.
func mountREST(mux *http.ServeMux, with served) (*restapi.API, error) {
	api := restapi.New(mux, bodyLimits)

	transactions := database.NewTransactions(with.pool)
	files := filespostgres.NewFileRepository(with.pool)
	setupCodes := consultancypostgres.NewSetupCodeRepository(with.pool)
	consultancies := consultancypostgres.NewConsultancyRepository(with.pool)
	setup := consultancyapplication.SetupDependencies{
		Transactions:  transactions,
		SetupCodes:    setupCodes,
		Consultancies: consultancies,
		Users:         consultancypostgres.NewUserRepository(with.pool),
		Roles:         consultancypostgres.NewRoleRepository(with.pool),
		ConfirmUpload: filesapplication.NewConfirmUpload(transactions, files, with.storage, clock.System{}, rand.Reader),
		RecordEntry:   auditapplication.NewRecordEntry(auditpostgres.NewAuditLog(with.pool)),
		Clock:         clock.System{},
	}
	consultancyrest.Register(api, consultancyrest.UseCases{
		TellSetupState: consultancyapplication.NewTellSetupState(consultancypostgres.NewSetupStateReader(with.pool), clock.System{}),
		OpenSetup:      consultancyapplication.NewOpenSetup(setupCodes, consultancies, clock.System{}),
		StartSetupUpload: consultancyapplication.NewStartSetupUpload(
			setupCodes,
			consultancies,
			filesapplication.NewStartUpload(transactions, files, with.storage, clock.System{}, rand.Reader),
			clock.System{},
		),
		ConcludeSetup:         consultancyapplication.NewConcludeSetup(setup),
		RegisterAdministrator: consultancyapplication.NewRegisterAdministrator(setup),
	})

	browser := accessrest.NewSession(with.sessions)
	login := accessapplication.LoginCodeDependencies{
		Transactions: transactions,
		Requests:     accesspostgres.NewLoginCodeRequestRepository(with.pool),
		Session:      browser,
		Clock:        clock.System{},
	}
	activeUsers := consultancypostgres.NewActiveUserReader(with.pool)
	findActiveUserByID := consultancyapplication.NewFindActiveUserByID(activeUsers)
	logIn := accessapplication.NewLogIn(accessapplication.LogInDependencies{
		Transactions:       transactions,
		Session:            browser,
		FindActiveUserByID: findActiveUserByID,
		RecordEntry:        setup.RecordEntry,
		Clock:              clock.System{},
	})
	checkLoginCode, err := accessapplication.NewCheckLoginCode(login, logIn, with.loginCodeKey)
	if err != nil {
		return nil, err
	}
	providers, err := accessapplication.NewLoginProviders(with.loginProviders...)
	if err != nil {
		return nil, err
	}
	// Every operation that needs a user finds who acts through the session.
	api.AuthenticateWith(accessrest.NewActors(accessapplication.NewFindActor(browser, findActiveUserByID)))
	accessrest.Register(api, accessrest.UseCases{
		AskForLoginCode: accessapplication.NewAskForLoginCode(
			login,
			consultancyapplication.NewFindActiveUser(activeUsers),
			emailapplication.NewQueueEmail(
				transactions,
				emailpostgres.NewEmailRepository(with.pool),
				emailriver.NewDeliveryJobs(with.saver),
				clock.System{},
			),
			rand.Reader,
		),
		CheckLoginCode:     checkLoginCode,
		TellLoggedInUser:   accessapplication.NewTellLoggedInUser(findActiveUserByID),
		LogOut:             accessapplication.NewLogOut(browser),
		TellLoginState:     accessapplication.NewTellLoginState(providers, browser),
		StartProviderLogin: accessapplication.NewStartProviderLogin(providers, browser, rand.Reader),
		ConcludeProviderLogin: accessapplication.NewConcludeProviderLogin(accessapplication.ProviderLoginDependencies{
			Providers:      providers,
			Session:        browser,
			Links:          accesspostgres.NewProviderLinkRepository(with.pool),
			FindActiveUser: consultancyapplication.NewFindActiveUser(activeUsers),
			LogIn:          logIn,
			Clock:          clock.System{},
		}),
	})

	return api, nil
}
