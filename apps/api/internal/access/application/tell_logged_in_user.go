package application

import (
	"context"

	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// TellLoggedInUser tells the app who is logged in: the user it shows in its
// menu and whose preferred language it speaks. It asks for no scope.
type TellLoggedInUser struct {
	findActiveUserByID *consultancyapplication.FindActiveUserByID
}

// NewTellLoggedInUser builds the use case.
func NewTellLoggedInUser(findActiveUserByID *consultancyapplication.FindActiveUserByID) *TellLoggedInUser {
	return &TellLoggedInUser{findActiveUserByID: findActiveUserByID}
}

// LoggedInUser is the user logged in, as the app shows them: their name,
// their e-mail, their preferred language and the keys of their roles.
type LoggedInUser struct {
	Name              string
	Email             kernel.Email
	PreferredLanguage kernel.Language
	Roles             []string
}

// Execute returns the user who acts, and false when they can no longer log
// in, deactivated since their session was read.
func (u *TellLoggedInUser) Execute(ctx context.Context, actor kernel.Actor) (LoggedInUser, bool, error) {
	user, found, err := u.findActiveUserByID.Execute(ctx, consultancyapplication.UserID(actor.User()))
	if err != nil || !found {
		return LoggedInUser{}, false, err
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, string(role))
	}
	return LoggedInUser{
		Name:              user.Name,
		Email:             user.Email,
		PreferredLanguage: user.PreferredLanguage,
		Roles:             roles,
	}, true, nil
}
