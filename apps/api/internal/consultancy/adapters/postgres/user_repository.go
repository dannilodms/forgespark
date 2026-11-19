package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// UserRepository keeps the users in the users table, and the roles each one
// has in user_roles.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository returns the repository on this pool.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of users.
func (r *UserRepository) NextIdentifier(ctx context.Context) (domain.UserID, error) {
	value, err := r.queries(ctx).NextUserID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of a user: %w", err)
	}
	return domain.NewUserID(value)
}

// EmailInUse tells whether a user, deactivated or not, has this e-mail, and
// locks nothing.
func (r *UserRepository) EmailInUse(ctx context.Context, email kernel.Email) (bool, error) {
	inUse, err := r.queries(ctx).EmailInUse(ctx, email.String())
	if err != nil {
		return false, fmt.Errorf("look for the e-mail among the users: %w", err)
	}
	return inUse, nil
}

// Save writes the user whole, and gives them the roles they have. The error
// names the step that failed and never a value of the user: PostgreSQL keeps
// the row it refused in a detail that the error does not print.
func (r *UserRepository) Save(ctx context.Context, user *domain.User) error {
	details := user.Details()
	telephone, hasTelephone := details.Telephone()
	photo, hasPhoto := details.PhotoFileID()
	err := r.queries(ctx).SaveUser(ctx, SaveUserParams{
		ID:                int64(user.ID()),
		Name:              details.Name(),
		Email:             details.Email().String(),
		Country:           details.Country().String(),
		NationalID:        details.NationalID(),
		Telephone:         textIf(telephone.String(), hasTelephone),
		PreferredLanguage: details.PreferredLanguage().String(),
		PhotoFileID:       fileIf(photo, hasPhoto),
		Deactivated:       user.Deactivated(),
	})
	if err != nil {
		return fmt.Errorf("save the user: %w", err)
	}
	roles := make([]int64, 0, len(user.Roles()))
	for _, role := range user.Roles() {
		roles = append(roles, int64(role))
	}
	err = r.queries(ctx).GiveUserRoles(ctx, GiveUserRolesParams{UserID: int64(user.ID()), RoleIds: roles})
	if err != nil {
		return fmt.Errorf("give the user their roles: %w", err)
	}
	return nil
}
