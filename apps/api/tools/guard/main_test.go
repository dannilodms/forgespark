package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dannilodms/software-factory-manager/apps/api/tools/guard/guardtest"
)

const goMod = "module example.test/api\n\ngo 1.27.4\n"

func TestCodeThatBreaksARuleFailsAndIsListed(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"go.mod": goMod,
		"internal/consultancy/domain/consultancy.go": `package domain

import "example.test/api/internal/consultancy/application"

func Name(known bool) string {
	if known {
		return application.Name
	} else {
		return ""
	}
}
`,
	})
	var stdout, stderr bytes.Buffer

	status := run([]string{"code", root}, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, `internal/consultancy/domain/consultancy.go:3: imports "example.test/api/internal/consultancy/application": a domain package imports the standard library, kernel and the domain of its own area (Dependency rule in CODING_STANDARDS.md)
internal/consultancy/domain/consultancy.go:8: else: handle the exceptional case first and return (Control flow in CODING_STANDARDS.md)
`, stdout.String())
	assert.Equal(t, "guard: 2 findings\n", stderr.String())
}

func TestCodeThatKeepsTheRulesPasses(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"go.mod":                             goMod,
		"internal/kernel/country/country.go": "package country\n",
	})
	var stdout, stderr bytes.Buffer

	status := run([]string{"code", root}, &stdout, &stderr)

	assert.Equal(t, 0, status)
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestAVersionThatIsNotExactFailsAndIsListed(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"deploy/compose.base.yaml": "services:\n  postgres:\n    image: postgres:latest\n",
	})
	var stdout, stderr bytes.Buffer

	status := run([]string{"versions", root}, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stdout.String(), `deploy/compose.base.yaml:3: image "postgres:latest" names no exact version`)
	assert.Equal(t, "guard: 1 finding\n", stderr.String())
}

func TestExactVersionsPass(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"deploy/compose.base.yaml": "services:\n  postgres:\n    image: postgres:18.7-trixie\n",
	})
	var stdout, stderr bytes.Buffer

	status := run([]string{"versions", root}, &stdout, &stderr)

	assert.Equal(t, 0, status)
	assert.Empty(t, stdout.String())
}

func TestACheckThatCannotReadItsDirectoryFails(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer

	status := run([]string{"code", t.TempDir()}, &stdout, &stderr)

	assert.Equal(t, 2, status)
	assert.Contains(t, stderr.String(), "go.mod")
}

func TestAnUnknownCommandIsNotUnderstood(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"no command":       nil,
		"an unknown one":   {"lint"},
		"too many folders": {"code", ".", "."},
	}
	for name, arguments := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer

			status := run(arguments, &stdout, &stderr)

			assert.Equal(t, 2, status)
			assert.Contains(t, stderr.String(), "Usage: guard")
		})
	}
}
