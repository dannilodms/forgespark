package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
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
