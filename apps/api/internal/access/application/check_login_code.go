package application

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// CheckLoginCode checks the six digits typed in a browser against the login
// code request that browser asked for, and logs its user in when they are
// right (Login in CODING_STANDARDS.md, ADR-0011). The code is computed again
// from the request and compared in constant time: it is stored nowhere.
type CheckLoginCode struct {
	login LoginCodeDependencies
	logIn *LogIn
}

// NewCheckLoginCode builds the use case. The key of the dependencies is the
// one the delivery of the e-mail computed the code with.
func NewCheckLoginCode(login LoginCodeDependencies, logIn *LogIn) *CheckLoginCode {
	return &CheckLoginCode{login: login, logIn: logIn}
}

// CheckLoginCodeInput is the code as the person typed it.
type CheckLoginCodeInput struct {
	Code string
}

// Execute checks the code. The request is read from the session of the
// browser, so a code works only in the browser that asked for it. The right
// code, while its request works, ends the request and logs its user in: the
// session is renewed and keeps the user, and the audit log records the login
// with the login code as the way in.
//
// Anything else answers domain.LoginCodeInvalidError with the attempts left,
// and the third time domain.LoginCodeEndedError, which ends the request: a
// wrong code, the code of a request ended or expired, a browser that asked
// for no code or for an e-mail of no active user, whose attempts the session
// counts. The right code of a user deactivated since they asked ends the
// request and answers domain.LoginCodeEndedError, and logs nobody in.
func (u *CheckLoginCode) Execute(ctx context.Context, input CheckLoginCodeInput) error {
	id, err := u.login.Session.LoginCodeRequest(ctx)
	if err != nil {
		return err
	}
	var refusal error
	var noRequest bool
	loggedIn, err := u.logIn.Execute(ctx, kernel.LoginCodeWayIn, func(ctx context.Context) (domain.UserID, bool, error) {
		// Locked, so two checks of one request count their attempts one
		// after the other.
		request, found, err := u.login.Requests.Load(ctx, id)
		noRequest = !found
		if err != nil || !found {
			return 0, false, err
		}
		refusal = request.Check(input.Code, u.login.Key, u.login.Clock.Now())
		if err := u.login.Requests.Save(ctx, request); err != nil {
			return 0, false, err
		}
		return request.User(), refusal == nil, nil
	})
	if err != nil {
		return err
	}
	if noRequest {
		return u.checkWithoutRequest(ctx, input.Code)
	}
	if refusal == nil && !loggedIn {
		return domain.LoginCodeEndedError{}
	}
	return refusal
}

// checkWithoutRequest answers a code typed in a browser that holds no
// request as a wrong code for one, counting the attempts in its session.
func (u *CheckLoginCode) checkWithoutRequest(ctx context.Context, typed string) error {
	attempts, err := u.login.Session.AttemptsWithoutRequest(ctx)
	if err != nil {
		return err
	}
	standIn := domain.StandInLoginCodeRequest(attempts)
	refusal := standIn.Check(typed, u.login.Key, u.login.Clock.Now())
	if err := u.login.Session.RememberAttemptsWithoutRequest(ctx, standIn.Attempts()); err != nil {
		return err
	}
	return refusal
}
