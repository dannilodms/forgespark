package applicationtest

import (
	"context"
	"slices"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

// SetupStateReader reads the state of the setup from the in-memory
// repositories, as the query of the database reads it from the tables.
type SetupStateReader struct {
	consultancies *ConsultancyRepository
	setupCodes    *SetupCodeRepository
}

// NewSetupStateReader returns a reader over these repositories.
func NewSetupStateReader(consultancies *ConsultancyRepository, setupCodes *SetupCodeRepository) *SetupStateReader {
	return &SetupStateReader{consultancies: consultancies, setupCodes: setupCodes}
}

// Read tells whether a consultancy was saved and whether a setup code works
// at this moment.
func (r *SetupStateReader) Read(_ context.Context, at time.Time) (application.SetupState, error) {
	return application.SetupState{
		Concluded: len(r.consultancies.All()) > 0,
		SetupCodeWaiting: slices.ContainsFunc(r.setupCodes.All(), func(setupCode domain.SetupCode) bool {
			return setupCode.Works(at)
		}),
	}, nil
}
