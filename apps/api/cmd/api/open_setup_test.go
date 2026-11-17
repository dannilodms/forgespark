package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// received is the answer to a request: its status, its media type and its
// body, as text and as the JSON object it is.
type received struct {
	status    int
	mediaType string
	allow     string
	body      string
	fields    map[string]any
}

// sendJSON sends a request with a JSON body and reads the JSON it answers,
// when it answers a body.
func sendJSON(t *testing.T, method, url, body string) received {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	answer := received{
		status:    response.StatusCode,
		mediaType: response.Header.Get("Content-Type"),
		allow:     response.Header.Get("Allow"),
		body:      string(raw),
	}
	if response.StatusCode == http.StatusNoContent {
		return answer
	}
	require.NoError(t, json.Unmarshal(raw, &answer.fields), "the answer is not a JSON object: %s", raw)
	return answer
}

// openSetup asks the API what a setup code opens.
func openSetup(t *testing.T, api, code string) received {
	t.Helper()
	body, err := json.Marshal(map[string]string{"code": code})
	require.NoError(t, err)
	return sendJSON(t, http.MethodPost, api+"/api/setup/open", string(body))
}

// requireSetupCodeInvalid checks the one answer of a setup code that opens
// nothing, whatever the reason.
func requireSetupCodeInvalid(t *testing.T, answer received) {
	t.Helper()
	require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	traceID, _ := answer.fields["trace_id"].(string)
	assert.Regexp(t, `^[0-9a-f]{32}$`, traceID)
	assert.JSONEq(t, `{
		"title": "Unprocessable Entity",
		"status": 422,
		"detail": "the setup code is wrong, expired or was ended",
		"code": "setup_code_invalid",
		"params": {},
		"trace_id": "`+traceID+`"
	}`, answer.body)
}

func TestAValidSetupCodeAnswersWhatItOpensAndWhenItExpires(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))

	answer := openSetup(t, api, code)

	require.Equal(t, http.StatusOK, answer.status, answer.body)
	assert.Equal(t, "application/json", answer.mediaType)
	var expiresAt time.Time
	err := asOwner(t, temporary).QueryRow(context.Background(), "SELECT expires_at FROM setup_codes").Scan(&expiresAt)
	require.NoError(t, err)
	assert.JSONEq(t, `{"opens":"first_setup","expires_at":"`+expiresAt.UTC().Format(time.RFC3339Nano)+`"}`, answer.body)
	assert.WithinDuration(t, time.Now().Add(time.Hour), expiresAt, time.Minute)
}

func TestASetupCodeThatIsWrongExpiredOrEndedGetsOneAnswerAndChangesNothing(t *testing.T) {
	t.Parallel()
	cases := map[string]func(t *testing.T, temporary databasetest.Database, issued string) (typed string){
		"wrong": func(*testing.T, databasetest.Database, string) string {
			return "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ"
		},
		"not a setup code": func(*testing.T, databasetest.Database, string) string {
			return "open sesame"
		},
		"expired": func(t *testing.T, temporary databasetest.Database, issued string) string {
			_, err := asOwner(t, temporary).Exec(context.Background(),
				"UPDATE setup_codes SET issued_at = issued_at - interval '1 hour', expires_at = expires_at - interval '1 hour'")
			require.NoError(t, err)
			return issued
		},
		"ended": func(t *testing.T, temporary databasetest.Database, issued string) string {
			issueSetupCodeIn(t, temporary)
			return issued
		},
	}
	for name, arrange := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			temporary := databasetest.Migrated(t)
			_, issued := issueSetupCodeIn(t, temporary)
			typed := arrange(t, temporary, issued)
			owner := asOwner(t, temporary)
			before := everyRow(t, owner)
			api := startAPI(t, environmentOf(temporary))

			answer := openSetup(t, api, typed)

			requireSetupCodeInvalid(t, answer)
			assert.Equal(t, before, everyRow(t, owner), "checking the code changed what is stored")
		})
	}
}

func TestAWrongSetupCodeLeavesTheSetupCodeWorking(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))

	for range 3 {
		requireSetupCodeInvalid(t, openSetup(t, api, "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ"))
	}
	answer := openSetup(t, api, code)

	assert.Equal(t, http.StatusOK, answer.status, answer.body)
}

func TestASetupCodeIsAcceptedHoweverItIsTyped(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	// The command draws the code, so the letters O, I and L stand for a 0 or
	// a 1 only when the code has one. The use case tests cover that with a
	// code that is known.
	lookAlikes := strings.NewReplacer("0", "O", "1", "L")
	typed := map[string]string{
		"in lower case":               strings.ToLower(code),
		"without hyphens":             strings.ReplaceAll(code, "-", ""),
		"with spaces":                 " " + strings.ReplaceAll(code, "-", "  ") + " ",
		"with O and L for 0 and 1":    lookAlikes.Replace(code),
		"with o and i for 0 and 1":    strings.ReplaceAll(strings.ToLower(lookAlikes.Replace(code)), "l", "i"),
		"in lower case, run together": strings.ToLower(strings.ReplaceAll(code, "-", "")),
	}
	for name, form := range typed {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			answer := openSetup(t, api, form)

			require.Equal(t, http.StatusOK, answer.status, answer.body)
			assert.Equal(t, "first_setup", answer.fields["opens"])
		})
	}
}

func TestOpeningTheSetupRejectsWhatIsOutsideItsContract(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	cases := map[string]struct {
		path, body string
		status     int
		code       string
		location   string
	}{
		"an unknown property of the body": {
			path: "/api/setup/open", body: `{"code":"` + code + `","step":"consultancy"}`,
			status: http.StatusUnprocessableEntity, code: "validation_failed", location: "body.step",
		},
		"an unknown query parameter": {
			path: "/api/setup/open?step=consultancy", body: `{"code":"` + code + `"}`,
			status: http.StatusUnprocessableEntity, code: "validation_failed", location: "query.step",
		},
		"no code": {
			path: "/api/setup/open", body: `{}`,
			status: http.StatusUnprocessableEntity, code: "validation_failed", location: "body.code",
		},
		"an empty code": {
			path: "/api/setup/open", body: `{"code":""}`,
			status: http.StatusUnprocessableEntity, code: "validation_failed", location: "body.code",
		},
		"a code far too long": {
			path: "/api/setup/open", body: `{"code":"` + strings.Repeat(code, 4) + `"}`,
			status: http.StatusUnprocessableEntity, code: "validation_failed", location: "body.code",
		},
	}
	for name, refused := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			answer := sendJSON(t, http.MethodPost, api+refused.path, refused.body)

			require.Equal(t, refused.status, answer.status, answer.body)
			assert.Equal(t, "application/problem+json", answer.mediaType)
			assert.Equal(t, refused.code, answer.fields["code"])
			assert.Regexp(t, `^[0-9a-f]{32}$`, answer.fields["trace_id"])
			assert.Contains(t, answer.body, `"location":"`+refused.location+`"`)
			assert.NotContains(t, strings.ToLower(answer.body), strings.ToLower(code), "the answer repeats the setup code")
		})
	}
}

func TestABodyOverTheSizeLimitIsRejected(t *testing.T) {
	t.Parallel()
	api := startAPI(t, environmentOf(databasetest.Migrated(t)))
	// The limit is 1 MiB. The code alone is within its own limit, so only
	// the size of the body is wrong.
	body := `{"code":"ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ"` + strings.Repeat(" ", 1<<20) + `}`

	answer := sendJSON(t, http.MethodPost, api+"/api/setup/open", body)

	assert.Equal(t, http.StatusRequestEntityTooLarge, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Equal(t, "request_entity_too_large", answer.fields["code"])
	assert.Regexp(t, `^[0-9a-f]{32}$`, answer.fields["trace_id"])
}

func TestAnUnknownRouteAndAWrongMethodAnswerInTheErrorFormat(t *testing.T) {
	t.Parallel()
	api := startAPI(t, environmentOf(databasetest.Migrated(t)))

	unknown := sendJSON(t, http.MethodGet, api+"/api/setup/unknown", "")
	wrongMethod := sendJSON(t, http.MethodGet, api+"/api/setup/open", "")

	assert.Equal(t, http.StatusNotFound, unknown.status)
	assert.Equal(t, "application/problem+json", unknown.mediaType)
	assert.Equal(t, "not_found", unknown.fields["code"])
	assert.Regexp(t, `^[0-9a-f]{32}$`, unknown.fields["trace_id"])
	assert.Equal(t, http.StatusMethodNotAllowed, wrongMethod.status)
	assert.Equal(t, "application/problem+json", wrongMethod.mediaType)
	assert.Equal(t, "method_not_allowed", wrongMethod.fields["code"])
	assert.Equal(t, "POST", wrongMethod.allow)
	assert.Regexp(t, `^[0-9a-f]{32}$`, wrongMethod.fields["trace_id"])
}
