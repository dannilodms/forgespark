package applicationtest

import (
	"context"
	"errors"
	"slices"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// UserRepository keeps the users in memory, with their roles. It stores a
// copy of what it is given, as the database does.
type UserRepository struct {
	lastIdentifier int64
	users          []domain.User
	failure        error
}

// NewUserRepository returns a repository with no user.
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *UserRepository) FailWith(failure error) {
	r.failure = failure
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *UserRepository) NextIdentifier(context.Context) (domain.UserID, error) {
	r.lastIdentifier++
	return domain.NewUserID(r.lastIdentifier)
}

// EmailInUse tells whether a stored user, deactivated or not, has this
// e-mail.
func (r *UserRepository) EmailInUse(_ context.Context, email kernel.Email) (bool, error) {
	return slices.ContainsFunc(r.users, func(stored domain.User) bool {
		return stored.Details().Email() == email
	}), nil
}

// Save stores the user, in place of the one with the same identifier when
// there is one. Like the database, it refuses a second user with the same
// e-mail.
func (r *UserRepository) Save(_ context.Context, user *domain.User) error {
	if r.failure != nil {
		return r.failure
	}
	sameEmail := slices.ContainsFunc(r.users, func(stored domain.User) bool {
		return stored.ID() != user.ID() && stored.Details().Email() == user.Details().Email()
	})
	if sameEmail {
		return errors.New("another user has this e-mail")
	}
	position := slices.IndexFunc(r.users, func(stored domain.User) bool { return stored.ID() == user.ID() })
	if position < 0 {
		r.users = append(r.users, *user)
		return nil
	}
	r.users[position] = *user
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *UserRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.users)
	return func() { r.users = saved }
}

// All returns every user stored, in the order they were first saved.
func (r *UserRepository) All() []domain.User {
	return slices.Clone(r.users)
}
