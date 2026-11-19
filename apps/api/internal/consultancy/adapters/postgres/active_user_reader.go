package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// ActiveUserReader reads the users who can log in from the users table. It
// builds no aggregate and locks nothing.
type ActiveUserReader struct {
	pool *pgxpool.Pool
}

// NewActiveUserReader returns the reader on this pool.
func NewActiveUserReader(pool *pgxpool.Pool) *ActiveUserReader {
	return &ActiveUserReader{pool: pool}
}

// ActiveUserWith returns the user who has this e-mail and is not deactivated.
// The error names the step that failed and never the e-mail.
func (r *ActiveUserReader) ActiveUserWith(ctx context.Context, email kernel.Email) (application.ActiveUser, bool, error) {
	row, err := New(database.Within(ctx, r.pool)).FindActiveUser(ctx, email.String())
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ActiveUser{}, false, nil
	}
	if err != nil {
		return application.ActiveUser{}, false, fmt.Errorf("look for the user who logs in with an e-mail: %w", err)
	}
	stored, err := kernel.NewEmail(row.Email)
	if err != nil {
		return application.ActiveUser{}, false, fmt.Errorf("read the user %d: their e-mail is not valid", row.ID)
	}
	language, err := kernel.NewLanguage(row.PreferredLanguage)
	if err != nil {
		return application.ActiveUser{}, false, fmt.Errorf("read the user %d: their preferred language is not supported", row.ID)
	}
	id, err := domain.NewUserID(row.ID)
	if err != nil {
		return application.ActiveUser{}, false, err
	}
	return application.ActiveUser{ID: id, Email: stored, PreferredLanguage: language}, true, nil
}
