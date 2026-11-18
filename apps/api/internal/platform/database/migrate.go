// Package database holds what every domain area shares to reach PostgreSQL:
// the pool, the migrations and the check of the schema version. Rules:
// Database in CODING_STANDARDS.md.
package database

import (
	"context"
	"fmt"
	"io/fs"
	"path"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// Migrate applies the migrations the database lacks, in order, and returns
// the file name of each one it applied. A run with nothing pending applies
// nothing. The connection is the one of the migration role, and goose holds
// its lock during the run, so two runs never apply the same migration.
func Migrate(ctx context.Context, connectionURL config.PostgresURL, migrations fs.FS) ([]string, error) {
	connection, err := pgx.ParseConfig(connectionURL.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("the connection of the migration role is not valid")
	}
	db := stdlib.OpenDB(*connection)
	defer db.Close()

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("prepare the migration lock: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations, goose.WithSessionLocker(locker))
	if err != nil {
		return nil, fmt.Errorf("read the migrations: %w", err)
	}
	results, err := provider.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf("apply the migrations: %w", err)
	}

	applied := make([]string, 0, len(results))
	for _, result := range results {
		applied = append(applied, path.Base(result.Source.Path))
	}
	return applied, nil
}
