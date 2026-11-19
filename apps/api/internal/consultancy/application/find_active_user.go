package application

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

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

// ActiveUser is a user who can log in, as the login needs them: who they are,
// the e-mail a login code goes to and the language it is written in.
type ActiveUser struct {
	ID                domain.UserID
	Email             kernel.Email
	PreferredLanguage kernel.Language
}

// Execute returns the user who is not deactivated and has this e-mail, and
// false when there is none. It locks nothing.
func (u *FindActiveUser) Execute(ctx context.Context, email kernel.Email) (ActiveUser, bool, error) {
	return u.reader.ActiveUserWith(ctx, email)
}
