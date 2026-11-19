package applicationtest

import (
	"context"
	"fmt"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// ActiveUserReader reads the users who can log in from the in-memory
// repository, as the query of the database reads them from the table.
type ActiveUserReader struct {
	users *UserRepository
}

// NewActiveUserReader returns a reader over this repository. A test of the
// login registers a user with Save, and deactivates one by saving them
// deactivated.
func NewActiveUserReader(users *UserRepository) *ActiveUserReader {
	return &ActiveUserReader{users: users}
}

// RegisterUser stores a developer with this e-mail and preferred language,
// and fictitious data for the rest, for a test of another area that needs a
// user who can log in. It returns the identifier of the user.
func (r *UserRepository) RegisterUser(email, preferredLanguage string) (domain.UserID, error) {
	details, err := domain.NewUserDetails(domain.TypedUser{
		Name:              "Someone Of The Tests",
		Email:             email,
		Country:           "BR",
		NationalID:        "000.000.000-00",
		PreferredLanguage: preferredLanguage,
	})
	if err != nil {
		return 0, err
	}
	id, err := r.NextIdentifier(context.Background())
	if err != nil {
		return 0, err
	}
	user, err := domain.RegisterUser(id, details, domain.RoleID(2))
	if err != nil {
		return 0, err
	}
	return id, r.Save(context.Background(), user)
}

// DeactivateUser deactivates the stored user with this identifier, as the
// use case that deactivates a user will.
func (r *UserRepository) DeactivateUser(id domain.UserID) error {
	position := slices.IndexFunc(r.users, func(stored domain.User) bool { return stored.ID() == id })
	if position < 0 {
		return fmt.Errorf("no user %d", id)
	}
	r.users[position].Deactivate()
	return nil
}

// ActiveUserWith returns the stored user who has this e-mail and is not
// deactivated.
func (r *ActiveUserReader) ActiveUserWith(_ context.Context, email kernel.Email) (application.ActiveUser, bool, error) {
	users := r.users.All()
	position := slices.IndexFunc(users, func(user domain.User) bool {
		return user.Details().Email() == email && !user.Deactivated()
	})
	if position < 0 {
		return application.ActiveUser{}, false, nil
	}
	found := users[position]
	return application.ActiveUser{
		ID:                found.ID(),
		Email:             found.Details().Email(),
		PreferredLanguage: found.Details().PreferredLanguage(),
	}, true, nil
}
