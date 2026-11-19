package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// LoginCodeKeyMinSize is the shortest key a login code is computed with, in
// bytes: as long as the hash HMAC-SHA256 runs on.
const LoginCodeKeyMinSize = sha256.Size

// LoginCodeKey is the key the login codes are computed with, which comes from
// configuration and never from the database. Changing it ends only the codes
// not yet used.
type LoginCodeKey struct {
	key []byte
}

// NewLoginCodeKey holds a key of at least LoginCodeKeyMinSize bytes.
func NewLoginCodeKey(key []byte) (LoginCodeKey, error) {
	if len(key) < LoginCodeKeyMinSize {
		return LoginCodeKey{}, fmt.Errorf("the key of the login code has %d bytes, and at least %d are needed", len(key), LoginCodeKeyMinSize)
	}
	return LoginCodeKey{key: append([]byte(nil), key...)}, nil
}

// loginCodeDigits is how many digits a login code has.
const loginCodeDigits = 6

// Code computes the login code of the request: HMAC-SHA256 under the key,
// over the identifier of the request (eight bytes, big-endian) followed by its
// random value, reduced to six digits by the dynamic truncation of RFC 4226.
// The delivery of the e-mail computes it to write it, and the check computes
// it to compare: it is stored nowhere.
func (r *LoginCodeRequest) Code(key LoginCodeKey) LoginCode {
	mac := hmac.New(sha256.New, key.key)
	mac.Write(binary.BigEndian.AppendUint64(nil, uint64(r.id)))
	mac.Write(r.randomValue)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return LoginCode{digits: fmt.Sprintf("%0*d", loginCodeDigits, value%1_000_000)}
}

// LoginCode is the six digits a person receives by e-mail and types to log
// in. It is a secret: it prints as a placeholder, and only Reveal gives its
// digits.
type LoginCode struct {
	digits string
}

// Reveal returns the six digits. They go into the e-mail of the code and
// nowhere else.
func (c LoginCode) Reveal() string {
	return c.digits
}

const loginCodePlaceholder = "[login code]"

// String keeps the code out of whatever formats it: a log line, an error, a
// printed struct.
func (c LoginCode) String() string {
	return loginCodePlaceholder
}

// GoString does the same for the %#v verb.
func (c LoginCode) GoString() string {
	return loginCodePlaceholder
}

// MarshalText does the same for an encoder, JSON included.
func (c LoginCode) MarshalText() ([]byte, error) {
	return []byte(loginCodePlaceholder), nil
}
