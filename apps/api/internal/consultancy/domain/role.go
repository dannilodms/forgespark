package domain

import (
	"errors"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// RoleID identifies a role.
type RoleID int64

// NewRoleID reads an identifier that came from the sequence of roles.
func NewRoleID(value int64) (RoleID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of a role is a positive number")
	}
	return RoleID(value), nil
}

// RoleKey is the fixed key of a role the setup creates, which the frontend
// translates. A role a person creates will have a name instead, with the
// screen that creates roles.
type RoleKey string

const (
	// AdministratorRoleKey is the administrator role, which grants every
	// scope and cannot be changed or removed. It stores no scope.
	AdministratorRoleKey RoleKey = "administrator"
	// DeveloperRoleKey is the developer role, an ordinary role the setup
	// creates with no scope.
	DeveloperRoleKey RoleKey = "developer"
)

// Scopes returns the scopes the role with this key grants: every scope the
// system knows for the administrator role, by rule, and none for the
// developer role, which gets scopes as the areas of work are built.
func (k RoleKey) Scopes() []kernel.Scope {
	if k == AdministratorRoleKey {
		return kernel.EveryScope()
	}
	return []kernel.Scope{}
}

// ScopesOf returns the scopes a user with these roles acts with: those of
// every role, each once (ADR-0026). They are read on every request, so a
// change of role shows on the next one.
func ScopesOf(roles []RoleKey) []kernel.Scope {
	scopes := []kernel.Scope{}
	for _, role := range roles {
		scopes = append(scopes, role.Scopes()...)
	}
	slices.Sort(scopes)
	return slices.Compact(scopes)
}

// Role is a named set of scopes given to users. No scope exists yet, so a
// role holds its key only.
type Role struct {
	id  RoleID
	key RoleKey
}

// CreateAdministratorRole creates the administrator role.
func CreateAdministratorRole(id RoleID) *Role {
	return &Role{id: id, key: AdministratorRoleKey}
}

// CreateDeveloperRole creates the developer role, with no scope.
func CreateDeveloperRole(id RoleID) *Role {
	return &Role{id: id, key: DeveloperRoleKey}
}

// RestoreRole rebuilds a role from what a repository stored.
func RestoreRole(id RoleID, key RoleKey) *Role {
	return &Role{id: id, key: key}
}

// ID returns the identifier of the role.
func (r *Role) ID() RoleID {
	return r.id
}

// Key returns the fixed key of the role.
func (r *Role) Key() RoleKey {
	return r.key
}
