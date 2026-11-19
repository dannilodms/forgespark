package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
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
	assert.Equal(t, "Ana Souza", found.Name)
	assert.Equal(t, "ana.souza@fabrica.example", found.Email.String())
	assert.Equal(t, "pt-BR", found.PreferredLanguage.String())
	assert.Equal(t, []domain.RoleKey{domain.AdministratorRoleKey}, found.Roles)
}

// activeUserOf asks the reader for the user with this identifier.
func (s stored) activeUserOf(t *testing.T, id domain.UserID) (application.ActiveUser, bool) {
	t.Helper()
	found, ok, err := s.activeUsers.ActiveUserOf(context.Background(), id)
	require.NoError(t, err)
	return found, ok
}

// The session of a browser keeps the identifier of its user, and reads the
// user with their roles on every request.
func TestAUserWhoCanLogInIsFoundByTheirIdentifierWithTheKeysOfTheirRoles(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, developer := temporary.createRoles(t)
	ana, err := temporary.registerUser(t, administratorAna, developer.ID(), administrator.ID())
	require.NoError(t, err)
	bruno := administratorAna
	bruno.Name = "Bruno Lima"
	bruno.Email = "bruno.lima@fabrica.example"
	bruno.PreferredLanguage = "es"
	developerBruno, err := temporary.registerUser(t, bruno, developer.ID())
	require.NoError(t, err)

	foundAna, anaFound := temporary.activeUserOf(t, ana.ID())
	foundBruno, brunoFound := temporary.activeUserOf(t, developerBruno.ID())

	require.True(t, anaFound)
	assert.Equal(t, application.ActiveUser{
		ID:                ana.ID(),
		Name:              "Ana Souza",
		Email:             ana.Details().Email(),
		PreferredLanguage: ana.Details().PreferredLanguage(),
		PhotoFileID:       0,
		Roles:             []domain.RoleKey{domain.AdministratorRoleKey, domain.DeveloperRoleKey},
		Scopes:            nil,
	}, foundAna)
	require.True(t, brunoFound)
	assert.Equal(t, "Bruno Lima", foundBruno.Name)
	assert.Equal(t, "es", foundBruno.PreferredLanguage.String())
	assert.Equal(t, []domain.RoleKey{domain.DeveloperRoleKey}, foundBruno.Roles)
}

// The app shows the photo of the user logged in, which the files area keeps.
func TestAUserWhoCanLogInIsFoundWithTheFileOfTheirPhoto(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	typed := administratorAna
	typed.PhotoFileID = temporary.aFile(t, "user_photo")
	user, err := temporary.registerUser(t, typed, administrator.ID())
	require.NoError(t, err)

	byIdentifier, _ := temporary.activeUserOf(t, user.ID())
	byEmail, _ := temporary.activeUserWith(t, typed.Email)

	assert.Equal(t, typed.PhotoFileID, byIdentifier.PhotoFileID)
	assert.Equal(t, typed.PhotoFileID, byEmail.PhotoFileID)
}

func TestNoUserIsFoundForAnIdentifierNobodyHasOrForADeactivatedUser(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	user, err := temporary.registerUser(t, administratorAna, administrator.ID())
	require.NoError(t, err)

	_, unknown := temporary.activeUserOf(t, user.ID()+1)
	user.Deactivate()
	require.NoError(t, temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		return temporary.users.Save(ctx, user)
	}))
	_, deactivated := temporary.activeUserOf(t, user.ID())

	assert.False(t, unknown, "a user was found for an identifier nobody has")
	assert.False(t, deactivated, "a deactivated user was found")
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
