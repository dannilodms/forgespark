package domain

import (
	"crypto/subtle"
	"time"
)

// LoginCodeMaxAttempts is how many wrong codes a request takes: the third
// ends it.
const LoginCodeMaxAttempts = 3

// Works tells whether the code of the request still logs its user in at this
// moment: the request was not ended, by a new request, by its use or by its
// third wrong code, and its ten minutes have not passed. It is computed,
// never stored.
func (r *LoginCodeRequest) Works(at time.Time) bool {
	return !r.ended && at.Before(r.expiresAt) && r.attempts < LoginCodeMaxAttempts
}

// Check checks the code typed for the request at this moment, comparing it
// with the code of the request in constant time. The right code, while the
// request works, ends the request, which is used, and returns nil: the user
// of the request logs in.
//
// Anything else is a wrong attempt, and is answered the same way whether the
// code was wrong or the request no longer works, so the answer tells nothing
// of why: LoginCodeInvalidError with the attempts left, and, at the third,
// LoginCodeEndedError, which ends the request.
func (r *LoginCodeRequest) Check(typed string, key LoginCodeKey, at time.Time) error {
	// Compared before the request is asked whether it works, so the time of
	// the answer is the same either way.
	right := r.Code(key).matches(typed)
	if right && r.Works(at) {
		r.End(at)
		return nil
	}
	r.attempts = min(r.attempts+1, LoginCodeMaxAttempts)
	if r.attempts == LoginCodeMaxAttempts {
		r.End(at)
		return LoginCodeEndedError{}
	}
	return LoginCodeInvalidError{AttemptsLeft: LoginCodeMaxAttempts - r.attempts}
}

// StandInLoginCodeRequest stands in for the request of a browser that holds
// none: it asked for a code for an e-mail of no active user, or asked for no
// code, or its request is gone. It never works, so every code typed for it is
// a wrong attempt, answered as one for a request that exists: the answer tells
// nothing of the e-mail (ADR-0011). Nothing of it is saved; the session of
// the browser keeps the attempts made with it.
func StandInLoginCodeRequest(attempts int) *LoginCodeRequest {
	return &LoginCodeRequest{id: NoLoginCodeRequest, attempts: min(max(attempts, 0), LoginCodeMaxAttempts), ended: true}
}

// matches tells, in constant time, whether the digits typed are the code.
func (c LoginCode) matches(typed string) bool {
	return subtle.ConstantTimeCompare([]byte(c.digits), []byte(typed)) == 1
}

// LoginCodeInvalidError is the answer to a login code that does not log in:
// a wrong code, or the code of a request that no longer works, or a code
// typed in a browser that asked for none. The person may try again while
// attempts are left.
type LoginCodeInvalidError struct {
	AttemptsLeft int
}

func (LoginCodeInvalidError) Error() string {
	return "the login code is not valid; another may be typed while attempts are left, which `attempts_left` counts"
}

// Code returns the stable code an inbound adapter answers with.
func (LoginCodeInvalidError) Code() string {
	return "login_code_invalid"
}

// Params returns how many attempts are left, which the text of the code
// shows.
func (e LoginCodeInvalidError) Params() map[string]any {
	return map[string]any{"attempts_left": e.AttemptsLeft}
}

// LoginCodeEndedError is the answer to the third wrong attempt: the code no
// longer works, and the person asks for a new one.
type LoginCodeEndedError struct{}

func (LoginCodeEndedError) Error() string {
	return "the login code ended after three wrong attempts: ask for a new one"
}

// Code returns the stable code an inbound adapter answers with.
func (LoginCodeEndedError) Code() string {
	return "login_code_ended"
}
