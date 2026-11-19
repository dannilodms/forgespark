package kernel

import "slices"

// Scope is what a role or an access token allows: read or write in one area
// of the system (ADR-0026). The list of scopes is fixed in the code, and a
// scope arrives with the use case that checks it.
type Scope string

// everyScope is the list of scopes the system knows. None exists yet: the
// first one arrives with the first use case that checks one, in the effort of
// the consultancy settings and the users.
var everyScope = []Scope{}

// EveryScope returns every scope the system knows, which the administrator
// role grants.
func EveryScope() []Scope {
	return slices.Clone(everyScope)
}

// Actor is who acts in a use case: a user, by identifier, and the scopes they
// act with. The inbound adapter that authenticated the caller builds it, from
// the session of a browser and the roles of its user, and passes it to the use
// case, which checks the scopes it needs (Inbound adapters in
// CODING_STANDARDS.md). An access token, later, will be one more part of it.
type Actor struct {
	user   int64
	scopes []Scope
}

// NewActor is the user with this identifier, acting with these scopes.
func NewActor(user int64, scopes []Scope) Actor {
	return Actor{user: user, scopes: slices.Clone(scopes)}
}

// User returns the identifier of the user who acts.
func (a Actor) User() int64 {
	return a.user
}

// Scopes returns the scopes the actor acts with.
func (a Actor) Scopes() []Scope {
	return slices.Clone(a.scopes)
}

// Holds tells whether the actor may act with the scope.
func (a Actor) Holds(scope Scope) bool {
	return slices.Contains(a.scopes, scope)
}
