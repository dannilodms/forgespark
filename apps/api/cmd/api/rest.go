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
// the file storage, the saving of jobs and the sessions of the browsers.
// Generating the spec passes none of them: the use cases are built and never
// run.
type served struct {
	pool     *pgxpool.Pool
	storage  *filess3.Storage
	saver    *jobs.Saver
	sessions *session.Sessions
}

// mountREST builds the REST adapter on the mux and registers every operation
// of every domain area: it wires each use case to its adapters, by hand.
//
// The OpenAPI spec is generated from what this function registers, so an
// operation that is served is in the spec.
func mountREST(mux *http.ServeMux, with served) *restapi.API {
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

	login := accessapplication.LoginCodeDependencies{
		Transactions: transactions,
		Requests:     accesspostgres.NewLoginCodeRequestRepository(with.pool),
		Session:      accessrest.NewSession(with.sessions),
		Clock:        clock.System{},
	}
	accessrest.Register(api, accessrest.UseCases{
		AskForLoginCode: accessapplication.NewAskForLoginCode(
			login,
			consultancyapplication.NewFindActiveUser(consultancypostgres.NewActiveUserReader(with.pool)),
			emailapplication.NewQueueEmail(
				transactions,
				emailpostgres.NewEmailRepository(with.pool),
				emailriver.NewDeliveryJobs(with.saver),
				clock.System{},
			),
			rand.Reader,
		),
	})

	return api
}
