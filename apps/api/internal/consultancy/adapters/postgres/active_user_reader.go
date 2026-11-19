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

// ActiveUserReader reads the users who can log in from the users table, with
// the keys of their roles. It builds no aggregate and locks nothing.
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
	return activeUserOf(activeUserRow(row))
}

// ActiveUserOf returns the user with this identifier when they are not
// deactivated.
func (r *ActiveUserReader) ActiveUserOf(ctx context.Context, id domain.UserID) (application.ActiveUser, bool, error) {
	row, err := New(database.Within(ctx, r.pool)).FindActiveUserByID(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ActiveUser{}, false, nil
	}
	if err != nil {
		return application.ActiveUser{}, false, fmt.Errorf("look for the user of a session: %w", err)
	}
	return activeUserOf(activeUserRow(row))
}

// activeUserRow is the row both queries answer with.
type activeUserRow FindActiveUserRow

func activeUserOf(row activeUserRow) (application.ActiveUser, bool, error) {
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
	roles := make([]domain.RoleKey, 0, len(row.RoleKeys))
	for _, key := range row.RoleKeys {
		roles = append(roles, domain.RoleKey(key))
	}
	return application.ActiveUser{
		ID:                id,
		Name:              row.Name,
		Email:             stored,
		PreferredLanguage: language,
		Roles:             roles,
		Scopes:            nil,
	}, true, nil
}
