package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// The paths of the session of a browser, past asking for a login code.
const (
	sessionPath   = "/api/session"
	checkCodePath = "/api/session/login_code/check"
	logOutPath    = "/api/session/log_out"
)

// sendFrom sends a request from a browser that keeps its cookies, with a JSON
// body when there is one and these headers, and reads the answer with its
// cookies.
func sendFrom(t *testing.T, client *http.Client, method, address, body string, headers map[string]string) asked {
	t.Helper()
	request, err := http.NewRequest(method, address, strings.NewReader(body))
	require.NoError(t, err)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	answer := asked{
		received: received{status: response.StatusCode, mediaType: response.Header.Get("Content-Type"), body: string(raw)},
		header:   response.Header,
		cookies:  response.Cookies(),
	}
	_ = json.Unmarshal(raw, &answer.fields)
	return answer
}

// checkCode types the code in the browser.
func checkCode(t *testing.T, client *http.Client, api, code string) asked {
	t.Helper()
	body, err := json.Marshal(map[string]string{"code": code})
	require.NoError(t, err)
	return sendFrom(t, client, http.MethodPost, api+checkCodePath, string(body), nil)
}

// whoIsLoggedIn asks the API who is logged in to the session of the browser.
func whoIsLoggedIn(t *testing.T, client *http.Client, api string) asked {
	t.Helper()
	return sendFrom(t, client, http.MethodGet, api+sessionPath, "", nil)
}

// loggedIn is a browser where Ana, the administrator the setup registered,
// logged in with a login code, and the answer to her code.
func loggedIn(t *testing.T, temporary databasetest.Database, api string) (*http.Client, asked) {
	t.Helper()
	withAUser(t, temporary, api, "Ana.Souza@Fabrica.example")
	browser := aBrowser(t)
	require.Equal(t, http.StatusAccepted, askForLoginCode(t, browser, api, "ana.souza@fabrica.example").status)
	checked := checkCode(t, browser, api, codeOfTheRequest(t, temporary))
	require.Equal(t, http.StatusNoContent, checked.status, checked.body)
	return browser, checked
}

// withCookie is a browser that holds this cookie of a session, and nothing
// else.
func withCookie(t *testing.T, api string, cookie *http.Cookie) *http.Client {
	t.Helper()
	browser := aBrowser(t)
	address, err := url.Parse(api)
	require.NoError(t, err)
	browser.Jar.SetCookies(address, []*http.Cookie{{Name: cookie.Name, Value: cookie.Value}})
	return browser
}

func requireNotLoggedIn(t *testing.T, answer asked) {
	t.Helper()
	require.Equal(t, http.StatusUnauthorized, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Equal(t, "unauthorized", answer.fields["code"])
}

// The right code logs the user in: the answer sets the cookie again with a
// new token, the session tells who is logged in, and the audit log records the
// login with its way in.
func TestTheRightCodeLogsTheUserInWithANewCookie(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	withAUser(t, temporary, api, "Ana.Souza@Fabrica.example")
	browser := aBrowser(t)
	asking := askForLoginCode(t, browser, api, "ana.souza@fabrica.example")
	before := sessionCookie(t, asking)

	checked := checkCode(t, browser, api, codeOfTheRequest(t, temporary))

	require.Equal(t, http.StatusNoContent, checked.status, checked.body)
	assert.Empty(t, checked.body)
	cookie := sessionCookie(t, checked)
	assert.NotEqual(t, before.Value, cookie.Value, "the login kept the token the session had before it")
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.False(t, cookie.Secure, "the app of development is served over HTTP")
	assert.InDelta(t, (7 * 24 * time.Hour).Seconds(), float64(cookie.MaxAge), 60)

	who := whoIsLoggedIn(t, browser, api)
	require.Equal(t, http.StatusOK, who.status, who.body)
	assert.Equal(t, "application/json", who.mediaType)
	assert.JSONEq(t, `{
		"name": "Ana Souza",
		"email": "ana.souza@fabrica.example",
		"preferred_language": "pt-BR",
		"role_keys": ["administrator"]
	}`, who.body)
	requireNotLoggedIn(t, whoIsLoggedIn(t, withCookie(t, api, before), api))

	owner := asOwner(t, temporary)
	var entry string
	require.NoError(t, owner.QueryRow(context.Background(), `
		SELECT concat_ws(' ', operation, actor_kind, (actor_id = users.id)::text, subject_kind, (subject_id = users.id)::text, way_in)
		FROM audit_log JOIN users ON users.email = 'ana.souza@fabrica.example'
		WHERE operation = 'logged_in'`).Scan(&entry))
	assert.Equal(t, "logged_in user true user true login_code", entry)
	var ended bool
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT ended_at IS NOT NULL FROM login_code_requests").Scan(&ended))
	assert.True(t, ended, "the code works more than once")
}

func TestAWrongCodeAnswersTheAttemptsLeftAndTheThirdEndsTheCode(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	withAUser(t, temporary, api, "ana.souza@fabrica.example")
	browser := aBrowser(t)
	require.Equal(t, http.StatusAccepted, askForLoginCode(t, browser, api, "ana.souza@fabrica.example").status)
	right := codeOfTheRequest(t, temporary)
	wrong := "000000"
	if right == wrong {
		wrong = "111111"
	}

	answers := []asked{checkCode(t, browser, api, wrong), checkCode(t, browser, api, wrong), checkCode(t, browser, api, wrong)}
	afterwards := checkCode(t, browser, api, right)

	for position, attemptsLeft := range []float64{2, 1} {
		answer := answers[position]
		require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
		assert.Equal(t, "application/problem+json", answer.mediaType)
		assert.Equal(t, "login_code_invalid", answer.fields["code"])
		assert.Equal(t, map[string]any{"attempts_left": attemptsLeft}, answer.fields["params"])
	}
	for _, answer := range []asked{answers[2], afterwards} {
		require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
		assert.Equal(t, "login_code_ended", answer.fields["code"])
	}
	requireNotLoggedIn(t, whoIsLoggedIn(t, browser, api))
	assert.Equal(t, 2, countOf(t, asOwner(t, temporary), "audit_log"), "a login was recorded: the code issued and the setup are the only entries")
}

// A code that is not six digits does not follow the contract, and is no
// attempt. The answer never repeats what was typed.
func TestACodeThatIsNotSixDigitsIsRefusedOnItsField(t *testing.T) {
	t.Parallel()
	api := startAPI(t, environmentOf(databasetest.Migrated(t)))

	for typed, code := range map[string]string{"": "required", "12345": "invalid", "1234567": "invalid", "12345a": "invalid"} {
		answer := checkCode(t, aBrowser(t), api, typed)

		require.Equal(t, http.StatusUnprocessableEntity, answer.status, "%q: %s", typed, answer.body)
		assert.Equal(t, "validation_failed", answer.fields["code"], typed)
		errors, _ := answer.fields["errors"].([]any)
		require.Len(t, errors, 1, typed)
		field, _ := errors[0].(map[string]any)
		assert.Equal(t, "body.code", field["location"], typed)
		assert.Equal(t, code, field["code"], typed)
		assertNotRepeated(t, answer.body, typed)
	}
}

// assertNotRepeated checks that an answer does not quote what was typed.
func assertNotRepeated(t *testing.T, body, typed string) {
	t.Helper()
	if typed == "" {
		return
	}
	assert.NotContains(t, body, typed, "the answer repeats what was typed")
}

func TestWhoIsLoggedInAnswers401WithNoSession(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser := aBrowser(t)
	require.Equal(t, http.StatusAccepted, askForLoginCode(t, browser, api, "bruno.lima@fabrica.example").status)

	never := whoIsLoggedIn(t, aBrowser(t), api)
	beforeTheLogin := whoIsLoggedIn(t, browser, api)
	madeUp := whoIsLoggedIn(t, withCookie(t, api, &http.Cookie{Name: "session", Value: "a-token-nobody-was-given"}), api)

	for _, answer := range []asked{never, beforeTheLogin, madeUp} {
		requireNotLoggedIn(t, answer)
	}
	assert.Equal(t, 1, countOf(t, asOwner(t, temporary), "sessions"), "the session before the login was ended")
}

// Deactivating a user takes effect at their next request: it answers 401,
// and their session is ended, its row removed and its cookie expired.
func TestADeactivatedUserIsRefusedAndTheirSessionEnds(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser, _ := loggedIn(t, temporary, api)
	owner := asOwner(t, temporary)
	_, err := owner.Exec(context.Background(), "UPDATE users SET deactivated = true")
	require.NoError(t, err)

	refused := whoIsLoggedIn(t, browser, api)

	requireNotLoggedIn(t, refused)
	expired := sessionCookie(t, refused)
	assert.Empty(t, expired.Value)
	assert.Negative(t, expired.MaxAge, "the cookie of the ended session was not expired")
	assert.Zero(t, countOf(t, owner, "sessions"), "the session of the deactivated user is still stored")
}

func TestLoggingOutEndsTheSession(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser, checked := loggedIn(t, temporary, api)
	cookie := sessionCookie(t, checked)

	out := sendFrom(t, browser, http.MethodPost, api+logOutPath, "", map[string]string{"Origin": appOrigin()})

	require.Equal(t, http.StatusNoContent, out.status, out.body)
	assert.Negative(t, sessionCookie(t, out).MaxAge, "the cookie was not expired")
	requireNotLoggedIn(t, whoIsLoggedIn(t, browser, api))
	requireNotLoggedIn(t, whoIsLoggedIn(t, withCookie(t, api, cookie), api))
	assert.Zero(t, countOf(t, asOwner(t, temporary), "sessions"))
	again := sendFrom(t, aBrowser(t), http.MethodPost, api+logOutPath, "", nil)
	assert.Equal(t, http.StatusNoContent, again.status, "logging out with no session: %s", again.body)
}

// A page of another site cannot act with the cookie of a logged-in user: a
// POST that carries it is refused with 403 when it comes from another origin.
func TestARequestThatChangesStateFromAnotherOriginWithTheCookieIsRefused(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, environmentOf(temporary))
	browser, _ := loggedIn(t, temporary, api)

	fromElsewhere := sendFrom(t, browser, http.MethodPost, api+logOutPath, "", map[string]string{"Origin": "https://elsewhere.example"})
	crossSite := sendFrom(t, browser, http.MethodPost, api+checkCodePath, `{"code": "000000"}`, map[string]string{"Sec-Fetch-Site": "cross-site"})

	for _, refused := range []asked{fromElsewhere, crossSite} {
		require.Equal(t, http.StatusForbidden, refused.status, refused.body)
		assert.Equal(t, "application/problem+json", refused.mediaType)
		assert.Equal(t, "forbidden", refused.fields["code"])
		assert.Regexp(t, `^[0-9a-f]{32}$`, refused.fields["trace_id"])
	}
	assert.Equal(t, http.StatusOK, whoIsLoggedIn(t, browser, api).status, "the refused logout ended the session")
	withoutCookie := sendFrom(t, aBrowser(t), http.MethodPost, api+loginCodePath, `{"email": "bruno.lima@fabrica.example"}`,
		map[string]string{"Origin": "https://elsewhere.example"})
	assert.Equal(t, http.StatusAccepted, withoutCookie.status, "a request with no cookie was checked: %s", withoutCookie.body)
	fromTheApp := sendFrom(t, browser, http.MethodPost, api+logOutPath, "", map[string]string{"Origin": appOrigin(), "Sec-Fetch-Site": "same-origin"})
	assert.Equal(t, http.StatusNoContent, fromTheApp.status, fromTheApp.body)
}
