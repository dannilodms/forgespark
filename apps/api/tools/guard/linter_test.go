package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lint runs the linter of the module, as `task api:lint` does, on the
// packages of a pattern, and returns what it printed and whether it failed.
// The fixtures are packages under testdata, which `./...` leaves out.
func lint(t *testing.T, pattern string) (string, error) {
	t.Helper()
	command := exec.Command("go", "tool", "-modfile=golangci-lint.mod", "golangci-lint", "run", pattern)
	// The test runs in the directory of its package: the module is two up.
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestTheLinterFailsOnWhatItsConfigurationRejects(t *testing.T) {
	t.Parallel()

	output, err := lint(t, "./tools/guard/testdata/linter/...")

	var failure *exec.ExitError
	require.ErrorAs(t, err, &failure, output)
	assert.Equal(t, 1, failure.ExitCode(), output)
	// A file that gofmt accepts and gofumpt does not.
	assert.Contains(t, output, "unformatted.go:7:1: File is not properly formatted (gofumpt)")
	// A rule of the standard set.
	assert.Contains(t, output, "unchecked.go:8:10: Error return value of `os.Mkdir` is not checked (errcheck)")
	// A rule silenced without its name and without a reason.
	assert.Contains(t, output, "silenced.go:9:24: directive `//nolint` should mention specific linter")
	// Code that keeps the rules, and a generated file that does not.
	assert.NotContains(t, output, "clean.go")
	assert.NotContains(t, output, "generated.go")
	assert.Contains(t, output, "3 issues:")
}
