package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// loginTag groups the operations of the login in the spec.
const loginTag = "Login"

// loginCodePath is where the login code of the session of a browser is asked
// for. Checking it is a POST to loginCodePath + "/check". Caddy limits both
// for each IP address.
const loginCodePath = restapi.Prefix + "/session/login_code"

// The operations of the login code need no logged-in user: they are how one
// logs in.
func registerLoginCode(api *restapi.API, useCases UseCases) {
	restapi.Register(api, huma.Operation{
		OperationID:   "ask_for_login_code",
		Method:        http.MethodPost,
		Path:          loginCodePath,
		DefaultStatus: http.StatusAccepted,
		Summary:       "Ask for a login code",
		Description: "Sends a login code to the e-mail when it belongs to an active user, and answers the same way, " +
			"with no body, for every e-mail, so that nobody learns from it who is a user. The code is six digits, " +
			"works for 10 minutes, once, and only in the browser that asked for it, which the session remembers: " +
			"the answer sets its cookie. A new code ends the one before. One code a minute and five an hour are " +
			"sent to each e-mail; past them the operation answers `login_code_limit_reached`.",
		Tags: []string{loginTag},
	}, restapi.Failures{
		http.StatusUnprocessableEntity: {
			domain.LoginCodeLimitReachedError{},
			restapi.InField(kernel.RequiredError{}),
			restapi.InField(kernel.TooLongError{}),
			restapi.InField(kernel.EmailInvalidError{}),
		},
		http.StatusTooManyRequests: {restapi.LimitedByCaddy},
	}, askForLoginCode{useCase: useCases.AskForLoginCode}.handle)
}

type askForLoginCode struct {
	useCase *application.AskForLoginCode
}

type askForLoginCodeInput struct {
	Body AskForLoginCodeRequest
}

// AskForLoginCodeRequest carries the e-mail as the person typed it.
type AskForLoginCodeRequest struct {
	Email string `json:"email" minLength:"1" format:"email" doc:"The e-mail the person logs in with, in any case."`
}

// TransformSchema sets the limit of the e-mail of kernel.
func (AskForLoginCodeRequest) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	email := schema.Properties["email"]
	maxLength := kernel.EmailMaxLength
	email.MaxLength = &maxLength
	email.PrecomputeMessages()
	return schema
}

// askForLoginCodeOutput has no body: the operation answers 202.
type askForLoginCodeOutput struct{}

func (o askForLoginCode) handle(ctx context.Context, input *askForLoginCodeInput) (*askForLoginCodeOutput, error) {
	err := o.useCase.Execute(ctx, application.AskForLoginCodeInput{Email: input.Body.Email})
	if err != nil {
		return nil, err
	}
	return &askForLoginCodeOutput{}, nil
}
