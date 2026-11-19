package rest

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
)

// sessionPath is the session of the browser that calls.
const sessionPath = restapi.Prefix + "/session"

// Actors finds who acts in each request, for the operations of every area
// that need a user: the use case reads the session of the browser of the
// request.
type Actors struct {
	findActor *application.FindActor
}

// NewActors returns the actors the use case finds.
func NewActors(findActor *application.FindActor) *Actors {
	return &Actors{findActor: findActor}
}

// ActorOf returns the user logged in to the session of the request, with the
// scopes of their roles.
func (a *Actors) ActorOf(ctx context.Context) (kernel.Actor, bool, error) {
	return a.findActor.Execute(ctx)
}

// The operations of the session of a logged-in user: who it is, and logging
// out.
func registerSession(api *restapi.API, useCases UseCases) {
	restapi.RegisterForUser(api, huma.Operation{
		OperationID: "tell_logged_in_user",
		Method:      http.MethodGet,
		Path:        sessionPath,
		Summary:     "Tell who is logged in",
		Description: "Answers the user logged in to the session of this browser, as the app shows them, with a URL " +
			"that reads their photo from the file storage for fifteen minutes, and the keys of their roles. The app " +
			"reads it again sooner than that, so the photo never breaks. It asks for no scope. With no user logged in, " +
			"the session missing, ended, or of a user deactivated since, it answers 401; the session of a deactivated " +
			"user is ended.",
		Tags: []string{loginTag},
	}, nil, tellLoggedInUser{useCase: useCases.TellLoggedInUser}.handle)

	restapi.Register(api, huma.Operation{
		OperationID:   "log_out",
		Method:        http.MethodPost,
		Path:          sessionPath + "/log_out",
		DefaultStatus: http.StatusNoContent,
		Summary:       "Log out",
		Description: "Ends the session of this browser, whether a user is logged in to it or not: the answer " +
			"expires its cookie, and the session opens nothing any more.",
		Tags: []string{loginTag},
	}, restapi.Failures{
		http.StatusInternalServerError: {},
	}, logOut{useCase: useCases.LogOut}.handle)
}

type tellLoggedInUser struct {
	useCase *application.TellLoggedInUser
}

type tellLoggedInUserOutput struct {
	Body LoggedInUser
}

// LoggedInUser is the user logged in to the session.
type LoggedInUser struct {
	Name              string   `json:"name" doc:"The name of the user."`
	Email             string   `json:"email" format:"email" doc:"The e-mail the user logs in with, in lowercase."`
	PhotoURL          string   `json:"photo_url,omitempty" format:"uri" doc:"A URL that reads the photo of the user from the file storage, signed for its public host name and valid for fifteen minutes. Left out when the user has no photo."`
	PreferredLanguage language `json:"preferred_language" doc:"The supported language the user reads the system in."`
	RoleKeys          []string `json:"role_keys" enum:"administrator,developer" nullable:"false" doc:"The keys of the roles of the user, sorted, which the app translates."`
}

func (o tellLoggedInUser) handle(ctx context.Context, actor kernel.Actor, _ *struct{}) (*tellLoggedInUserOutput, error) {
	user, found, err := o.useCase.Execute(ctx, actor)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, restapi.NotLoggedIn
	}
	return &tellLoggedInUserOutput{Body: LoggedInUser{
		Name:              user.Name,
		Email:             user.Email.String(),
		PhotoURL:          user.PhotoURL,
		PreferredLanguage: language(user.PreferredLanguage.String()),
		RoleKeys:          user.Roles,
	}}, nil
}

// language is a supported language. Its schema lists the supported languages
// of kernel, so the generated client knows them.
type language string

// Schema lists the supported languages.
func (language) Schema(huma.Registry) *huma.Schema {
	tags := kernel.SupportedLanguages()
	enum := make([]any, 0, len(tags))
	for _, tag := range tags {
		enum = append(enum, tag)
	}
	return &huma.Schema{Type: huma.TypeString, Enum: enum}
}

type logOut struct {
	useCase *application.LogOut
}

// logOutOutput has no body: the operation answers 204.
type logOutOutput struct{}

func (o logOut) handle(ctx context.Context, _ *struct{}) (*logOutOutput, error) {
	if err := o.useCase.Execute(ctx); err != nil {
		return nil, err
	}
	return &logOutOutput{}, nil
}
