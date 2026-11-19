package domain

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// LoginProviderKey names a login provider turned on: in the paths of a login
// through it, in its provider links and, as the way in, in the audit log. It
// is a key (kernel.Key), and never the way in of the login code.
type LoginProviderKey string

// NewLoginProviderKey reads the key of a login provider.
func NewLoginProviderKey(text string) (LoginProviderKey, error) {
	wayIn, err := kernel.NewProviderWayIn(text)
	if err != nil {
		return "", fmt.Errorf("the key of a login provider: %w", err)
	}
	return LoginProviderKey(wayIn), nil
}

func (k LoginProviderKey) String() string {
	return string(k)
}

// WayIn is the way in the audit log records for a login through this
// provider: its key.
func (k LoginProviderKey) WayIn() kernel.WayIn {
	return kernel.WayIn(k)
}

// ReturnPath is the path of the app a browser returns to once it logs in: the
// page it was at when its session ended. It is a path of the app and nothing
// else, so that a link crafted elsewhere cannot send the browser to another
// site after the login.
type ReturnPath string

// RootPath is the root of the app, where a browser lands after a login with
// no path to return to.
const RootPath ReturnPath = "/"

// returnPathMaxLength is the longest path kept.
const returnPathMaxLength = 2048

// NewReturnPath reads the path to return to. Anything but a path of the app
// returns the root: an address of another site, a path the browser would
// read as one (`//host`, `/\host`), a path of the API, or a text with a
// space, a control character or more than 2048 characters.
func NewReturnPath(text string) ReturnPath {
	if !isPathOfTheApp(text) {
		return RootPath
	}
	return ReturnPath(text)
}

func isPathOfTheApp(text string) bool {
	if len(text) > returnPathMaxLength || !strings.HasPrefix(text, "/") || strings.HasPrefix(text, "//") {
		return false
	}
	if strings.ContainsAny(text, "\\ ") || strings.ContainsFunc(text, isControl) {
		return false
	}
	parsed, err := url.Parse(text)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" {
		return false
	}
	return parsed.Path != "/api" && !strings.HasPrefix(parsed.Path, "/api/")
}

func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f
}

func (p ReturnPath) String() string {
	return string(p)
}

// secretSize is the size of the state, the nonce and the verifier of a login
// through a provider, in bytes of the random source.
const secretSize = 32

// ProviderLogin is a login through a provider in progress, which the session
// of the browser keeps from its start to the callback of the provider: which
// provider it went to, the state the callback must bring back, the nonce the
// ID token must carry, the verifier of PKCE, and the path of the app to
// return to. The callback uses it once.
type ProviderLogin struct {
	provider LoginProviderKey
	state    string
	nonce    string
	verifier string
	returnTo ReturnPath
}

// StartProviderLogin starts a login through the provider, with a state, a
// nonce and a verifier of 32 bytes each, drawn from the random source and
// written in base64url without padding: the verifier is then 43 characters,
// as RFC 7636 asks.
func StartProviderLogin(provider LoginProviderKey, returnTo ReturnPath, random io.Reader) (ProviderLogin, error) {
	drawn := make([]byte, 3*secretSize)
	if _, err := io.ReadFull(random, drawn); err != nil {
		return ProviderLogin{}, fmt.Errorf("read the random source: %w", err)
	}
	return ProviderLogin{
		provider: provider,
		state:    base64.RawURLEncoding.EncodeToString(drawn[:secretSize]),
		nonce:    base64.RawURLEncoding.EncodeToString(drawn[secretSize : 2*secretSize]),
		verifier: base64.RawURLEncoding.EncodeToString(drawn[2*secretSize:]),
		returnTo: returnTo,
	}, nil
}

// StoredProviderLogin is a login through a provider as the session kept it.
type StoredProviderLogin struct {
	Provider string
	State    string
	Nonce    string
	Verifier string
	ReturnTo string
}

// RestoreProviderLogin rebuilds a login through a provider the session kept.
func RestoreProviderLogin(stored StoredProviderLogin) (ProviderLogin, error) {
	provider, err := NewLoginProviderKey(stored.Provider)
	if err != nil {
		return ProviderLogin{}, err
	}
	if stored.State == "" || stored.Nonce == "" || stored.Verifier == "" {
		return ProviderLogin{}, errors.New("a login through a provider has a state, a nonce and a verifier")
	}
	return ProviderLogin{
		provider: provider,
		state:    stored.State,
		nonce:    stored.Nonce,
		verifier: stored.Verifier,
		returnTo: NewReturnPath(stored.ReturnTo),
	}, nil
}

// Accepts tells whether a callback of this provider with this state belongs
// to the login: the state is compared in constant time. A callback of another
// provider, or with another state, was not sent by the provider for this
// browser, and logs nobody in.
func (l ProviderLogin) Accepts(provider LoginProviderKey, state string) bool {
	sameState := subtle.ConstantTimeCompare([]byte(l.state), []byte(state)) == 1
	return sameState && provider == l.provider
}

// Provider returns the provider the login went to.
func (l ProviderLogin) Provider() LoginProviderKey {
	return l.provider
}

// State returns the state the callback brings back.
func (l ProviderLogin) State() string {
	return l.state
}

// Nonce returns the nonce the ID token carries.
func (l ProviderLogin) Nonce() string {
	return l.nonce
}

// Verifier returns the verifier of PKCE, which only the exchange of the code
// sends to the provider.
func (l ProviderLogin) Verifier() string {
	return l.verifier
}

// ReturnTo returns the path of the app to return to once logged in.
func (l ProviderLogin) ReturnTo() ReturnPath {
	return l.returnTo
}

// subjectMaxLength is the longest subject OpenID Connect allows.
const subjectMaxLength = 255

// ProviderSubject is the identifier a login provider gives one account,
// which never changes and is never given to another account: the `sub` of
// its ID token.
type ProviderSubject string

// NewProviderSubject reads the subject of an account at a provider: one to
// 255 characters.
func NewProviderSubject(text string) (ProviderSubject, error) {
	if text == "" || len(text) > subjectMaxLength {
		return "", fmt.Errorf("the subject of an account at a login provider has 1 to %d characters", subjectMaxLength)
	}
	return ProviderSubject(text), nil
}

func (s ProviderSubject) String() string {
	return string(s)
}

// ProviderLink records that a user holds one account at a login provider,
// from the first login through it (ADR-0011). Later logins through that
// provider find the user by the link, whatever e-mail the account has by
// then. A user has one link at each provider at most, and the login never
// changes a link.
type ProviderLink struct {
	provider LoginProviderKey
	subject  ProviderSubject
	user     UserID
	linkedAt time.Time
}

// LinkProvider links the account of the provider to the user, at the moment
// of their first login through it.
func LinkProvider(provider LoginProviderKey, subject ProviderSubject, user UserID, at time.Time) *ProviderLink {
	return &ProviderLink{provider: provider, subject: subject, user: user, linkedAt: at}
}

// StoredProviderLink is a provider link as a repository stored it.
type StoredProviderLink struct {
	Provider string
	Subject  string
	User     int64
	LinkedAt time.Time
}

// RestoreProviderLink rebuilds a stored provider link.
func RestoreProviderLink(stored StoredProviderLink) (*ProviderLink, error) {
	provider, err := NewLoginProviderKey(stored.Provider)
	if err != nil {
		return nil, err
	}
	subject, err := NewProviderSubject(stored.Subject)
	if err != nil {
		return nil, err
	}
	if stored.User <= 0 {
		return nil, errors.New("a provider link is the link of a user")
	}
	return LinkProvider(provider, subject, UserID(stored.User), stored.LinkedAt), nil
}

// Provider returns the provider of the account.
func (l *ProviderLink) Provider() LoginProviderKey {
	return l.provider
}

// Subject returns the account at the provider.
func (l *ProviderLink) Subject() ProviderSubject {
	return l.subject
}

// User returns the user who holds the account.
func (l *ProviderLink) User() UserID {
	return l.user
}

// LinkedAt returns when the account was linked: the first login through it.
func (l *ProviderLink) LinkedAt() time.Time {
	return l.linkedAt
}
