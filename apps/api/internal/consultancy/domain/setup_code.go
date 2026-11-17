// Package domain holds the aggregates of the consultancy area: the
// consultancy, its users and roles, and the setup code that opens the setup.
package domain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// SetupCodeLifetime is how long a setup code works after it is issued.
const SetupCodeLifetime = time.Hour

// SetupCodeID identifies a setup code.
type SetupCodeID int64

// NewSetupCodeID reads an identifier that came from the sequence of setup
// codes.
func NewSetupCodeID(value int64) (SetupCodeID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of a setup code is a positive number")
	}
	return SetupCodeID(value), nil
}

// SetupCodeOpens tells what a setup code opens.
type SetupCodeOpens string

// OpensFirstSetup is the registration of the consultancy and of its first
// administrator.
const OpensFirstSetup SetupCodeOpens = "first_setup"

// SetupCode is the record of a setup code issued: its hash and when it was
// issued, when it expires and when it was ended. The code itself is never
// part of it. Whether the code still works follows from the three times and
// is never stored.
type SetupCode struct {
	id        SetupCodeID
	hash      SetupCodeHash
	issuedAt  time.Time
	expiresAt time.Time
	endedAt   time.Time
	ended     bool
}

// IssueSetupCode creates a setup code from the random source. It returns the
// record to save and the code itself, which is given to whoever installs the
// system and kept nowhere.
func IssueSetupCode(id SetupCodeID, issuedAt time.Time, random io.Reader) (*SetupCode, PlainSetupCode, error) {
	plain, err := newPlainSetupCode(random)
	if err != nil {
		return nil, PlainSetupCode{}, err
	}
	setupCode := &SetupCode{
		id:        id,
		hash:      plain.Hash(),
		issuedAt:  issuedAt,
		expiresAt: issuedAt.Add(SetupCodeLifetime),
	}
	return setupCode, plain, nil
}

// ID returns the identifier of the setup code.
func (c *SetupCode) ID() SetupCodeID {
	return c.id
}

// Hash returns the hash the setup code is found by.
func (c *SetupCode) Hash() SetupCodeHash {
	return c.hash
}

// IssuedAt returns when the setup code was issued.
func (c *SetupCode) IssuedAt() time.Time {
	return c.issuedAt
}

// ExpiresAt returns when the setup code stops working by itself.
func (c *SetupCode) ExpiresAt() time.Time {
	return c.expiresAt
}

// End makes the setup code stop working from this moment on. A setup code
// already ended keeps the moment it was first ended at.
func (c *SetupCode) End(at time.Time) {
	if c.ended {
		return
	}
	c.endedAt = at
	c.ended = true
}

// EndedAt returns when the setup code was ended, and false while it was not.
func (c *SetupCode) EndedAt() (time.Time, bool) {
	return c.endedAt, c.ended
}

// SetupCodeHash is the SHA-256 hash of a setup code in its normalised form.
// It is all the system stores of a code.
type SetupCodeHash [sha256.Size]byte

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

// Hash returns the hash of the normalised code.
func (p PlainSetupCode) Hash() SetupCodeHash {
	return sha256.Sum256([]byte(p.normalised))
}
