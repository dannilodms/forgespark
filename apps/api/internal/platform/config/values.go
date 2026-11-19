package config

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/mail"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var errNotValid = errors.New("not valid")

// ListenAddress is the host and port a server listens on, as host:port. The
// host may be empty, which means every interface.
type ListenAddress string

// UnmarshalText reads a listen address from the text of a variable.
func (a *ListenAddress) UnmarshalText(text []byte) error {
	_, port, err := net.SplitHostPort(string(text))
	if err != nil {
		return errNotValid
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return errNotValid
	}
	*a = ListenAddress(text)
	return nil
}

func (a ListenAddress) String() string {
	return string(a)
}

// HTTPURL is the address of a service reached over HTTP, as an http or https
// URL. The zero value is no address: the variable was left empty.
type HTTPURL struct {
	address string
}

// UnmarshalText reads an address from the text of a variable.
func (u *HTTPURL) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))
	if err != nil {
		return errNotValid
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errNotValid
	}
	if parsed.Host == "" {
		return errNotValid
	}
	u.address = string(text)
	return nil
}

// IsSet tells whether the variable held an address.
func (u HTTPURL) IsSet() bool {
	return u.address != ""
}

func (u HTTPURL) String() string {
	return u.address
}

// PostgresURL is a connection URL of PostgreSQL. It carries a password, so it
// prints as a placeholder: only ConnectionString gives the URL itself.
type PostgresURL struct {
	connectionString string
}

// UnmarshalText reads a connection URL from the text of a variable. It accepts
// what the pool accepts, in the URL form only.
func (u *PostgresURL) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))
	if err != nil {
		return errNotValid
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errNotValid
	}
	if _, err := pgxpool.ParseConfig(string(text)); err != nil {
		return errNotValid
	}
	u.connectionString = string(text)
	return nil
}

// NewPostgresURL reads a connection URL that comes from somewhere other than
// a variable, such as a temporary database of the tests, as a variable is
// read.
func NewPostgresURL(text string) (PostgresURL, error) {
	var connection PostgresURL
	err := connection.UnmarshalText([]byte(text))
	return connection, err
}

// GoString keeps the URL out of a %#v as well.
func (u PostgresURL) GoString() string {
	return u.String()
}

// ConnectionString returns the URL, password included. It goes to the
// database driver and nowhere else.
func (u PostgresURL) ConnectionString() string {
	return u.connectionString
}

func (u PostgresURL) String() string {
	return "[PostgreSQL connection URL]"
}

// Origin is the address a browser reaches a site at: a scheme, a host and
// maybe a port, with nothing after them.
type Origin string

// UnmarshalText reads an origin from the text of a variable. One slash at
// the end is accepted and dropped.
func (o *Origin) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))
	if err != nil {
		return errNotValid
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errNotValid
	}
	bare := url.URL{Scheme: parsed.Scheme, Host: parsed.Host}
	if parsed.Host == "" || strings.TrimSuffix(string(text), "/") != bare.String() {
		return errNotValid
	}
	*o = Origin(bare.String())
	return nil
}

func (o Origin) String() string {
	return string(o)
}

// IsHTTPS tells whether the site is served over HTTPS.
func (o Origin) IsHTTPS() bool {
	return strings.HasPrefix(string(o), "https://")
}

// Mailbox is an e-mail address whose owner receives the e-mail of several
// addresses derived from it, each with a tag after a plus sign in the local
// part (`name+tag@example.com`), as the mailbox of the seed does.
type Mailbox struct {
	localPart string
	domain    string
}

// UnmarshalText reads a mailbox from the text of a variable: one address and
// nothing around it, no name and no angle brackets.
func (m *Mailbox) UnmarshalText(text []byte) error {
	address := string(text)
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Name != "" || parsed.Address != address {
		return errNotValid
	}
	localPart, domain, _ := strings.Cut(address, "@")
	m.localPart, m.domain = localPart, domain
	return nil
}

// Subaddress returns the address that reaches the mailbox with this tag:
// `name+tag@example.com`.
func (m Mailbox) Subaddress(tag string) string {
	return m.localPart + "+" + tag + "@" + m.domain
}

func (m Mailbox) String() string {
	return m.localPart + "@" + m.domain
}

// EmailAddress is one e-mail address and nothing around it, no name and no
// angle brackets, such as the address the e-mail of the system is sent from.
type EmailAddress string

// UnmarshalText reads an e-mail address from the text of a variable.
func (a *EmailAddress) UnmarshalText(text []byte) error {
	address := string(text)
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Name != "" || parsed.Address != address {
		return errNotValid
	}
	*a = EmailAddress(address)
	return nil
}

func (a EmailAddress) String() string {
	return string(a)
}

// EmailSender is what the worker sends e-mail through.
type EmailSender string

// The senders of e-mail.
const (
	// SMTPSender sends to an SMTP server, Mailpit outside production.
	SMTPSender EmailSender = "smtp"
	// ResendSender sends through the API of Resend, in production.
	ResendSender EmailSender = "resend"
)

// UnmarshalText reads a sender from the text of a variable.
func (s *EmailSender) UnmarshalText(text []byte) error {
	sender := EmailSender(text)
	if sender != SMTPSender && sender != ResendSender {
		return errNotValid
	}
	*s = sender
	return nil
}

// ServerAddress is where a server that the binary calls listens, as
// host:port, with a host.
type ServerAddress struct {
	host string
	port int
}

// UnmarshalText reads a server address from the text of a variable.
func (a *ServerAddress) UnmarshalText(text []byte) error {
	host, port, err := net.SplitHostPort(string(text))
	if err != nil || host == "" {
		return errNotValid
	}
	number, err := strconv.ParseUint(port, 10, 16)
	if err != nil || number == 0 {
		return errNotValid
	}
	a.host, a.port = host, int(number)
	return nil
}

// Host returns the host of the server.
func (a ServerAddress) Host() string {
	return a.host
}

// Port returns the port of the server.
func (a ServerAddress) Port() int {
	return a.port
}

func (a ServerAddress) String() string {
	return net.JoinHostPort(a.host, strconv.Itoa(a.port))
}

// KeyMinSize is the shortest Key, in bytes.
const KeyMinSize = 32

// Key is a secret key of at least KeyMinSize bytes, written in hexadecimal,
// such as the key the login codes are computed with (`openssl rand -hex 32`
// writes one). It prints as a placeholder: only Reveal gives its bytes.
type Key struct {
	bytes []byte
}

// UnmarshalText reads a key from the text of a variable.
func (k *Key) UnmarshalText(text []byte) error {
	decoded, err := hex.DecodeString(string(text))
	if err != nil || len(decoded) < KeyMinSize {
		return errNotValid
	}
	k.bytes = decoded
	return nil
}

// Reveal returns the bytes of the key. They go to what computes with the key
// and nowhere else.
func (k Key) Reveal() []byte {
	return append([]byte(nil), k.bytes...)
}

func (k Key) String() string {
	return "[secret]"
}

// GoString keeps the key out of a %#v as well.
func (k Key) GoString() string {
	return "[secret]"
}

// Secret is a value that opens something, such as the key the API signs with
// for the file storage. It prints as a placeholder: only Reveal gives it.
type Secret struct {
	value string
}

// NewSecret holds a secret that comes from somewhere other than a variable,
// such as the key the tests read.
func NewSecret(value string) Secret {
	return Secret{value: value}
}

// UnmarshalText reads a secret from the text of a variable.
func (s *Secret) UnmarshalText(text []byte) error {
	s.value = string(text)
	return nil
}

// Reveal returns the secret. It goes to the library that uses it and
// nowhere else.
func (s Secret) Reveal() string {
	return s.value
}

func (s Secret) String() string {
	return "[secret]"
}

// GoString keeps the secret out of a %#v as well.
func (s Secret) GoString() string {
	return "[secret]"
}

// WebhookSecret is the secret a service signs the webhooks it sends with, as
// Svix writes it for Resend: `whsec_` followed by the key in base64. It
// prints as a placeholder: only Key gives the key. The zero value is no
// secret: the variable was left empty.
type WebhookSecret struct {
	key []byte
}

// NewWebhookSecret reads a secret that comes from somewhere other than a
// variable, such as the one the tests sign with, as a variable is read.
func NewWebhookSecret(text string) (WebhookSecret, error) {
	var secret WebhookSecret
	err := secret.UnmarshalText([]byte(text))
	return secret, err
}

// UnmarshalText reads a secret from the text of a variable.
func (s *WebhookSecret) UnmarshalText(text []byte) error {
	encoded, prefixed := strings.CutPrefix(string(text), "whsec_")
	key, err := base64.StdEncoding.DecodeString(encoded)
	if !prefixed || err != nil || len(key) == 0 {
		return errNotValid
	}
	s.key = key
	return nil
}

// IsSet tells whether the variable held a secret.
func (s WebhookSecret) IsSet() bool {
	return len(s.key) > 0
}

// Key returns the key the webhooks are signed with. It goes to the check of
// a signature and nowhere else.
func (s WebhookSecret) Key() []byte {
	return slices.Clone(s.key)
}

func (s WebhookSecret) String() string {
	return "[secret]"
}

// GoString keeps the secret out of a %#v as well.
func (s WebhookSecret) GoString() string {
	return "[secret]"
}
