package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// setupTag groups the operations of the setup in the spec.
const setupTag = "Setup"

// The operations of the setup need no session: the setup code allows them.
// It travels in the body of each one, never in a path, a query parameter or
// a header, so that it reaches no log, no span and no browser history.
func registerSetup(api *restapi.API, useCases UseCases) {
	restapi.Register(api, huma.Operation{
		OperationID: "open_setup",
		Method:      http.MethodPost,
		Path:        restapi.Prefix + "/setup/open",
		Summary:     "Open the setup with a setup code",
		Description: "Checks a setup code and tells what it opens and when it expires. It saves nothing, and a wrong code changes nothing.",
		Tags:        []string{setupTag},
	}, restapi.Failures{
		http.StatusUnprocessableEntity: {domain.SetupCodeInvalidError{}},
	}, openSetup{useCase: useCases.OpenSetup}.handle)
}

type openSetup struct {
	useCase *application.OpenSetup
}

type openSetupInput struct {
	Body OpenSetupRequest
}

// OpenSetupRequest carries the setup code as the person typed it.
type OpenSetupRequest struct {
	Code string `json:"code" minLength:"1" maxLength:"64" doc:"The setup code, as typed. Case, hyphens and spaces do not count, and O, I and L are read as 0, 1 and 1."`
}

type openSetupOutput struct {
	Body OpenedSetup
}

// OpenedSetup is what a setup code that works opens.
type OpenedSetup struct {
	Opens     string    `json:"opens" enum:"first_setup" doc:"What the setup code opens."`
	ExpiresAt time.Time `json:"expires_at" doc:"When the setup code stops working."`
}

func (o openSetup) handle(ctx context.Context, input *openSetupInput) (*openSetupOutput, error) {
	opened, err := o.useCase.Execute(ctx, application.OpenSetupInput{Code: input.Body.Code})
	if err != nil {
		return nil, err
	}
	return &openSetupOutput{Body: OpenedSetup{
		Opens:     string(opened.Opens),
		ExpiresAt: opened.ExpiresAt.UTC(),
	}}, nil
}
