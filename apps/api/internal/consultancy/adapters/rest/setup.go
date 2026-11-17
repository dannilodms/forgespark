package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// setupTag groups the operations of the setup in the spec.
const setupTag = "Setup"

// The operations of the setup need no session: the setup code allows them.
// It travels in the body of each one, never in a path, a query parameter or
// a header, so that it reaches no log, no span and no browser history.
func registerSetup(api *restapi.API, useCases UseCases) {
	restapi.Register(api, huma.Operation{
		OperationID: "tell_setup_state",
		Method:      http.MethodGet,
		Path:        restapi.Prefix + "/setup",
		Summary:     "Tell the state of the setup",
		Description: "Tells whether the setup is concluded and whether a setup code is waiting. It needs no session.",
		Tags:        []string{setupTag},
	}, restapi.Failures{
		// No rule can be broken. Listing the status gives the answer of a
		// failure its description, as in the other operations.
		http.StatusInternalServerError: {},
	}, tellSetupState{useCase: useCases.TellSetupState}.handle)

	restapi.Register(api, huma.Operation{
		OperationID: "open_setup",
		Method:      http.MethodPost,
		Path:        restapi.Prefix + "/setup/open",
		Summary:     "Open the setup with a setup code",
		Description: "Checks a setup code and tells what it opens and when it expires. It saves nothing, and a wrong code changes nothing.",
		Tags:        []string{setupTag},
	}, restapi.Failures{
		http.StatusUnprocessableEntity: {domain.SetupCodeInvalidError{}},
		http.StatusTooManyRequests:     {restapi.LimitedByCaddy},
	}, openSetup{useCase: useCases.OpenSetup}.handle)

	restapi.Register(api, huma.Operation{
		OperationID: "conclude_setup",
		Method:      http.MethodPost,
		Path:        restapi.Prefix + "/setup/conclude",
		Summary:     "Conclude the first setup",
		Description: "Registers the consultancy, its first administrator and the two first roles together, and ends the setup code. " +
			"Every field is checked before anything is saved, and nothing is saved when one breaks a rule.",
		Tags: []string{setupTag},
	}, restapi.Failures{
		http.StatusUnprocessableEntity: append([]restapi.Failure{domain.SetupCodeInvalidError{}}, fieldRules...),
		http.StatusConflict:            {domain.SetupAlreadyConcludedError{}},
		http.StatusTooManyRequests:     {restapi.LimitedByCaddy},
	}, concludeSetup{useCase: useCases.ConcludeSetup}.handle)
}

// fieldRules are the rules a field of the consultancy or of an administrator
// may break, as the domain checks them. The contract declares what it can of
// them, for a client to check before it sends: a required text, its length,
// the form of an e-mail, of a telephone and of a website.
var fieldRules = []restapi.Failure{
	restapi.InField(kernel.RequiredError{}),
	restapi.InField(kernel.TooLongError{}),
	restapi.InField(kernel.CountryUnknownError{}),
	restapi.InField(kernel.CurrencyUnknownError{}),
	restapi.InField(kernel.TimeZoneUnknownError{}),
	restapi.InField(kernel.LanguageUnsupportedError{}),
	restapi.InField(kernel.EmailInvalidError{}),
	restapi.InField(kernel.TelephoneInvalidError{}),
	restapi.InField(kernel.WebsiteInvalidError{}),
}

type tellSetupState struct {
	useCase *application.TellSetupState
}

type tellSetupStateOutput struct {
	Body SetupState
}

// SetupState is the state of the setup.
type SetupState struct {
	Concluded        bool `json:"concluded" doc:"Whether the setup is concluded: the consultancy and its first administrator exist."`
	SetupCodeWaiting bool `json:"setup_code_waiting" doc:"Whether a setup code is issued and still works, before the setup is concluded or after."`
}

func (o tellSetupState) handle(ctx context.Context, _ *struct{}) (*tellSetupStateOutput, error) {
	state, err := o.useCase.Execute(ctx)
	if err != nil {
		return nil, err
	}
	return &tellSetupStateOutput{Body: SetupState{
		Concluded:        state.Concluded,
		SetupCodeWaiting: state.SetupCodeWaiting,
	}}, nil
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
	Opens     string    `json:"opens" enum:"first_setup,new_administrator" doc:"What the setup code opens: the first setup while the setup is not concluded, and the registration of a new administrator once it is."`
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

type concludeSetup struct {
	useCase *application.ConcludeSetup
}

type concludeSetupInput struct {
	Body ConcludeSetupRequest
}

// ConcludeSetupRequest carries the setup code and everything typed in the
// setup.
type ConcludeSetupRequest struct {
	Code          string             `json:"code" minLength:"1" maxLength:"64" doc:"The setup code, as typed, as in open_setup."`
	Consultancy   SetupConsultancy   `json:"consultancy" doc:"The consultancy."`
	Administrator SetupAdministrator `json:"administrator" doc:"Its first administrator, who receives the administrator role."`
}

// concludeSetupOutput has no body: the operation answers 204.
type concludeSetupOutput struct{}

func (o concludeSetup) handle(ctx context.Context, input *concludeSetupInput) (*concludeSetupOutput, error) {
	err := o.useCase.Execute(ctx, application.ConcludeSetupInput{
		Code:          input.Body.Code,
		Consultancy:   input.Body.Consultancy.typed(),
		Administrator: input.Body.Administrator.typed(),
	})
	if err != nil {
		return nil, err
	}
	return &concludeSetupOutput{}, nil
}
