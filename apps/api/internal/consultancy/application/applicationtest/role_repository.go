package applicationtest

import (
	"context"
	"slices"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
)

// RoleRepository keeps the roles in memory. It stores a copy of what it is
// given, as the database does.
type RoleRepository struct {
	lastIdentifier int64
	roles          []domain.Role
	failure        error
}

// NewRoleRepository returns a repository with no role.
func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *RoleRepository) FailWith(failure error) {
	r.failure = failure
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *RoleRepository) NextIdentifier(context.Context) (domain.RoleID, error) {
	r.lastIdentifier++
	return domain.NewRoleID(r.lastIdentifier)
}

// FindByKey returns a copy of the role with this key.
func (r *RoleRepository) FindByKey(_ context.Context, key domain.RoleKey) (*domain.Role, bool, error) {
	position := slices.IndexFunc(r.roles, func(stored domain.Role) bool { return stored.Key() == key })
	if position < 0 {
		return nil, false, nil
	}
	found := r.roles[position]
	return &found, true, nil
}

// Save stores the role, in place of the one with the same identifier when
// there is one.
func (r *RoleRepository) Save(_ context.Context, role *domain.Role) error {
	if r.failure != nil {
		return r.failure
	}
	position := slices.IndexFunc(r.roles, func(stored domain.Role) bool { return stored.ID() == role.ID() })
	if position < 0 {
		r.roles = append(r.roles, *role)
		return nil
	}
	r.roles[position] = *role
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *RoleRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.roles)
	return func() { r.roles = saved }
}

// All returns every role stored, in the order they were first saved.
func (r *RoleRepository) All() []domain.Role {
	return slices.Clone(r.roles)
}
