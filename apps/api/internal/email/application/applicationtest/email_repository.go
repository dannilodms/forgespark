package applicationtest

import (
	"context"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
)

// EmailRepository keeps the e-mails in memory. It stores a copy of what it is
// given and hands out copies, as the database does.
type EmailRepository struct {
	lastIdentifier int64
	emails         []domain.Email
	failure        error
}

// NewEmailRepository returns a repository with no e-mail.
func NewEmailRepository() *EmailRepository {
	return &EmailRepository{}
}

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *EmailRepository) FailWith(failure error) {
	r.failure = failure
}

// NextIdentifier reserves the next identifier, starting at 1.
func (r *EmailRepository) NextIdentifier(context.Context) (domain.EmailID, error) {
	r.lastIdentifier++
	return domain.NewEmailID(r.lastIdentifier)
}

// Load returns a copy of the e-mail with this identifier. A context that
// ended is refused, as the database refuses a query on it.
func (r *EmailRepository) Load(ctx context.Context, id domain.EmailID) (*domain.Email, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	email, found := r.Find(id)
	if !found {
		return nil, false, nil
	}
	return &email, true, nil
}

// Save stores the e-mail, in place of the one with the same identifier when
// there is one. A context that ended is refused, as by the database.
func (r *EmailRepository) Save(ctx context.Context, email *domain.Email) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.failure != nil {
		return r.failure
	}
	position := slices.IndexFunc(r.emails, func(stored domain.Email) bool { return stored.ID() == email.ID() })
	if position < 0 {
		r.emails = append(r.emails, *email)
		return nil
	}
	r.emails[position] = *email
	return nil
}

// Snapshot lets the repository take part in the in-memory transactions.
func (r *EmailRepository) Snapshot() (restore func()) {
	saved := slices.Clone(r.emails)
	return func() { r.emails = saved }
}

// All returns every e-mail stored, in the order they were first saved.
func (r *EmailRepository) All() []domain.Email {
	return slices.Clone(r.emails)
}

// Find returns the e-mail with this identifier, and false when there is none.
func (r *EmailRepository) Find(id domain.EmailID) (domain.Email, bool) {
	position := slices.IndexFunc(r.emails, func(stored domain.Email) bool { return stored.ID() == id })
	if position < 0 {
		return domain.Email{}, false
	}
	return r.emails[position], true
}
