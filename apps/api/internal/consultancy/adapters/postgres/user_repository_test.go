package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

var administratorAna = domain.TypedUser{
	Name:              "Ana Souza",
	Email:             "Ana.Souza@Fabrica.example",
	Country:           "BR",
	NationalID:        "123.456.789-09",
	Telephone:         "+55 11 98888 7777",
	PreferredLanguage: "pt-BR",
}

// createRoles saves the administrator role and the developer role, in a
// transaction of their own, and returns them.
func (s stored) createRoles(t *testing.T) (administrator, developer *domain.Role) {
	t.Helper()
	err := s.transactions.Run(context.Background(), func(ctx context.Context) error {
		administratorID, err := s.roles.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		developerID, err := s.roles.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		administrator = domain.CreateAdministratorRole(administratorID)
		developer = domain.CreateDeveloperRole(developerID)
		return errors.Join(s.roles.Save(ctx, administrator), s.roles.Save(ctx, developer))
	})
	require.NoError(t, err)
	return administrator, developer
}

// registerUser saves a user with these roles, in a transaction of its own.
func (s stored) registerUser(t *testing.T, typed domain.TypedUser, roles ...domain.RoleID) (*domain.User, error) {
	t.Helper()
	details, err := domain.NewUserDetails(typed)
	require.NoError(t, err)
	var user *domain.User
	err = s.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := s.users.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		user, err = domain.RegisterUser(id, details, roles...)
		if err != nil {
			return err
		}
		return s.users.Save(ctx, user)
	})
	return user, err
}

func TestTheTwoRolesOfTheSetupAreSavedWithTheirKeys(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)

	administrator, developer := temporary.createRoles(t)

	rows, err := temporary.pool.Query(context.Background(), "SELECT id, key FROM roles ORDER BY id")
	require.NoError(t, err)
	type role struct {
		ID  int64
		Key string
	}
	roles, err := pgx.CollectRows(rows, pgx.RowToStructByPos[role])
	require.NoError(t, err)
	assert.Equal(t, []role{
		{ID: int64(administrator.ID()), Key: "administrator"},
		{ID: int64(developer.ID()), Key: "developer"},
	}, roles)
}

func TestTheDatabaseRefusesASecondRoleWithTheSameKey(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	temporary.createRoles(t)
	const uniqueViolation = "23505"

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := temporary.roles.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		return temporary.roles.Save(ctx, domain.CreateAdministratorRole(id))
	})

	assert.Equal(t, uniqueViolation, sqlState(t, err))
}

func TestAUserIsSavedAsTypedWithTheirEmailInLowercaseAndTheirRoles(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)

	user, err := temporary.registerUser(t, administratorAna, administrator.ID())

	require.NoError(t, err)
	row := temporary.row(t, "SELECT * FROM users")
	assert.Equal(t, map[string]any{
		"id":                 int64(user.ID()),
		"name":               "Ana Souza",
		"email":              "ana.souza@fabrica.example",
		"country":            "BR",
		"national_id":        "123.456.789-09",
		"telephone":          "+5511988887777",
		"preferred_language": "pt-BR",
		"photo_file_id":      nil,
		"deactivated":        false,
	}, row)
	assert.Equal(t, map[string]any{"user_id": int64(user.ID()), "role_id": int64(administrator.ID())},
		temporary.row(t, "SELECT user_id, role_id FROM user_roles"))
}

func TestAUserIsSavedWithTheFileOfTheirPhoto(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	typed := administratorAna
	typed.PhotoFileID = temporary.aFile(t, "user_photo")

	_, err := temporary.registerUser(t, typed, administrator.ID())

	require.NoError(t, err)
	assert.Equal(t, int64(typed.PhotoFileID), temporary.row(t, "SELECT photo_file_id FROM users")["photo_file_id"])
}

func TestTheDatabaseRefusesAPhotoThatIsNoFile(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	typed := administratorAna
	typed.PhotoFileID = 987654
	const foreignKeyViolation = "23503"

	_, err := temporary.registerUser(t, typed, administrator.ID())

	assert.Equal(t, foreignKeyViolation, sqlState(t, err))
}

func TestAUserWhoIsSavedAgainKeepsOneRowForEachRole(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	user, err := temporary.registerUser(t, administratorAna, administrator.ID())
	require.NoError(t, err)

	err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		return temporary.users.Save(ctx, user)
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), temporary.row(t, "SELECT count(*) FROM user_roles")["count"])
}

func TestTheDatabaseRefusesASecondUserWithTheSameEmail(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	_, err := temporary.registerUser(t, administratorAna, administrator.ID())
	require.NoError(t, err)
	const uniqueViolation = "23505"
	another := administratorAna
	another.Name = "Ana Maria Souza"
	another.Email = "ANA.SOUZA@fabrica.example"

	_, err = temporary.registerUser(t, another, administrator.ID())

	assert.Equal(t, uniqueViolation, sqlState(t, err))
	assert.NotContains(t, err.Error(), "ana.souza", "the error carries the e-mail")
	assert.Equal(t, int64(1), temporary.row(t, "SELECT count(*) FROM users")["count"])
}

// emailInUse asks the repository whether a user has the e-mail, as typed.
func (s stored) emailInUse(t *testing.T, typed string) bool {
	t.Helper()
	email, err := kernel.NewEmail(typed)
	require.NoError(t, err)
	inUse, err := s.users.EmailInUse(context.Background(), email)
	require.NoError(t, err)
	return inUse
}

func TestAnEmailIsInUseOnceAUserHasIt(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	require.False(t, temporary.emailInUse(t, administratorAna.Email), "an e-mail no user has")

	_, err := temporary.registerUser(t, administratorAna, administrator.ID())

	require.NoError(t, err)
	assert.True(t, temporary.emailInUse(t, "ANA.SOUZA@fabrica.example"))
	assert.False(t, temporary.emailInUse(t, "bruno.lima@fabrica.example"))
}

func TestTheEmailOfADeactivatedUserIsInUse(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	administrator, _ := temporary.createRoles(t)
	user, err := temporary.registerUser(t, administratorAna, administrator.ID())
	require.NoError(t, err)
	user.Deactivate()
	require.NoError(t, temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		return temporary.users.Save(ctx, user)
	}))
	require.Equal(t, true, temporary.row(t, "SELECT deactivated FROM users")["deactivated"])

	assert.True(t, temporary.emailInUse(t, administratorAna.Email))
}

func TestTheRoleOfTheSetupIsFoundByItsKey(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	_, notFound, err := temporary.roles.FindByKey(context.Background(), domain.AdministratorRoleKey)
	require.NoError(t, err)
	require.False(t, notFound, "a role was found before the setup created one")
	administrator, developer := temporary.createRoles(t)

	for _, created := range []*domain.Role{administrator, developer} {
		found, ok, err := temporary.roles.FindByKey(context.Background(), created.Key())
		require.NoError(t, err)
		require.True(t, ok, "no role %s", created.Key())
		assert.Equal(t, created, found)
	}
}
