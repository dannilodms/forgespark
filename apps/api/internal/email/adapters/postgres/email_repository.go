// Package postgres keeps the e-mail records in PostgreSQL. What an e-mail
// says beyond its subject is written when it is delivered and is not kept.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// EmailRepository keeps the e-mails in the emails table.
type EmailRepository struct {
	pool *pgxpool.Pool
}

// NewEmailRepository returns the repository on this pool.
func NewEmailRepository(pool *pgxpool.Pool) *EmailRepository {
	return &EmailRepository{pool: pool}
}

func (r *EmailRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of e-mails.
func (r *EmailRepository) NextIdentifier(ctx context.Context) (domain.EmailID, error) {
	value, err := r.queries(ctx).NextEmailID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of an e-mail: %w", err)
	}
	return domain.NewEmailID(value)
}

// Load loads the e-mail and locks its row until the transaction ends.
func (r *EmailRepository) Load(ctx context.Context, id domain.EmailID) (*domain.Email, bool, error) {
	row, err := r.queries(ctx).LoadEmail(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("load an e-mail: %w", err)
	}
	language, err := kernel.NewLanguage(row.Language)
	if err != nil {
		return nil, false, fmt.Errorf("load the e-mail %d: its language is not supported", row.ID)
	}
	return domain.RestoreEmail(domain.StoredEmail{
		ID:        domain.EmailID(row.ID),
		Recipient: row.Recipient,
		Subject:   row.Subject,
		Template:  domain.Template(row.Template),
		Language:  language,
		Reference: domain.RestoreReference(domain.ReferenceKind(row.ReferenceKind), row.ReferenceID),
		Status:    domain.Status(row.Status),
		QueuedAt:  row.QueuedAt.UTC(),
		MessageID: valueOf(row.MessageID),
		SentAt:    inUTC(row.SentAt),
		FailedAt:  inUTC(row.FailedAt),
		Failure:   domain.FailureCode(valueOf(row.FailureCode)),
	}), true, nil
}

// Save writes the e-mail whole: a new row, or every column of the row it
// already has.
func (r *EmailRepository) Save(ctx context.Context, email *domain.Email) error {
	params := SaveEmailParams{
		ID:            int64(email.ID()),
		Recipient:     email.Recipient().String(),
		Subject:       email.Subject(),
		Template:      string(email.Template()),
		Language:      email.Language().String(),
		ReferenceKind: string(email.Reference().Kind()),
		ReferenceID:   email.Reference().Identifier(),
		Status:        string(email.Status()),
		QueuedAt:      email.QueuedAt(),
		MessageID:     nil,
		SentAt:        nil,
		FailedAt:      nil,
		FailureCode:   nil,
	}
	if messageID, sent := email.MessageID(); sent {
		params.MessageID = &messageID
	}
	if sentAt, sent := email.SentAt(); sent {
		params.SentAt = &sentAt
	}
	if code, failedAt, failed := email.Failure(); failed {
		failure := string(code)
		params.FailureCode = &failure
		params.FailedAt = &failedAt
	}
	if err := r.queries(ctx).SaveEmail(ctx, params); err != nil {
		return fmt.Errorf("save an e-mail: %w", err)
	}
	return nil
}

func valueOf(text *string) string {
	if text == nil {
		return ""
	}
	return *text
}

// inUTC gives an instant PostgreSQL answered in the time zone of the session
// in UTC, and nothing when there is none.
func inUTC(at *time.Time) *time.Time {
	if at == nil {
		return nil
	}
	utc := at.UTC()
	return &utc
}
