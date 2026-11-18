package application

import (
	"context"
	"errors"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// RegisterAdministrator registers a new administrator with a setup code
// issued after the setup was concluded: the way back in when no
// administrator can log in. It changes no user that exists.
type RegisterAdministrator struct {
	setup savingWithSetupCode
}

// NewRegisterAdministrator builds the use case.
func NewRegisterAdministrator(dependencies SetupDependencies) *RegisterAdministrator {
	return &RegisterAdministrator{setup: savingWithSetupCode{
		SetupDependencies: dependencies,
		opens:             domain.OpensNewAdministrator,
		operation:         auditapplication.AdministratorRegistered,
	}}
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
// setup, and named on its field when the files area refuses it. Then, in one
// transaction, it locks the setup code, checks that it still works and that
// the setup is concluded, refuses an e-mail that belongs to a user, creates
// the user with the administrator role, ends the setup code and records the
// registration in the audit log.
func (u *RegisterAdministrator) Execute(ctx context.Context, input RegisterAdministratorInput) error {
	var invalid kernel.InvalidFieldsError
	details, err := domain.NewUserDetails(input.Administrator)
	invalid.Include("administrator", err)
	if err := invalid.Err(); err != nil {
		return err
	}

	photo, hasPhoto := details.PhotoFileID()
	images := []chosenImage{{field: "administrator.photo_file_id", id: photo, chosen: hasPhoto, purpose: UserPhoto}}
	return u.setup.save(ctx, input.Code, images, func(ctx context.Context) (*domain.User, error) {
		return u.registerUser(ctx, details)
	})
}

// registerUser creates the user with the administrator role, unless the
// e-mail belongs to a user already.
func (u *RegisterAdministrator) registerUser(ctx context.Context, details domain.UserDetails) (*domain.User, error) {
	inUse, err := u.setup.Users.EmailInUse(ctx, details.Email())
	if err != nil {
		return nil, err
	}
	if inUse {
		return nil, domain.EmailInUseError{}
	}
	administratorRole, found, err := u.setup.Roles.FindByKey(ctx, domain.AdministratorRoleKey)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("the setup is concluded and the administrator role does not exist")
	}
	return u.setup.registerUser(ctx, details, administratorRole.ID())
}
