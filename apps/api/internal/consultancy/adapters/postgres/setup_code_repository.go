// Package postgres keeps the aggregates of the consultancy area in
// PostgreSQL. Its SQL is in the .sql files beside this one, and sqlc
// generates the functions that run it: the files named sqlc_*.go and
// *.sql.go, which are never edited.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// SetupCodeRepository keeps the setup codes in the setup_codes table.
type SetupCodeRepository struct {
	pool *pgxpool.Pool
}

// NewSetupCodeRepository returns the repository on this pool.
func NewSetupCodeRepository(pool *pgxpool.Pool) *SetupCodeRepository {
	return &SetupCodeRepository{pool: pool}
}

// queries runs the SQL in the transaction of the use case when the context
// carries one.
func (r *SetupCodeRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of setup codes.
func (r *SetupCodeRepository) NextIdentifier(ctx context.Context) (domain.SetupCodeID, error) {
	value, err := r.queries(ctx).NextSetupCodeID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of a setup code: %w", err)
	}
	return domain.NewSetupCodeID(value)
}

// LoadNotEnded loads the setup codes that were not ended and locks their
// rows until the transaction ends.
func (r *SetupCodeRepository) LoadNotEnded(ctx context.Context) ([]*domain.SetupCode, error) {
	rows, err := r.queries(ctx).LoadSetupCodesNotEnded(ctx)
	if err != nil {
		return nil, fmt.Errorf("load the setup codes not ended: %w", err)
	}
	setupCodes := make([]*domain.SetupCode, 0, len(rows))
	var unreadable error
	for _, row := range rows {
		setupCode, err := setupCodeOf(row)
		unreadable = errors.Join(unreadable, err)
		setupCodes = append(setupCodes, setupCode)
	}
	if unreadable != nil {
		return nil, unreadable
	}
	return setupCodes, nil
}

// FindByHash finds the setup code with this hash, ended or not, and locks
// nothing.
func (r *SetupCodeRepository) FindByHash(ctx context.Context, hash domain.SetupCodeHash) (*domain.SetupCode, bool, error) {
	row, err := r.queries(ctx).FindSetupCodeByHash(ctx, hash[:])
	return foundSetupCode(row, err, "find the setup code by its hash")
}

// LoadByHash loads the setup code with this hash, ended or not, and locks its
// row until the transaction ends.
func (r *SetupCodeRepository) LoadByHash(ctx context.Context, hash domain.SetupCodeHash) (*domain.SetupCode, bool, error) {
	row, err := r.queries(ctx).LoadSetupCodeByHash(ctx, hash[:])
	return foundSetupCode(row, err, "load the setup code by its hash")
}

// foundSetupCode rebuilds the setup code a query found, and answers false
// when it found none.
func foundSetupCode(row SetupCode, err error, doing string) (*domain.SetupCode, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", doing, err)
	}
	setupCode, err := setupCodeOf(row)
	if err != nil {
		return nil, false, err
	}
	return setupCode, true, nil
}

// Save writes the setup code whole: a new row, or every column of the row it
// already has.
func (r *SetupCodeRepository) Save(ctx context.Context, setupCode *domain.SetupCode) error {
	hash := setupCode.Hash()
	err := r.queries(ctx).SaveSetupCode(ctx, SaveSetupCodeParams{
		ID:        int64(setupCode.ID()),
		CodeHash:  hash[:],
		IssuedAt:  setupCode.IssuedAt(),
		ExpiresAt: setupCode.ExpiresAt(),
		EndedAt:   endedAtOf(setupCode),
	})
	if err != nil {
		return fmt.Errorf("save the setup code: %w", err)
	}
	return nil
}

func endedAtOf(setupCode *domain.SetupCode) *time.Time {
	endedAt, ended := setupCode.EndedAt()
	if !ended {
		return nil
	}
	return &endedAt
}

// setupCodeOf rebuilds the aggregate from its row. PostgreSQL answers an
// instant in the time zone of the session, and the domain keeps UTC.
func setupCodeOf(row SetupCode) (*domain.SetupCode, error) {
	id, err := domain.NewSetupCodeID(row.ID)
	if err != nil {
		return nil, err
	}
	hash, err := domain.SetupCodeHashFromBytes(row.CodeHash)
	if err != nil {
		return nil, err
	}
	return domain.RestoreSetupCode(id, hash, row.IssuedAt.UTC(), row.ExpiresAt.UTC(), inUTC(row.EndedAt)), nil
}

func inUTC(instant *time.Time) *time.Time {
	if instant == nil {
		return nil
	}
	utc := instant.UTC()
	return &utc
}
