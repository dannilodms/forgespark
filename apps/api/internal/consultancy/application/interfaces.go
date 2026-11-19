// Package application holds the use cases of the consultancy area and the
// interfaces they need. A use case is a type built once, with what it needs,
// and run with Execute.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// Transactions opens and closes the transaction of a use case that writes.
type Transactions interface {
	// Run calls the operation inside one transaction, and passes it the
	// context every repository must receive to take part. The transaction is
	// committed when the operation returns nil and rolled back, with nothing
	// saved, when it returns an error.
	Run(ctx context.Context, operation func(ctx context.Context) error) error
}

// Clock tells the current time. A use case never reads the time of the
// machine by itself, so a test decides what time it is.
type Clock interface {
	Now() time.Time
}

// SetupCodeRepository keeps the setup codes.
type SetupCodeRepository interface {
	// NextIdentifier reserves the identifier of a setup code that does not
	// exist yet.
	NextIdentifier(ctx context.Context) (domain.SetupCodeID, error)
	// LoadNotEnded loads every setup code that was not ended, expired or
	// not, and locks each one until the transaction ends.
	LoadNotEnded(ctx context.Context) ([]*domain.SetupCode, error)
	// FindByHash returns the setup code with this hash, ended or not, and
	// false when there is none. It locks nothing: it serves a use case that
	// checks a setup code and saves nothing.
	FindByHash(ctx context.Context, hash domain.SetupCodeHash) (*domain.SetupCode, bool, error)
	// LoadByHash loads the setup code with this hash, ended or not, and false
	// when there is none. It locks the setup code until the transaction ends,
	// so two operations that use the same code run one after the other and
	// the second sees whether the first ended it.
	LoadByHash(ctx context.Context, hash domain.SetupCodeHash) (*domain.SetupCode, bool, error)
	// Save stores a setup code whole, whether it is new or changed.
	Save(ctx context.Context, setupCode *domain.SetupCode) error
}

// ConsultancyRepository keeps the consultancy, of which there is one.
type ConsultancyRepository interface {
	// NextIdentifier reserves the identifier of the consultancy.
	NextIdentifier(ctx context.Context) (domain.ConsultancyID, error)
	// Exists tells whether the consultancy was registered, which is whether
	// the setup is concluded. It locks nothing.
	Exists(ctx context.Context) (bool, error)
	// Save stores the consultancy whole, whether it is new or changed.
	Save(ctx context.Context, consultancy *domain.Consultancy) error
}

// UserRepository keeps the users, with the roles each one has.
type UserRepository interface {
	// NextIdentifier reserves the identifier of a user that does not exist
	// yet.
	NextIdentifier(ctx context.Context) (domain.UserID, error)
	// EmailInUse tells whether a user, deactivated or not, has this e-mail.
	// It locks nothing: the use cases that register a user run under the
	// lock of the one setup code that works, so two of them do not overlap.
	EmailInUse(ctx context.Context, email kernel.Email) (bool, error)
	// Save stores a user whole, with their roles, whether they are new or
	// changed.
	Save(ctx context.Context, user *domain.User) error
}

// RoleRepository keeps the roles.
type RoleRepository interface {
	// NextIdentifier reserves the identifier of a role that does not exist
	// yet.
	NextIdentifier(ctx context.Context) (domain.RoleID, error)
	// FindByKey returns the role the setup created with this key, and false
	// when there is none. It locks nothing: a role with a key is never
	// removed, and a user who receives it changes nothing of it.
	FindByKey(ctx context.Context, key domain.RoleKey) (*domain.Role, bool, error)
	// Save stores a role whole, whether it is new or changed.
	Save(ctx context.Context, role *domain.Role) error
}

// ActiveUserReader reads the users who can log in, shaped for the login and
// the session, with the keys of their roles, sorted. It leaves the scopes to
// the use case, which reads them from the roles.
type ActiveUserReader interface {
	// ActiveUserWith returns the user who has this e-mail and is not
	// deactivated, and false when there is none. It locks nothing.
	ActiveUserWith(ctx context.Context, email kernel.Email) (ActiveUser, bool, error)
	// ActiveUserOf returns the user with this identifier when they are not
	// deactivated, and false when there is none. It locks nothing.
	ActiveUserOf(ctx context.Context, id domain.UserID) (ActiveUser, bool, error)
}

// SetupStateReader reads the state of the setup, shaped for whoever asks.
type SetupStateReader interface {
	// Read tells whether the setup is concluded and whether a setup code
	// works at this moment. It locks nothing.
	Read(ctx context.Context, at time.Time) (SetupState, error)
}
