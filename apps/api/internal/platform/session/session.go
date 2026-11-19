// Package session keeps the session of each browser (Login in
// CODING_STANDARDS.md, ADR-0011): the browser holds an opaque cookie that
// JavaScript cannot read, and what the session holds lives in PostgreSQL,
// through scs and its store pgxstore. A session exists before the login too,
// for what the login needs between two requests of one browser.
//
// The scs session manager is embedded, so an adapter reads and writes the
// session of a request with its methods (Put, GetInt64, Remove, RenewToken,
// Destroy) on the context of the request. What each key means belongs to the
// adapter that writes it.
package session

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

const (
	// CookieName names the cookie of the session.
	CookieName = "session"
	// IdleTimeout ends a session that was not used for this long.
	IdleTimeout = 7 * 24 * time.Hour
	// Lifetime ends a session this long after it started, used or not. The
	// login renews the token of the session, which starts it again.
	Lifetime = 30 * 24 * time.Hour
)

// Settings is what the sessions take from configuration.
type Settings struct {
	// Secure sends the cookie over HTTPS only. It is on wherever the app is
	// served over HTTPS, and off in development, which has no certificate.
	Secure bool
	// AppOrigin is the origin the app is served from, as APP_ORIGIN names
	// it: a request that changes state with the cookie must come from it.
	AppOrigin config.Origin
}

// Sessions are the sessions of the browsers, kept in the sessions table.
type Sessions struct {
	*scs.SessionManager

	appOrigin config.Origin
}

// New builds the sessions on the pool. The store removes no session that
// ended: an ended session is never read, and the daily cleanup of the worker
// removes it (Jobs in CODING_STANDARDS.md), so the API runs no query of its
// own outside a request.
//
// The cookie is HttpOnly and SameSite=Lax, so the callback of a login
// provider, a navigation from another site, carries it. Its token is 256
// random bits, and the table keeps its SHA-256 only. A session ends after
// IdleTimeout without use and, in any case, Lifetime after it started.
func New(pool *pgxpool.Pool, settings Settings) *Sessions {
	// An interval of zero starts no cleanup of the store.
	store := pgxstore.NewWithCleanupInterval(pool, 0)
	manager := scs.New()
	manager.Store = store
	manager.IdleTimeout = IdleTimeout
	manager.Lifetime = Lifetime
	manager.HashTokenInStore = true
	manager.Cookie.Name = CookieName
	manager.Cookie.Path = "/"
	manager.Cookie.HttpOnly = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = settings.Secure
	manager.Cookie.Persist = true
	// A session that cannot be read or saved fails the request in the error
	// format, with its raw error in the logs.
	manager.ErrorFunc = restapi.WriteFailure
	return &Sessions{SessionManager: manager, appOrigin: settings.AppOrigin}
}

// Under loads the session of each request whose path is under the prefix,
// and saves it, with its cookie, before the answer leaves. The other paths,
// liveness and readiness, touch no session, so they never wait on the
// database.
//
// Before that, it refuses a request from another site that would change
// state with the cookie (fromAnotherSite), with 403 in the error format: the
// session is not read, and no handler runs.
func (s *Sessions) Under(prefix string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != prefix && !strings.HasPrefix(r.URL.Path, prefix+"/") {
			next.ServeHTTP(w, r)
			return
		}
		if s.fromAnotherSite(r) {
			restapi.WriteRefusal(w, r, http.StatusForbidden,
				"a request that changes state with the cookie of a session must come from the origin of the app")
			return
		}
		s.LoadAndSave(routeWrittenBack(r, next)).ServeHTTP(w, r)
	})
}

// readOnly are the methods that change no state, which are not checked.
var readOnly = []string{http.MethodGet, http.MethodHead, http.MethodOptions}

// fromAnotherSite tells a request that changes state with the cookie of a
// session and does not come from the app (Sessions in API conventions,
// ADR-0011): its Origin header names another origin, or it has none and its
// Sec-Fetch-Site header says `cross-site`. A request with no cookie is not
// checked: the bounce webhook carries none, nor will an access token. A
// client that sends neither header, such as curl, is not a browser another
// site can drive, and passes.
func (s *Sessions) fromAnotherSite(r *http.Request) bool {
	if slices.Contains(readOnly, r.Method) {
		return false
	}
	if _, err := r.Cookie(CookieName); err != nil {
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		// Read as the origin of the app is, so the case of its letters and a
		// default port do not tell one apart from the other. An opaque origin
		// (`null`) is no origin of the app.
		sent, err := config.NewOrigin(origin)
		return err != nil || sent != s.appOrigin
	}
	return r.Header.Get("Sec-Fetch-Site") == "cross-site"
}

// routeWrittenBack passes on the request that carries the session, and
// writes the route the mux matched on it back to the request this handler
// received, which the telemetry of the request reads: the mux writes the
// route on the request it is given, and that one is a copy.
func routeWrittenBack(received *http.Request, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, withSession *http.Request) {
		defer func() { received.Pattern = withSession.Pattern }()
		next.ServeHTTP(w, withSession)
	})
}
