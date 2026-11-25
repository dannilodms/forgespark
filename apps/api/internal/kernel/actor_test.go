package kernel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

func TestAnActorIsTheUserWithTheScopesItWasGiven(t *testing.T) {
	scopes := []kernel.Scope{"consultancy:read", "users:write"}

	actor := kernel.NewActor(7, scopes)
	scopes[0] = "credentials:read"

	assert.Equal(t, int64(7), actor.User())
	assert.Equal(t, []kernel.Scope{"consultancy:read", "users:write"}, actor.Scopes(), "the actor follows a change of the list it was given")
	assert.True(t, actor.Holds("users:write"))
	assert.False(t, actor.Holds("credentials:read"), "the actor holds a scope it was not given")
}

func TestAnActorWithNoScopeHoldsNone(t *testing.T) {
	actor := kernel.NewActor(8, nil)

	assert.Empty(t, actor.Scopes())
	assert.False(t, actor.Holds("consultancy:read"))
}

// No use case checks a scope yet, so the system knows none: a scope arrives
// with the use case that checks it (ADR-0026).
func TestTheSystemKnowsNoScopeYet(t *testing.T) {
	assert.Empty(t, kernel.EveryScope())
}
