package application

import (
	"context"
	"errors"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// SetupDependencies are what the two use cases that register a user with a
// setup code are built with: concluding the setup and registering an
// administrator.
type SetupDependencies struct {
	Transactions  Transactions
	SetupCodes    SetupCodeRepository
	Consultancies ConsultancyRepository
	Users         UserRepository
	Roles         RoleRepository
	ConfirmUpload *filesapplication.ConfirmUpload
	RecordEntry   *auditapplication.RecordEntry
	Clock         Clock
}

// savingWithSetupCode is what concluding the setup and registering an
// administrator share: a setup code that must open what the operation is,
// the images chosen, a user registered, and the setup code ended and its
// use recorded, in one transaction.
type savingWithSetupCode struct {
	SetupDependencies
	// opens is what the setup code must open for the operation.
	opens domain.SetupCodeOpens
	// operation is what the audit log records.
	operation auditapplication.Operation
}

// chosenImage is an image chosen in the setup, named as the field that holds
// it, and what it is for.
type chosenImage struct {
	field   string
	id      domain.FileID
	chosen  bool
	purpose FilePurpose
}

// save checks the setup code typed, without a lock, and that it opens the
// operation: a code that does not reaches nothing else. It then confirms the
// images chosen, before the transaction (Writes in CODING_STANDARDS.md). In
// one transaction it locks the setup code and checks it again, registers
// the user that register returns, ends the setup code and records the
// operation.
func (s savingWithSetupCode) save(
	ctx context.Context,
	typed string,
	images []chosenImage,
	register func(ctx context.Context) (*domain.User, error),
) error {
	setupCode, err := findWorkingSetupCode(ctx, s.SetupCodes, typed, s.Clock.Now())
	if err != nil {
		return err
	}
	if err := opensAtThisMoment(ctx, s.Consultancies, s.opens); err != nil {
		return err
	}
	if err := s.confirmImages(ctx, images); err != nil {
		return err
	}
	return s.Transactions.Run(ctx, func(ctx context.Context) error {
		return s.inTransaction(ctx, setupCode.Hash(), register)
	})
}

func (s savingWithSetupCode) inTransaction(
	ctx context.Context,
	hash domain.SetupCodeHash,
	register func(ctx context.Context) (*domain.User, error),
) error {
	now := s.Clock.Now()

	setupCode, err := lockWorkingSetupCode(ctx, s.SetupCodes, hash, now)
	if err != nil {
		return err
	}
	if err := opensAtThisMoment(ctx, s.Consultancies, s.opens); err != nil {
		return err
	}
	user, err := register(ctx)
	if err != nil {
		return err
	}
	setupCode.End(now)
	if err := s.SetupCodes.Save(ctx, setupCode); err != nil {
		return err
	}
	return s.RecordEntry.Execute(ctx, auditapplication.RecordEntryInput{
		OccurredAt: now,
		Operation:  s.operation,
		Actor:      auditapplication.SetupCodeUsed(int64(setupCode.ID())),
		Subject:    auditapplication.User(int64(user.ID())),
	})
}

// confirmImages confirms each image chosen, outside the transaction of the
// operation: each reads the storage. A file the files area refuses is named
// as the field that holds it.
func (s savingWithSetupCode) confirmImages(ctx context.Context, images []chosenImage) error {
	var invalid kernel.InvalidFieldsError
	var failed error
	for _, image := range images {
		refused, err := confirmImage(ctx, s.ConfirmUpload, image)
		failed = errors.Join(failed, err)
		invalid.AddField(image.field, refused)
	}
	if failed != nil {
		return failed
	}
	return invalid.Err()
}

// registerUser creates a user with the role given.
func (s savingWithSetupCode) registerUser(ctx context.Context, details domain.UserDetails, role domain.RoleID) (*domain.User, error) {
	id, err := s.Users.NextIdentifier(ctx)
	if err != nil {
		return nil, err
	}
	user, err := domain.RegisterUser(id, details, role)
	if err != nil {
		return nil, err
	}
	if err := s.Users.Save(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
