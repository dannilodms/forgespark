// Package databasetest gives a test its own database in the PostgreSQL of
// development, so tests run in parallel and never touch the development
// database. Migrated gives one with every migration applied, which is what an
// adapter test wants. New gives an empty one, for the tests of the migrations
// themselves.
package databasetest

import (
	"context"
	"crypto/rand"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	// The driver pgtestdb opens its connections with.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// Database is a temporary database, reached by the two roles of the system.
type Database struct {
	// MigrationURL connects as the migration role, which owns the database.
	MigrationURL string
	// APIURL connects as the API role.
	APIURL string
}

// MigrationConnection is MigrationURL as the configuration holds a
// connection, for what takes one: the migrations and the pool.
func (d Database) MigrationConnection(t testing.TB) config.PostgresURL {
	t.Helper()
	return connection(t, d.MigrationURL)
}

// APIConnection is APIURL as the configuration holds a connection, for the
// pool.
func (d Database) APIConnection(t testing.TB) config.PostgresURL {
	t.Helper()
	return connection(t, d.APIURL)
}

func connection(t testing.TB, url string) config.PostgresURL {
	t.Helper()
	read, err := config.NewPostgresURL(url)
	if err != nil {
		t.Fatalf("the connection of a temporary database is not valid: %v", err)
	}
	return read
}

// Migrated creates a database with every migration applied, which lives until
// the test ends. pgtestdb clones it from a template that the migrations built
// once: the template is built again only when a migration file changes, so a
// test pays for a copy and never for the migrations.
//
// The connections come from the variables `task api:test` loads: the
// superuser creates and removes the databases, and the other two give the
// roles.
func Migrated(t testing.TB) Database {
	t.Helper()
	admin := parse(t, "TEST_DATABASE_ADMIN_URL")
	migrationURL := parse(t, "MIGRATION_DATABASE_URL")
	apiURL := parse(t, "DATABASE_URL")
	adminPassword, _ := admin.User.Password()
	migrationPassword, _ := migrationURL.User.Password()

	instance := pgtestdb.Custom(t, pgtestdb.Config{
		DriverName: "pgx",
		Host:       admin.Hostname(),
		Port:       admin.Port(),
		User:       admin.User.Username(),
		Password:   adminPassword,
		Database:   strings.TrimPrefix(admin.Path, "/"),
		Options:    admin.RawQuery,
		// The template and each copy belong to the migration role, as the
		// development database does. The role already exists, so pgtestdb
		// creates none.
		TestRole: &pgtestdb.Role{
			Username:     migrationURL.User.Username(),
			Password:     migrationPassword,
			Capabilities: "NOSUPERUSER NOCREATEDB NOCREATEROLE",
		},
		ForceTerminateConnections: true,
	}, migrator{})

	return Database{
		MigrationURL: onDatabase(migrationURL, instance.Database),
		APIURL:       onDatabase(apiURL, instance.Database),
	}
}

// New creates an empty database that lives until the test ends: no migration
// was applied to it. The connections come from the same variables as the
// ones of Migrated.
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

// variable is a variable of the process, which the configuration reads.
func variable(t testing.TB, name string) string {
	t.Helper()
	value := config.Environment()[name]
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
