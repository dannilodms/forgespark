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
// with no database behind it, and returns the spec of what it registered.
func generatedSpec(t *testing.T) *huma.OpenAPI {
	t.Helper()
	return mountREST(http.NewServeMux(), nil, nil).OpenAPI()
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

// The setup code travels in the body of each operation of the setup. An
// operation that took it any other way would write it to the address, which
// the browser history, the logs and the spans keep.
func TestTheSetupCodeIsInNoPathQueryParameterOrHeaderOfTheSpec(t *testing.T) {
	t.Parallel()
	paths := generatedSpec(t).Paths
	require.Contains(t, paths, "/api/setup/open")

	for path, item := range paths {
		assertTakesNothingOutsideTheBody(t, path, item)
	}
}

// assertTakesNothingOutsideTheBody checks a path of the setup: it has no
// variable, and its operations take no parameter.
func assertTakesNothingOutsideTheBody(t *testing.T, path string, item *huma.PathItem) {
	t.Helper()
	if !strings.HasPrefix(path, "/api/setup") {
		return
	}
	assert.NotContains(t, path, "{", "a path of the setup takes a value: %s", path)
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

	require.Len(t, posts, 3, "the POST operations of the setup")
	for path, item := range posts {
		limited := item.Post.Responses["429"]
		require.NotNil(t, limited, "POST %s does not declare the 429 of Caddy", path)
		require.Contains(t, limited.Content, "application/problem+json", "the 429 of POST %s", path)
		assert.Equal(t, "#/components/schemas/Problem", limited.Content["application/problem+json"].Schema.Ref)
		assert.Contains(t, limited.Description, "`too_many_requests`", "the 429 of POST %s", path)
	}
	assert.NotContains(t, paths["/api/setup"].Get.Responses, "429", "Caddy does not limit reading the state of the setup")
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
