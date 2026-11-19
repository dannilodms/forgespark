package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log/slog"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// LoginCodeKeyMinSize is the shortest key a login code is computed with, in
// bytes: as long as the hash HMAC-SHA256 runs on.
const LoginCodeKeyMinSize = sha256.Size

// LoginCodeKey is the key the login codes are computed with, which comes from
// configuration and never from the database. Changing it ends only the codes
// not yet used, and starts the counts of the limits again. It is a secret: it
// prints as a placeholder, and only the HMACs of this package read its bytes.
type LoginCodeKey struct {
	key []byte
	// addresses is the key the limits count the addresses under, derived from
	// the key so that no HMAC of an address is ever an HMAC of a code.
	addresses []byte
}

// askedAddressLabel is what the key the addresses are counted under is
// derived from, under the key of the login code.
const askedAddressLabel = "the addresses the limits of the login code count"

// NewLoginCodeKey holds a key of at least LoginCodeKeyMinSize bytes.
func NewLoginCodeKey(key []byte) (LoginCodeKey, error) {
	if len(key) < LoginCodeKeyMinSize {
		return LoginCodeKey{}, fmt.Errorf("the key of the login code has %d bytes, and at least %d are needed", len(key), LoginCodeKeyMinSize)
	}
	return LoginCodeKey{key: append([]byte(nil), key...), addresses: hmacOf(key, []byte(askedAddressLabel))}, nil
}

func hmacOf(key []byte, parts ...[]byte) []byte {
	mac := hmac.New(sha256.New, key)
	for _, part := range parts {
		mac.Write(part)
	}
	return mac.Sum(nil)
}

const keyPlaceholder = "[secret]"

// String keeps the key out of whatever formats it: a log line, an error, a
// printed struct.
func (k LoginCodeKey) String() string {
	return keyPlaceholder
}

// GoString does the same for the %#v verb.
func (k LoginCodeKey) GoString() string {
	return keyPlaceholder
}

// MarshalText does the same for an encoder, JSON included.
func (k LoginCodeKey) MarshalText() ([]byte, error) {
	return []byte(keyPlaceholder), nil
}

// LogValue does the same for a logger.
func (k LoginCodeKey) LogValue() slog.Value {
	return slog.StringValue(keyPlaceholder)
}

// AskedAddress is the address a login code is asked for as the limits count
// it: by the HMAC-SHA256 of the address, as it reads in lowercase, under a key
// derived from the key of the login code. The table of the limits holds no
// address, and a dump of it tells none without the key.
func (k LoginCodeKey) AskedAddress(email kernel.Email) AskedAddress {
	return AskedAddress{digest: hmacOf(k.addresses, []byte(email.String()))}
}

// AskedAddress is an address a login code was asked for, kept by its keyed
// digest and never as the address (LoginCodeKey.AskedAddress).
type AskedAddress struct {
	digest []byte
}

// Digest returns the 32 bytes the limits count the address by.
func (a AskedAddress) Digest() []byte {
	return append([]byte(nil), a.digest...)
}

// loginCodeDigits is how many digits a login code has.
const loginCodeDigits = 6

// Code computes the login code of the request: HMAC-SHA256 under the key,
// over the identifier of the request (eight bytes, big-endian) followed by its
// random value, reduced to six digits by the dynamic truncation of RFC 4226.
// The delivery of the e-mail computes it to write it, and the check computes
// it to compare: it is stored nowhere.
func (r *LoginCodeRequest) Code(key LoginCodeKey) LoginCode {
	sum := hmacOf(key.key, binary.BigEndian.AppendUint64(nil, uint64(r.id)), r.randomValue)

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
