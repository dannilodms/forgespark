package main

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	emailpostgres "github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/postgres"
	emailriver "github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/river"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/smtp/smtptest"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	emailapplicationtest "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/clock"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry/telemetrytest"
)

// The test of sensitive values (Telemetry in CODING_STANDARDS.md). It runs
// the commands of the binary through a scenario in which every sensitive
// value the system handles goes by, with the logs at their lowest level, and
// then searches the logs and the exported spans for each value.
//
// A ticket that brings a new kind of sensitive value (a setup code, an
// e-mail, a name) adds it to TestTheLogsAndTheSpansCarryNothingSensitive, or
// to a scenario of its own below when the value needs a migrated database or
// a job to go by:
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
	environment := withFileStorage(map[string]string{
		"API_LISTEN_ADDRESS":          "127.0.0.1:0",
		"LOG_LEVEL":                   "debug",
		"OTEL_EXPORTER_OTLP_ENDPOINT": collector.URL,
		"DATABASE_URL": "postgres://api:" + guarded.add("the password of the API role", "guarded-password-of-the-api-role") +
			"@127.0.0.1:1/software_factory?sslmode=disable",
		"MIGRATION_DATABASE_URL": "postgres://migration:" + guarded.add("the password of the migration role", "guarded-password-of-the-migration-role") +
			"@127.0.0.1:1/software_factory?sslmode=disable",
	})
	environment["FILE_STORAGE_SECRET_ACCESS_KEY"] = guarded.add("the secret key of the file storage", "guarded-secret-key-of-the-file-storage")
	withEmailSender(environment)

	// The migrations fail to apply.
	var migrateOutput, migrateLogs output
	require.Equal(t, 1, run(context.Background(), []string{"migrate"}, environment, &migrateOutput, &migrateLogs))

	// The worker fails to reach PostgreSQL when it starts.
	var workerOutput, workerLogs output
	require.Equal(t, 1, run(context.Background(), []string{"worker"}, environment, &workerOutput, &workerLogs))

	// Two requests carry a sensitive value in each place a request has one:
	// readiness, which fails and writes its line but opens no span, and an
	// operation under /api, which opens its span and refuses the unknown
	// query parameter.
	api := launchAPI(t, environment)
	query := "?code=" + guarded.add("a value in a query string", "guarded-value-in-a-query-string")
	headers := []string{
		"Cookie", "session=" + guarded.add("a session cookie", "guarded-session-cookie"),
		"Authorization", "Bearer " + guarded.add("an access token", "guarded-access-token"),
	}
	body := guarded.add("a value in a body", "guarded-value-in-a-body")
	failed := newCaller().send(http.MethodGet, api.url+"/readyz"+query, headers, body)
	require.Equal(t, http.StatusServiceUnavailable, failed.status)
	refused := newCaller().send(http.MethodGet, api.url+"/api/setup"+query, headers, body)
	require.Equal(t, http.StatusUnprocessableEntity, refused.status)
	// Concluding the setup carries the data of a person in its body, which
	// the use case checks before it fails to reach PostgreSQL.
	setup := setupBody("ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ")
	object(setup, "administrator")["name"] = guarded.add("the name of a person", "Guarded Name Of A Person")
	object(setup, "administrator")["email"] = guarded.add("the e-mail of a person", "guarded.email@example.com")
	object(setup, "administrator")["national_id"] = guarded.add("the national ID of a person", "GUARDED-NATIONAL-ID")
	unreachable := concludeSetup(t, api.url, setup)
	require.Equal(t, http.StatusInternalServerError, unreachable.status, unreachable.body)
	api.signal()
	require.Equal(t, 0, <-api.status)

	logs := migrateLogs.String() + workerLogs.String() + api.logs.String()
	require.Contains(t, logs, "api: apply the migrations", "the scenario logged nothing about the migrations")
	require.Contains(t, logs, "the worker stopped", "the scenario logged nothing about the worker")
	require.Contains(t, logs, "request failed", "the scenario logged no failed request")
	require.Contains(t, logs, "/api/setup/conclude", "the scenario logged no failed conclusion of the setup")
	require.Contains(t, collector.Text(), "GET /api/setup", "the scenario exported no span")
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

// An SMTP server repeats the address it refuses, in angle brackets and in the
// case it was sent in, and go-mail lists it after the answer.
func TestTheTestOfSensitiveValuesFindsAnAddressInTheAnswerOfAnEmailServer(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	guarded.add("the address of a recipient", "ana.souza@fabrica.example")

	for _, line := range []string{
		`{"msg":"Job errored; retrying","error":"550 5.1.1 <Ana.Souza@Fabrica.example>: Recipient address rejected"}`,
		`{"msg":"Job errored","error":"sending SMTP RCPT TO command, affected recipient(s): ana.souza@fabrica.example"}`,
	} {
		assert.Equal(t, []string{"the address of a recipient"}, guarded.foundIn(line), line)
	}
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
	tooLong := openSetup(t, api.url, strings.Repeat(code, 4))
	require.Equal(t, http.StatusUnprocessableEntity, tooLong.status)
	// A caller that sends the code where it must never go, in the address.
	inTheAddress := sendJSON(t, http.MethodPost, api.url+"/api/setup/open?code="+code, `{"code":"`+code+`"}`)
	require.Equal(t, http.StatusUnprocessableEntity, inTheAddress.status)
	require.Contains(t, inTheAddress.body, `"location":"query.code"`)
	wrongMethod := sendJSON(t, http.MethodPut, api.url+"/api/setup/open", `{"code":"`+code+`"}`)
	require.Equal(t, http.StatusMethodNotAllowed, wrongMethod.status)
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
	for _, answer := range []received{refused, tooLong, inTheAddress, wrongMethod, failed} {
		assert.Empty(t, guarded.foundIn(answer.body), "an answer repeats the setup code")
	}
}

// An image of the setup goes by the API as a URL signed with the key of the
// file storage: whoever holds the URL can send to it until it expires. This
// scenario starts uploads, sends an image, concludes the setup with it and
// fails a start with a 5xx, on a migrated database and the Garage of
// development. It then searches the logs and the spans for the secret key of
// the storage, the signature of each URL and the setup code.
func TestTheLogsAndTheSpansOfAnUploadNeverCarryTheSignatureOrTheKeyOfTheStorage(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	collector := telemetrytest.NewCollector(t)
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	guarded.add("the setup code", code)
	environment := environmentOf(temporary)
	environment["LOG_LEVEL"] = "debug"
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL
	guarded.add("the secret key of the file storage", environment["FILE_STORAGE_SECRET_ACCESS_KEY"])
	api := launchAPI(t, environment)

	signed := func(purpose string) (fileID, signedURL string) {
		t.Helper()
		started := startUpload(t, api.url, uploadBody(code, purpose, "image/png", 2048))
		require.Equal(t, http.StatusCreated, started.status, started.body)
		signedURL, _ = started.fields["upload_url"].(string)
		parsed, err := url.Parse(signedURL)
		require.NoError(t, err)
		guarded.add("the signature of a URL", parsed.Query().Get("X-Amz-Signature"))
		fileID, _ = started.fields["file_id"].(string)
		return fileID, signedURL
	}
	logo, logoURL := signed("consultancy_logo")
	require.Equal(t, http.StatusOK, sendToTheStorage(t, logoURL, "image/png", aPNG(2048)))
	neverSent, _ := signed("user_photo")
	// Refused: the photo was never sent, so the storage answers 404 to the
	// question about it.
	body := setupBody(code)
	object(body, "consultancy")["logo_file_id"] = logo
	object(body, "administrator")["photo_file_id"] = neverSent
	require.Equal(t, http.StatusUnprocessableEntity, concludeSetup(t, api.url, body).status)
	// Concluded with the logo, which the storage is asked about.
	delete(object(body, "administrator"), "photo_file_id")
	require.Equal(t, http.StatusNoContent, concludeSetup(t, api.url, body).status)
	// A 5xx: the API role may not write files.
	_, err := asOwner(t, temporary).Exec(context.Background(), "REVOKE INSERT ON files FROM api")
	require.NoError(t, err)
	_, newCode := issueSetupCodeIn(t, temporary)
	guarded.add("the second setup code", newCode)
	failed := startUpload(t, api.url, uploadBody(newCode, "user_photo", "image/png", 2048))
	require.Equal(t, http.StatusInternalServerError, failed.status, failed.body)
	api.signal()
	require.Equal(t, 0, <-api.status)

	logs := api.logs.String()
	require.Contains(t, logs, "permission denied for table files", "the scenario logged no raw error")
	require.Contains(t, collector.Text(), "POST /api/setup/uploads", "the scenario exported no span of the operation")
	require.Contains(t, collector.Text(), "HeadObject", "the scenario exported no span of a call to the storage")
	assert.Empty(t, guarded.foundIn(logs), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
}

// Concluding the setup carries the data of a person: a name, an e-mail, a
// national ID and a telephone. This scenario concludes it on a migrated
// database through each way it can go: refused by the contract, refused by a
// rule of the domain, failed with a 5xx whose raw error is logged (a
// permission taken away, and a row PostgreSQL refuses for a duplicate
// e-mail), concluded, and tried again with the code it ended. It then
// searches the logs, the spans, the answers and the audit log for each value.
func TestTheLogsTheSpansAndTheAuditLogOfConcludingTheSetupNeverCarryTheDataOfAPerson(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	collector := telemetrytest.NewCollector(t)
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	guarded.add("the setup code", code)
	environment := environmentOf(temporary)
	environment["LOG_LEVEL"] = "debug"
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL
	api := launchAPI(t, environment)
	owner := asOwner(t, temporary)

	// A value the system rewrites is registered as typed and as stored.
	setup := setupBody(code)
	administrator := object(setup, "administrator")
	administrator["name"] = guarded.add("the name of the administrator", "Guarded Administrator Souza")
	administrator["email"] = guarded.add("the e-mail as typed", "Guarded.Administrator@Fabrica.example")
	guarded.add("the e-mail as stored", "guarded.administrator@fabrica.example")
	administrator["national_id"] = guarded.add("the national ID", "GUARDED-123.456.789-09")
	administrator["telephone"] = guarded.add("the telephone as typed", "+55 11 97531 8642")
	guarded.add("the telephone as stored", "+5511975318642")
	object(setup, "consultancy")["tax_id"] = guarded.add("the tax ID", "GUARDED-12.345.678/0001-90")

	answers := []received{}
	send := func(body map[string]any, status int) {
		t.Helper()
		answer := concludeSetup(t, api.url, body)
		require.Equal(t, status, answer.status, answer.body)
		answers = append(answers, answer)
	}
	// Refused by the contract: an e-mail that is two, and a name too long.
	refusedByTheContract := setupBody(code)
	object(refusedByTheContract, "administrator")["email"] = "ana@fabrica.example, guarded.administrator@fabrica.example"
	object(refusedByTheContract, "administrator")["national_id"] = strings.Repeat("GUARDED-123.456.789-09", 3)
	send(refusedByTheContract, http.StatusUnprocessableEntity)
	// Refused by a rule of the domain, with every other value of the person.
	refusedByTheDomain := setupBody(code)
	maps.Copy(object(refusedByTheDomain, "administrator"), administrator)
	object(refusedByTheDomain, "administrator")["country"] = "ZZ"
	send(refusedByTheDomain, http.StatusUnprocessableEntity)
	// A 5xx: the API role may not write users.
	_, err := owner.Exec(context.Background(), "REVOKE INSERT ON users FROM api")
	require.NoError(t, err)
	send(setup, http.StatusInternalServerError)
	_, err = owner.Exec(context.Background(), "GRANT INSERT ON users TO api")
	require.NoError(t, err)
	// A 5xx: a user with the same e-mail, written past the use case.
	_, err = owner.Exec(context.Background(), `INSERT INTO users (id, name, email, country, national_id, preferred_language, deactivated)
		VALUES (-1, 'Someone Else', 'guarded.administrator@fabrica.example', 'BR', '1', 'en-US', false)`)
	require.NoError(t, err)
	send(setup, http.StatusInternalServerError)
	_, err = owner.Exec(context.Background(), "DELETE FROM users WHERE id = -1")
	require.NoError(t, err)
	// Concluded, and tried again with the code it ended.
	send(setup, http.StatusNoContent)
	send(setup, http.StatusUnprocessableEntity)
	api.signal()
	require.Equal(t, 0, <-api.status)

	logs := api.logs.String()
	require.Contains(t, logs, "permission denied for table users", "the scenario logged no raw error of a permission")
	require.Contains(t, logs, "users_email_key", "the scenario logged no raw error of a duplicate e-mail")
	require.Contains(t, collector.Text(), "POST /api/setup/conclude", "the scenario exported no span of the operation")
	require.Contains(t, collector.Text(), "INSERT INTO users", "the scenario exported no span of the query that saves the user")
	assert.Empty(t, guarded.foundIn(logs), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
	for _, answer := range answers {
		assert.Empty(t, guarded.foundIn(answer.body), "an answer repeats a sensitive value: %s", answer.body)
	}
	var auditLog string
	require.NoError(t, owner.QueryRow(context.Background(),
		"SELECT string_agg(row_to_json(audit_log)::text, E'\\n') FROM audit_log").Scan(&auditLog))
	require.Contains(t, auditLog, "setup_concluded", "the scenario recorded no conclusion")
	assert.Empty(t, guarded.foundIn(auditLog), "the audit log carries a sensitive value")
}

// Registering an administrator, the way back in, carries the data of a person
// and may carry the e-mail of a user that exists. This scenario concludes the
// setup, issues a second code and registers on a migrated database through
// each way the operation can go: refused by the contract, refused by a rule
// of the domain, refused for an e-mail in use, failed with a 5xx whose raw
// error is logged, registered, and tried again with the code it ended. It
// then searches the logs, the spans, the answers and the audit log for each
// value.
func TestTheLogsTheSpansAndTheAuditLogOfRegisteringAnAdministratorNeverCarryTheDataOfAPerson(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	collector := telemetrytest.NewCollector(t)
	temporary := databasetest.Migrated(t)
	environment := environmentOf(temporary)
	environment["LOG_LEVEL"] = "debug"
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = collector.URL
	api := launchAPI(t, environment)
	owner := asOwner(t, temporary)
	_, first := issueSetupCodeIn(t, temporary)
	guarded.add("the first setup code", first)
	require.Equal(t, http.StatusNoContent, concludeSetup(t, api.url, setupBody(first)).status)
	_, code := issueSetupCodeIn(t, temporary)
	guarded.add("the setup code of the way back in", code)

	// A value the system rewrites is registered as typed and as stored.
	body := administratorBody(code)
	administrator := object(body, "administrator")
	administrator["name"] = guarded.add("the name of the new administrator", "Guarded Bruno Lima")
	administrator["email"] = guarded.add("the e-mail as typed", "Guarded.Bruno@Fabrica.example")
	guarded.add("the e-mail as stored", "guarded.bruno@fabrica.example")
	administrator["national_id"] = guarded.add("the national ID", "GUARDED-987 654 321")
	administrator["telephone"] = guarded.add("the telephone as typed", "+351 21 864 2097")
	guarded.add("the telephone as stored", "+351218642097")

	answers := []received{}
	send := func(body map[string]any, status int) {
		t.Helper()
		answer := registerAdministrator(t, api.url, body)
		require.Equal(t, status, answer.status, answer.body)
		answers = append(answers, answer)
	}
	// Refused by the contract: an e-mail that is two, and a national ID too
	// long.
	refusedByTheContract := administratorBody(code)
	object(refusedByTheContract, "administrator")["email"] = "bruno@fabrica.example, guarded.bruno@fabrica.example"
	object(refusedByTheContract, "administrator")["national_id"] = strings.Repeat("GUARDED-987 654 321", 3)
	send(refusedByTheContract, http.StatusUnprocessableEntity)
	// Refused by a rule of the domain, with every other value of the person.
	refusedByTheDomain := administratorBody(code)
	maps.Copy(object(refusedByTheDomain, "administrator"), administrator)
	object(refusedByTheDomain, "administrator")["country"] = "ZZ"
	send(refusedByTheDomain, http.StatusUnprocessableEntity)
	// Refused: the e-mail of the user that exists, typed in another case.
	inUse := administratorBody(code)
	maps.Copy(object(inUse, "administrator"), administrator)
	object(inUse, "administrator")["email"] = guarded.add("the e-mail of a user that exists, as typed", "ANA.Souza@Fabrica.example")
	guarded.add("the e-mail of a user that exists, as stored", theFirstAdministrator)
	send(inUse, http.StatusConflict)
	// A 5xx: the API role may not write users.
	_, err := owner.Exec(context.Background(), "REVOKE INSERT ON users FROM api")
	require.NoError(t, err)
	send(body, http.StatusInternalServerError)
	_, err = owner.Exec(context.Background(), "GRANT INSERT ON users TO api")
	require.NoError(t, err)
	// Registered, and tried again with the code it ended.
	send(body, http.StatusCreated)
	send(body, http.StatusUnprocessableEntity)
	api.signal()
	require.Equal(t, 0, <-api.status)

	logs := api.logs.String()
	require.Contains(t, logs, "permission denied for table users", "the scenario logged no raw error of a permission")
	require.Contains(t, collector.Text(), "POST /api/setup/administrators", "the scenario exported no span of the operation")
	require.Contains(t, collector.Text(), "FROM users WHERE email", "the scenario exported no span of the query that looks for the e-mail")
	require.Contains(t, collector.Text(), "INSERT INTO users", "the scenario exported no span of the query that saves the user")
	assert.Empty(t, guarded.foundIn(logs), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
	for _, answer := range answers {
		assert.Empty(t, guarded.foundIn(answer.body), "an answer repeats a sensitive value: %s", answer.body)
	}
	var auditLog string
	require.NoError(t, owner.QueryRow(context.Background(),
		"SELECT string_agg(row_to_json(audit_log)::text, E'\n') FROM audit_log").Scan(&auditLog))
	require.Contains(t, auditLog, "administrator_registered", "the scenario recorded no registration")
	assert.Empty(t, guarded.foundIn(auditLog), "the audit log carries a sensitive value")
}

// An e-mail record carries the address of its recipient, and the job that
// delivers it the identifier of the record alone. This scenario queues an
// e-mail on a migrated database, inside the span of a request, and runs its
// job as the worker does, with the worker's wiring and logs at their lowest
// level, against an SMTP server that refuses the recipient and repeats the
// address in its answer, as servers do. It then searches the logs, the spans
// and the jobs River keeps, their arguments and their errors, for the address
// and for the code the e-mail carried.
//
// No area gives the worker the content of an e-mail yet (the login will), so
// the job runs with a source of login codes of the test, on the rest of the
// wiring of the worker, instead of through the command.
func TestTheLogsTheSpansAndTheJobsOfSendingAnEmailNeverCarryTheAddressOrTheContent(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	recipient := guarded.add("the address of the recipient", "Guarded.Recipient@Fabrica.example")
	code := guarded.add("the code the e-mail carries", "guarded-login-code-482915")
	collector := telemetrytest.NewCollector(t)
	logs := &output{}
	logger := telemetry.NewLogger(logs, slog.LevelDebug)
	tracing, err := telemetry.StartTracing("api", collector.URL, logger)
	require.NoError(t, err)
	temporary := databasetest.Migrated(t)
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            temporary.APIConnection(t),
		MaxSize:                  4,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: time.Minute,
		TracerProvider:           tracing.Provider,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Queued inside the span of a request, as the login will.
	saver, err := jobs.NewSaver(pool, tracing.Provider, logger)
	require.NoError(t, err)
	queue := emailapplication.NewQueueEmail(
		database.NewTransactions(pool), emailpostgres.NewEmailRepository(pool), emailriver.NewDeliveryJobs(saver), clock.System{})
	english, err := kernel.NewLanguage("en-US")
	require.NoError(t, err)
	requestCtx, request := tracing.Provider.Tracer("test").Start(context.Background(), "POST /api/login/codes")
	err = database.NewTransactions(pool).Run(requestCtx, func(ctx context.Context) error {
		return queue.Execute(ctx, emailapplication.QueueEmailInput{
			Recipient: recipient,
			Template:  emailapplication.LoginCode,
			Language:  english,
			Reference: emailapplication.LoginCodeRequest(1),
		})
	})
	require.NoError(t, err)
	request.End()

	// The job runs on the configuration of the worker, pointed at a server
	// that refuses the recipient. The worker also requires a file storage,
	// which this scenario never reaches.
	server := smtptest.NewRefusingServer(t)
	environment := newStandInStorage(map[string]time.Time{}, false).environment(t,
		withEmailSender(map[string]string{"DATABASE_URL": temporary.APIURL}))
	environment["EMAIL_SMTP_ADDRESS"] = server.Address()
	settings, err := config.Load(config.Worker, environment)
	require.NoError(t, err)
	loginCodes := emailapplicationtest.NewContentSource()
	loginCodes.Give(emailapplication.LoginCodeRequest(1), emailapplication.LoginCodeContent{Code: code, Lifetime: 10 * time.Minute})
	kinds := jobs.NewKinds()
	contents := emailapplication.ContentSources{emailapplication.LoginCode: loginCodes}
	jobs.Add(kinds, emailriver.NewDeliveryWorker(deliverEmail(pool, emailSender(settings, tracing.Provider), contents)))
	runner, err := jobs.NewRunner(pool, jobs.RunnerSettings{
		LightQueueMaxJobs: 1,
		HeavyQueueMaxJobs: 1,
		ShutdownDeadline:  2 * time.Second,
		Kinds:             kinds,
		TracerProvider:    tracing.Provider,
		Logger:            logger,
	})
	require.NoError(t, err)
	shutdown, signal := context.WithCancel(context.Background())
	stopped := make(chan error, 1)
	go func() { stopped <- runner.Run(shutdown) }()
	select {
	case <-server.Refused:
	case <-time.After(30 * time.Second):
		t.Fatalf("the job did not reach the SMTP server: %s", logs)
	}
	require.Eventually(t, func() bool {
		var failures int
		require.NoError(t, pool.QueryRow(context.Background(),
			"SELECT count(*) FROM river_job WHERE kind = 'deliver_email' AND errors IS NOT NULL").Scan(&failures))
		return failures > 0
	}, 30*time.Second, 50*time.Millisecond, "River kept no error of the job")
	signal()
	require.NoError(t, <-stopped)
	tracing.Stop()

	var jobsKept string
	require.NoError(t, pool.QueryRow(context.Background(),
		"SELECT string_agg(row_to_json(river_job)::text, E'\n') FROM river_job").Scan(&jobsKept))
	require.Contains(t, logs.String(), "Job errored", "the scenario logged no failed job")
	require.Contains(t, logs.String(), "RCPT TO", "the scenario logged no refusal of the SMTP server")
	require.Contains(t, jobsKept, "RCPT TO", "River kept no error of the job")
	require.Contains(t, collector.Text(), "river.work/deliver_email", "the scenario exported no span of the job")
	require.Contains(t, collector.Text(), "SMTP send", "the scenario exported no span of the sending")
	require.Contains(t, collector.Text(), "INSERT INTO emails", "the scenario exported no span of the query that records the e-mail")
	assert.Empty(t, guarded.foundIn(logs.String()), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
	assert.Empty(t, guarded.foundIn(jobsKept), "a job carries a sensitive value")
}

// The cleanup of the storage handles the key of every object. This scenario
// runs the worker on a migrated database with a storage that lists an object
// left behind and cuts the connection of its removal, whose URL holds the
// key, so the job fails and River logs the failure. It then searches the logs
// and the spans for the key.
func TestTheLogsAndTheSpansOfTheCleanupOfTheStorageNeverCarryTheKeyOfAnObject(t *testing.T) {
	t.Parallel()
	guarded := &sensitiveValues{}
	key := guarded.add("the key of an object", "0f1e2d3c4b5a69788796a5b4c3d2e1f0")
	collector := telemetrytest.NewCollector(t)
	storage := newStandInStorage(map[string]time.Time{key: time.Now().Add(-25 * time.Hour)}, true)
	environment := storage.environment(t, withEmailSender(map[string]string{
		"DATABASE_URL":                databasetest.Migrated(t).APIURL,
		"LOG_LEVEL":                   "debug",
		"OTEL_EXPORTER_OTLP_ENDPOINT": collector.URL,
	}))
	worker := launchWorker(t, environment)

	require.Equal(t, key, receiveWithin(t, storage.removals, 30*time.Second, "the removal of the object"))
	require.Eventually(t, func() bool { return strings.Contains(worker.logs.String(), "Job errored") },
		30*time.Second, 10*time.Millisecond, "the failed cleanup was not logged: %s", worker.logs)
	worker.signal()
	require.Equal(t, 0, <-worker.status, "%s", worker.logs)

	require.Contains(t, worker.logs.String(), "remove an object", "the logs do not say why the cleanup failed")
	require.Contains(t, collector.Text(), "river.work/clean_up_storage", "the scenario exported no span of the job")
	assert.Empty(t, guarded.foundIn(worker.logs.String()), "the logs carry a sensitive value")
	assert.Empty(t, guarded.foundIn(collector.Text()), "the spans carry a sensitive value")
}
