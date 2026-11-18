package application

import (
	"context"
	"errors"
	"time"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// RegisterAdministrator registers a new administrator with a setup code
// issued after the setup was concluded: the way back in when no
// administrator can log in. It changes no user that exists.
type RegisterAdministrator struct {
	transactions  Transactions
	setupCodes    SetupCodeRepository
	consultancies ConsultancyRepository
	users         UserRepository
	roles         RoleRepository
	confirmUpload *filesapplication.ConfirmUpload
	recordEntry   *auditapplication.RecordEntry
	clock         Clock
}

// NewRegisterAdministrator builds the use case.
func NewRegisterAdministrator(
	transactions Transactions,
	setupCodes SetupCodeRepository,
	consultancies ConsultancyRepository,
	users UserRepository,
	roles RoleRepository,
	confirmUpload *filesapplication.ConfirmUpload,
	recordEntry *auditapplication.RecordEntry,
	clock Clock,
) *RegisterAdministrator {
	return &RegisterAdministrator{
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

// RegisterAdministratorInput is what was typed: the setup code and the new
// administrator. The preferred language of the administrator is the language
// the setup screen was in.
type RegisterAdministratorInput struct {
	Code          string
	Administrator domain.TypedUser
}

// Execute registers the administrator. What was typed is checked first, and a
// kernel.InvalidFieldsError names each field that breaks a rule as the input
// is structured: `administrator.email`. With a setup code that works on a
// setup that is concluded, the photo chosen is confirmed, as in the first
// setup, and named on its field when the files area refuses it. Then, in one transaction, it locks the
// setup code, checks that it still works and that the setup is concluded,
// refuses an e-mail that belongs to a user, creates the user with the
// administrator role, ends the setup code and records the registration in
// the audit log.
func (u *RegisterAdministrator) Execute(ctx context.Context, input RegisterAdministratorInput) error {
	var invalid kernel.InvalidFieldsError
	details, err := domain.NewUserDetails(input.Administrator)
	invalid.Include("administrator", err)
	if err := invalid.Err(); err != nil {
		return err
	}
	setupCode, err := findWorkingSetupCode(ctx, u.setupCodes, input.Code, u.clock.Now())
	if err != nil {
		return err
	}
	if err := opensAtThisMoment(ctx, u.consultancies, domain.OpensNewAdministrator); err != nil {
		return err
	}
	if err := u.confirmPhoto(ctx, details); err != nil {
		return err
	}
	return u.transactions.Run(ctx, func(ctx context.Context) error {
		return u.register(ctx, setupCode.Hash(), details)
	})
}

// confirmPhoto confirms the photo chosen, outside the transaction of the
// registration: it reads the storage.
func (u *RegisterAdministrator) confirmPhoto(ctx context.Context, details domain.UserDetails) error {
	photo, hasPhoto := details.PhotoFileID()
	refused, err := confirmImage(ctx, u.confirmUpload, photo, hasPhoto, UserPhoto)
	if err != nil {
		return err
	}
	var invalid kernel.InvalidFieldsError
	invalid.Check("administrator.photo_file_id", refused)
	return invalid.Err()
}

func (u *RegisterAdministrator) register(ctx context.Context, hash domain.SetupCodeHash, details domain.UserDetails) error {
	now := u.clock.Now()

	setupCode, err := u.workingSetupCode(ctx, hash, now)
	if err != nil {
		return err
	}
	administrator, err := u.registerUser(ctx, details)
	if err != nil {
		return err
	}
	setupCode.End(now)
	if err := u.setupCodes.Save(ctx, setupCode); err != nil {
		return err
	}
	return u.recordEntry.Execute(ctx, auditapplication.RecordEntryInput{
		OccurredAt: now,
		Operation:  auditapplication.AdministratorRegistered,
		Actor:      auditapplication.SetupCodeUsed(int64(setupCode.ID())),
		Subject:    auditapplication.User(int64(administrator.ID())),
	})
}

// workingSetupCode loads the setup code typed, locked, and checks that it
// opens the registration of a new administrator at this moment: that it
// works and that the consultancy exists.
func (u *RegisterAdministrator) workingSetupCode(ctx context.Context, hash domain.SetupCodeHash, now time.Time) (*domain.SetupCode, error) {
	setupCode, err := lockWorkingSetupCode(ctx, u.setupCodes, hash, now)
	if err != nil {
		return nil, err
	}
	if err := opensAtThisMoment(ctx, u.consultancies, domain.OpensNewAdministrator); err != nil {
		return nil, err
	}
	return setupCode, nil
}

// registerUser creates the user with the administrator role, unless the
// e-mail belongs to a user already.
func (u *RegisterAdministrator) registerUser(ctx context.Context, details domain.UserDetails) (*domain.User, error) {
	inUse, err := u.users.EmailInUse(ctx, details.Email())
	if err != nil {
		return nil, err
	}
	if inUse {
		return nil, domain.EmailInUseError{}
	}
	administratorRole, found, err := u.roles.FindByKey(ctx, domain.AdministratorRoleKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("the setup is concluded and the administrator role does not exist")
	}
	id, err := u.users.NextIdentifier(ctx)
	if err != nil {
		return nil, err
	}
	administrator, err := domain.RegisterUser(id, details, administratorRole.ID())
	if err != nil {
		return nil, err
	}
	if err := u.users.Save(ctx, administrator); err != nil {
		return nil, err
	}
	return administrator, nil
}
