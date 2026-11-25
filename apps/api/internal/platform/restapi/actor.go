package restapi

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// Actors finds who acts in a request: the user logged in to the session of
// its browser, with the scopes of their roles (Inbound adapters in
// CODING_STANDARDS.md). The access area answers it, and the command that
// serves gives it to the API with AuthenticateWith.
type Actors interface {
	// ActorOf returns the actor of the request, and false when no user is
	// logged in to it.
	ActorOf(ctx context.Context) (kernel.Actor, bool, error)
}

// AuthenticateWith gives the API the way to find who acts in a request, which
// every operation registered with RegisterForUser asks on each request.
func (a *API) AuthenticateWith(actors Actors) {
	a.actors = actors
}

// NotLoggedIn is the answer, 401, of an operation that needs a user to a
// request with no user logged in. Every operation registered with
// RegisterForUser declares it, and its handler may return it.
var NotLoggedIn kernel.BrokenRule = notLoggedIn{}

type notLoggedIn struct{}

func (notLoggedIn) Error() string {
	return "no user is logged in to the session of the request: it has none, it ended, or its user was deactivated"
}

func (notLoggedIn) Code() string {
	return codeOfStatus(http.StatusUnauthorized)
}

// actorKey is the key of the actor in the context of a request.
type actorKey struct{}

// errNoActors is a programming error: an operation needs a user and the API
// was given no way to find one.
var errNoActors = errors.New("the API has no way to find who acts in a request: call AuthenticateWith")

// RegisterForUser adds an operation that needs a logged-in user, as Register
// does. Before the input is read, the API finds who acts in the request: with
// nobody logged in it answers NotLoggedIn, 401, and the handler does not run.
// The handler receives the actor, to pass to its use case, which checks the
// scopes it needs.
func RegisterForUser[I, O any](
	api *API,
	operation huma.Operation,
	failures Failures,
	handler func(context.Context, kernel.Actor, *I) (*O, error),
) {
	declared := Failures{}
	maps.Copy(declared, failures)
	declared[http.StatusUnauthorized] = append(slices.Clone(declared[http.StatusUnauthorized]), NotLoggedIn)
	operation.Middlewares = append(slices.Clone(operation.Middlewares), api.authenticate)

	Register(api, operation, declared, func(ctx context.Context, input *I) (*O, error) {
		actor, found := ctx.Value(actorKey{}).(kernel.Actor)
		if !found {
			return nil, errNoActors
		}
		return handler(ctx, actor, input)
	})
}

// authenticate finds who acts in the request and passes them on in its
// context, or answers 401 when nobody is logged in, or 500.
func (a *API) authenticate(ctx huma.Context, next func(huma.Context)) {
	actor, found, err := a.actorOf(ctx.Context())
	if err != nil {
		// The raw error goes to the logs, and nothing of it to the answer.
		_ = huma.WriteErr(a.API, ctx, http.StatusInternalServerError, "find who acts", err)
		return
	}
	if !found {
		_ = huma.WriteErr(a.API, ctx, http.StatusUnauthorized, NotLoggedIn.Error())
		return
	}
	next(huma.WithValue(ctx, actorKey{}, actor))
}

func (a *API) actorOf(ctx context.Context) (kernel.Actor, bool, error) {
	if a.actors == nil {
		return kernel.Actor{}, false, errNoActors
	}
	return a.actors.ActorOf(ctx)
}
