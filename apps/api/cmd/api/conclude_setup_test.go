package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// setupBody is the body of concluding the setup, every field filled in, as a
// person types it.
func setupBody(code string) map[string]any {
	return map[string]any{
		"code": code,
		"consultancy": map[string]any{
			"name":       "Fábrica de Software",
			"legal_name": "Fábrica de Software Ltda.",
			"country":    "BR",
			"tax_id":     "12.345.678/0001-90",
			"time_zone":  "America/Sao_Paulo",
			"currency":   "BRL",
			"email":      "contato@fabrica.example",
			"telephone":  "+55 11 3333 4444",
			"website":    "https://fabrica.example",
			"address": map[string]any{
				"line1":       "Avenida Paulista, 1000",
				"line2":       "Sala 101",
				"city":        "São Paulo",
				"region":      "SP",
				"postal_code": "01310-100",
			},
		},
		"administrator": map[string]any{
			"name":               "Ana Souza",
			"email":              "Ana.Souza@Fabrica.example",
			"country":            "BR",
			"national_id":        "123.456.789-09",
			"telephone":          "+55 11 98888 7777",
			"preferred_language": "pt-BR",
		},
	}
}

// object returns the object of the body a dotted path names: `consultancy`,
// `consultancy.address`.
func object(body map[string]any, path string) map[string]any {
	for _, name := range strings.Split(path, ".") {
		body = body[name].(map[string]any)
	}
	return body
}

// concludeSetup posts the body to the API.
func concludeSetup(t *testing.T, api string, body map[string]any) received {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	return sendJSON(t, http.MethodPost, api+"/api/setup/conclude", string(encoded))
}

// setupState asks the API for the state of the setup.
func setupState(t *testing.T, api string) string {
	t.Helper()
	answer := sendJSON(t, http.MethodGet, api+"/api/setup", "")
	require.Equal(t, http.StatusOK, answer.status, answer.body)
	assert.Equal(t, "application/json", answer.mediaType)
	return answer.body
}

// countOf counts the rows of a table, as the owner of the tables.
func countOf(t *testing.T, owner *pgx.Conn, table string) int {
	t.Helper()
	var count int
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT count(*) FROM "+pgx.Identifier{table}.Sanitize()).Scan(&count))
	return count
}

func TestConcludingTheSetupAnswers204AndSavesEverythingTogether(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	owner := asOwner(t, temporary)

	answer := concludeSetup(t, api, setupBody(code))

	require.Equal(t, http.StatusNoContent, answer.status, answer.body)
	assert.Empty(t, answer.body)
	var consultancy, administrator string
	require.NoError(t, owner.QueryRow(context.Background(),
		"SELECT row_to_json(consultancy)::jsonb - 'id' FROM consultancy").Scan(&consultancy))
	assert.JSONEq(t, `{
		"name": "Fábrica de Software", "legal_name": "Fábrica de Software Ltda.", "country": "BR",
		"tax_id": "12.345.678/0001-90", "time_zone": "America/Sao_Paulo", "currency": "BRL",
		"email": "contato@fabrica.example", "telephone": "+551133334444", "website": "https://fabrica.example",
		"address_line1": "Avenida Paulista, 1000", "address_line2": "Sala 101", "address_city": "São Paulo",
		"address_region": "SP", "address_postal_code": "01310-100"
	}`, consultancy)
	require.NoError(t, owner.QueryRow(context.Background(),
		"SELECT row_to_json(users)::jsonb - 'id' FROM users").Scan(&administrator))
	assert.JSONEq(t, `{
		"name": "Ana Souza", "email": "ana.souza@fabrica.example", "country": "BR",
		"national_id": "123.456.789-09", "telephone": "+5511988887777", "preferred_language": "pt-BR",
		"deactivated": false
	}`, administrator, "the e-mail typed in upper case is stored in lowercase")
}

func TestTheAdministratorHoldsTheAdministratorRoleAndNoUserHoldsTheDeveloperRole(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))

	require.Equal(t, http.StatusNoContent, concludeSetup(t, api, setupBody(code)).status)

	rows, err := asOwner(t, temporary).Query(context.Background(), `
		SELECT roles.key, coalesce(string_agg(users.email, ','), '')
		FROM roles
		LEFT JOIN user_roles ON user_roles.role_id = roles.id
		LEFT JOIN users ON users.id = user_roles.user_id
		GROUP BY roles.key
		ORDER BY roles.key`)
	require.NoError(t, err)
	type holders struct {
		Role, Users string
	}
	held, err := pgx.CollectRows(rows, pgx.RowToStructByPos[holders])
	require.NoError(t, err)
	assert.Equal(t, []holders{
		{Role: "administrator", Users: "ana.souza@fabrica.example"},
		{Role: "developer", Users: ""},
	}, held)
}

func TestConcludingTheSetupEndsTheSetupCodeAndLeavesOneEntryInTheAuditLog(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))

	require.Equal(t, http.StatusNoContent, concludeSetup(t, api, setupBody(code)).status)

	rows, err := asOwner(t, temporary).Query(context.Background(), `
		SELECT audit_log.operation, audit_log.actor_kind, audit_log.subject_kind,
		       audit_log.actor_id = setup_codes.id, audit_log.subject_id = users.id,
		       audit_log.occurred_at = setup_codes.ended_at
		FROM audit_log, setup_codes, users
		WHERE audit_log.operation = 'setup_concluded'`)
	require.NoError(t, err)
	type entry struct {
		Operation, ActorKind, SubjectKind                       string
		ActedByTheSetupCode, OnTheAdministrator, WhenItWasEnded bool
	}
	entries, err := pgx.CollectRows(rows, pgx.RowToStructByPos[entry])
	require.NoError(t, err)
	assert.Equal(t, []entry{{
		Operation: "setup_concluded", ActorKind: "setup_code", SubjectKind: "user",
		ActedByTheSetupCode: true, OnTheAdministrator: true, WhenItWasEnded: true,
	}}, entries)
	assert.Equal(t, 2, countOf(t, asOwner(t, temporary), "audit_log"), "one entry for the code issued, one for the setup")
}

func TestTheStateOfTheSetupIsNotConcludedBeforeAndConcludedAfter(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))

	newInstall := setupState(t, api)
	_, code := issueSetupCodeIn(t, temporary)
	codeWaiting := setupState(t, api)
	require.Equal(t, http.StatusNoContent, concludeSetup(t, api, setupBody(code)).status)
	concluded := setupState(t, api)
	issueSetupCodeIn(t, temporary)
	wayBackIn := setupState(t, api)

	assert.JSONEq(t, `{"concluded":false,"setup_code_waiting":false}`, newInstall)
	assert.JSONEq(t, `{"concluded":false,"setup_code_waiting":true}`, codeWaiting)
	assert.JSONEq(t, `{"concluded":true,"setup_code_waiting":false}`, concluded)
	assert.JSONEq(t, `{"concluded":true,"setup_code_waiting":true}`, wayBackIn)
}

func TestASetupCodeThatExpiredIsNotWaiting(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	issueSetupCodeIn(t, temporary)
	_, err := asOwner(t, temporary).Exec(context.Background(),
		"UPDATE setup_codes SET issued_at = issued_at - interval '1 hour', expires_at = expires_at - interval '1 hour'")
	require.NoError(t, err)
	api := startAPI(t, environmentOf(temporary))

	assert.JSONEq(t, `{"concluded":false,"setup_code_waiting":false}`, setupState(t, api))
}

func TestConcludingTwiceWithTheSameCodeSavesOnceAndTheSecondIsToldTheCodeIsInvalid(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	require.Equal(t, http.StatusNoContent, concludeSetup(t, api, setupBody(code)).status)
	again := setupBody(code)
	object(again, "administrator")["email"] = "bruno.lima@fabrica.example"

	answer := concludeSetup(t, api, again)

	requireSetupCodeInvalid(t, answer)
	owner := asOwner(t, temporary)
	assert.Equal(t, 1, countOf(t, owner, "consultancy"))
	assert.Equal(t, 1, countOf(t, owner, "users"))
	assert.Equal(t, 2, countOf(t, owner, "roles"))
}

func TestConcludingTheFirstSetupOnceItIsConcludedAnswers409(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, first := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	require.Equal(t, http.StatusNoContent, concludeSetup(t, api, setupBody(first)).status)
	_, second := issueSetupCodeIn(t, temporary)
	owner := asOwner(t, temporary)
	before := everyRow(t, owner)

	answer := concludeSetup(t, api, setupBody(second))

	require.Equal(t, http.StatusConflict, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Equal(t, "setup_already_concluded", answer.fields["code"])
	assert.Equal(t, before, everyRow(t, owner), "a refused setup changed what is stored")
	assert.Equal(t, "new_administrator", openSetup(t, api, second).fields["opens"],
		"the code issued after the setup does not open the registration of a new administrator")
}

// The last step of concluding, the entry of the audit log, fails: what the
// steps before it wrote is rolled back with it.
func TestAFailureAtTheLastStepOfConcludingLeavesNothingSavedAndTheSetupCodeWorking(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	owner := asOwner(t, temporary)
	_, err := owner.Exec(context.Background(), "REVOKE INSERT ON audit_log FROM api")
	require.NoError(t, err)
	api := startAPI(t, environmentOf(temporary))

	answer := concludeSetup(t, api, setupBody(code))

	require.Equal(t, http.StatusInternalServerError, answer.status, answer.body)
	for _, table := range []string{"consultancy", "users", "roles", "user_roles"} {
		assert.Equal(t, 0, countOf(t, owner, table), "the failed setup left rows in %s", table)
	}
	assert.Equal(t, http.StatusOK, openSetup(t, api, code).status, "the setup code stopped working")
}

// postedAnswer posts a JSON body and says what came back: the status, and the
// code of a failure. It fails no test, so a goroutine may call it.
func postedAnswer(url string, body []byte) string {
	response, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err.Error()
	}
	defer response.Body.Close()
	var failure struct {
		Code string `json:"code"`
	}
	// A 204 has no body to read.
	_ = json.NewDecoder(response.Body).Decode(&failure)
	return strings.TrimSpace(strconv.Itoa(response.StatusCode) + " " + failure.Code)
}

// Two people who opened the setup with the same code conclude at the same
// moment. The setup code is locked by the first, so the second waits, then
// finds it ended.
func TestConcludingRequestsAtTheSameMomentSaveOneConsultancy(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	const requests = 6
	answers := make([]string, requests)
	start := make(chan struct{})
	var concluding sync.WaitGroup

	for each := range requests {
		body := setupBody(code)
		object(body, "administrator")["email"] = strings.Repeat("a", each+1) + "@fabrica.example"
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		concluding.Go(func() {
			<-start
			answers[each] = postedAnswer(api+"/api/setup/conclude", encoded)
		})
	}
	close(start)
	concluding.Wait()

	outcomes := map[string]int{}
	for _, answer := range answers {
		outcomes[answer]++
	}
	assert.Equal(t, map[string]int{"204": 1, "422 setup_code_invalid": requests - 1}, outcomes)
	owner := asOwner(t, temporary)
	assert.Equal(t, 1, countOf(t, owner, "consultancy"))
	assert.Equal(t, 1, countOf(t, owner, "users"))
	assert.Equal(t, 2, countOf(t, owner, "roles"))
	assert.Equal(t, 1, countOf(t, owner, "user_roles"))
}
