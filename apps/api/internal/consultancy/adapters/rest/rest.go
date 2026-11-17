// Package rest serves the use cases of the consultancy area over REST. Each
// operation parses its input, calls one use case and maps the result. The
// types of the bodies are here, and the OpenAPI spec is generated from them.
package rest

import (
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// UseCases are the use cases of the area that have a REST operation.
type UseCases struct {
	TellSetupState   *application.TellSetupState
	OpenSetup        *application.OpenSetup
	StartSetupUpload *application.StartSetupUpload
	ConcludeSetup    *application.ConcludeSetup
}

// Register adds the operations of the area to the API.
func Register(api *restapi.API, useCases UseCases) {
	registerSetup(api, useCases)
}
