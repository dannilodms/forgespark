// Package restapi is what every REST operation stands on: Huma on the router
// of the standard library, the error format on every failure, strict input,
// and the limits on a request body. A domain area adds its operations with
// Register, from its `adapters/rest` package. Rules: API conventions and
// Errors in CODING_STANDARDS.md.
package restapi

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

const (
	title = "My Software Factory Manager"
	// The contract carries no version: every consumer ships with the backend
	// (ADR-0009).
	version = "unversioned"
)

// Limits bound the body of a request.
type Limits struct {
	// BodyBytes is the largest body an operation reads.
	BodyBytes int64
	// BodyReadTimeout is how long a caller has to send its body.
	BodyReadTimeout time.Duration
}

// API is the REST adapter of the binary. Its OpenAPI spec is generated from
// the operations registered on it.
type API struct {
	huma.API
	limits Limits
}

// New builds the REST adapter on the mux. It serves no documentation page, no
// copy of the spec and no schema: the committed spec is the reference.
func New(mux *http.ServeMux, limits Limits) *API {
	config := huma.DefaultConfig(title, version)
	config.DocsPath = ""
	config.OpenAPIPath = ""
	config.SchemasPath = ""
	// The default hook adds a `$schema` property to every body, whose value
	// repeats the host the caller named. No response repeats what the caller
	// sent.
	config.CreateHooks = nil
	config.RejectUnknownQueryParameters = true

	api := &API{API: humago.New(mux, config), limits: limits}
	api.UseMiddleware(api.recoverPanic)
	mountUnrouted(mux)
	return api
}
