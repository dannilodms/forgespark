// Package postgres keeps the records of the files in PostgreSQL. The bytes of
// a file are in the file storage, never here.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/files/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// FileRepository keeps the files in the files table.
type FileRepository struct {
	pool *pgxpool.Pool
}

// NewFileRepository returns the repository on this pool.
func NewFileRepository(pool *pgxpool.Pool) *FileRepository {
	return &FileRepository{pool: pool}
}

func (r *FileRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of files.
func (r *FileRepository) NextIdentifier(ctx context.Context) (domain.FileID, error) {
	value, err := r.queries(ctx).NextFileID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of a file: %w", err)
	}
	return domain.NewFileID(value)
}

// Load loads the file and locks its row until the transaction ends.
func (r *FileRepository) Load(ctx context.Context, id domain.FileID) (*domain.File, bool, error) {
	row, err := r.queries(ctx).LoadFile(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("load a file: %w", err)
	}
	return domain.RestoreFile(
		domain.FileID(row.ID),
		domain.ObjectKey(row.ObjectKey),
		domain.Purpose(row.Purpose),
		domain.MediaType(row.MediaType),
		row.Size,
		row.StartedAt.UTC(),
		inUTC(row.ConfirmedAt),
	), true, nil
}

// Save writes the file whole: a new row, or every column of the row it
// already has.
func (r *FileRepository) Save(ctx context.Context, file *domain.File) error {
	var confirmedAt *time.Time
	if at, confirmed := file.ConfirmedAt(); confirmed {
		confirmedAt = &at
	}
	err := r.queries(ctx).SaveFile(ctx, SaveFileParams{
		ID:          int64(file.ID()),
		ObjectKey:   string(file.ObjectKey()),
		Purpose:     string(file.Purpose()),
		MediaType:   string(file.Declared().MediaType()),
		Size:        file.Declared().Size(),
		StartedAt:   file.StartedAt(),
		ConfirmedAt: confirmedAt,
	})
	if err != nil {
		return fmt.Errorf("save a file: %w", err)
	}
	return nil
}

// RemoveAbandoned deletes the files whose upload started before this instant
// and was never confirmed. A row a confirmation holds is deleted, or not,
// once the confirmation ends: PostgreSQL then reads it again.
func (r *FileRepository) RemoveAbandoned(ctx context.Context, startedBefore time.Time) error {
	if err := r.queries(ctx).RemoveAbandonedFiles(ctx, startedBefore); err != nil {
		return fmt.Errorf("remove the abandoned files: %w", err)
	}
	return nil
}

// ConfirmedKeys returns those of the keys a confirmed file names, in their
// order.
func (r *FileRepository) ConfirmedKeys(ctx context.Context, among []domain.ObjectKey) ([]domain.ObjectKey, error) {
	keys := make([]string, 0, len(among))
	for _, key := range among {
		keys = append(keys, string(key))
	}
	rows, err := r.queries(ctx).ConfirmedObjectKeys(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("read the keys of the confirmed files: %w", err)
	}
	confirmed := make([]domain.ObjectKey, 0, len(rows))
	for _, key := range rows {
		confirmed = append(confirmed, domain.ObjectKey(key))
	}
	return confirmed, nil
}

// ConfirmedKey returns the key of the file with this identifier when it was
// confirmed, and locks nothing.
func (r *FileRepository) ConfirmedKey(ctx context.Context, id domain.FileID) (domain.ObjectKey, bool, error) {
	key, err := r.queries(ctx).ConfirmedObjectKeyOf(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read the key of a confirmed file: %w", err)
	}
	return domain.ObjectKey(key), true, nil
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
