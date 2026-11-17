package config

import (
	"errors"
	"net"
	"net/url"
	"strconv"

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

// ConnectionString returns the URL, password included. It goes to the
// database driver and nowhere else.
func (u PostgresURL) ConnectionString() string {
	return u.connectionString
}

func (u PostgresURL) String() string {
	return "[PostgreSQL connection URL]"
}
