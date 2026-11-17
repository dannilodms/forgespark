// Package application holds the use cases of the consultancy area and the
// interfaces they need. A use case is a type built once, with what it needs,
// and run with Execute.
package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
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
	// Save stores a setup code whole, whether it is new or changed.
	Save(ctx context.Context, setupCode *domain.SetupCode) error
}
