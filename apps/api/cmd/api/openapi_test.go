package main

import (
	"flag"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// committedSpec is the OpenAPI spec of the REST contract, generated from the
// operations the binary serves and committed. The client of the app is
// generated from this file, and the lint and the breaking-change report read
// it. The test runs in the directory of its package: the module is two up.
const committedSpec = "../../openapi.yaml"

// writeSpec makes the test below write the spec instead of comparing it:
// `task api:spec` passes it.
var writeSpec = flag.Bool("write-spec", false, "write the OpenAPI spec generated from the code to "+committedSpec)

// generatedSpec builds the REST adapter as the command that serves does,
// with no database behind it and a key of the login code of zeros, and
// returns the spec of what it registered.
func generatedSpec(t *testing.T) *huma.OpenAPI {
	t.Helper()
	api, err := mountREST(http.NewServeMux(), served{})
	require.NoError(t, err)
	return api.OpenAPI()
}

func TestTheCommittedSpecMatchesTheCode(t *testing.T) {
	generated, err := generatedSpec(t).YAML()
	require.NoError(t, err)
	if *writeSpec {
		require.NoError(t, os.WriteFile(committedSpec, generated, 0o644))
		return
	}

	committed, err := os.ReadFile(committedSpec)

	require.NoError(t, err)
	line, inFile, fromCode := firstDifference(string(committed), string(generated))
	require.Zero(t, line, "apps/api/openapi.yaml is not what the code generates: run `task api:spec` and commit the file.\n"+
		"The first difference is at line %d.\n  in the file:   %s\n  from the code: %s", line, inFile, fromCode)
}

// firstDifference finds the first line two texts differ at, counted from 1,
// and returns it with what each text has there. It returns 0 when the texts
// are the same.
func firstDifference(one, other string) (line int, inOne, inOther string) {
	if one == other {
		return 0, "", ""
	}
	// A text that ends before the other differs from it where it ends.
	these := append(strings.Split(one, "\n"), "(the end)")
	those := append(strings.Split(other, "\n"), "(the end)")
	position := 0
	for position < len(these)-1 && position < len(those)-1 && these[position] == those[position] {
		position++
	}
	return position + 1, these[position], those[position]
}

func TestTheSpecIsTheSameEachTimeItIsGenerated(t *testing.T) {
	t.Parallel()
	first, err := generatedSpec(t).YAML()
	require.NoError(t, err)

	for range 5 {
		again, err := generatedSpec(t).YAML()
		require.NoError(t, err)
		require.Equal(t, string(first), string(again))
	}
}

// The setup code travels in the body of each operation of the setup, the
// e-mail and the login code in the body of each operation of the login code,
// and the token of the session in its cookie alone. An operation that took
// one any other way would write it to the address, which the browser history,
// the logs and the spans keep, or to a header the contract names.
//
// The two navigations of a login through a provider are the one exception,
// which the next test checks: the browser follows them, so what they take is
// in their path and their query.
func TestNoSetupCodeEmailLoginCodeOrTokenIsInAPathQueryParameterOrHeaderOfTheSpec(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths
	for _, path := range []string{
		"/api/setup/open", "/api/session", "/api/session/login", "/api/session/login_code",
		"/api/session/login_code/check", "/api/session/log_out",
	} {
		require.Contains(t, paths, path)
	}

	for path, item := range paths {
		assertTakesNothingOutsideTheBody(t, path, item)
	}
}

// providerLoginPaths are the start and the callback of a login through a
// provider.
var providerLoginPaths = []string{
	"/api/session/login_providers/{key}/start",
	"/api/session/login_providers/{key}/callback",
}

// The start takes the key of the provider in its path and the path of the
// app to return to in its query; the callback takes the key in its path and
// what OAuth defines in its query: the code, which works once and only with
// the verifier the session keeps, the state, and the failure the provider
// answers. No e-mail, no token and no header: the e-mail of a refusal is read
// from the login state, and the tokens never leave the API. Both answer a
// redirect, with its Location.
func TestTheLoginThroughAProviderTakesTheKeyInItsPathAndNoEmailOrToken(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths
	takes := map[string][]string{
		providerLoginPaths[0]: {"path key", "query return_to"},
		providerLoginPaths[1]: {"path key", "query code", "query state", "query error", "query error_description"},
	}
	redirects := map[string]string{providerLoginPaths[0]: "302", providerLoginPaths[1]: "303"}

	for path, parameters := range takes {
		require.Contains(t, paths, path)
		item := paths[path]
		assert.Empty(t, item.Parameters, path)
		require.Len(t, operationsOf(item), 1, "%s has another operation", path)
		assert.ElementsMatch(t, parameters, parametersOf(item.Get), path)
		redirect := item.Get.Responses[redirects[path]]
		require.NotNil(t, redirect, "%s answers no redirect", path)
		assert.Contains(t, redirect.Headers, "Location", path)
		assert.Contains(t, item.Get.Responses["404"].Description, "`not_found`", path)
	}
}

// The callback checks nothing of what the provider writes in its query, so
// it never answers `validation_failed`, and its spec declares no 422; the
// start checks the length of its path to return to, and declares it.
func TestTheCallbackOfAProviderDeclaresNoFailedValidation(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths

	assert.NotContains(t, paths[providerLoginPaths[1]].Get.Responses, "422")
	assert.Contains(t, paths[providerLoginPaths[0]].Get.Responses, "422")
}

// parametersOf lists where each parameter of the operation is and its name:
// `query return_to`.
func parametersOf(operation *huma.Operation) []string {
	taken := []string{}
	for _, parameter := range operation.Parameters {
		taken = append(taken, parameter.In+" "+parameter.Name)
	}
	return taken
}

// assertTakesNothingOutsideTheBody checks a path of the setup or of the
// session, but the two of a login through a provider: it has no variable,
// and its operations take no parameter.
func assertTakesNothingOutsideTheBody(t *testing.T, path string, item *huma.PathItem) {
	t.Helper()
	if !strings.HasPrefix(path, "/api/setup") && !strings.HasPrefix(path, "/api/session") {
		return
	}
	if slices.Contains(providerLoginPaths, path) {
		return
	}
	assert.NotContains(t, path, "{", "a path takes a value: %s", path)
	assert.Empty(t, item.Parameters, "the operations of %s take a parameter outside the body", path)
	for _, operation := range operationsOf(item) {
		assert.Empty(t, operation.Parameters, "%s %s takes a parameter outside the body", operation.Method, path)
	}
}

// operationsOf lists the operations a path has.
func operationsOf(item *huma.PathItem) []*huma.Operation {
	every := []*huma.Operation{item.Get, item.Put, item.Post, item.Delete, item.Options, item.Head, item.Patch, item.Trace}
	return slices.DeleteFunc(every, func(operation *huma.Operation) bool { return operation == nil })
}

// Caddy limits the POST operations of the setup for each IP address and
// answers 429 before the request reaches the API. The spec declares it, so the
// client generated from it knows the answer.
func TestEveryPostOperationOfTheSetupDeclaresTheLimitOfCaddy(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths
	posts := maps.Clone(paths)
	maps.DeleteFunc(posts, func(path string, item *huma.PathItem) bool {
		return !strings.HasPrefix(path, "/api/setup") || item.Post == nil
	})

	require.Len(t, posts, 4, "the POST operations of the setup")
	for path, item := range posts {
		limited := item.Post.Responses["429"]
		require.NotNil(t, limited, "POST %s does not declare the 429 of Caddy", path)
		require.Contains(t, limited.Content, "application/problem+json", "the 429 of POST %s", path)
		assert.Equal(t, "#/components/schemas/Problem", limited.Content["application/problem+json"].Schema.Ref)
		assert.Contains(t, limited.Description, "`too_many_requests`", "the 429 of POST %s", path)
	}
	assert.NotContains(t, paths["/api/setup"].Get.Responses, "429", "Caddy does not limit reading the state of the setup")
}

// Caddy limits the two operations of the login code for each IP address, in
// a zone of their own: asking for a code and checking it. The other
// operations of the session have no limit.
func TestEveryOperationOfTheLoginCodeDeclaresTheLimitOfCaddy(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths
	posts := maps.Clone(paths)
	maps.DeleteFunc(posts, func(path string, item *huma.PathItem) bool {
		return !strings.HasPrefix(path, "/api/session/login_code") || item.Post == nil
	})

	require.Len(t, posts, 2, "the POST operations of the login code")
	for path, item := range posts {
		limited := item.Post.Responses["429"]
		require.NotNil(t, limited, "POST %s does not declare the 429 of Caddy", path)
		assert.Contains(t, limited.Description, "`too_many_requests`", "the 429 of POST %s", path)
	}
	assert.NotContains(t, paths["/api/session/log_out"].Post.Responses, "429")
	assert.NotContains(t, paths["/api/session"].Get.Responses, "429")
}

// Who is logged in and the consultancy need a session, and their spec says
// so: 401 with the code the app reads as a session that ended.
func TestWhoIsLoggedInAndTheConsultancyDeclareThe401OfNoSession(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths
	for path, operationID := range map[string]string{"/api/session": "tell_logged_in_user", "/api/consultancy": "tell_consultancy"} {
		require.Contains(t, paths, path)
		read := paths[path].Get
		require.NotNil(t, read, "GET %s", path)

		require.Contains(t, read.Responses, "401", "GET %s", path)
		assert.Contains(t, read.Responses["401"].Description, "`unauthorized`", "GET %s", path)
		assert.Equal(t, operationID, read.OperationID)
	}
}

func TestEveryFailureOfAnOperationIsDescribedByTheErrorFormat(t *testing.T) {
	t.Parallel()
	spec := generatedSpec(t)
	open := spec.Paths["/api/setup/open"].Post
	require.NotNil(t, open)

	for _, status := range []string{"422", "500"} {
		require.Contains(t, open.Responses, status)
		content := open.Responses[status].Content
		require.Contains(t, content, "application/problem+json", "the %s of open_setup", status)
		assert.Equal(t, "#/components/schemas/Problem", content["application/problem+json"].Schema.Ref)
	}
	problem := spec.Components.Schemas.Map()["Problem"]
	require.NotNil(t, problem)
	assert.ElementsMatch(t, []string{"title", "status", "code", "params", "trace_id"}, problem.Required)
	assert.Contains(t, open.Responses["422"].Description, "setup_code_invalid")
	assert.Contains(t, open.Responses["422"].Description, "validation_failed")
}
