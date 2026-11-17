package application

import (
	"context"
	"time"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// ConcludeSetup concludes the first setup: it registers the consultancy, its
// first administrator and the two first roles together, and ends the setup
// code that allowed it. Nothing of them is saved before, so a setup left
// halfway leaves nothing behind.
type ConcludeSetup struct {
	transactions  Transactions
	setupCodes    SetupCodeRepository
	consultancies ConsultancyRepository
	users         UserRepository
	roles         RoleRepository
	confirmUpload *filesapplication.ConfirmUpload
	recordEntry   *auditapplication.RecordEntry
	clock         Clock
}

// NewConcludeSetup builds the use case.
func NewConcludeSetup(
	transactions Transactions,
	setupCodes SetupCodeRepository,
	consultancies ConsultancyRepository,
	users UserRepository,
	roles RoleRepository,
	confirmUpload *filesapplication.ConfirmUpload,
	recordEntry *auditapplication.RecordEntry,
	clock Clock,
) *ConcludeSetup {
	return &ConcludeSetup{
		transactions:  transactions,
		setupCodes:    setupCodes,
		consultancies: consultancies,
		users:         users,
		roles:         roles,
		confirmUpload: confirmUpload,
		recordEntry:   recordEntry,
		clock:         clock,
	}
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
// setup code that works, it then confirms the images chosen, which may have
// been sent under an earlier code, and names the one the files area refuses.
// Then, in one transaction, it locks the setup code, checks that it still
// works and that no consultancy exists, creates the administrator role and the
// developer role, the consultancy and the administrator with the
// administrator role, ends the setup code and records the conclusion in the
// audit log.
func (u *ConcludeSetup) Execute(ctx context.Context, input ConcludeSetupInput) error {
	var invalid kernel.InvalidFieldsError
	consultancy, err := domain.NewConsultancyDetails(input.Consultancy)
	invalid.Include("consultancy", err)
	administrator, err := domain.NewUserDetails(input.Administrator)
	invalid.Include("administrator", err)
	if err := invalid.Err(); err != nil {
		return err
	}
	setupCode, err := findWorkingSetupCode(ctx, u.setupCodes, input.Code, u.clock.Now())
	if err != nil {
		return err
	}
	if err := u.confirmImages(ctx, consultancy, administrator); err != nil {
		return err
	}

	return u.transactions.Run(ctx, func(ctx context.Context) error {
		return u.conclude(ctx, setupCode.Hash(), consultancy, administrator)
	})
}

// confirmImages confirms the logo and the photo chosen, outside the
// transaction of the setup: each reads the storage. A file the files area
// refuses is named as the field that holds it.
func (u *ConcludeSetup) confirmImages(ctx context.Context, consultancy domain.ConsultancyDetails, administrator domain.UserDetails) error {
	logo, hasLogo := consultancy.LogoFileID()
	logoRefused, err := confirmImage(ctx, u.confirmUpload, logo, hasLogo, ConsultancyLogo)
	if err != nil {
		return err
	}
	photo, hasPhoto := administrator.PhotoFileID()
	photoRefused, err := confirmImage(ctx, u.confirmUpload, photo, hasPhoto, UserPhoto)
	if err != nil {
		return err
	}
	var invalid kernel.InvalidFieldsError
	invalid.Check("consultancy.logo_file_id", logoRefused)
	invalid.Check("administrator.photo_file_id", photoRefused)
	return invalid.Err()
}

func (u *ConcludeSetup) conclude(
	ctx context.Context,
	hash domain.SetupCodeHash,
	consultancyDetails domain.ConsultancyDetails,
	administratorDetails domain.UserDetails,
) error {
	now := u.clock.Now()

	setupCode, err := u.workingSetupCode(ctx, hash, now)
	if err != nil {
		return err
	}
	administratorRole, err := u.createRoles(ctx)
	if err != nil {
		return err
	}
	if err := u.registerConsultancy(ctx, consultancyDetails); err != nil {
		return err
	}
	administrator, err := u.registerAdministrator(ctx, administratorDetails, administratorRole)
	if err != nil {
		return err
	}
	setupCode.End(now)
	if err := u.setupCodes.Save(ctx, setupCode); err != nil {
		return err
	}
	return u.recordEntry.Execute(ctx, auditapplication.RecordEntryInput{
		OccurredAt: now,
		Operation:  auditapplication.SetupConcluded,
		Actor:      auditapplication.SetupCodeUsed(int64(setupCode.ID())),
		Subject:    auditapplication.User(int64(administrator.ID())),
	})
}

// workingSetupCode loads the setup code typed, locked, and checks that it
// opens the first setup at this moment: that it works and that no
// consultancy exists yet.
func (u *ConcludeSetup) workingSetupCode(ctx context.Context, hash domain.SetupCodeHash, now time.Time) (*domain.SetupCode, error) {
	setupCode, err := lockWorkingSetupCode(ctx, u.setupCodes, hash, now)
	if err != nil {
		return nil, err
	}
	concluded, err := u.consultancies.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if concluded {
		return nil, domain.SetupAlreadyConcludedError{}
	}
	return setupCode, nil
}

// createRoles creates the administrator role and the developer role, and
// returns the first.
func (u *ConcludeSetup) createRoles(ctx context.Context) (domain.RoleID, error) {
	administratorID, err := u.roles.NextIdentifier(ctx)
	if err != nil {
		return 0, err
	}
	if err := u.roles.Save(ctx, domain.CreateAdministratorRole(administratorID)); err != nil {
		return 0, err
	}
	developerID, err := u.roles.NextIdentifier(ctx)
	if err != nil {
		return 0, err
	}
	if err := u.roles.Save(ctx, domain.CreateDeveloperRole(developerID)); err != nil {
		return 0, err
	}
	return administratorID, nil
}

func (u *ConcludeSetup) registerConsultancy(ctx context.Context, details domain.ConsultancyDetails) error {
	id, err := u.consultancies.NextIdentifier(ctx)
	if err != nil {
		return err
	}
	return u.consultancies.Save(ctx, domain.RegisterConsultancy(id, details))
}

func (u *ConcludeSetup) registerAdministrator(ctx context.Context, details domain.UserDetails, administratorRole domain.RoleID) (*domain.User, error) {
	id, err := u.users.NextIdentifier(ctx)
	if err != nil {
		return nil, err
	}
	administrator, err := domain.RegisterUser(id, details, administratorRole)
	if err != nil {
		return nil, err
	}
	if err := u.users.Save(ctx, administrator); err != nil {
		return nil, err
	}
	return administrator, nil
}
