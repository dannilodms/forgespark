// Package domain holds the aggregates of the access area: the login code
// request, and the rules of the login code and of its limits (Login in
// CODING_STANDARDS.md, ADR-0011). The code itself is never part of a record:
// it is computed from the request whenever it is needed.
package domain

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

// LoginCodeLifetime is how long a login code works after it is asked for.
const LoginCodeLifetime = 10 * time.Minute

// LoginCodeRequestKeptFor is how long a login code request is kept once it
// ended or expired: no code works with it, and the delivery of its e-mail,
// which reads it, has long run out of attempts.
const LoginCodeRequestKeptFor = 24 * time.Hour

// randomValueSize is the size of the random value of a request, in bytes.
const randomValueSize = 32

// LoginCodeRequestID identifies a login code request.
type LoginCodeRequestID int64

// NoLoginCodeRequest is what a browser that asked for a code for an e-mail of
// no active user is left with: no request was saved, and no code works with
// it. The session keeps it as it keeps a request, so that asking answers the
// same way for every e-mail.
const NoLoginCodeRequest LoginCodeRequestID = 0

// NewLoginCodeRequestID reads an identifier that came from the sequence of
// login code requests.
func NewLoginCodeRequestID(value int64) (LoginCodeRequestID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of a login code request is a positive number")
	}
	return LoginCodeRequestID(value), nil
}

// UserID identifies a user, whom the consultancy area keeps. The access area
// holds a user by this identifier only.
type UserID int64

// LoginCodeRequest is the record of a login code asked for: whose it is, the
// random value its code is computed from, when it was asked for and expires,
// the attempts made with it and when it was ended. The code is never part of
// it. Whether the code still works follows from the times and the attempts,
// and is never stored.
type LoginCodeRequest struct {
	id          LoginCodeRequestID
	user        UserID
	randomValue []byte
	askedAt     time.Time
	expiresAt   time.Time
	attempts    int
	endedAt     time.Time
	ended       bool
}

// AskForLoginCode creates the request of a login code for the user, with a
// random value drawn from the random source.
func AskForLoginCode(id LoginCodeRequestID, user UserID, askedAt time.Time, random io.Reader) (*LoginCodeRequest, error) {
	value := make([]byte, randomValueSize)
	if _, err := io.ReadFull(random, value); err != nil {
		return nil, fmt.Errorf("read the random source: %w", err)
	}
	return &LoginCodeRequest{
		id:          id,
		user:        user,
		randomValue: value,
		askedAt:     askedAt,
		expiresAt:   askedAt.Add(LoginCodeLifetime),
	}, nil
}

// StoredLoginCodeRequest is a login code request as a repository stored it.
// EndedAt is nil while the request was not ended.
type StoredLoginCodeRequest struct {
	ID          LoginCodeRequestID
	User        UserID
	RandomValue []byte
	AskedAt     time.Time
	ExpiresAt   time.Time
	Attempts    int
	EndedAt     *time.Time
}

// RestoreLoginCodeRequest rebuilds a request from what a repository stored. It
// checks nothing that was checked when the request was saved.
func RestoreLoginCodeRequest(stored StoredLoginCodeRequest) *LoginCodeRequest {
	request := &LoginCodeRequest{
		id:          stored.ID,
		user:        stored.User,
		randomValue: bytes.Clone(stored.RandomValue),
		askedAt:     stored.AskedAt,
		expiresAt:   stored.ExpiresAt,
		attempts:    stored.Attempts,
	}
	if stored.EndedAt != nil {
		request.End(*stored.EndedAt)
	}
	return request
}

// End makes the code of the request stop working from this moment on, as a
// new request of the same user does. A request already ended keeps the moment
// it was first ended at.
func (r *LoginCodeRequest) End(at time.Time) {
	if r.ended {
		return
	}
	r.endedAt = at
	r.ended = true
}

// ID returns the identifier of the request.
func (r *LoginCodeRequest) ID() LoginCodeRequestID { return r.id }

// User returns the user the request is for.
func (r *LoginCodeRequest) User() UserID { return r.user }

// RandomValue returns the random value the code is computed from, for a
// repository to store.
func (r *LoginCodeRequest) RandomValue() []byte { return bytes.Clone(r.randomValue) }

// AskedAt returns when the code was asked for.
func (r *LoginCodeRequest) AskedAt() time.Time { return r.askedAt }

// ExpiresAt returns when the code stops working by itself.
func (r *LoginCodeRequest) ExpiresAt() time.Time { return r.expiresAt }

// Attempts returns how many wrong codes were typed for the request.
func (r *LoginCodeRequest) Attempts() int { return r.attempts }

// EndedAt returns when the request was ended, and false while it was not.
func (r *LoginCodeRequest) EndedAt() (time.Time, bool) { return r.endedAt, r.ended }
