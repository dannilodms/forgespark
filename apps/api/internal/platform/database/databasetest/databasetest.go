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
// test pays for a copy and never for the migrations. The copy also gets the
// rights on the database that the migrations set on the template, which
// PostgreSQL does not copy.
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
	keepTheRightsOfTheTemplate(t, admin.String(), instance.Database)

	return Database{
		MigrationURL: onDatabase(migrationURL, instance.Database),
		APIURL:       onDatabase(apiURL, instance.Database),
	}
}

// keepTheRightsOfTheTemplate gives a copy the rights on the database that the
// migrations set on its template, such as the one the API role lost to
// create temporary tables. PostgreSQL copies everything in a template but
// those: a new database, a copy included, starts with the rights PostgreSQL
// gives by default. Only the superuser writes them this way, as pgtestdb
// marks a template.
func keepTheRightsOfTheTemplate(t testing.TB, adminURL, copied string) {
	t.Helper()
	// pgtestdb names a copy after its template: `<template>_inst_<random>`.
	template, _, isACopy := strings.Cut(copied, "_inst_")
	if !isACopy {
		t.Fatalf("pgtestdb no longer names a copy %q after its template; databasetest has to find the template another way.", copied)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	connection, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("PostgreSQL does not answer. Start it with `task services`. %v", err)
	}
	defer connection.Close(ctx)
	copiedRights, err := connection.Exec(ctx, `
		UPDATE pg_database SET datacl = template.datacl
		FROM pg_database AS template
		WHERE pg_database.datname = $1 AND template.datname = $2`, copied, template)
	if err != nil {
		t.Fatalf("give %s the rights of its template: %v", copied, err)
	}
	if copiedRights.RowsAffected() != 1 {
		t.Fatalf("the template of %s, %s, does not exist.", copied, template)
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
