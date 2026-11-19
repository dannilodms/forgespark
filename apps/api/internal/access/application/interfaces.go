// Package application holds the use cases of the access area and the
// interfaces they need: asking for a login code, and the content of its
// e-mail, which the delivery of the e-mail area asks for. A use case is a
// type built once, with what it needs, and run with Execute.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

// Transactions opens and closes the transaction of a use case that writes.
type Transactions interface {
	// Run calls the operation inside one transaction, and passes it the
	// context every repository must receive to take part. The transaction is
	// committed when the operation returns nil and rolled back, with nothing
	// saved, when it returns an error. The use cases of other areas it calls
	// join it.
	Run(ctx context.Context, operation func(ctx context.Context) error) error
}

// Clock tells the current time. A use case never reads the time of the
// machine by itself, so a test decides what time it is.
type Clock interface {
	Now() time.Time
}

// LoginCodeRequestRepository keeps the login code requests.
type LoginCodeRequestRepository interface {
	// NextIdentifier reserves the identifier of a request that does not
	// exist yet.
	NextIdentifier(ctx context.Context) (domain.LoginCodeRequestID, error)
	// LoadNotEndedOf loads every request of the user that was not ended,
	// expired or not, and locks each one until the transaction ends. A user
	// has one at most.
	LoadNotEndedOf(ctx context.Context, user domain.UserID) ([]*domain.LoginCodeRequest, error)
	// CountAskedSince counts the requests of the user asked for after this
	// moment, ended or not. It locks nothing.
	CountAskedSince(ctx context.Context, user domain.UserID, since time.Time) (int, error)
	// Find returns the request with this identifier, ended or not, and false
	// when there is none. It locks nothing: it serves the delivery of the
	// e-mail of the code, which saves nothing of the request.
	Find(ctx context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error)
	// Save stores a request whole, whether it is new or changed.
	Save(ctx context.Context, request *domain.LoginCodeRequest) error
}

// Session is the session of the browser that sent the request, which exists
// before the login: what the login keeps there between two requests of that
// browser. The context of the request carries it.
type Session interface {
	// RememberLoginCodeRequest keeps the login code request this browser
	// asked for, in place of any earlier one. domain.NoLoginCodeRequest is
	// kept for an e-mail of no active user, the same way.
	RememberLoginCodeRequest(ctx context.Context, id domain.LoginCodeRequestID) error
}
