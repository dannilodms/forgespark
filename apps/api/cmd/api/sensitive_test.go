package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry/telemetrytest"
)

// The test of sensitive values (Telemetry in CODING_STANDARDS.md). It runs
// the commands of the binary through a scenario in which every sensitive
// value the system handles goes by, with the logs at their lowest level, and
// then searches the logs and the exported spans for each value.
//
// A ticket that brings a new kind of sensitive value (a setup code, an
// e-mail, a name) adds it to TestTheLogsAndTheSpansCarryNothingSensitive:
//
//  1. Add to the scenario the step that carries the value: a command of the
//     binary, or a request to the API. Prefer a step that fails, because a
//     failure is what gets logged.
//  2. Register the value with guarded.add. A value the system rewrites is
//     registered in each form: as typed and as stored. A value a command
//     makes up, such as a setup code, is read from its output and registered
//     after the command ran.
//  3. Append the error output of a new command to the logs the test
//     searches. The logs and the spans of the API are already there.
//
// Write the values in plain letters, digits and hyphens. The search ignores
// case and also looks for the form a value takes in a query string.

// sensitiveValues are the values the logs and the spans must never carry.
type sensitiveValues struct {
	values []sensitiveValue
}

type sensitiveValue struct {
	// what names the value in the failure of the test, which never shows
	// the value itself.
	what  string
	value string
}

// add registers a value and returns it, for the scenario to use.
func (s *sensitiveValues) add(what, value string) string {
	s.values = append(s.values, sensitiveValue{what: what, value: value})
	return value
}

// foundIn names each registered value the text carries.
func (s *sensitiveValues) foundIn(text string) []string {
	found := []string{}
	lowered := strings.ToLower(text)
	for _, guarded := range s.values {
		found = append(found, guarded.namedWhenIn(lowered)...)
	}
	return found
}

func (v sensitiveValue) namedWhenIn(lowered string) []string {
	value := strings.ToLower(v.value)
	if !strings.Contains(lowered, value) && !strings.Contains(lowered, strings.ToLower(url.QueryEscape(v.value))) {
		return nil
	}
	return []string{v.what}
}

func TestTheLogsAndTheSpansCarryNothingSensitive(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	collector := telemetrytest.NewCollector(t)
	// Nothing answers on port 1 of the machine, so every step that reaches
	// PostgreSQL fails, and its raw error is logged.
	environment := map[string]string{
		"API_LISTEN_ADDRESS":          "127.0.0.1:0",
		"LOG_LEVEL":                   "debug",
		"OTEL_EXPORTER_OTLP_ENDPOINT": collector.URL,
		"DATABASE_URL": "postgres://api:" + guarded.add("the password of the API role", "guarded-password-of-the-api-role") +
			"@127.0.0.1:1/software_factory?sslmode=disable",
		"MIGRATION_DATABASE_URL": "postgres://migration:" + guarded.add("the password of the migration role", "guarded-password-of-the-migration-role") +
			"@127.0.0.1:1/software_factory?sslmode=disable",
	}

	// The migrations fail to apply.
	var migrateOutput, migrateLogs output
	require.Equal(t, 1, run(context.Background(), []string{"migrate"}, environment, &migrateOutput, &migrateLogs))

	// A request fails, with a sensitive value in each place a request has one.
	api := launchAPI(t, environment)
	failed := newCaller().send(http.MethodGet,
		api.url+"/readyz?code="+guarded.add("a value in a query string", "guarded-value-in-a-query-string"),
		[]string{
			"Cookie", "session=" + guarded.add("a session cookie", "guarded-session-cookie"),
			"Authorization", "Bearer " + guarded.add("an access token", "guarded-access-token"),
		},
		guarded.add("a value in a body", "guarded-value-in-a-body"),
	)
	require.Equal(t, http.StatusServiceUnavailable, failed.status)
	api.signal()
	require.Equal(t, 0, <-api.status)

	logs := migrateLogs.String() + api.logs.String()
	require.Contains(t, logs, "api: apply the migrations", "the scenario logged nothing about the migrations")
	require.Contains(t, logs, "request failed", "the scenario logged no failed request")
	require.Contains(t, collector.Text(), "GET /readyz", "the scenario exported no span")
	assert.Empty(t, guarded.foundIn(logs), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
}

func TestTheTestOfSensitiveValuesFailsWhenAGuardedValueIsWrittenToTheLogs(t *testing.T) {
	t.Parallel()
	cases := map[string]func(logger *slog.Logger){
		"in an attribute": func(logger *slog.Logger) { logger.Info("code checked", "code", "K7M2X-9QW4T-B8N6R-C3V5Z") },
		"in a message":    func(logger *slog.Logger) { logger.Info("checked K7M2X-9QW4T-B8N6R-C3V5Z") },
		"in a raw error": func(logger *slog.Logger) {
			logger.Error("request failed", "error", errors.New("no code K7M2X-9QW4T-B8N6R-C3V5Z"))
		},
		"in another case": func(logger *slog.Logger) { logger.Info("code checked", "code", "k7m2x-9qw4t-b8n6r-c3v5z") },
		"inside a query string": func(logger *slog.Logger) {
			logger.Info("request", "url", "/api/setup?code=K7M2X-9QW4T-B8N6R-C3V5Z&step=1")
		},
	}
	for name, write := range cases {
		t.Run(name, func(t *testing.T) {
			guarded := &sensitiveValues{}
			guarded.add("the setup code", "K7M2X-9QW4T-B8N6R-C3V5Z")
			guarded.add("the e-mail of the administrator", "ana.souza@example.com")
			var logs output

			write(telemetry.NewLogger(&logs, slog.LevelDebug))

			assert.Equal(t, []string{"the setup code"}, guarded.foundIn(logs.String()))
		})
	}
}

func TestTheTestOfSensitiveValuesFindsAValueWrittenAsAQueryStringWritesIt(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	guarded.add("the e-mail of the administrator", "ana.souza@example.com")

	found := guarded.foundIn(`{"msg":"request","query":"email=ana.souza%40example.com"}`)

	assert.Equal(t, []string{"the e-mail of the administrator"}, found)
}

// The scenario of the test above runs with PostgreSQL out of reach, and a
// setup code only exists once the database stored its hash. This one issues a
// code on a migrated database, with the logs at their lowest level and the
// spans collected, and then searches both for the code the command printed.
func TestTheLogsAndTheSpansOfIssuingASetupCodeNeverCarryTheCode(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	collector := telemetrytest.NewCollector(t)
	environment := issueEnvironment(databasetest.Migrated(t))
	environment["LOG_LEVEL"] = "debug"
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL

	var printed, logs output
	require.Equal(t, 0, run(context.Background(), []string{"issue-setup-code"}, environment, &printed, &logs), logs.String())

	// The command makes the code up, so it is read from what the command
	// printed, and registered as printed and as it is hashed.
	code := printedSetupCode.FindString(printed.String())
	require.NotEmpty(t, code, "the command printed no setup code")
	guarded.add("the setup code", code)
	guarded.add("the setup code without its hyphens", strings.ReplaceAll(code, "-", ""))
	require.Contains(t, collector.Text(), "INSERT INTO setup_codes", "the scenario exported no span of the query that stores the code")
	require.Contains(t, collector.Text(), "INSERT INTO audit_log", "the scenario exported no span of the query that records the operation")
	assert.Empty(t, guarded.foundIn(logs.String()), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
}

// A setup code reaches the API in the body of a request. This scenario types
// one in each form a person may, a wrong one, and one in a request the API
// refuses, on a migrated database, and ends with a request that fails with a
// 5xx, whose raw error is logged. It then searches the logs and the spans for
// every form that was sent.
func TestTheLogsAndTheSpansOfOpeningTheSetupNeverCarryTheCode(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	collector := telemetrytest.NewCollector(t)
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	environment := environmentOf(temporary)
	environment["LOG_LEVEL"] = "debug"
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL
	api := launchAPI(t, environment)

	// Each form is a different text in a log, so each one is registered. The
	// search ignores case.
	asPrinted := guarded.add("the setup code", code)
	withoutHyphens := guarded.add("the setup code without its hyphens", strings.ReplaceAll(code, "-", ""))
	withSpaces := guarded.add("the setup code typed with spaces", strings.ReplaceAll(code, "-", " "))
	withLookAlikes := guarded.add("the setup code typed with O and L for 0 and 1",
		strings.NewReplacer("0", "O", "1", "L", "-", " ").Replace(strings.ToLower(code)))
	wrong := guarded.add("a wrong setup code", "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ")
	for _, typed := range []string{asPrinted, withoutHyphens, withSpaces, withLookAlikes} {
		require.Equal(t, http.StatusOK, openSetup(t, api.url, typed).status)
	}
	require.Equal(t, http.StatusUnprocessableEntity, openSetup(t, api.url, wrong).status)
	// Refused before the use case runs: an unknown property, a code longer
	// than the contract allows, an unknown query parameter and a wrong method.
	refused := sendJSON(t, http.MethodPost, api.url+"/api/setup/open", `{"code":"`+code+`","step":"`+code+`"}`)
	require.Equal(t, http.StatusUnprocessableEntity, refused.status)
	refused = openSetup(t, api.url, strings.Repeat(code, 4))
	require.Equal(t, http.StatusUnprocessableEntity, refused.status)
	refused = sendJSON(t, http.MethodPut, api.url+"/api/setup/open", `{"code":"`+code+`"}`)
	require.Equal(t, http.StatusMethodNotAllowed, refused.status)
	// The API role loses the right to read the setup codes, so the query of
	// the next request fails and its raw error is logged.
	_, err := asOwner(t, temporary).Exec(context.Background(), "REVOKE SELECT ON setup_codes FROM api")
	require.NoError(t, err)
	failed := openSetup(t, api.url, code)
	require.Equal(t, http.StatusInternalServerError, failed.status, failed.body)
	api.signal()
	require.Equal(t, 0, <-api.status)

	logs := api.logs.String()
	require.Contains(t, logs, "request failed", "the scenario logged no failed request")
	require.Contains(t, logs, "permission denied for table setup_codes", "the scenario logged no raw error")
	require.Contains(t, collector.Text(), "POST /api/setup/open", "the scenario exported no span of the operation")
	require.Contains(t, collector.Text(), "FROM setup_codes", "the scenario exported no span of the query that finds the code")
	assert.Empty(t, guarded.foundIn(logs), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
	for _, answer := range []received{refused, failed} {
		assert.Empty(t, guarded.foundIn(answer.body), "an answer repeats the setup code")
	}
}
