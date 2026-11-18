package application

import (
	"context"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// ConcludeSetup concludes the first setup: it registers the consultancy, its
// first administrator and the two first roles together, and ends the setup
// code that allowed it. Nothing of them is saved before, so a setup left
// halfway leaves nothing behind.
type ConcludeSetup struct {
	setup savingWithSetupCode
}

// NewConcludeSetup builds the use case.
func NewConcludeSetup(dependencies SetupDependencies) *ConcludeSetup {
	return &ConcludeSetup{setup: savingWithSetupCode{
		SetupDependencies: dependencies,
		opens:             domain.OpensFirstSetup,
		operation:         auditapplication.SetupConcluded,
	}}
}

// ConcludeSetupInput is what was typed in the setup: the setup code, the
// consultancy and its first administrator, with the images chosen for them.
// The preferred language of the administrator is the language the setup
// screen was in.
type ConcludeSetupInput struct {
	Code          string
	Consultancy   domain.TypedConsultancy
	Administrator domain.TypedUser
}

// Execute validates everything that was typed first, and answers with a
// kernel.InvalidFieldsError naming each field that breaks a rule. With a
// setup code that works and opens the first setup, it then confirms the
// images chosen, which may have been sent under an earlier code, and names
// the one the files area refuses: a code that opens nothing, or opens the
// registration of a new administrator, reaches no image. Then, in one
// transaction, it locks the setup code, checks that it still works and that
// no consultancy exists, creates the administrator role and the developer
// role, the consultancy and the administrator with the administrator role,
// ends the setup code and records the conclusion in the audit log.
func (u *ConcludeSetup) Execute(ctx context.Context, input ConcludeSetupInput) error {
	var invalid kernel.InvalidFieldsError
	consultancy, err := domain.NewConsultancyDetails(input.Consultancy)
	invalid.Include("consultancy", err)
	administrator, err := domain.NewUserDetails(input.Administrator)
	invalid.Include("administrator", err)
	if err := invalid.Err(); err != nil {
		return err
	}

	logo, hasLogo := consultancy.LogoFileID()
	photo, hasPhoto := administrator.PhotoFileID()
	images := []chosenImage{
		{field: "consultancy.logo_file_id", id: logo, chosen: hasLogo, purpose: ConsultancyLogo},
		{field: "administrator.photo_file_id", id: photo, chosen: hasPhoto, purpose: UserPhoto},
	}
	return u.setup.save(ctx, input.Code, images, func(ctx context.Context) (*domain.User, error) {
		administratorRole, err := u.createRoles(ctx)
		if err != nil {
			return nil, err
		}
		if err := u.registerConsultancy(ctx, consultancy); err != nil {
			return nil, err
		}
		return u.setup.registerUser(ctx, administrator, administratorRole)
	})
}

// createRoles creates the administrator role and the developer role, and
// returns the first.
func (u *ConcludeSetup) createRoles(ctx context.Context) (domain.RoleID, error) {
	roles := u.setup.Roles
	administratorID, err := roles.NextIdentifier(ctx)
	if err != nil {
		return 0, err
	}
	if err := roles.Save(ctx, domain.CreateAdministratorRole(administratorID)); err != nil {
		return 0, err
	}
	developerID, err := roles.NextIdentifier(ctx)
	if err != nil {
		return 0, err
	}
	if err := roles.Save(ctx, domain.CreateDeveloperRole(developerID)); err != nil {
		return 0, err
	}
	return administratorID, nil
}

func (u *ConcludeSetup) registerConsultancy(ctx context.Context, details domain.ConsultancyDetails) error {
	consultancies := u.setup.Consultancies
	id, err := consultancies.NextIdentifier(ctx)
	if err != nil {
		return err
	}
	return consultancies.Save(ctx, domain.RegisterConsultancy(id, details))
}
