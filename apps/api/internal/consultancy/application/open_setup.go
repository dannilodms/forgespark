package application

import (
	"context"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
)

// OpenSetup checks a setup code that a person typed and tells what it opens
// and when it expires. It saves nothing, so a wrong code changes nothing and
// the setup code keeps working.
type OpenSetup struct {
	setupCodes    SetupCodeRepository
	consultancies ConsultancyRepository
	clock         Clock
}

// NewOpenSetup builds the use case.
func NewOpenSetup(setupCodes SetupCodeRepository, consultancies ConsultancyRepository, clock Clock) *OpenSetup {
	return &OpenSetup{setupCodes: setupCodes, consultancies: consultancies, clock: clock}
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

// Execute answers what the setup code opens: the first setup while no
// consultancy exists, and the registration of a new administrator once one
// does.
func (u *OpenSetup) Execute(ctx context.Context, input OpenSetupInput) (OpenedSetup, error) {
	setupCode, err := findWorkingSetupCode(ctx, u.setupCodes, input.Code, u.clock.Now())
	if err != nil {
		return OpenedSetup{}, err
	}
	concluded, err := u.consultancies.Exists(ctx)
	if err != nil {
		return OpenedSetup{}, err
	}
	return OpenedSetup{Opens: domain.SetupCodeOpensWhen(concluded), ExpiresAt: setupCode.ExpiresAt()}, nil
}
