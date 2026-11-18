package domain

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// setupCodeAlphabet is Crockford's Base32: the digits and the letters
// without I, L, O and U, which read like other characters.
const setupCodeAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

const (
	setupCodeLength    = 20
	setupCodeGroupSize = 5
)

// PlainSetupCode is a setup code as whoever installs the system reads and
// types it. It is a secret: it prints as a placeholder, and only Reveal
// gives its characters.
type PlainSetupCode struct {
	// normalised holds the 20 characters, upper case and with no hyphen.
	normalised string
}

// newPlainSetupCode draws a setup code from the random source: 20 characters
// of five bits each, about 100 bits. Each character takes the five low bits
// of one random byte, so every character is as likely as any other.
func newPlainSetupCode(random io.Reader) (PlainSetupCode, error) {
	drawn := make([]byte, setupCodeLength)
	if _, err := io.ReadFull(random, drawn); err != nil {
		return PlainSetupCode{}, fmt.Errorf("read the random source: %w", err)
	}
	for i, each := range drawn {
		drawn[i] = setupCodeAlphabet[each&0b11111]
	}
	return PlainSetupCode{normalised: string(drawn)}, nil
}

// ReadPlainSetupCode reads a setup code as a person typed it. Case, hyphens
// and spaces do not count, and the letters that are never printed because
// they read like a digit are taken for that digit: O for 0, I and L for 1.
//
// A text that is not a setup code after that is answered as a wrong code is.
func ReadPlainSetupCode(typed string) (PlainSetupCode, error) {
	normalised := strings.Map(readAsPrinted, strings.ToUpper(typed))
	if len(normalised) != setupCodeLength || strings.Trim(normalised, setupCodeAlphabet) != "" {
		return PlainSetupCode{}, SetupCodeInvalidError{}
	}
	return PlainSetupCode{normalised: normalised}, nil
}

// readAsPrinted gives the character of a printed setup code that a typed,
// upper-case character stands for, and drops the ones that separate groups.
func readAsPrinted(typed rune) rune {
	switch {
	case typed == '-' || unicode.IsSpace(typed):
		return -1
	case typed == 'O':
		return '0'
	case typed == 'I' || typed == 'L':
		return '1'
	}
	return typed
}

// Reveal returns the code in four groups of five characters, as it is shown
// to whoever installs the system. It goes to the output of the command that
// issued the code and nowhere else.
func (p PlainSetupCode) Reveal() string {
	groups := make([]string, 0, setupCodeLength/setupCodeGroupSize)
	for start := 0; start < len(p.normalised); start += setupCodeGroupSize {
		groups = append(groups, p.normalised[start:start+setupCodeGroupSize])
	}
	return strings.Join(groups, "-")
}

// Hash returns the hash of the normalised code, which is what is stored.
func (p PlainSetupCode) Hash() SetupCodeHash {
	return sha256.Sum256([]byte(p.normalised))
}

const setupCodePlaceholder = "[setup code]"

// String keeps the code out of whatever formats it: a log line, an error, a
// printed struct.
func (p PlainSetupCode) String() string {
	return setupCodePlaceholder
}

// GoString does the same for the %#v verb.
func (p PlainSetupCode) GoString() string {
	return setupCodePlaceholder
}

// MarshalText does the same for an encoder, JSON included.
func (p PlainSetupCode) MarshalText() ([]byte, error) {
	return []byte(setupCodePlaceholder), nil
}

// SetupCodeHash is the SHA-256 hash of a setup code in its normalised form.
// It is all the system stores of a code.
type SetupCodeHash [sha256.Size]byte

// SetupCodeHashFromBytes reads a hash that a repository stored.
func SetupCodeHashFromBytes(stored []byte) (SetupCodeHash, error) {
	if len(stored) != sha256.Size {
		return SetupCodeHash{}, fmt.Errorf("the hash of a setup code has %d bytes, not %d", sha256.Size, len(stored))
	}
	return SetupCodeHash(stored), nil
}
