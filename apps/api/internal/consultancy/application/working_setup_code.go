package application

import (
	"context"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
)

// findWorkingSetupCode reads the setup code a person typed and finds it,
// locking nothing, and answers SetupCodeInvalidError when it is wrong, expired
// or ended: one answer for the three.
func findWorkingSetupCode(ctx context.Context, setupCodes SetupCodeRepository, typed string, now time.Time) (*domain.SetupCode, error) {
	plain, err := domain.ReadPlainSetupCode(typed)
	if err != nil {
		return nil, err
	}
	setupCode, found, err := setupCodes.FindByHash(ctx, plain.Hash())
	if err != nil {
		return nil, err
	}
	if !found || !setupCode.Works(now) {
		return nil, domain.SetupCodeInvalidError{}
	}
	return setupCode, nil
}

// lockWorkingSetupCode loads the setup code with this hash, locked until the
// transaction ends, so that two operations with the same code run one after
// the other and the second sees whether the first ended it. It answers
// SetupCodeInvalidError when the code does not work at this moment.
func lockWorkingSetupCode(ctx context.Context, setupCodes SetupCodeRepository, hash domain.SetupCodeHash, now time.Time) (*domain.SetupCode, error) {
	setupCode, found, err := setupCodes.LoadByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if !found || !setupCode.Works(now) {
		return nil, domain.SetupCodeInvalidError{}
	}
	return setupCode, nil
}

// opensWithWorkingSetupCode checks, without a lock, that the setup code found
// working still opens the operation. Finding the code and asking whether the
// setup is concluded are two reads, and an operation with the same code may
// commit between them: it ends the code and changes the setup. So when the
// setup no longer opens the operation, the code is read again, and a code
// that stopped working answers SetupCodeInvalidError, as it would have a
// moment later, instead of SetupAlreadyConcludedError or
// SetupNotConcludedError.
func opensWithWorkingSetupCode(
	ctx context.Context,
	setupCodes SetupCodeRepository,
	consultancies ConsultancyRepository,
	setupCode *domain.SetupCode,
	operation domain.SetupCodeOpens,
	now time.Time,
) error {
	refused := opensAtThisMoment(ctx, consultancies, operation)
	if refused == nil {
		return nil
	}
	again, found, err := setupCodes.FindByHash(ctx, setupCode.Hash())
	if err != nil {
		return err
	}
	if !found || !again.Works(now) {
		return domain.SetupCodeInvalidError{}
	}
	return refused
}

// opensAtThisMoment checks that a setup code that works opens, at this
// moment, what the operation is: the first setup while no consultancy
// exists, the registration of a new administrator once one does. It answers
// SetupAlreadyConcludedError or SetupNotConcludedError otherwise.
func opensAtThisMoment(ctx context.Context, consultancies ConsultancyRepository, operation domain.SetupCodeOpens) error {
	concluded, err := consultancies.Exists(ctx)
	if err != nil {
		return err
	}
	if domain.SetupCodeOpensWhen(concluded) == operation {
		return nil
	}
	if concluded {
		return domain.SetupAlreadyConcludedError{}
	}
	return domain.SetupNotConcludedError{}
}
