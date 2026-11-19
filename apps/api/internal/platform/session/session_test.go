package session_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/session"
)

const day = 24 * time.Hour

// site serves, behind the sessions, a path that writes a number in the
// session, one that reads it back, and one outside the prefix of the API.
type site struct {
	server *httptest.Server
	pool   *pgxpool.Pool
}

func newSite(t *testing.T, pool *pgxpool.Pool, settings session.Settings) site {
	t.Helper()
	sessions := session.New(pool, settings)
	t.Cleanup(sessions.Stop)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/numbers", func(w http.ResponseWriter, r *http.Request) {
		number, err := strconv.ParseInt(r.URL.Query().Get("number"), 10, 64)
		require.NoError(t, err)
		sessions.Put(r.Context(), "number", number)
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("GET /api/numbers", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, strconv.FormatInt(sessions.GetInt64(r.Context(), "number"), 10))
	})
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(sessions.Under("/api", mux))
	t.Cleanup(server.Close)
	return site{server: server, pool: pool}
}

// migratedPool is a pool of the API role on a migrated temporary database.
func migratedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            databasetest.Migrated(t).APIConnection(t),
		MaxSize:                  4,
		StatementTimeout:         10 * time.Second,
		IdleInTransactionTimeout: time.Minute,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

// browser is a client that keeps the cookies it receives, as a browser does.
func browser(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &http.Client{Jar: jar, Timeout: 30 * time.Second}
}

func send(t *testing.T, client *http.Client, method, url string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, nil)
	require.NoError(t, err)
	response, err := client.Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	return response
}

func sessionCookieOf(t *testing.T, response *http.Response) *http.Cookie {
	t.Helper()
	cookies := response.Cookies()
	position := slices.IndexFunc(cookies, func(cookie *http.Cookie) bool { return cookie.Name == session.CookieName })
	require.GreaterOrEqual(t, position, 0, "the answer sets no cookie %q: %v", session.CookieName, response.Header.Values("Set-Cookie"))
	return cookies[position]
}

func TestTheSessionIsAnOpaqueCookieJavaScriptCannotReadThatLastsSevenDaysWithoutUse(t *testing.T) {
	t.Parallel()
	served := newSite(t, migratedPool(t), session.Settings{Secure: false})

	written := send(t, browser(t), http.MethodPost, served.server.URL+"/api/numbers?number=37")

	require.Equal(t, http.StatusAccepted, written.StatusCode)
	cookie := sessionCookieOf(t, written)
	assert.True(t, cookie.HttpOnly, "JavaScript can read the cookie")
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.False(t, cookie.Secure, "the cookie is sent over HTTPS only in development")
	assert.InDelta(t, (7 * day).Seconds(), float64(cookie.MaxAge), 60, "the cookie does not outlive seven days without use")
	assert.NotContains(t, cookie.Value, "37", "the cookie carries what the session holds")
}

func TestTheCookieGoesOverHTTPSOnlyWhenTheSettingsSaySo(t *testing.T) {
	t.Parallel()
	served := newSite(t, migratedPool(t), session.Settings{Secure: true})

	written := send(t, browser(t), http.MethodPost, served.server.URL+"/api/numbers?number=37")

	assert.True(t, sessionCookieOf(t, written).Secure)
}

func TestTheSessionIsReadBackWithItsCookieAndNotWithoutIt(t *testing.T) {
	t.Parallel()
	served := newSite(t, migratedPool(t), session.Settings{Secure: false})
	ana := browser(t)
	send(t, ana, http.MethodPost, served.server.URL+"/api/numbers?number=37")

	again := send(t, ana, http.MethodGet, served.server.URL+"/api/numbers")
	elsewhere := send(t, browser(t), http.MethodGet, served.server.URL+"/api/numbers")

	assert.Equal(t, "37", body(t, again))
	assert.Equal(t, "0", body(t, elsewhere), "another browser reads the session of the first")
	assert.Empty(t, elsewhere.Header.Values("Set-Cookie"), "a session that holds nothing was saved")
}

func body(t *testing.T, response *http.Response) string {
	t.Helper()
	read, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return string(read)
}

// The table keeps the SHA-256 of the token, so a dump of the database opens
// no session. The row ends after seven days without use, and the data holds
// the end of the session, thirty days after it started.
func TestTheStoreKeepsTheHashOfTheTokenAndBothLivesOfTheSession(t *testing.T) {
	t.Parallel()
	pool := migratedPool(t)
	served := newSite(t, pool, session.Settings{Secure: false})
	started := time.Now()

	cookie := sessionCookieOf(t, send(t, browser(t), http.MethodPost, served.server.URL+"/api/numbers?number=37"))

	var token string
	var data []byte
	var expiry time.Time
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT token, data, expiry FROM sessions").Scan(&token, &data, &expiry))
	assert.NotEqual(t, cookie.Value, token, "the table keeps the token of the cookie as it is")
	assert.NotContains(t, token, cookie.Value)
	assert.WithinDuration(t, started.Add(7*day), expiry, time.Minute)
	deadline, values, err := scs.GobCodec{}.Decode(data)
	require.NoError(t, err)
	assert.WithinDuration(t, started.Add(30*day), deadline, time.Minute)
	assert.Equal(t, map[string]any{"number": int64(37)}, values)
}

func TestAPathOutsideThePrefixTouchesNoSession(t *testing.T) {
	t.Parallel()
	// Port 1 of the machine has no PostgreSQL behind it: a session that is
	// read fails.
	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            unreachable(t),
		MaxSize:                  1,
		StatementTimeout:         time.Second,
		IdleInTransactionTimeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	served := newSite(t, pool, session.Settings{Secure: false})
	withACookie := browser(t)
	withACookie.Jar.SetCookies(mustParse(t, served.server.URL), []*http.Cookie{{Name: session.CookieName, Value: "a-token"}})

	alive := send(t, withACookie, http.MethodGet, served.server.URL+"/livez")
	underTheAPI := send(t, withACookie, http.MethodGet, served.server.URL+"/api/numbers")

	assert.Equal(t, http.StatusOK, alive.StatusCode, "liveness read the session")
	assert.Equal(t, http.StatusInternalServerError, underTheAPI.StatusCode)
	assert.Equal(t, "application/problem+json", underTheAPI.Header.Get("Content-Type"))
	var problem map[string]any
	require.NoError(t, json.NewDecoder(underTheAPI.Body).Decode(&problem))
	assert.Equal(t, "internal_server_error", problem["code"])
	assert.Nil(t, problem["detail"], "the answer repeats the raw error")
}

// unreachable is a connection to port 1 of the machine, where no PostgreSQL
// answers.
func unreachable(t *testing.T) config.PostgresURL {
	t.Helper()
	connection, err := config.NewPostgresURL("postgres://api:api@127.0.0.1:1/software_factory?sslmode=disable&connect_timeout=1")
	require.NoError(t, err)
	return connection
}

func mustParse(t *testing.T, address string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(address)
	require.NoError(t, err)
	return parsed
}
