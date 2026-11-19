package kernel

import (
	"errors"
	"fmt"
	"regexp"
)

// KeyMaxLength is the longest key.
const KeyMaxLength = 64

// keyForm is the form of a key: words of lowercase letters and digits joined
// by one underscore, starting with a letter.
var keyForm = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

// Key is a name the system gives something in snake_case, such as the key of
// a login provider, the way in of a login, or the code of an answer of a
// service: lowercase letters and digits, in words joined by one underscore,
// starting with a letter, KeyMaxLength at most. Being in that form, it carries
// nothing a person typed, so a log line can repeat it.
type Key string

// errNotAKey is the text of a key that is not in the form.
var errNotAKey = fmt.Errorf("a key is lowercase letters and digits, in words joined by one underscore, starting with a letter, %d at most", KeyMaxLength)

// NewKey reads a key.
func NewKey(text string) (Key, error) {
	if len(text) > KeyMaxLength || !keyForm.MatchString(text) {
		return "", errNotAKey
	}
	return Key(text), nil
}

// WayIn is how a user logs in: with a login code, LoginCodeWayIn, or through
// a login provider, named by its key. The audit log records it with each
// login. The access area logs in, and the audit area records it.
type WayIn string

// LoginCodeWayIn is a login with a login code.
const LoginCodeWayIn WayIn = "login_code"

// NewWayIn reads the way in of a login: a key. A login provider is
// configured, so any key is one.
func NewWayIn(text string) (WayIn, error) {
	key, err := NewKey(text)
	return WayIn(key), err
}

// NewProviderWayIn reads the key of a login provider, which is the way in of
// a login through it: a key, and never the way in of the login code.
func NewProviderWayIn(text string) (WayIn, error) {
	wayIn, err := NewWayIn(text)
	if err != nil {
		return "", err
	}
	if wayIn == LoginCodeWayIn {
		return "", errors.New("the key of a login provider is never the way in of the login code")
	}
	return wayIn, nil
}
