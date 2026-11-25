// Package postgres keeps the login code requests and the provider links in
// PostgreSQL. The code of a request is not kept: it is computed from the
// request with a key the database never holds.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/forgespark/apps/api/internal/access/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
)

// LoginCodeRequestRepository keeps the login code requests in the
// login_code_requests table.
type LoginCodeRequestRepository struct {
	pool *pgxpool.Pool
}

// NewLoginCodeRequestRepository returns the repository on this pool.
func NewLoginCodeRequestRepository(pool *pgxpool.Pool) *LoginCodeRequestRepository {
	return &LoginCodeRequestRepository{pool: pool}
}

func (r *LoginCodeRequestRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of requests.
func (r *LoginCodeRequestRepository) NextIdentifier(ctx context.Context) (domain.LoginCodeRequestID, error) {
	value, err := r.queries(ctx).NextLoginCodeRequestID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of a login code request: %w", err)
	}
	return domain.NewLoginCodeRequestID(value)
}

// LoadNotEndedOf loads the requests of the user that were not ended, and
// locks their rows until the transaction ends.
func (r *LoginCodeRequestRepository) LoadNotEndedOf(ctx context.Context, user domain.UserID) ([]*domain.LoginCodeRequest, error) {
	rows, err := r.queries(ctx).LoadNotEndedLoginCodeRequestsOf(ctx, int64(user))
	if err != nil {
		return nil, fmt.Errorf("load the login code requests of a user: %w", err)
	}
	requests := make([]*domain.LoginCodeRequest, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, restored(row))
	}
	return requests, nil
}

// Find returns the request with this identifier, and locks nothing.
func (r *LoginCodeRequestRepository) Find(ctx context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error) {
	row, err := r.queries(ctx).FindLoginCodeRequest(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("find a login code request: %w", err)
	}
	return restored(row), true, nil
}

// Load returns the request with this identifier, and locks its row until the
// transaction ends.
func (r *LoginCodeRequestRepository) Load(ctx context.Context, id domain.LoginCodeRequestID) (*domain.LoginCodeRequest, bool, error) {
	row, err := r.queries(ctx).LoadLoginCodeRequest(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("load a login code request: %w", err)
	}
	return restored(row), true, nil
}

// Save writes the request whole: a new row, or every column of the row it
// already has.
func (r *LoginCodeRequestRepository) Save(ctx context.Context, request *domain.LoginCodeRequest) error {
	params := SaveLoginCodeRequestParams{
		ID:          int64(request.ID()),
		UserID:      int64(request.User()),
		RandomValue: request.RandomValue(),
		AskedAt:     request.AskedAt(),
		ExpiresAt:   request.ExpiresAt(),
		Attempts:    int32(request.Attempts()),
		EndedAt:     nil,
	}
	if endedAt, ended := request.EndedAt(); ended {
		params.EndedAt = &endedAt
	}
	if err := r.queries(ctx).SaveLoginCodeRequest(ctx, params); err != nil {
		return fmt.Errorf("save a login code request: %w", err)
	}
	return nil
}

// restored rebuilds a request from its row, with its instants in UTC.
func restored(row LoginCodeRequest) *domain.LoginCodeRequest {
	stored := domain.StoredLoginCodeRequest{
		ID:          domain.LoginCodeRequestID(row.ID),
		User:        domain.UserID(row.UserID),
		RandomValue: row.RandomValue,
		AskedAt:     row.AskedAt.UTC(),
		ExpiresAt:   row.ExpiresAt.UTC(),
		Attempts:    int(row.Attempts),
		EndedAt:     nil,
	}
	if row.EndedAt != nil {
		endedAt := row.EndedAt.UTC()
		stored.EndedAt = &endedAt
	}
	return domain.RestoreLoginCodeRequest(stored)
}

// RemoveEndedOrExpiredBefore removes the requests that ended, or expired,
// before the moment.
func (r *LoginCodeRequestRepository) RemoveEndedOrExpiredBefore(ctx context.Context, before time.Time) error {
	if err := r.queries(ctx).RemoveLoginCodeRequestsEndedOrExpiredBefore(ctx, before); err != nil {
		return fmt.Errorf("remove the old login code requests: %w", err)
	}
	return nil
}
