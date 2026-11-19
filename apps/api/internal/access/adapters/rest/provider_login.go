package rest

import (
	"context"
	"net/http"
	"net/url"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// loginPath is the login of the session of the browser, before anyone is
// logged in: its state.
const loginPath = sessionPath + "/login"

// loginProvidersPath is where a login through each login provider starts
// and concludes: under the key of the provider, `/start` and `/callback`.
const loginProvidersPath = sessionPath + "/login_providers"

// CallbackPath is the path of the callback of the provider with this key,
// which the provider sends the browser back to: the redirect URL of the
// client of the system at the provider is the origin of the app followed by
// this path.
func CallbackPath(key string) string {
	return loginProvidersPath + "/" + key + "/callback"
}

// loginScreenPath is the login screen of the app, which a login through a
// provider that logged nobody in returns to, with its notice.
const loginScreenPath = "/login"

// The notices of the login screen a login through a provider ends with.
const (
	loginNotConcludedNotice    = "login_not_concluded"
	noActiveUserNotice         = "no_active_user"
	anotherAccountLinkedNotice = "another_account_linked"
)

// The login through a provider is two navigations of the browser, which no
// client calls: the start, which sends the browser to the provider, and the
// callback, which the provider sends it back to. Both need no logged-in user,
// and both answer a redirect. The login state is read by the login screen.
func registerProviderLogin(api *restapi.API, useCases UseCases) {
	restapi.Register(api, huma.Operation{
		OperationID: "tell_login_state",
		Method:      http.MethodGet,
		Path:        loginPath,
		Summary:     "Tell the state of the login",
		Description: "Answers what the login screen needs before anyone is logged in: the login providers turned " +
			"on, one button each, and the e-mail the last login through a provider in this browser was refused for, " +
			"because no active user has it, which the screen shows with the notice `no_active_user`: the e-mail " +
			"travels in this answer and in no address. It needs no session, and reading it forgets nothing.",
		Tags: []string{loginTag},
	}, restapi.Failures{
		http.StatusInternalServerError: {},
	}, tellLoginState{useCase: useCases.TellLoginState}.handle)

	restapi.Register(api, huma.Operation{
		OperationID:   "start_provider_login",
		Method:        http.MethodGet,
		Path:          loginProvidersPath + "/{key}/start",
		DefaultStatus: http.StatusFound,
		Summary:       "Start a login through a login provider",
		Description: "The browser navigates here from the button of a login provider; the generated client does not " +
			"call it. It answers a redirect to the provider, with a state, a nonce and the challenge of a PKCE " +
			"verifier, which the session of this browser keeps, with the path to return to: the answer sets its " +
			"cookie. When the provider cannot be reached, it redirects to the login screen with the notice " +
			"`login_not_concluded`. A key of no provider turned on answers `not_found`.",
		Tags: []string{loginTag},
	}, restapi.Failures{
		http.StatusNotFound: {restapi.NotFound},
	}, startProviderLogin{useCase: useCases.StartProviderLogin}.handle)

	restapi.Register(api, huma.Operation{
		OperationID:   "conclude_provider_login",
		Method:        http.MethodGet,
		Path:          loginProvidersPath + "/{key}/callback",
		DefaultStatus: http.StatusSeeOther,
		// The provider writes the query of its callback, and may add
		// parameters of its own: they are not refused, and the use case
		// judges what it reads. It has no body. So it never answers
		// `validation_failed`, and declares no 422.
		SkipValidateParams: true,
		SkipValidateBody:   true,
		Summary:            "Conclude a login through a login provider",
		Description: "The login provider sends the browser back here; the generated client does not call it. The " +
			"state must be the one of the login this browser started, which is used once. The code is exchanged " +
			"with the verifier, and the ID token validated with its nonce; the user is found by the provider link " +
			"of the account, or else by the e-mail the token says is verified, and the account is linked. It " +
			"redirects to the path to return to, logged in, with the cookie of the session set again with a new " +
			"token. A verified e-mail of no active user redirects to the login screen with the notice " +
			"`no_active_user`, the e-mail kept for `tell_login_state`; an account that is not linked, whose " +
			"verified e-mail is the one of a user who has another account of the provider linked, redirects with " +
			"`another_account_linked`; a cancel or an error at the provider, a code it refuses, an e-mail it does " +
			"not say is verified, or a state that is not the one of this browser redirects with " +
			"`login_not_concluded`. The query parameters are the ones OAuth defines, and any other the provider " +
			"adds is not refused. A key of no provider turned on answers `not_found`.",
		Tags: []string{loginTag},
	}, restapi.Failures{
		http.StatusNotFound: {restapi.NotFound},
	}, concludeProviderLogin{useCase: useCases.ConcludeProviderLogin}.handle)
}

type tellLoginState struct {
	useCase *application.TellLoginState
}

type tellLoginStateOutput struct {
	Body LoginState
}

// LoginState is what the login screen needs before the login.
type LoginState struct {
	LoginProviders []TurnedOnLoginProvider `json:"login_providers" nullable:"false" doc:"The login providers turned on, in the order the screen shows them, none included."`
	RefusedEmail   string                  `json:"refused_email,omitempty" format:"email" doc:"The e-mail the last login through a provider in this browser was refused for, because no active user has it. Absent when it was not refused."`
}

// TurnedOnLoginProvider is a login provider turned on, as the login screen
// draws its button.
type TurnedOnLoginProvider struct {
	Key         string `json:"key" doc:"The key of the provider, which names it in the path of its login."`
	DisplayName string `json:"display_name" doc:"The name the button shows."`
}

func (o tellLoginState) handle(ctx context.Context, _ *struct{}) (*tellLoginStateOutput, error) {
	state, err := o.useCase.Execute(ctx)
	if err != nil {
		return nil, err
	}
	providers := make([]TurnedOnLoginProvider, 0, len(state.Providers))
	for _, provider := range state.Providers {
		providers = append(providers, TurnedOnLoginProvider{Key: provider.Key, DisplayName: provider.DisplayName})
	}
	output := &tellLoginStateOutput{Body: LoginState{LoginProviders: providers, RefusedEmail: ""}}
	if state.Refused {
		output.Body.RefusedEmail = state.RefusedEmail.String()
	}
	return output, nil
}

// redirectOutput is the answer of a navigation of the browser: where it goes
// next.
type redirectOutput struct {
	Location string `header:"Location" doc:"Where the browser goes next."`
}

type startProviderLogin struct {
	useCase *application.StartProviderLogin
}

type startProviderLoginInput struct {
	Key      string `path:"key" doc:"The key of the login provider, as the login state names it."`
	ReturnTo string `query:"return_to" maxLength:"2048" doc:"The path of the app to return to once logged in. Anything but a path of the app returns to its root."`
}

func (o startProviderLogin) handle(ctx context.Context, input *startProviderLoginInput) (*redirectOutput, error) {
	started, found, err := o.useCase.Execute(ctx, application.StartProviderLoginInput{Provider: input.Key, ReturnTo: input.ReturnTo})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, restapi.NotFound
	}
	if started.AuthorizationURL == "" {
		return &redirectOutput{Location: loginScreen(loginNotConcludedNotice, started.ReturnTo)}, nil
	}
	return &redirectOutput{Location: started.AuthorizationURL}, nil
}

type concludeProviderLogin struct {
	useCase *application.ConcludeProviderLogin
}

type concludeProviderLoginInput struct {
	Key              string `path:"key" doc:"The key of the login provider."`
	Code             string `query:"code" doc:"The authorization code the provider gives, used once and with the verifier the session keeps."`
	State            string `query:"state" doc:"The state the login was started with."`
	Error            string `query:"error" doc:"The code of the failure the provider answers, when the person cancelled or the provider failed."`
	ErrorDescription string `query:"error_description" doc:"What the provider says of its failure. It is not read."`
}

func (o concludeProviderLogin) handle(ctx context.Context, input *concludeProviderLoginInput) (*redirectOutput, error) {
	concluded, found, err := o.useCase.Execute(ctx, application.ConcludeProviderLoginInput{
		Provider: input.Key,
		Code:     input.Code,
		State:    input.State,
		Failed:   input.Error != "",
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, restapi.NotFound
	}
	switch concluded.Outcome {
	case application.LoggedInThroughProvider:
		return &redirectOutput{Location: concluded.ReturnTo.String()}, nil
	case application.NoActiveUser:
		return &redirectOutput{Location: loginScreen(noActiveUserNotice, concluded.ReturnTo)}, nil
	case application.AnotherAccountLinked:
		return &redirectOutput{Location: loginScreen(anotherAccountLinkedNotice, concluded.ReturnTo)}, nil
	}
	return &redirectOutput{Location: loginScreen(loginNotConcludedNotice, concluded.ReturnTo)}, nil
}

// loginScreen is the address of the login screen with the notice, and the
// path to return to when there is one, so a new attempt returns there too.
func loginScreen(notice string, returnTo domain.ReturnPath) string {
	query := url.Values{"notice": {notice}}
	if returnTo != domain.RootPath {
		query.Set("return_to", returnTo.String())
	}
	return loginScreenPath + "?" + query.Encode()
}
