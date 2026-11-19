// Package rest serves the use cases of the access area over REST, and keeps
// what the login writes in the session of the browser. Each operation parses
// its input, calls one use case and maps the result. The types of the bodies
// are here, and the OpenAPI spec is generated from them.
//
// The login lives under the session of the browser, `/api/session`, which
// exists once for each browser. The two operations of the login code are
// under `/api/session/login_code`, which Caddy limits for each IP address:
// asking for a code (`POST /api/session/login_code`) and checking it
// (`POST /api/session/login_code/check`). An e-mail, a code or a token
// travels in a body, never in a path, a query parameter or a header.
package rest

import (
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// UseCases are the use cases of the area that have a REST operation.
type UseCases struct {
	AskForLoginCode *application.AskForLoginCode
}

// Register adds the operations of the area to the API.
func Register(api *restapi.API, useCases UseCases) {
	registerLoginCode(api, useCases)
}
