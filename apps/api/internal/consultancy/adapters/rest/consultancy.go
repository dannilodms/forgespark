package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// consultancyTag groups the operations of the consultancy in the spec.
const consultancyTag = "Consultancy"

// The consultancy exists once, so its path is singular. Reading it needs a
// logged-in user and no scope (ADR-0026).
func registerConsultancy(api *restapi.API, useCases UseCases) {
	restapi.RegisterForUser(api, huma.Operation{
		OperationID: "tell_consultancy",
		Method:      http.MethodGet,
		Path:        restapi.Prefix + "/consultancy",
		Summary:     "Tell the consultancy",
		Description: "Answers the consultancy as the shell of the app shows it: its name, and a URL that reads its logo " +
			"from the file storage for fifteen minutes. The app reads it again sooner than that, so the logo never " +
			"breaks. Any logged-in user reads it: it asks for no scope.",
		Tags: []string{consultancyTag},
	}, restapi.Failures{
		// No rule can be broken. Listing the status gives the answer of a
		// failure its description, as in the other operations.
		http.StatusInternalServerError: {},
	}, tellConsultancy{useCase: useCases.TellConsultancy}.handle)
}

type tellConsultancy struct {
	useCase *application.TellConsultancy
}

type tellConsultancyOutput struct {
	Body Consultancy
}

// Consultancy is the consultancy as the shell of the app shows it.
type Consultancy struct {
	Name    string `json:"name" doc:"The name of the consultancy."`
	LogoURL string `json:"logo_url,omitempty" format:"uri" doc:"A URL that reads the logo of the consultancy from the file storage, signed for its public host name and valid for fifteen minutes. Left out when the consultancy has no logo."`
}

func (o tellConsultancy) handle(ctx context.Context, actor kernel.Actor, _ *struct{}) (*tellConsultancyOutput, error) {
	consultancy, err := o.useCase.Execute(ctx, actor)
	if err != nil {
		return nil, err
	}
	return &tellConsultancyOutput{Body: Consultancy{Name: consultancy.Name, LogoURL: consultancy.LogoURL}}, nil
}
