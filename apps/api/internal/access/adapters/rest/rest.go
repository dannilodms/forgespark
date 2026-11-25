// Package rest serves the use cases of the access area over REST, keeps what
// the login writes in the session of the browser, and tells the REST adapter
// who acts in each request. Each operation parses its input, calls one use
// case and maps the result. The types of the bodies are here, and the OpenAPI
// spec is generated from them.
//
// The login lives under the session of the browser, `/api/session`, which
// exists once for each browser. The two operations of the login code are
// under `/api/session/login_code`, which Caddy limits for each IP address:
// asking for a code (`POST /api/session/login_code`) and checking it
// (`POST /api/session/login_code/check`). Who is logged in is
// `GET /api/session`, and logging out `POST /api/session/log_out`. An e-mail,
// a code or a token travels in a body or in the cookie of the session, never
// in a path, a query parameter or a header of the contract.
//
// A login through a login provider is two navigations of the browser:
// `GET /api/session/login_providers/{key}/start`, which takes the path to
// return to as a query parameter and sends the browser to the provider, and
// `GET /api/session/login_providers/{key}/callback`, which the provider sends
// it back to with the code and the state of OAuth in the query: the one
// exception, which OAuth sets. The login screen reads the providers turned
// on, and the e-mail of the last refusal, at `GET /api/session/login`.
package rest

import (
	"github.com/dannilodms/forgespark/apps/api/internal/access/application"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/restapi"
)

// UseCases are the use cases of the area that have a REST operation.
type UseCases struct {
	AskForLoginCode  *application.AskForLoginCode
	CheckLoginCode   *application.CheckLoginCode
	TellLoggedInUser *application.TellLoggedInUser
	LogOut           *application.LogOut

	TellLoginState        *application.TellLoginState
	StartProviderLogin    *application.StartProviderLogin
	ConcludeProviderLogin *application.ConcludeProviderLogin
}

// Register adds the operations of the area to the API.
func Register(api *restapi.API, useCases UseCases) {
	registerLoginCode(api, useCases)
	registerSession(api, useCases)
	registerProviderLogin(api, useCases)
}
