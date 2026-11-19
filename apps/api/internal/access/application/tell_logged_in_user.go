package application

import (
	"context"

	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// TellLoggedInUser tells the app who is logged in: the user it shows in its
// menu, with their photo, and whose preferred language it speaks. It asks for
// no scope.
type TellLoggedInUser struct {
	findActiveUserByID *consultancyapplication.FindActiveUserByID
	signDownload       *filesapplication.SignDownload
}

// NewTellLoggedInUser builds the use case. The files area signs the URL the
// browser reads the photo from.
func NewTellLoggedInUser(
	findActiveUserByID *consultancyapplication.FindActiveUserByID,
	signDownload *filesapplication.SignDownload,
) *TellLoggedInUser {
	return &TellLoggedInUser{findActiveUserByID: findActiveUserByID, signDownload: signDownload}
}

// LoggedInUser is the user logged in, as the app shows them: their name,
// their e-mail, their photo, their preferred language and the keys of their
// roles.
type LoggedInUser struct {
	Name  string
	Email kernel.Email
	// PhotoURL reads the photo from the file storage for fifteen minutes, and
	// is empty when the user has no photo.
	PhotoURL          string
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
	photoURL, err := u.signDownload.URLOf(ctx, (*filesapplication.FileID)(user.PhotoFileID))
	if err != nil {
		return LoggedInUser{}, false, err
	}
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, string(role))
	}
	return LoggedInUser{
		Name:              user.Name,
		Email:             user.Email,
		PhotoURL:          photoURL,
		PreferredLanguage: user.PreferredLanguage,
		Roles:             roles,
	}, true, nil
}
