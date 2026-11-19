package application

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
)

// LogInDependencies are what logging a user in needs: the transactions, the
// session of the browser, the users who can log in, the audit log and the
// clock.
type LogInDependencies struct {
	Transactions       Transactions
	Session            Session
	FindActiveUserByID *consultancyapplication.FindActiveUserByID
	RecordEntry        *auditapplication.RecordEntry
	Clock              Clock
}

// LogIn is the step every way in ends with: the login code checked, and the
// login through a provider (ADR-0011). It is no operation of its own: each
// way in calls it.
type LogIn struct {
	with LogInDependencies
}

// NewLogIn builds the step.
func NewLogIn(with LogInDependencies) *LogIn {
	return &LogIn{with: with}
}

// FindUser is what a way in does inside the transaction of the login: it
// checks what the person gave, saves what that changes, and returns the user
// to log in, or false when it refuses. A refusal is no error, so what it
// saved is kept, such as an attempt counted.
type FindUser func(ctx context.Context) (domain.UserID, bool, error)

// Execute runs the way in, then logs in the user it found. In one
// transaction, the way in finds the user and, when that user is active, the
// audit log records `logged_in` with the way in, so a login is never without
// its record. Once the transaction commits, the session of the browser is
// renewed and keeps the user.
//
// It returns false, and logs nobody in, when the way in refused or the user it
// found cannot log in: deactivated, or gone.
func (u *LogIn) Execute(ctx context.Context, way domain.WayIn, findUser FindUser) (bool, error) {
	var user domain.UserID
	var active bool
	err := u.with.Transactions.Run(ctx, func(ctx context.Context) error {
		found, accepted, err := findUser(ctx)
		if err != nil || !accepted {
			return err
		}
		user = found
		active, err = u.record(ctx, user, way)
		return err
	})
	if err != nil || !active {
		return false, err
	}
	return true, u.with.Session.LogIn(ctx, user)
}

// record records the login of the user, when they can log in, and tells
// whether they can.
func (u *LogIn) record(ctx context.Context, user domain.UserID, way domain.WayIn) (bool, error) {
	_, active, err := u.with.FindActiveUserByID.Execute(ctx, consultancyapplication.UserID(user))
	if err != nil || !active {
		return false, err
	}
	return true, u.with.RecordEntry.Execute(ctx, auditapplication.RecordEntryInput{
		OccurredAt: u.with.Clock.Now(),
		Operation:  auditapplication.LoggedIn,
		Actor:      auditapplication.ActingUser(int64(user)),
		Subject:    auditapplication.User(int64(user)),
		WayIn:      auditapplication.WayIn(way),
	})
}
