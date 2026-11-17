package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5/pgxpool"

	auditpostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/adapters/postgres"
	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	consultancypostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/adapters/postgres"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	consultancydomain "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/clock"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// setupPath is where the app shows the setup, after its origin.
const setupPath = "/setup"

// commandTracer names the tracer that opens the span of a command.
const commandTracer = "github.com/dannilodms/software-factory-manager/apps/api/cmd/api"

// issueSetupCode issues a setup code and prints it, with when it stops
// working, what it opens and the address of the setup.
//
// The code goes to the output of the command, which is the terminal of
// whoever ran it and not the output of the container, and nowhere else. The
// logs, on the error output, hold a failed export of traces and nothing of
// the code. A span holds the SQL text of a query and never its parameters,
// and the database keeps the hash.
//
// When it returns, the span of the command ends, the pool is closed and then
// the last spans are sent, in that order.
func issueSetupCode(ctx context.Context, environment map[string]string, stdout, stderr io.Writer) error {
	settings, err := config.Load(config.IssueSetupCode, environment)
	if err != nil {
		return err
	}

	logger := telemetry.NewLogger(stderr, settings.LogLevel)
	tracing, err := telemetry.StartTracing("api", settings.TelemetryExportURL.String(), logger)
	if err != nil {
		return err
	}
	defer tracing.Stop()

	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            settings.DatabaseURL.ConnectionString(),
		MaxSize:                  settings.DatabasePoolMaxSize,
		StatementTimeout:         settings.DatabaseStatementTimeout,
		IdleInTransactionTimeout: settings.DatabaseIdleInTransactionTimeout,
		TracerProvider:           tracing.Provider,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	// One span around the work, so the queries of the command are one trace.
	ctx, span := tracing.Provider.Tracer(commandTracer).Start(ctx, "issue-setup-code")
	defer span.End()
	if err := checkSchema(ctx, pool, settings); err != nil {
		return err
	}

	issue := consultancyapplication.NewIssueSetupCode(
		database.NewTransactions(pool),
		consultancypostgres.NewSetupCodeRepository(pool),
		consultancypostgres.NewConsultancyRepository(pool),
		auditapplication.NewRecordEntry(auditpostgres.NewAuditLog(pool)),
		clock.System{},
		rand.Reader,
	)
	issued, err := issue.Execute(ctx)
	if err != nil {
		return fmt.Errorf("issue the setup code: %w", err)
	}

	fmt.Fprintf(stdout, issuedSetupCodeText,
		issued.Code.Reveal(),
		openedBy(issued.Opens),
		issued.ExpiresAt.UTC().Format("2006-01-02 15:04:05 MST"),
		settings.AppOrigin.String()+setupPath,
	)
	return nil
}

const issuedSetupCodeText = `Setup code:        %s
It opens:          %s
It stops working:  %s
Address:           %s

Type the code at the address. A new setup code ends this one.
`

// openedBy says in words what a setup code opens. The way back in also says
// that it leaves the users that exist as they are.
func openedBy(opens consultancydomain.SetupCodeOpens) string {
	if opens == consultancydomain.OpensNewAdministrator {
		return "the registration of a new administrator. The users that exist are not changed."
	}
	return "the first setup, which registers the consultancy and its first administrator"
}

// checkSchema stops the command, with what to do, when PostgreSQL does not
// answer or a migration is pending.
//
// It waits as long as one statement may run, and not the timeout of a health
// check: a probe is asked again every few seconds and gives up early, while
// this runs once, for a person who is waiting. A server that refuses the
// connection still fails at once.
func checkSchema(ctx context.Context, pool *pgxpool.Pool, settings config.Config) error {
	ctx, cancel := context.WithTimeout(ctx, settings.DatabaseStatementTimeout)
	defer cancel()

	err := database.CheckSchema(ctx, pool, migrations.Files)
	var version database.SchemaVersionError
	if errors.As(err, &version) && version.Applied < version.Expected {
		return fmt.Errorf("%w. Apply the migrations with the migrate command first", err)
	}
	return err
}
