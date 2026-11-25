package application

import (
	"context"
	"errors"

	filesapplication "github.com/dannilodms/forgespark/apps/api/internal/files/application"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// errNoConsultancy is a failure of the system: nobody logs in before the
// setup registered the consultancy, which is never removed.
var errNoConsultancy = errors.New("no consultancy is registered, and a user is logged in")

// TellConsultancy tells any logged-in user the consultancy as the shell of
// the app shows it: its name and its logo. It asks for no scope (ADR-0026).
type TellConsultancy struct {
	reader       ConsultancyReader
	signDownload *filesapplication.SignDownload
}

// NewTellConsultancy builds the use case. The files area signs the URL the
// browser reads the logo from.
func NewTellConsultancy(reader ConsultancyReader, signDownload *filesapplication.SignDownload) *TellConsultancy {
	return &TellConsultancy{reader: reader, signDownload: signDownload}
}

// ShownConsultancy is the consultancy as the shell shows it.
type ShownConsultancy struct {
	Name string
	// LogoURL reads the logo from the file storage for fifteen minutes, and is
	// empty when the consultancy has no logo.
	LogoURL string
}

// Execute reads the consultancy for the user who acts, whatever their scopes.
func (u *TellConsultancy) Execute(ctx context.Context, _ kernel.Actor) (ShownConsultancy, error) {
	consultancy, found, err := u.reader.Read(ctx)
	if err != nil {
		return ShownConsultancy{}, err
	}
	if !found {
		return ShownConsultancy{}, errNoConsultancy
	}
	logoURL, err := u.signDownload.URLOf(ctx, (*filesapplication.FileID)(consultancy.LogoFileID))
	if err != nil {
		return ShownConsultancy{}, err
	}
	return ShownConsultancy{Name: consultancy.Name, LogoURL: logoURL}, nil
}
