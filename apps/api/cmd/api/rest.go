package main

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	consultancypostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/adapters/postgres"
	consultancyrest "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/adapters/rest"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/clock"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// bodyLimits bound the body of every REST operation: a megabyte, sent within
// five seconds. A file never passes through the API, so a body is JSON.
var bodyLimits = restapi.Limits{BodyBytes: 1 << 20, BodyReadTimeout: 5 * time.Second}

// mountREST builds the REST adapter on the mux and registers every operation
// of every domain area: it wires each use case to its adapters, by hand.
//
// The OpenAPI spec is generated from what this function registers, so an
// operation that is served is in the spec. Generating the spec calls it with
// no pool: the use cases are built and never run.
func mountREST(mux *http.ServeMux, pool *pgxpool.Pool) *restapi.API {
	api := restapi.New(mux, bodyLimits)

	setupCodes := consultancypostgres.NewSetupCodeRepository(pool)
	consultancyrest.Register(api, consultancyrest.UseCases{
		OpenSetup: consultancyapplication.NewOpenSetup(setupCodes, clock.System{}),
	})

	return api
}
