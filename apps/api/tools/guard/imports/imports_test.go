package imports_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/guardtest"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/imports"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

const goMod = "module example.test/api\n\ngo 1.27.4\n"

// source writes a Go file of the package that imports the paths.
func source(pkg string, paths ...string) string {
	text := "package " + pkg + "\n\nimport (\n"
	for _, path := range paths {
		text += "\t_ \"" + path + "\"\n"
	}
	return text + ")\n"
}

// Each case is one file under internal and one import the dependency rule
// forbids for the place of that file.
func TestAForbiddenImportIsRejected(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		file     string
		imported string
	}{
		"domain to application": {
			file:     "internal/consultancy/domain/consultancy.go",
			imported: "example.test/api/internal/consultancy/application",
		},
		"domain to an adapter": {
			file:     "internal/consultancy/domain/consultancy.go",
			imported: "example.test/api/internal/consultancy/adapters/postgres",
		},
		"domain to platform": {
			file:     "internal/consultancy/domain/consultancy.go",
			imported: "example.test/api/internal/platform/database",
		},
		"domain to a third-party library": {
			file:     "internal/consultancy/domain/consultancy.go",
			imported: "github.com/jackc/pgx/v5",
		},
		"domain to the domain of another area": {
			file:     "internal/consultancy/domain/consultancy.go",
			imported: "example.test/api/internal/files/domain",
		},
		"application to an adapter": {
			file:     "internal/consultancy/application/conclude.go",
			imported: "example.test/api/internal/consultancy/adapters/postgres",
		},
		"application to the domain of another area": {
			file:     "internal/consultancy/application/conclude.go",
			imported: "example.test/api/internal/files/domain",
		},
		"application to a third-party library": {
			file:     "internal/consultancy/application/conclude.go",
			imported: "github.com/jackc/pgx/v5",
		},
		"an adapter to another adapter of its area": {
			file:     "internal/consultancy/adapters/rest/setup.go",
			imported: "example.test/api/internal/consultancy/adapters/postgres",
		},
		"an adapter to an adapter of another area": {
			file:     "internal/consultancy/adapters/postgres/consultancy.go",
			imported: "example.test/api/internal/files/adapters/postgres",
		},
		"an adapter to the application of another area": {
			file:     "internal/consultancy/adapters/rest/setup.go",
			imported: "example.test/api/internal/files/application",
		},
		"an adapter to the domain of another area": {
			file:     "internal/consultancy/adapters/rest/setup.go",
			imported: "example.test/api/internal/files/domain",
		},
		"a test to the domain of another area": {
			file:     "internal/consultancy/application/conclude_test.go",
			imported: "example.test/api/internal/files/domain",
		},
		"kernel to a third-party library": {
			file:     "internal/kernel/language/language.go",
			imported: "golang.org/x/text/language",
		},
		"kernel to platform": {
			file:     "internal/kernel/language/language.go",
			imported: "example.test/api/internal/platform/config",
		},
		"platform to a domain area": {
			file:     "internal/platform/health/health.go",
			imported: "example.test/api/internal/consultancy/application",
		},
		"a package that is outside the layout": {
			file:     "internal/platform/health/health.go",
			imported: "example.test/api/internal/consultancy/helpers",
		},
	}
	for name, forbidden := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := guardtest.Tree(t, map[string]string{
				"go.mod":       goMod,
				forbidden.file: source("guarded", "context", forbidden.imported),
			})

			findings, err := imports.Check(root)

			require.NoError(t, err)
			require.Len(t, findings, 1)
			assert.Equal(t, forbidden.file, findings[0].File)
			assert.Equal(t, 5, findings[0].Line)
			assert.Contains(t, findings[0].Message, `imports "`+forbidden.imported+`"`)
		})
	}
}

// The tree holds every kind of import the dependency rule allows, in areas
// and packages the repository does not have yet.
func TestEveryAllowedImportPasses(t *testing.T) {
	t.Parallel()
	const module = "example.test/api/internal/"
	root := guardtest.Tree(t, map[string]string{
		"go.mod": goMod,
		"internal/kernel/country/country.go": source("country",
			"errors",
			module+"kernel/language",
		),
		"internal/kernel/country/country_test.go": source("country_test",
			"github.com/stretchr/testify/assert",
		),
		"internal/platform/database/pool.go": source("database",
			"context",
			"github.com/jackc/pgx/v5/pgxpool",
			module+"platform/config",
			module+"kernel/language",
		),
		"internal/platform/database/migrate_test.go": source("database_test",
			"example.test/api/migrations",
		),
		"internal/consultancy/domain/consultancy.go": source("domain",
			"net/mail",
			module+"kernel/country",
			module+"consultancy/domain/setupcode",
		),
		"internal/consultancy/domain/consultancy_test.go": source("domain_test",
			"github.com/stretchr/testify/require",
			module+"consultancy/domain",
		),
		"internal/consultancy/application/conclude.go": source("application",
			"context",
			module+"consultancy/domain",
			module+"kernel/country",
			module+"platform/clock",
			module+"files/application",
			module+"audit/application",
		),
		"internal/consultancy/application/conclude_test.go": source("application_test",
			"github.com/stretchr/testify/require",
			module+"consultancy/application",
			module+"consultancy/application/memory",
		),
		"internal/consultancy/adapters/postgres/consultancy.go": source("postgres",
			"github.com/jackc/pgx/v5",
			module+"consultancy/application",
			module+"consultancy/domain",
			module+"consultancy/adapters/postgres/sqlc",
			module+"kernel/country",
			module+"platform/database",
		),
		"internal/consultancy/adapters/postgres/sqlc/queries.sql.go": "// Code generated by sqlc. DO NOT EDIT.\n\n" + source("sqlc",
			"github.com/jackc/pgx/v5",
		),
		"internal/consultancy/adapters/postgres/consultancy_test.go": source("postgres_test",
			"example.test/api/migrations",
			module+"platform/database/databasetest",
		),
		"internal/consultancy/adapters/rest/setup.go": source("rest",
			"github.com/danielgtaylor/huma/v2",
			module+"consultancy/application",
		),
		"cmd/api/serve.go": source("main",
			module+"consultancy/adapters/postgres",
			module+"consultancy/adapters/rest",
			module+"files/adapters/s3",
		),
	})

	findings, err := imports.Check(root)

	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestAPackageOutsideTheLayoutIsRejectedOnce(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"go.mod":                               goMod,
		"internal/consultancy/helpers/one.go":  source("helpers"),
		"internal/consultancy/helpers/two.go":  source("helpers"),
		"internal/consultancy/adapters/all.go": source("adapters"),
		"internal/shared.go":                   source("internal"),
	})

	findings, err := imports.Check(root)

	require.NoError(t, err)
	require.Len(t, findings, 3)
	assert.Equal(t, "internal", findings[0].File)
	assert.Equal(t, "internal/consultancy/adapters", findings[1].File)
	assert.Equal(t, "internal/consultancy/helpers", findings[2].File)
	assert.Contains(t, findings[2].Message, "is outside the layout")
}

func TestGeneratedCodeIsJudgedLikeAnyOther(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"go.mod": goMod,
		"internal/consultancy/domain/generated.go": "// Code generated by a tool. DO NOT EDIT.\n\n" + source("domain",
			"github.com/jackc/pgx/v5",
		),
	})

	findings, err := imports.Check(root)

	require.NoError(t, err)
	assert.Len(t, findings, 1)
}

func TestFixturesAreNotJudged(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"go.mod": goMod,
		"internal/consultancy/domain/testdata/fixture.go": source("fixture", "github.com/jackc/pgx/v5"),
	})

	findings, err := imports.Check(root)

	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestAModuleWithNoInternalCodePasses(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{"go.mod": goMod, "cmd/api/main.go": source("main")})

	findings, err := imports.Check(root)

	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestDomainImportingApplicationIsRejected(t *testing.T) {
	t.Parallel()
	root := guardtest.Tree(t, map[string]string{
		"go.mod": goMod,
		"internal/consultancy/domain/consultancy.go": source("domain",
			"errors",
			"example.test/api/internal/consultancy/application",
		),
		"internal/consultancy/application/conclude.go": source("application"),
	})

	findings, err := imports.Check(root)

	require.NoError(t, err)
	assert.Equal(t, []report.Finding{{
		File:    "internal/consultancy/domain/consultancy.go",
		Line:    5,
		Message: `imports "example.test/api/internal/consultancy/application": a domain package imports the standard library, kernel and the domain of its own area (Dependency rule in CODING_STANDARDS.md)`,
	}}, findings)
}
