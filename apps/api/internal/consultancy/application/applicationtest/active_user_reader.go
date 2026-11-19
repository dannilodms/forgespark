package applicationtest

import (
	"context"
	"fmt"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// The roles the setup creates on an empty database, by their identifiers, as
// the users RegisterUser and RegisterAdministrator store receive them.
const (
	administratorRole domain.RoleID = 1
	developerRole     domain.RoleID = 2
)

// ActiveUserReader reads the users who can log in from the in-memory
// repository, as the query of the database reads them from the table.
type ActiveUserReader struct {
	users *UserRepository
	roles map[domain.RoleID]domain.RoleKey
}

// NewActiveUserReader returns a reader over this repository, where the roles
// are the two the setup creates: the administrator role is 1 and the
// developer role is 2. A test of the login registers a user with
// RegisterUser or RegisterAdministrator, and deactivates one with
// DeactivateUser.
func NewActiveUserReader(users *UserRepository) *ActiveUserReader {
	return &ActiveUserReader{
		users: users,
		roles: map[domain.RoleID]domain.RoleKey{
			administratorRole: domain.AdministratorRoleKey,
			developerRole:     domain.DeveloperRoleKey,
		},
	}
}

// RegisterUser stores a developer with this e-mail and preferred language,
// and fictitious data for the rest, for a test of another area that needs a
// user who can log in. It returns the identifier of the user.
func (r *UserRepository) RegisterUser(email, preferredLanguage string) (domain.UserID, error) {
	return r.register("Someone Of The Tests", email, preferredLanguage, 0, developerRole)
}

// RegisterAdministrator stores an administrator with this name, e-mail and
// preferred language, as RegisterUser stores a developer.
func (r *UserRepository) RegisterAdministrator(name, email, preferredLanguage string) (domain.UserID, error) {
	return r.register(name, email, preferredLanguage, 0, administratorRole)
}

// RegisterAdministratorWithPhoto stores an administrator as
// RegisterAdministrator does, with the file of their photo.
func (r *UserRepository) RegisterAdministratorWithPhoto(name, email, preferredLanguage string, photo domain.FileID) (domain.UserID, error) {
	return r.register(name, email, preferredLanguage, photo, administratorRole)
}

func (r *UserRepository) register(name, email, preferredLanguage string, photo domain.FileID, role domain.RoleID) (domain.UserID, error) {
	details, err := domain.NewUserDetails(domain.TypedUser{
		Name:              name,
		Email:             email,
		Country:           "BR",
		NationalID:        "000.000.000-00",
		PreferredLanguage: preferredLanguage,
		PhotoFileID:       photo,
	})
	if err != nil {
		return 0, err
	}
	id, err := r.NextIdentifier(context.Background())
	if err != nil {
		return 0, err
	}
	user, err := domain.RegisterUser(id, details, role)
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
	return r.activeUser(func(user domain.User) bool { return user.Details().Email() == email })
}

// ActiveUserOf returns the stored user with this identifier when they are not
// deactivated.
func (r *ActiveUserReader) ActiveUserOf(_ context.Context, id domain.UserID) (application.ActiveUser, bool, error) {
	return r.activeUser(func(user domain.User) bool { return user.ID() == id })
}

func (r *ActiveUserReader) activeUser(wanted func(domain.User) bool) (application.ActiveUser, bool, error) {
	users := r.users.All()
	position := slices.IndexFunc(users, func(user domain.User) bool { return wanted(user) && !user.Deactivated() })
	if position < 0 {
		return application.ActiveUser{}, false, nil
	}
	found := users[position]
	roles := []domain.RoleKey{}
	for _, role := range found.Roles() {
		roles = append(roles, r.roles[role])
	}
	slices.Sort(roles)
	return application.ActiveUser{
		ID:                found.ID(),
		Name:              found.Details().Name(),
		Email:             found.Details().Email(),
		PreferredLanguage: found.Details().PreferredLanguage(),
		PhotoFileID:       fileOrNone(found.Details().PhotoFileID()),
		Roles:             roles,
	}, true, nil
}
