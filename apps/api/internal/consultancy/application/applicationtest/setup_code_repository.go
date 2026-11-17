package applicationtest

import (
	"context"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

// SetupCodeRepository keeps the setup codes in memory. It stores a copy of
// what it is given and hands out copies, so a change to a setup code shows
// only after it is saved, as with the database.
type SetupCodeRepository struct {
	lastIdentifier int64
	setupCodes     []domain.SetupCode
}

// NewSetupCodeRepository returns a repository with no setup code.
func NewSetupCodeRepository() *SetupCodeRepository {
	return &SetupCodeRepository{}
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *SetupCodeRepository) NextIdentifier(context.Context) (domain.SetupCodeID, error) {
	r.lastIdentifier++
	return domain.NewSetupCodeID(r.lastIdentifier)
}

// Save stores the setup code, in place of the one with the same identifier
// when there is one.
func (r *SetupCodeRepository) Save(_ context.Context, setupCode *domain.SetupCode) error {
	position := slices.IndexFunc(r.setupCodes, func(stored domain.SetupCode) bool {
		return stored.ID() == setupCode.ID()
	})
	if position < 0 {
		r.setupCodes = append(r.setupCodes, *setupCode)
		return nil
	}
	r.setupCodes[position] = *setupCode
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *SetupCodeRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.setupCodes)
	return func() { r.setupCodes = saved }
}

// All returns every setup code stored, in the order they were first saved.
func (r *SetupCodeRepository) All() []domain.SetupCode {
	return slices.Clone(r.setupCodes)
}

// LoadNotEnded returns a copy of every setup code that was not ended.
func (r *SetupCodeRepository) LoadNotEnded(context.Context) ([]*domain.SetupCode, error) {
	var notEnded []*domain.SetupCode
	for _, stored := range r.setupCodes {
		notEnded = appendNotEnded(notEnded, stored)
	}
	return notEnded, nil
}

func appendNotEnded(notEnded []*domain.SetupCode, stored domain.SetupCode) []*domain.SetupCode {
	if _, ended := stored.EndedAt(); ended {
		return notEnded
	}
	return append(notEnded, &stored)
}
