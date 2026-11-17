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

// SetupCodeIssued is recorded when whoever installs the system issues a
// setup code. Its actor is Installer and its subject is the SetupCode issued.
const SetupCodeIssued = domain.SetupCodeIssued

// Installer is whoever installs the system, who has no identifier.
func Installer() Actor {
	return domain.Installer()
}

// SetupCode is the setup code with this identifier, as what was acted on.
func SetupCode(identifier int64) Subject {
	return domain.SetupCode(identifier)
}
