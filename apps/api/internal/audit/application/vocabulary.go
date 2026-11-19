package application

import "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/domain"

// Another domain area imports this package and no other of the audit log, so
// the words it needs to describe an operation are named here.
type (
	// Operation is a sensitive operation the audit log records.
	Operation = domain.Operation
	// Actor is who performed an operation.
	Actor = domain.Actor
	// Subject is what an operation acted on.
	Subject = domain.Subject
	// Entry is one record of the audit log.
	Entry = domain.Entry
)

const (
	// SetupCodeIssued is recorded when whoever installs the system issues a
	// setup code. Its actor is Installer and its subject is the SetupCode
	// issued.
	SetupCodeIssued = domain.SetupCodeIssued
	// SetupConcluded is recorded when the setup is concluded. Its actor is
	// the SetupCodeUsed and its subject is the User created, the first
	// administrator.
	SetupConcluded = domain.SetupConcluded
	// AdministratorRegistered is recorded when a setup code issued after the
	// setup registers a new administrator. Its actor is the SetupCodeUsed and
	// its subject is the User created.
	AdministratorRegistered = domain.AdministratorRegistered
	// LoggedIn is recorded when a user logs in. Its actor is the ActingUser
	// who logged in, its subject is the same User, and it has a WayIn.
	LoggedIn = domain.LoggedIn
)

// Installer is whoever installs the system, who has no identifier.
func Installer() Actor {
	return domain.Installer()
}

// SetupCodeUsed is the setup code with this identifier, as who acted.
func SetupCodeUsed(identifier int64) Actor {
	return domain.SetupCodeUsed(identifier)
}

// ActingUser is the user with this identifier, as who acted.
func ActingUser(identifier int64) Actor {
	return domain.ActingUser(identifier)
}

// SetupCode is the setup code with this identifier, as what was acted on.
func SetupCode(identifier int64) Subject {
	return domain.SetupCode(identifier)
}

// User is the user with this identifier, as what was acted on.
func User(identifier int64) Subject {
	return domain.User(identifier)
}
