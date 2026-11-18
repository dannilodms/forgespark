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
	// description tells a reader of the spec what holds for every operation.
	description = "The REST contract of the system, generated from the Go code and committed.\n\n" +
		"Every failure answers `application/problem+json` with the schema `Problem`. " +
		"Its `code` is stable and is what a client translates, its `params` fill the translated text, " +
		"and its `trace_id` finds the request in the logs. " +
		"Each operation lists, under its statuses, the codes of the rules it can break. " +
		"Any operation may also answer `bad_request` (400, the body is not JSON), `not_found` (404), " +
		"`method_not_allowed` (405), `request_timeout` (408, the body took too long to arrive), " +
		"`request_entity_too_large` (413), `unsupported_media_type` (415), `validation_failed` (422), " +
		"`too_many_requests` (429) and `internal_server_error` (500). " +
		"When the API cannot be reached, Caddy answers in the same format, with `bad_gateway` (502), " +
		"`service_unavailable` (503) or `gateway_timeout` (504), and the `trace_id` of the trace it started.\n\n" +
		"An unknown property of a body and an unknown query parameter are rejected with `validation_failed`."
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
	config.Info.Description = description
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
