package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// activeUserWith asks the reader for the user who logs in with the e-mail, as
// typed.
func (s stored) activeUserWith(t *testing.T, typed string) (application.ActiveUser, bool) {
	t.Helper()
	email, err := kernel.NewEmail(typed)
	require.NoError(t, err)
	found, ok, err := s.activeUsers.ActiveUserWith(context.Background(), email)
	require.NoError(t, err)
	return found, ok
}

func TestAUserWhoCanLogInIsFoundByTheirEmailInAnyCase(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	user, err := temporary.registerUser(t, administratorAna, administrator.ID())
	require.NoError(t, err)

	found, ok := temporary.activeUserWith(t, "ANA.Souza@fabrica.EXAMPLE")

	require.True(t, ok)
	assert.Equal(t, user.ID(), found.ID)
	assert.Equal(t, "ana.souza@fabrica.example", found.Email.String())
	assert.Equal(t, "pt-BR", found.PreferredLanguage.String())
}

func TestNoUserIsFoundForAnEmailNobodyHasOrForADeactivatedUser(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	user, err := temporary.registerUser(t, administratorAna, administrator.ID())
	require.NoError(t, err)

	_, unknown := temporary.activeUserWith(t, "bruno.lima@fabrica.example")
	user.Deactivate()
	require.NoError(t, temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		return temporary.users.Save(ctx, user)
	}))
	_, deactivated := temporary.activeUserWith(t, administratorAna.Email)

	assert.False(t, unknown, "a user was found for an e-mail nobody has")
	assert.False(t, deactivated, "a deactivated user was found")
}
