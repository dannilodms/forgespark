package main

import (
	"context"
	"crypto/sha256"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
)

// A setup code as it is printed: four groups of five characters of
// Crockford's Base32, which has no I, L, O or U.
var printedSetupCode = regexp.MustCompile(`\b[0-9A-HJKMNP-TV-Z]{5}(?:-[0-9A-HJKMNP-TV-Z]{5}){3}\b`)

// issueEnvironment gives the two variables `issue-setup-code` reads, pointed
// at a temporary database.
func issueEnvironment(temporary databasetest.Database) map[string]string {
	return map[string]string{
		"DATABASE_URL": temporary.APIURL,
		"APP_ORIGIN":   "https://factory.example.com",
	}
}

// issueSetupCodeIn runs the command, expects it to succeed and returns what
// it printed and the setup code found there.
func issueSetupCodeIn(t *testing.T, temporary databasetest.Database) (printed, code string) {
	t.Helper()
	var stdout, stderr output
	status := run(context.Background(), []string{"issue-setup-code"}, issueEnvironment(temporary), &stdout, &stderr)
	require.Equal(t, 0, status, stderr.String())
	require.Empty(t, stderr.String(), "the command wrote to its error output")
	codes := printedSetupCode.FindAllString(stdout.String(), -1)
	require.Len(t, codes, 1, "the command printed: %s", stdout.String())
	return stdout.String(), codes[0]
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

// everyRow returns every row of every table of the database, as text.
func everyRow(t *testing.T, owner *pgx.Conn) string {
	t.Helper()
	ctx := context.Background()
	rows, err := owner.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")
	require.NoError(t, err)
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	require.Contains(t, tables, "setup_codes")
	require.Contains(t, tables, "audit_log")

	var everything strings.Builder
	for _, table := range tables {
		var content string
		statement := "SELECT coalesce(string_agg(row_to_json(each)::text, E'\\n'), '') FROM " + pgx.Identifier{table}.Sanitize() + " AS each"
		require.NoError(t, owner.QueryRow(ctx, statement).Scan(&content))
		everything.WriteString(table + ":\n" + content + "\n")
	}
	return everything.String()
}

func TestIssueSetupCodePrintsTheCodeWhenItStopsWorkingWhatItOpensAndTheAddressOfTheSetup(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	before := time.Now()

	printed, code := issueSetupCodeIn(t, temporary)

	assert.Len(t, code, 23)
	assert.NotRegexp(t, `[ILOU]`, code)
	assert.Contains(t, printed, "the first setup")
	assert.Contains(t, printed, "https://factory.example.com/setup")
	stops := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} UTC`).FindString(printed)
	require.NotEmpty(t, stops, "the command printed no time: %s", printed)
	stopsAt, err := time.Parse("2006-01-02 15:04:05 MST", stops)
	require.NoError(t, err)
	assert.WithinRange(t, stopsAt, before.Add(time.Hour-2*time.Second), time.Now().Add(time.Hour+2*time.Second))
}

func TestTwoSetupCodesIssuedInARowDifferAndTheFirstIsEnded(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	_, first := issueSetupCodeIn(t, temporary)
	_, second := issueSetupCodeIn(t, temporary)

	assert.NotEqual(t, first, second)
	rows, err := asOwner(t, temporary).Query(context.Background(), "SELECT code_hash, ended_at IS NOT NULL FROM setup_codes ORDER BY id")
	require.NoError(t, err)
	type storedSetupCode struct {
		Hash  []byte
		Ended bool
	}
	stored, err := pgx.CollectRows(rows, pgx.RowToStructByPos[storedSetupCode])
	require.NoError(t, err)
	// What is stored of a code is the SHA-256 hash of its 20 characters.
	hashOf := func(code string) []byte {
		hash := sha256.Sum256([]byte(strings.ReplaceAll(code, "-", "")))
		return hash[:]
	}
	assert.Equal(t, []storedSetupCode{
		{Hash: hashOf(first), Ended: true},
		{Hash: hashOf(second), Ended: false},
	}, stored)
}

func TestTheSetupCodeReachesTheOutputOfTheCommandAndNothingElse(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	printed, code := issueSetupCodeIn(t, temporary)

	normalised := strings.ReplaceAll(code, "-", "")
	assert.Equal(t, 1, strings.Count(printed, code), "the code is printed once")
	assert.NotContains(t, printed, normalised)
	stored := everyRow(t, asOwner(t, temporary))
	for _, form := range []string{code, normalised, strings.ToLower(code), strings.ToLower(normalised)} {
		assert.NotContains(t, stored, form, "the database holds the setup code")
	}
}

func TestEachSetupCodeIssuedLeavesOneEntryInTheAuditLog(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)

	issueSetupCodeIn(t, temporary)
	issueSetupCodeIn(t, temporary)

	rows, err := asOwner(t, temporary).Query(context.Background(), `
		SELECT audit_log.operation, audit_log.actor_kind, audit_log.actor_id IS NULL, audit_log.subject_kind,
		       audit_log.occurred_at = setup_codes.issued_at
		FROM setup_codes
		LEFT JOIN audit_log ON audit_log.subject_id = setup_codes.id
		ORDER BY setup_codes.id`)
	require.NoError(t, err)
	type entry struct {
		Operation, ActorKind string
		NoActorID            bool
		SubjectKind          string
		AtTheTimeOfTheCode   bool
	}
	entries, err := pgx.CollectRows(rows, pgx.RowToStructByPos[entry])
	require.NoError(t, err)
	issued := entry{Operation: "setup_code_issued", ActorKind: "installer", NoActorID: true, SubjectKind: "setup_code", AtTheTimeOfTheCode: true}
	assert.Equal(t, []entry{issued, issued}, entries)
}

func TestAFailureToWriteTheAuditEntryLeavesNoSetupCodeStoredAndPrintsNone(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	owner := asOwner(t, temporary)
	_, earlier := issueSetupCodeIn(t, temporary)
	_, err := owner.Exec(context.Background(), "REVOKE INSERT ON audit_log FROM api")
	require.NoError(t, err)
	var stdout, stderr output

	status := run(context.Background(), []string{"issue-setup-code"}, issueEnvironment(temporary), &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stderr.String(), "api: issue the setup code: append the entry to the audit log")
	assert.Empty(t, stdout.String(), "a setup code that was not stored was printed")
	var stored, notEnded int
	err = owner.QueryRow(context.Background(), "SELECT count(*), count(*) FILTER (WHERE ended_at IS NULL) FROM setup_codes").Scan(&stored, &notEnded)
	require.NoError(t, err)
	assert.Equal(t, 1, stored, "the setup code of the failed command was stored")
	assert.Equal(t, 1, notEnded, "the failed command ended the earlier setup code %s", earlier)
}

func TestIssueSetupCodeFailsWhenPostgreSQLDoesNotAnswer(t *testing.T) {
	t.Parallel()
	var stdout, stderr output
	environment := map[string]string{
		// Port 1 of the machine has no PostgreSQL behind it.
		"DATABASE_URL": "postgres://api:hunter2@127.0.0.1:1/forgespark?sslmode=disable",
		"APP_ORIGIN":   "https://factory.example.com",
	}

	status := run(context.Background(), []string{"issue-setup-code"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stderr.String(), "api: PostgreSQL does not answer")
	assert.NotContains(t, stderr.String(), "hunter2")
	assert.Empty(t, stdout.String())
}

func TestIssueSetupCodeFailsWhileAMigrationIsPending(t *testing.T) {
	t.Parallel()
	temporary := databasetest.New(t)
	var stdout, stderr output

	status := run(context.Background(), []string{"issue-setup-code"}, issueEnvironment(temporary), &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Regexp(t, `^api: the database schema is at version 0 and this binary expects version \d+\. Apply the migrations with the migrate command first\n$`, stderr.String())
	assert.Empty(t, stdout.String())
}

func TestIssueSetupCodeRefusesToRunWithoutTheOriginOfTheApp(t *testing.T) {
	var stdout, stderr output
	environment := map[string]string{
		"DATABASE_URL": "postgres://api:hunter2@localhost:5432/forgespark",
	}

	status := run(context.Background(), []string{"issue-setup-code"}, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Equal(t, "api: configuration: APP_ORIGIN is required\n", stderr.String())
	assert.Empty(t, stdout.String())
}

// A health check is asked again every few seconds, so it gives up early. The
// command runs once, for a person who is waiting: a slow first connection is
// not a PostgreSQL that does not answer.
func TestIssueSetupCodeIsNotBoundByTheTimeoutOfAHealthCheck(t *testing.T) {
	t.Parallel()
	environment := issueEnvironment(databasetest.Migrated(t))
	environment["HEALTH_CHECK_TIMEOUT"] = "1ns"
	var stdout, stderr output

	status := run(context.Background(), []string{"issue-setup-code"}, environment, &stdout, &stderr)

	assert.Equal(t, 0, status, stderr.String())
	assert.Regexp(t, printedSetupCode, stdout.String())
}
