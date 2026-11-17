// Package databasetest gives a test its own empty database in the PostgreSQL
// of development, so tests run in parallel and never touch the development
// database.
package databasetest

import (
	"context"
	"crypto/rand"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// Database is a temporary database, reached by the two roles of the system.
// It starts empty: no migration was applied to it.
type Database struct {
	// MigrationURL connects as the migration role, which owns the database.
	MigrationURL string
	// APIURL connects as the API role.
	APIURL string
}

// New creates a database that lives until the test ends. The connections
// come from the variables `task api:test` loads: the superuser creates and
// removes the database, and the other two give the roles.
func New(t testing.TB) Database {
	t.Helper()
	adminURL := variable(t, "TEST_DATABASE_ADMIN_URL")
	migrationURL := parse(t, "MIGRATION_DATABASE_URL")
	apiURL := parse(t, "DATABASE_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	name := "test_" + strings.ToLower(rand.Text())
	owner := migrationURL.User.Username()
	execute(t, ctx, adminURL, "CREATE DATABASE "+quote(name)+" OWNER "+quote(owner))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		execute(t, ctx, adminURL, "DROP DATABASE "+quote(name)+" WITH (FORCE)")
	})

	return Database{
		MigrationURL: onDatabase(migrationURL, name),
		APIURL:       onDatabase(apiURL, name),
	}
}

func variable(t testing.TB, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Fatalf("%s is not set. Run the tests with `task api:test`, which loads it from .env.example.", name)
	}
	return value
}

func parse(t testing.TB, name string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(variable(t, name))
	if err != nil {
		t.Fatalf("%s is not a URL.", name)
	}
	return parsed
}

func execute(t testing.TB, ctx context.Context, adminURL, statement string) {
	t.Helper()
	connection, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("PostgreSQL does not answer. Start it with `task services`. %v", err)
	}
	defer connection.Close(ctx)
	if _, err := connection.Exec(ctx, statement); err != nil {
		t.Fatalf("%s: %v", statement, err)
	}
}

func onDatabase(connection *url.URL, name string) string {
	moved := *connection
	moved.Path = "/" + name
	return moved.String()
}

func quote(identifier string) string {
	return pgx.Identifier{identifier}.Sanitize()
}
