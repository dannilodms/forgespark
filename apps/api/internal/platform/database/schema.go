package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

// CheckSchema tells whether PostgreSQL answers and its schema is the version
// this binary expects, which is the newest of the migrations it carries. It
// is the readiness check, and what a command runs before it uses the
// database. A nil error means both hold.
func CheckSchema(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS) error {
	expected, err := newestVersion(migrations)
	if err != nil {
		return err
	}
	applied, err := appliedVersion(ctx, pool)
	if err != nil {
		return fmt.Errorf("PostgreSQL does not answer: %w", err)
	}
	if applied != expected {
		return fmt.Errorf("the database schema is at version %d and this binary expects version %d", applied, expected)
	}
	return nil
}

func newestVersion(migrations fs.FS) (int64, error) {
	files, err := fs.Glob(migrations, "*.sql")
	if err != nil {
		return 0, fmt.Errorf("list the migrations: %w", err)
	}
	var newest int64
	for _, file := range files {
		version, err := goose.NumericComponent(file)
		if err != nil {
			return 0, fmt.Errorf("read the version of the migration %s: %w", file, err)
		}
		newest = max(newest, version)
	}
	return newest, nil
}

// appliedVersion reads the newest version in the table goose keeps. A
// database that never ran a migration has no such table and is at version 0.
func appliedVersion(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var version int64
	err := pool.QueryRow(ctx, "SELECT max(version_id) FROM "+goose.DefaultTablename).Scan(&version)
	if isUndefinedTable(err) {
		return 0, nil
	}
	return version, err
}

func isUndefinedTable(err error) bool {
	const undefinedTable = "42P01"
	var failure *pgconn.PgError
	return errors.As(err, &failure) && failure.Code == undefinedTable
}
