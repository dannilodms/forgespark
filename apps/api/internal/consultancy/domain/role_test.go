package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// A user acts with the scopes of all their roles (ADR-0026): the
// administrator role grants every scope, one added later included, and the
// developer role none yet.
func TestAUserActsWithTheScopesOfTheirRoles(t *testing.T) {
	cases := map[string]struct {
		roles  []domain.RoleKey
		scopes []kernel.Scope
	}{
		"an administrator":                    {roles: []domain.RoleKey{domain.AdministratorRoleKey}, scopes: kernel.EveryScope()},
		"a developer":                         {roles: []domain.RoleKey{domain.DeveloperRoleKey}, scopes: []kernel.Scope{}},
		"an administrator who is a developer": {roles: []domain.RoleKey{domain.DeveloperRoleKey, domain.AdministratorRoleKey}, scopes: kernel.EveryScope()},
		"a user with no role":                 {roles: nil, scopes: []kernel.Scope{}},
	}
	for name, each := range cases {
		t.Run(name, func(t *testing.T) {
			assert.ElementsMatch(t, each.scopes, domain.ScopesOf(each.roles))
			assert.NotNil(t, domain.ScopesOf(each.roles))
		})
	}
}
