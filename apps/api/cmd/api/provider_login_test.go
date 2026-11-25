package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/access/adapters/oidc/oidctest"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
)

// These tests log in through the Dex of development, the local login
// provider, which `task services` starts. Dex sends the browser back to the
// callback at APP_ORIGIN, the one address its client accepts; the tests take
// the code and the state from there to the API under test.

// The path where the login state is read.
const loginStatePath = "/api/session/login"

var localLoginProviderVariables = []string{
	"LOCAL_LOGIN_PROVIDER_KEY", "LOCAL_LOGIN_PROVIDER_DISPLAY_NAME", "LOCAL_LOGIN_PROVIDER_ISSUER_URL",
	"LOCAL_LOGIN_PROVIDER_CLIENT_ID", "LOCAL_LOGIN_PROVIDER_CLIENT_SECRET",
}

// withLocalLoginProvider turns on the local login provider of development
// in an environment.
func withLocalLoginProvider(t *testing.T, environment map[string]string) map[string]string {
	t.Helper()
	for _, variable := range localLoginProviderVariables {
		environment[variable] = requiredVariable(t, variable)
	}
	return environment
}

// startPathOf is where a login through the provider starts.
func startPathOf(key string) string {
	return "/api/session/login_providers/" + key + "/start"
}

// callbackOf is the address Dex sends the browser back to.
func callbackOf(t *testing.T) string {
	t.Helper()
	return appOrigin() + "/api/session/login_providers/" + requiredVariable(t, "LOCAL_LOGIN_PROVIDER_KEY") + "/callback"
}

// stayingPut is the browser, following no redirect: the test reads where
// each answer sends it.
func stayingPut(browser *http.Client) *http.Client {
	put := *browser
	put.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &put
}

// toTheAPI is the callback Dex sent the browser back to, at the API under
// test.
func toTheAPI(t *testing.T, api string, callback *url.URL) string {
	t.Helper()
	return api + callback.Path + "?" + callback.RawQuery
}

// throughDex starts a login through the local provider from the browser,
// with the path to return to, logs the person in at Dex, and brings the
// browser back to the callback of the API. It returns the answer of the
// callback, which is not followed.
func throughDex(t *testing.T, browser *http.Client, api string, person oidctest.Person, returnTo string) asked {
	t.Helper()
	start := api + startPathOf(requiredVariable(t, "LOCAL_LOGIN_PROVIDER_KEY")) + "?" + url.Values{"return_to": {returnTo}}.Encode()
	callback := oidctest.LogIn(t, browser, start, person, callbackOf(t))
	return sendFrom(t, stayingPut(browser), http.MethodGet, toTheAPI(t, api, callback), "", nil)
}

func requireRedirectTo(t *testing.T, answer asked, location string) {
	t.Helper()
	require.Equal(t, http.StatusSeeOther, answer.status, answer.body)
	assert.Equal(t, location, answer.header.Get("Location"))
}

func loginStateOf(t *testing.T, browser *http.Client, api string) asked {
	t.Helper()
	return sendFrom(t, browser, http.MethodGet, api+loginStatePath, "", nil)
}

// The first login through Dex finds the administrator by the e-mail Dex
// says is verified, links the account, logs her in with a new cookie and
// sends the browser to the path it was at; the audit log records the login
// with the key of the provider as the way in.
func TestALoginThroughTheLocalProviderLogsTheUserInAndReturnsToThePathOfTheApp(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	browser := aBrowser(t)

	concluded := throughDex(t, browser, api, oidctest.Administrator, "/projects?tab=files")

	requireRedirectTo(t, concluded, "/projects?tab=files")
	cookie := sessionCookie(t, concluded)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	who := whoIsLoggedIn(t, browser, api)
	require.Equal(t, http.StatusOK, who.status, who.body)
	assert.Equal(t, oidctest.Administrator.Email, who.fields["email"])
	owner := asOwner(t, temporary)
	var entry string
	require.NoError(t, owner.QueryRow(context.Background(), `
		SELECT concat_ws(' ', operation, actor_kind, (actor_id = users.id)::text, way_in)
		FROM audit_log JOIN users ON users.email = $1
		WHERE operation = 'logged_in'`, oidctest.Administrator.Email).Scan(&entry))
	assert.Equal(t, "logged_in user true local", entry)
	var links int
	require.NoError(t, owner.QueryRow(context.Background(), `
		SELECT count(*) FROM provider_links JOIN users ON users.id = provider_links.user_id
		WHERE provider_key = 'local' AND users.email = $1`, oidctest.Administrator.Email).Scan(&links))
	assert.Equal(t, 1, links)
}

// Once linked, the account logs its user in by the link: the user's e-mail
// changed since, and the second login creates no other link.
func TestALaterLoginThroughTheLocalProviderFindsTheUserByTheLink(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	requireRedirectTo(t, throughDex(t, aBrowser(t), api, oidctest.Administrator, "/"), "/")
	owner := asOwner(t, temporary)
	_, err := owner.Exec(context.Background(), `UPDATE users SET email = 'ana.renamed@fabrica.example' WHERE email = $1`, oidctest.Administrator.Email)
	require.NoError(t, err)
	browser := aBrowser(t)

	concluded := throughDex(t, browser, api, oidctest.Administrator, "/")

	requireRedirectTo(t, concluded, "/")
	who := whoIsLoggedIn(t, browser, api)
	require.Equal(t, http.StatusOK, who.status, who.body)
	assert.Equal(t, "ana.renamed@fabrica.example", who.fields["email"])
	var links int
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT count(*) FROM provider_links").Scan(&links))
	assert.Equal(t, 1, links)
}

// A user who has another account of the provider linked is known there by
// that one: the account of Dex with their e-mail logs nobody in, and the
// browser goes back to the login with a notice of its own (story 17).
func TestAnAccountWithTheEmailOfAUserWhoLinkedAnotherReturnsToTheLoginWithItsNotice(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	owner := asOwner(t, temporary)
	_, err := owner.Exec(context.Background(), `INSERT INTO provider_links (provider_key, subject, user_id, linked_at)
		SELECT 'local', 'another-account-at-dex', id, now() FROM users WHERE email = $1`, oidctest.Administrator.Email)
	require.NoError(t, err)
	browser := aBrowser(t)

	concluded := throughDex(t, browser, api, oidctest.Administrator, "/projects")

	requireRedirectTo(t, concluded, "/login?notice=another_account_linked&return_to=%2Fprojects")
	requireNotLoggedIn(t, whoIsLoggedIn(t, browser, api))
	assert.NotContains(t, loginStateOf(t, browser, api).body, "refused_email")
}

// The developer of Dex is no user: the browser goes back to the login with
// the notice, the e-mail in no address, and the login state tells it to this
// browser alone.
func TestAVerifiedEmailOfNoActiveUserReturnsToTheLoginAndTheLoginStateTellsIt(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	browser := aBrowser(t)

	concluded := throughDex(t, browser, api, oidctest.Developer, "/projects")

	requireRedirectTo(t, concluded, "/login?notice=no_active_user&return_to=%2Fprojects")
	requireNotLoggedIn(t, whoIsLoggedIn(t, browser, api))
	state := loginStateOf(t, browser, api)
	require.Equal(t, http.StatusOK, state.status, state.body)
	assert.JSONEq(t, `{
		"login_providers": [{"key": "local", "display_name": "`+os.Getenv("LOCAL_LOGIN_PROVIDER_DISPLAY_NAME")+`"}],
		"refused_email": "`+oidctest.Developer.Email+`"
	}`, state.body)
	again := loginStateOf(t, browser, api)
	assert.Equal(t, state.body, again.body, "reading the login state forgot the refusal")
	elsewhere := loginStateOf(t, aBrowser(t), api)
	assert.NotContains(t, elsewhere.body, "refused_email")
	var links int
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(), "SELECT count(*) FROM provider_links").Scan(&links))
	assert.Zero(t, links)
}

// The login state needs no session: a browser with no cookie reads the
// providers turned on, and none when the provider is off; the read sets no
// cookie.
func TestTheLoginStateNeedsNoSession(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	on := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	off := startAPI(t, environmentOf(temporary))

	withProvider := loginStateOf(t, aBrowser(t), on)
	withNone := loginStateOf(t, aBrowser(t), off)

	require.Equal(t, http.StatusOK, withProvider.status, withProvider.body)
	assert.JSONEq(t, `{"login_providers": [{"key": "local", "display_name": "`+os.Getenv("LOCAL_LOGIN_PROVIDER_DISPLAY_NAME")+`"}]}`, withProvider.body)
	assert.Empty(t, withProvider.cookies, "reading the login state created a session")
	require.Equal(t, http.StatusOK, withNone.status, withNone.body)
	assert.JSONEq(t, `{"login_providers": []}`, withNone.body)
}

// The start sends the browser to Dex with a state, a nonce and a challenge,
// and sets the cookie of the session, which keeps them; a path to return to
// that is not a path of the app returns to the root.
func TestStartingALoginSendsTheBrowserToTheProviderWithTheCookieOfTheSession(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	browser := aBrowser(t)

	started := sendFrom(t, stayingPut(browser), http.MethodGet,
		api+startPathOf("local")+"?return_to="+url.QueryEscape("https://elsewhere.example/"), "", nil)

	require.Equal(t, http.StatusFound, started.status, started.body)
	sent, err := url.Parse(started.header.Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, requiredVariable(t, "LOCAL_LOGIN_PROVIDER_ISSUER_URL")+"/auth", sent.Scheme+"://"+sent.Host+sent.Path)
	for _, parameter := range []string{"state", "nonce", "code_challenge"} {
		assert.NotEmpty(t, sent.Query().Get(parameter), parameter)
	}
	sessionCookie(t, started)
	callback := oidctest.LogIn(t, browser, started.header.Get("Location"), oidctest.Administrator, callbackOf(t))
	requireRedirectTo(t, sendFrom(t, stayingPut(browser), http.MethodGet, toTheAPI(t, api, callback), "", nil), "/")
}

// A callback that brings a state this browser was not given, a cancel at the
// provider, or a callback in a browser that started nothing logs nobody in
// and returns to the login with its notice.
func TestACallbackThatIsNotTheOneOfThisBrowserLogsNobodyIn(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	start := api + startPathOf("local") + "?return_to=%2Fprojects"

	cases := map[string]func(t *testing.T, browser *http.Client) string{
		"another state": func(t *testing.T, browser *http.Client) string {
			callback := oidctest.LogIn(t, browser, start, oidctest.Administrator, callbackOf(t))
			query := callback.Query()
			query.Set("state", "a-state-made-up-elsewhere")
			return api + callback.Path + "?" + query.Encode()
		},
		"a cancel at the provider": func(t *testing.T, browser *http.Client) string {
			started := sendFrom(t, stayingPut(browser), http.MethodGet, start, "", nil)
			sent, err := url.Parse(started.header.Get("Location"))
			require.NoError(t, err)
			query := url.Values{"error": {"access_denied"}, "error_description": {"The person cancelled."}, "state": {sent.Query().Get("state")}}
			return api + "/api/session/login_providers/local/callback?" + query.Encode()
		},
		"a browser that started nothing": func(t *testing.T, _ *http.Client) string {
			return toTheAPI(t, api, oidctest.LogIn(t, aBrowser(t), start, oidctest.Administrator, callbackOf(t)))
		},
	}
	for name, callbackIn := range cases {
		// One after the other, so the audit log is read once they all ran.
		t.Run(name, func(t *testing.T) {
			browser := aBrowser(t)
			address := callbackIn(t, browser)

			concluded := sendFrom(t, stayingPut(browser), http.MethodGet, address, "", nil)

			require.Equal(t, http.StatusSeeOther, concluded.status, concluded.body)
			location := concluded.header.Get("Location")
			assert.Contains(t, []string{"/login?notice=login_not_concluded&return_to=%2Fprojects", "/login?notice=login_not_concluded"}, location)
			requireNotLoggedIn(t, whoIsLoggedIn(t, browser, api))
		})
	}
	var logins int
	require.NoError(t, asOwner(t, temporary).QueryRow(context.Background(), "SELECT count(*) FROM audit_log WHERE operation = 'logged_in'").Scan(&logins))
	assert.Zero(t, logins)
}

// The callback may carry parameters a provider adds of its own: they are not
// refused.
func TestACallbackWithParametersOfTheProviderIsNotRefused(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	withAUser(t, temporary, api, oidctest.Administrator.Email)
	browser := aBrowser(t)
	callback := oidctest.LogIn(t, browser, api+startPathOf("local"), oidctest.Administrator, callbackOf(t))
	query := callback.Query()
	query.Set("authuser", "0")
	query.Set("prompt", "none")

	concluded := sendFrom(t, stayingPut(browser), http.MethodGet, api+callback.Path+"?"+query.Encode(), "", nil)

	requireRedirectTo(t, concluded, "/")
}

func TestAProviderThatIsNotTurnedOnIsNotFound(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	api := startAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	off := startAPI(t, environmentOf(temporary))

	for _, address := range []string{
		api + startPathOf("google"),
		api + "/api/session/login_providers/google/callback?code=a&state=b",
		off + startPathOf("local"),
		off + "/api/session/login_providers/local/callback?code=a&state=b",
	} {
		answer := sendFrom(t, stayingPut(aBrowser(t)), http.MethodGet, address, "", nil)

		require.Equal(t, http.StatusNotFound, answer.status, address)
		assert.Equal(t, "application/problem+json", answer.mediaType)
		assert.Equal(t, "not_found", answer.fields["code"])
	}
}

// A provider that does not answer sends the browser back to the login with
// the notice, and the API names the provider in its log.
func TestAProviderThatCannotBeReachedReturnsToTheLogin(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	environment := withLocalLoginProvider(t, environmentOf(temporary))
	// Nothing answers on port 1 of the machine.
	environment["LOCAL_LOGIN_PROVIDER_ISSUER_URL"] = "http://127.0.0.1:1"
	served := launchAPI(t, environment)
	t.Cleanup(func() {
		served.signal()
		<-served.status
	})

	started := sendFrom(t, stayingPut(aBrowser(t)), http.MethodGet, served.url+startPathOf("local")+"?return_to=%2Fprojects", "", nil)

	require.Equal(t, http.StatusFound, started.status, started.body)
	assert.Equal(t, "/login?notice=login_not_concluded&return_to=%2Fprojects", started.header.Get("Location"))
	assert.Contains(t, served.messages(t), "a login through a provider failed")
}

func TestTheAPINamesTheLocalLoginProviderInItsLogWhenItIsOn(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	on := launchAPI(t, withLocalLoginProvider(t, environmentOf(temporary)))
	off := launchAPI(t, environmentOf(temporary))
	t.Cleanup(func() {
		on.signal()
		off.signal()
		<-on.status
		<-off.status
	})

	lines := on.lines(t)
	position := slices.IndexFunc(lines, func(line map[string]any) bool { return line["msg"] == "the local login provider is on" })
	require.GreaterOrEqual(t, position, 0, "no line names the local login provider: %v", on.messages(t))
	assert.Equal(t, "local", lines[position]["login_provider"])
	assert.Equal(t, requiredVariable(t, "LOCAL_LOGIN_PROVIDER_ISSUER_URL"), lines[position]["issuer"])
	assert.NotContains(t, off.messages(t), "the local login provider is on")
}

func TestTheAPIRefusesToStartWithTheLocalLoginProviderHalfSet(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	environment := withLocalLoginProvider(t, environmentOf(temporary))
	environment["LOCAL_LOGIN_PROVIDER_CLIENT_SECRET"] = ""
	var stdout, stderr output

	status := run(context.Background(), nil, environment, &stdout, &stderr)

	assert.Equal(t, 1, status)
	assert.Contains(t, stderr.String(), "LOCAL_LOGIN_PROVIDER_CLIENT_SECRET is required")
}
