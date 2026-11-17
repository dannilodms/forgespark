package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// RoleRepository keeps the roles in the roles table.
type RoleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository returns the repository on this pool.
func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of roles.
func (r *RoleRepository) NextIdentifier(ctx context.Context) (domain.RoleID, error) {
	value, err := r.queries(ctx).NextRoleID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of a role: %w", err)
	}
	return domain.NewRoleID(value)
}

// Save writes the role whole: a new row, or every column of the row it
// already has.
func (r *RoleRepository) Save(ctx context.Context, role *domain.Role) error {
	err := r.queries(ctx).SaveRole(ctx, SaveRoleParams{ID: int64(role.ID()), Key: string(role.Key())})
	if err != nil {
		return fmt.Errorf("save the role: %w", err)
	}
	return nil
}
