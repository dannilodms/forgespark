package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database"
)

// ConsultancyRepository keeps the consultancy in the consultancy table, which
// holds one row at most.
type ConsultancyRepository struct {
	pool *pgxpool.Pool
}

// NewConsultancyRepository returns the repository on this pool.
func NewConsultancyRepository(pool *pgxpool.Pool) *ConsultancyRepository {
	return &ConsultancyRepository{pool: pool}
}

func (r *ConsultancyRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// NextIdentifier takes the next value of the sequence of the consultancy.
func (r *ConsultancyRepository) NextIdentifier(ctx context.Context) (domain.ConsultancyID, error) {
	value, err := r.queries(ctx).NextConsultancyID(ctx)
	if err != nil {
		return 0, fmt.Errorf("reserve the identifier of the consultancy: %w", err)
	}
	return domain.NewConsultancyID(value)
}

// Exists tells whether the consultancy was registered. It locks nothing.
func (r *ConsultancyRepository) Exists(ctx context.Context) (bool, error) {
	exists, err := r.queries(ctx).ConsultancyExists(ctx)
	if err != nil {
		return false, fmt.Errorf("tell whether the consultancy exists: %w", err)
	}
	return exists, nil
}

// Save writes the consultancy whole: a new row, or every column of the row
// it already has.
func (r *ConsultancyRepository) Save(ctx context.Context, consultancy *domain.Consultancy) error {
	details := consultancy.Details()
	email, hasEmail := details.Email()
	telephone, hasTelephone := details.Telephone()
	website, hasWebsite := details.Website()
	address, hasAddress := details.Address()
	logo, hasLogo := details.LogoFileID()
	err := r.queries(ctx).SaveConsultancy(ctx, SaveConsultancyParams{
		ID:                int64(consultancy.ID()),
		Name:              details.Name(),
		LegalName:         details.LegalName(),
		Country:           details.Country().String(),
		TaxID:             details.TaxID(),
		TimeZone:          details.TimeZone().String(),
		Currency:          details.Currency().String(),
		Email:             textIf(email.String(), hasEmail),
		Telephone:         textIf(telephone.String(), hasTelephone),
		Website:           textIf(website.String(), hasWebsite),
		AddressLine1:      textIf(address.Line1(), hasAddress),
		AddressLine2:      textIf(address.Line2(), hasAddress),
		AddressCity:       textIf(address.City(), hasAddress),
		AddressRegion:     textIf(address.Region(), hasAddress),
		AddressPostalCode: textIf(address.PostalCode(), hasAddress),
		LogoFileID:        fileIf(logo, hasLogo),
	})
	if err != nil {
		return fmt.Errorf("save the consultancy: %w", err)
	}
	return nil
}

// fileIf gives the file to store in a column that may be empty: nothing when
// none was chosen.
func fileIf(id domain.FileID, present bool) *int64 {
	if !present {
		return nil
	}
	value := int64(id)
	return &value
}

// textIf gives the text to store in a column that may be empty: nothing when
// the value is absent, and nothing for an empty part of a value either.
func textIf(text string, present bool) *string {
	if !present || text == "" {
		return nil
	}
	return &text
}
