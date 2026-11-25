package main

import (
	"context"
	"fmt"
	"io"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/forgespark/apps/api/migrations"
)

// migrate applies the migrations the database lacks, as the migration role,
// and prints each one. It is its own step, run before the API starts. It
// writes no line of its own on the error output but its failure; a line a
// library writes there is a JSON line of its logger.
func migrate(ctx context.Context, environment map[string]string, stdout, stderr io.Writer) error {
	settings, err := config.Load(config.Migrate, environment)
	if err != nil {
		return err
	}
	telemetry.RouteTheStandardLogger(telemetry.NewLogger(stderr, settings.LogLevel))

	applied, err := database.Migrate(ctx, settings.MigrationDatabaseURL, migrations.Files)
	if err != nil {
		return err
	}
	for _, migration := range applied {
		fmt.Fprintf(stdout, "applied %s\n", migration)
	}
	if len(applied) == 0 {
		fmt.Fprintln(stdout, "no migration to apply")
	}
	return nil
}
