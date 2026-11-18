package main

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// output collects what the command writes.
type output struct {
	mu   sync.Mutex
	text bytes.Buffer
}

func (o *output) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text.Write(p)
}

func (o *output) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text.String()
}

// seedEnvironment gives the two variables the seed reads, pointed at a
// temporary database.
func seedEnvironment(temporary databasetest.Database) map[string]string {
	return map[string]string{
		"DATABASE_URL": temporary.APIURL,
		"SEED_MAILBOX": "seed@fabrica.example",
	}
}

// seedIn runs the seed on the database, expects it to succeed and returns
// what it printed.
func seedIn(t *testing.T, temporary databasetest.Database) string {
	t.Helper()
	var stdout, stderr output
	status := run(context.Background(), nil, seedEnvironment(temporary), &stdout, &stderr)
	require.Equal(t, 0, status, stderr.String())
	require.Empty(t, stderr.String(), "the seed wrote to its error output")
	return stdout.String()
}

// asOwner connects as the migration role, which owns every table and so
// reads what the API role wrote.
func asOwner(t *testing.T, temporary databasetest.Database) *pgx.Conn {
	t.Helper()
	connection, err := pgx.Connect(context.Background(), temporary.MigrationURL)
	require.NoError(t, err)
	t.Cleanup(func() { connection.Close(context.Background()) })
	return connection
}

// registeredUser is a user as the seed leaves them, with the keys of their
// roles.
type registeredUser struct {
	Name, Email, PreferredLanguage string
	Deactivated                    bool
	Roles                          []string
}

func usersOf(t *testing.T, owner *pgx.Conn) []registeredUser {
	t.Helper()
	rows, err := owner.Query(context.Background(), `
		SELECT users.name, users.email, users.preferred_language, users.deactivated,
		       array_agg(roles.key ORDER BY roles.key)
		FROM users
		JOIN user_roles ON user_roles.user_id = users.id
		JOIN roles ON roles.id = user_roles.role_id
		GROUP BY users.id ORDER BY users.id`)
	require.NoError(t, err)
	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[registeredUser])
	require.NoError(t, err)
	return users
}

// seededData returns what the seed leaves in the database, with its
// identifiers and without what is the moment of the run: the times, and the
// hash of the setup code, which is random.
func seededData(t *testing.T, owner *pgx.Conn) string {
	t.Helper()
	ctx := context.Background()
	statements := []string{
		"SELECT row_to_json(each)::text FROM (SELECT * FROM consultancy ORDER BY id) AS each",
		"SELECT row_to_json(each)::text FROM (SELECT * FROM roles ORDER BY id) AS each",
		"SELECT row_to_json(each)::text FROM (SELECT * FROM users ORDER BY id) AS each",
		"SELECT row_to_json(each)::text FROM (SELECT * FROM user_roles ORDER BY user_id, role_id) AS each",
		"SELECT row_to_json(each)::text FROM (SELECT id, ended_at IS NOT NULL AS ended FROM setup_codes ORDER BY id) AS each",
		"SELECT row_to_json(each)::text FROM (SELECT id, operation, actor_kind, actor_id, subject_kind, subject_id FROM audit_log ORDER BY id) AS each",
	}
	var everything strings.Builder
	for _, statement := range statements {
		rows, err := owner.Query(ctx, statement)
		require.NoError(t, err)
		lines, err := pgx.CollectRows(rows, pgx.RowTo[string])
		require.NoError(t, err)
		everything.WriteString(strings.Join(lines, "\n") + "\n")
	}
	return everything.String()
}

// everyRow returns every row of every table of the database, as text, the
// times included.
func everyRow(t *testing.T, owner *pgx.Conn) string {
	t.Helper()
	ctx := context.Background()
	rows, err := owner.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")
	require.NoError(t, err)
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	require.Contains(t, tables, "consultancy")

	var everything strings.Builder
	for _, table := range tables {
		var content string
		statement := "SELECT coalesce(string_agg(row_to_json(each)::text, E'\\n'), '') FROM " + pgx.Identifier{table}.Sanitize() + " AS each"
		require.NoError(t, owner.QueryRow(ctx, statement).Scan(&content))
		everything.WriteString(table + ":\n" + content + "\n")
	}
	return everything.String()
}

func TestTwoRunsOnEmptyDatabasesProduceTheSameDataWithTheSameIdentifiers(t *testing.T) {
	t.Parallel()
	first := databasetest.Migrated(t)
	second := databasetest.Migrated(t)

	seedIn(t, first)
	seedIn(t, second)

	firstData := seededData(t, asOwner(t, first))
	assert.Equal(t, firstData, seededData(t, asOwner(t, second)))
	assert.Contains(t, firstData, `"id":1,"name":"Fábrica Exemplo"`)
	assert.Contains(t, firstData, `"id":1,"name":"Ana Exemplo"`)
}

func TestASecondRunOnASeededDatabaseChangesNothingAndSucceeds(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	seedIn(t, temporary)
	owner := asOwner(t, temporary)
	before := everyRow(t, owner)

	printed := seedIn(t, temporary)

	assert.Equal(t, before, everyRow(t, owner))
	assert.Contains(t, printed, "changes nothing")
}

func TestTheOutputOfTheSeedCarriesNoEmailAddress(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	printed := seedIn(t, temporary)

	assert.NotContains(t, printed, "@")
	assert.Contains(t, printed, "SEED_MAILBOX", "the output does not say how to log in")
}

func TestTheSeedFailsWhileAMigrationIsPending(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	var stdout, stderr output

	status := run(context.Background(), nil, seedEnvironment(temporary), &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Regexp(t, `^seed: the database schema is at version 0 and this binary expects version \d+\. Apply the migrations with the migrate command of the API first\n$`, stderr.String())
	assert.Empty(t, stdout.String())
}

func TestTheSeedRefusesToRunWithoutTheMailbox(t *testing.T) {
	var stdout, stderr output
	environment := map[string]string{
		"DATABASE_URL": "postgres://api:hunter2@localhost:5432/software_factory",
	}

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "seed: configuration: SEED_MAILBOX is required\n", stderr.String())
	assert.Empty(t, stdout.String())
}

func TestTheSeedFailsWhenPostgreSQLDoesNotAnswer(t *testing.T) {
	t.Parallel()
	var stdout, stderr output
	environment := map[string]string{
		// Port 1 of the machine has no PostgreSQL behind it.
		"DATABASE_URL": "postgres://api:hunter2@127.0.0.1:1/software_factory?sslmode=disable",
		"SEED_MAILBOX": "seed@fabrica.example",
	}

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stderr.String(), "seed: PostgreSQL does not answer")
	assert.NotContains(t, stderr.String(), "hunter2")
	assert.Empty(t, stdout.String())
}

func TestAnArgumentIsRefusedWithTheUsage(t *testing.T) {
	var stdout, stderr output

	status := run(context.Background(), []string{"everything"}, map[string]string{}, &stdout, &stderr)

	assert.Equal(t, 2, status)
	assert.Contains(t, stderr.String(), `seed: unexpected argument "everything"`)
	assert.Contains(t, stderr.String(), "Usage: seed")
	assert.Empty(t, stdout.String())
}

func TestAskingForHelpPrintsTheUsageAndSucceeds(t *testing.T) {
	var stdout, stderr output

	status := run(context.Background(), []string{"-h"}, map[string]string{}, &stdout, &stderr)

	assert.Equal(t, 0, status)
	assert.Contains(t, stderr.String(), "Usage: seed")
}

func TestTheSeedRegistersTheConsultancyAndItsAdministratorAndConcludesTheSetup(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	printed := seedIn(t, temporary)

	owner := asOwner(t, temporary)
	var consultancy, legalName string
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT name, legal_name FROM consultancy").Scan(&consultancy, &legalName))
	assert.Equal(t, "Fábrica Exemplo", consultancy)
	assert.Equal(t, "Fábrica Exemplo Software Ltda.", legalName)
	assert.Equal(t, []registeredUser{{
		Name: "Ana Exemplo", Email: "seed+admin@fabrica.example", PreferredLanguage: "pt-BR",
		Deactivated: false, Roles: []string{"administrator"},
	}}, usersOf(t, owner))
	var waiting int
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT count(*) FROM setup_codes WHERE ended_at IS NULL").Scan(&waiting))
	assert.Equal(t, 0, waiting, "a setup code is still waiting after the seed")
	assert.Contains(t, printed, "Fábrica Exemplo")
}
