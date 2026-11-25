package application

import (
	"context"

	consultancyapplication "github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// FindActor finds who acts in a request: the user logged in to the session of
// its browser, with the scopes of their roles, read on every request
// (ADR-0026). The REST adapter runs it for every operation that needs a user,
// and passes the actor to the use case of the operation.
type FindActor struct {
	session            Session
	findActiveUserByID *consultancyapplication.FindActiveUserByID
}

// NewFindActor builds the use case.
func NewFindActor(session Session, findActiveUserByID *consultancyapplication.FindActiveUserByID) *FindActor {
	return &FindActor{session: session, findActiveUserByID: findActiveUserByID}
}

// Execute returns the actor of the request, and false when nobody is logged
// in to its session. A session whose user was deactivated, or no longer
// exists, is ended, and has no actor: deactivating a user takes effect on
// their next request.
func (u *FindActor) Execute(ctx context.Context) (kernel.Actor, bool, error) {
	user, loggedIn, err := u.session.User(ctx)
	if err != nil || !loggedIn {
		return kernel.Actor{}, false, err
	}
	active, found, err := u.findActiveUserByID.Execute(ctx, consultancyapplication.UserID(user))
	if err != nil {
		return kernel.Actor{}, false, err
	}
	if !found {
		return kernel.Actor{}, false, u.session.End(ctx)
	}
	return kernel.NewActor(int64(active.ID), active.Scopes), true, nil
}
