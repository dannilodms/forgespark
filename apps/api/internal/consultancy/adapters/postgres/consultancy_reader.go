package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// ConsultancyReader reads the consultancy as the shell of the app shows it,
// in one query. It builds no aggregate and locks nothing.
type ConsultancyReader struct {
	pool *pgxpool.Pool
}

// NewConsultancyReader returns the reader on this pool.
func NewConsultancyReader(pool *pgxpool.Pool) *ConsultancyReader {
	return &ConsultancyReader{pool: pool}
}

// Read returns the name of the consultancy and the file of its logo, the
// file 0 when it has none.
func (r *ConsultancyReader) Read(ctx context.Context) (application.ConsultancyNameAndLogo, bool, error) {
	row, err := New(database.Within(ctx, r.pool)).ReadConsultancyNameAndLogo(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ConsultancyNameAndLogo{}, false, nil
	}
	if err != nil {
		return application.ConsultancyNameAndLogo{}, false, fmt.Errorf("read the name and the logo of the consultancy: %w", err)
	}
	return application.ConsultancyNameAndLogo{Name: row.Name, LogoFileID: fileOrNone(row.LogoFileID)}, true, nil
}

// fileOrNone is the file a column names, and the file 0 when it names none.
func fileOrNone(id *int64) domain.FileID {
	if id == nil {
		return 0
	}
	return domain.FileID(*id)
}
