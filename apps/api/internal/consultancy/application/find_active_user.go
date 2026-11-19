package application

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// The login of the access area holds a user by identifier, and answers the
// keys of their roles and their photo. An adapter of another area reaches
// this area only through this package, so the words it needs are named here.
type (
	// UserID identifies a user.
	UserID = domain.UserID
	// RoleKey is the fixed key of a role the setup creates.
	RoleKey = domain.RoleKey
	// FileID identifies a file of the files area, such as the photo of a
	// user.
	FileID = domain.FileID
)

// ActiveUser is a user who can log in, as the login and the session of a
// browser need them: who they are, the e-mail a login code goes to and the
// language it is written in, the name and the photo the app shows, the keys
// of their roles and the scopes those roles grant (ADR-0026).
type ActiveUser struct {
	ID                domain.UserID
	Name              string
	Email             kernel.Email
	PreferredLanguage kernel.Language
	// PhotoFileID is the file of their photo, and nil when they have none.
	PhotoFileID *domain.FileID
	Roles       []domain.RoleKey
	Scopes      []kernel.Scope
}

// withScopes gives the user the scopes of their roles.
func withScopes(user ActiveUser, found bool, err error) (ActiveUser, bool, error) {
	if err != nil || !found {
		return ActiveUser{}, false, err
	}
	user.Scopes = domain.ScopesOf(user.Roles)
	return user, true, nil
}

// FindActiveUser finds the user who logs in with an e-mail, for the login of
// the access area. A deactivated user cannot log in, so they are not found,
// as an e-mail that belongs to no user is not.
type FindActiveUser struct {
	reader ActiveUserReader
}

// NewFindActiveUser builds the use case.
func NewFindActiveUser(reader ActiveUserReader) *FindActiveUser {
	return &FindActiveUser{reader: reader}
}

// Execute returns the user who is not deactivated and has this e-mail, and
// false when there is none. It locks nothing.
func (u *FindActiveUser) Execute(ctx context.Context, email kernel.Email) (ActiveUser, bool, error) {
	return withScopes(u.reader.ActiveUserWith(ctx, email))
}

// FindActiveUserByID finds the user a session of the access area holds, on
// every request, with the scopes of their roles: a change of role shows on
// the next request, and a deactivated user, who is not found, is refused on
// it.
type FindActiveUserByID struct {
	reader ActiveUserReader
}

// NewFindActiveUserByID builds the use case.
func NewFindActiveUserByID(reader ActiveUserReader) *FindActiveUserByID {
	return &FindActiveUserByID{reader: reader}
}

// Execute returns the user with this identifier when they are not
// deactivated, and false when there is none. It locks nothing.
func (u *FindActiveUserByID) Execute(ctx context.Context, id UserID) (ActiveUser, bool, error) {
	return withScopes(u.reader.ActiveUserOf(ctx, id))
}
