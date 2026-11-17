// Package domain holds the aggregates of the consultancy area: the
// consultancy, its users and roles, and the setup code that opens the setup.
package domain

import (
	"errors"
	"io"
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

// SetupCodeOpens tells what a setup code opens. It is never stored: it
// follows from whether the setup is concluded.
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

// RestoreSetupCode rebuilds a setup code from what a repository stored.
// endedAt is nil while the setup code was not ended.
func RestoreSetupCode(id SetupCodeID, hash SetupCodeHash, issuedAt, expiresAt time.Time, endedAt *time.Time) *SetupCode {
	setupCode := &SetupCode{id: id, hash: hash, issuedAt: issuedAt, expiresAt: expiresAt}
	if endedAt != nil {
		setupCode.End(*endedAt)
	}
	return setupCode
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

// EndedAt returns when the setup code was ended, and false while it was not.
func (c *SetupCode) EndedAt() (time.Time, bool) {
	return c.endedAt, c.ended
}
