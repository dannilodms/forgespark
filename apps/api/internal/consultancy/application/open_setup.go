package application

import (
	"context"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

// OpenSetup checks a setup code that a person typed and tells what it opens
// and when it expires. It saves nothing, so a wrong code changes nothing and
// the setup code keeps working.
type OpenSetup struct {
	setupCodes SetupCodeRepository
	clock      Clock
}

// NewOpenSetup builds the use case.
func NewOpenSetup(setupCodes SetupCodeRepository, clock Clock) *OpenSetup {
	return &OpenSetup{setupCodes: setupCodes, clock: clock}
}

// OpenSetupInput carries the setup code as it was typed.
type OpenSetupInput struct {
	Code string
}

// OpenedSetup is what a setup code that works opens.
type OpenedSetup struct {
	Opens     domain.SetupCodeOpens
	ExpiresAt time.Time
}

// Execute answers what the setup code opens.
func (u *OpenSetup) Execute(ctx context.Context, input OpenSetupInput) (OpenedSetup, error) {
	plain, err := domain.ReadPlainSetupCode(input.Code)
	if err != nil {
		return OpenedSetup{}, err
	}
	setupCode, found, err := u.setupCodes.FindByHash(ctx, plain.Hash())
	if err != nil {
		return OpenedSetup{}, err
	}
	if !found || !setupCode.Works(u.clock.Now()) {
		return OpenedSetup{}, domain.SetupCodeInvalidError{}
	}
	return OpenedSetup{Opens: domain.OpensFirstSetup, ExpiresAt: setupCode.ExpiresAt()}, nil
}
