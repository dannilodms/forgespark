package applicationtest

import (
	"context"
	"errors"
	"slices"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
)

// ConsultancyRepository keeps the consultancy in memory. It stores a copy of
// what it is given, as the database does.
type ConsultancyRepository struct {
	lastIdentifier int64
	consultancies  []domain.Consultancy
	failure        error
	beforeExists   func()
}

// BeforeNextExists runs hook once, the next time Exists is asked, before it
// answers: another operation that commits between two reads of a use case.
func (r *ConsultancyRepository) BeforeNextExists(hook func()) {
	r.beforeExists = hook
}

// NewConsultancyRepository returns a repository with no consultancy.
func NewConsultancyRepository() *ConsultancyRepository {
	return &ConsultancyRepository{}
}

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *ConsultancyRepository) FailWith(failure error) {
	r.failure = failure
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *ConsultancyRepository) NextIdentifier(context.Context) (domain.ConsultancyID, error) {
	r.lastIdentifier++
	return domain.NewConsultancyID(r.lastIdentifier)
}

// Exists tells whether a consultancy was saved.
func (r *ConsultancyRepository) Exists(context.Context) (bool, error) {
	hook := r.beforeExists
	r.beforeExists = nil
	if hook != nil {
		hook()
	}
	return len(r.consultancies) > 0, nil
}

// Save stores the consultancy, in place of the one with the same identifier
// when there is one. Like the database, it refuses a second consultancy.
func (r *ConsultancyRepository) Save(_ context.Context, consultancy *domain.Consultancy) error {
	if r.failure != nil {
		return r.failure
	}
	position := slices.IndexFunc(r.consultancies, func(stored domain.Consultancy) bool {
		return stored.ID() == consultancy.ID()
	})
	if position >= 0 {
		r.consultancies[position] = *consultancy
		return nil
	}
	if len(r.consultancies) > 0 {
		return errors.New("there is one consultancy and it exists")
	}
	r.consultancies = append(r.consultancies, *consultancy)
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *ConsultancyRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.consultancies)
	return func() { r.consultancies = saved }
}

// All returns every consultancy stored: one at most.
func (r *ConsultancyRepository) All() []domain.Consultancy {
	return slices.Clone(r.consultancies)
}
