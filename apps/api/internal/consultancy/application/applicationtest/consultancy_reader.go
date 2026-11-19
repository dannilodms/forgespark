package applicationtest

import (
	"context"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
)

// ConsultancyReader reads the consultancy from the in-memory repository, as
// the query of the database reads it from the table.
type ConsultancyReader struct {
	consultancies *ConsultancyRepository
}

// NewConsultancyReader returns a reader over this repository.
func NewConsultancyReader(consultancies *ConsultancyRepository) *ConsultancyReader {
	return &ConsultancyReader{consultancies: consultancies}
}

// Read returns the name and the logo of the stored consultancy.
func (r *ConsultancyReader) Read(context.Context) (application.ConsultancyNameAndLogo, bool, error) {
	stored := r.consultancies.All()
	if len(stored) == 0 {
		return application.ConsultancyNameAndLogo{}, false, nil
	}
	details := stored[0].Details()
	logo, _ := details.LogoFileID()
	return application.ConsultancyNameAndLogo{Name: details.Name(), LogoFileID: logo}, true, nil
}
