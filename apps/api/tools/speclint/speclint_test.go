// Package speclint holds the lint of the OpenAPI spec: the ruleset vacuum
// runs and the function it needs for the plural paths. The tests run vacuum
// on specs that keep the conventions and on specs that break them.
package speclint

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lint runs vacuum on a spec, as `task api:spec:lint` does on the committed
// one, and returns what it printed and whether it failed.
func lint(t *testing.T, spec string) (string, error) {
	t.Helper()
	// `go tool` looks at every package of vacuum before it runs it, which
	// takes seconds, and minutes on a busy machine. `task check` and the
	// pipeline run these tests; `task api:test -- -short ./...` leaves them
	// out.
	if testing.Short() {
		t.Skip("runs vacuum, which is slow to start")
	}
	path, err := filepath.Abs(spec)
	require.NoError(t, err)
	command := exec.Command("go", "tool", "-modfile=vacuum.mod", "vacuum", "lint",
		"--no-update-check", "--no-banner", "--no-style", "--details", "--all-results", "--no-clip",
		"--ruleset", "tools/speclint/ruleset.yaml",
		"--functions", "tools/speclint/functions",
		path,
	)
	// The test runs in the directory of its package: the module is two up.
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestTheLintPassesASpecThatKeepsEveryConvention(t *testing.T) {
	t.Parallel()

	output, err := lint(t, "testdata/conventional.yaml")

	require.NoError(t, err, output)
	// The paths that pass: `/api/setup` and `/api/health`, which the ruleset
	// lists as singular, a `POST` that ends in a verb, and plural resources.
	// The one property that is not snake_case and passes is `$schema`.
	assert.Contains(t, output, "pluralPaths", "the function of the plural paths was not loaded")
}

func TestTheLintFailsOnEachConventionASpecBreaks(t *testing.T) {
	t.Parallel()

	output, err := lint(t, "testdata/unconventional.yaml")

	var failure *exec.ExitError
	require.ErrorAs(t, err, &failure, output)
	assert.Equal(t, 1, failure.ExitCode(), output)
	// A singular noun that is not on the list.
	assert.Contains(t, output, "the path /api/project has the singular segment `project`")
	// A singular noun at the end of a path that is not a `POST`, so it is no verb.
	assert.Contains(t, output, "the path /api/projects/{projectId}/detail has the singular segment `detail`")
	assert.Contains(t, output, "the property legalName is not in snake_case")
	// A property of an object inside an object.
	assert.Contains(t, output, "the property PostalCode is not in snake_case")
	assert.Contains(t, output, "the query parameter sortBy is not in snake_case")
	assert.Contains(t, output, "the path parameter projectId is not in snake_case")
	assert.Contains(t, output, "Failed with 6 errors")
}

// thing is a body of the API the next test generates a spec from.
type thing struct {
	LegalName string `json:"legal_name"`
}

// Huma adds a `$schema` property to every body, unless its schema links are
// turned off as the API of this repository does. This proves the lint on a
// spec Huma generated with them on, so turning them back on breaks no check.
func TestTheLintPassesTheSpecHumaGeneratesWithItsSchemaLinks(t *testing.T) {
	t.Parallel()
	api := humago.New(http.NewServeMux(), huma.DefaultConfig("A spec with the defaults of Huma", "unversioned"))
	huma.Register(api, huma.Operation{
		OperationID: "rename_thing",
		Method:      http.MethodPost,
		Path:        "/api/things/{thing_id}/rename",
	}, func(_ context.Context, input *struct {
		ThingID string `path:"thing_id"`
		Notify  bool   `query:"notify_owner"`
		Body    thing
	},
	) (*struct{ Body thing }, error) {
		return &struct{ Body thing }{Body: input.Body}, nil
	})
	generated, err := api.OpenAPI().YAML()
	require.NoError(t, err)
	require.Contains(t, string(generated), "$schema:", "Huma no longer adds the property this test is about")
	require.Contains(t, string(generated), "ErrorModel", "the spec does not carry the error model of Huma")
	spec := filepath.Join(t.TempDir(), "openapi.yaml")
	require.NoError(t, os.WriteFile(spec, generated, 0o600))

	output, err := lint(t, spec)

	require.NoError(t, err, output)
}
