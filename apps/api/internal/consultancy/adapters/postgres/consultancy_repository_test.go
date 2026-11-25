package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
)

var fullConsultancy = domain.TypedConsultancy{
	Name:      "Fábrica de Software",
	LegalName: "Fábrica de Software Ltda.",
	Country:   "BR",
	TaxID:     "12.345.678/0001-90",
	TimeZone:  "America/Sao_Paulo",
	Currency:  "BRL",
	Email:     "Contato@Fabrica.example",
	Telephone: "+55 11 3333 4444",
	Website:   "https://fabrica.example",
	Address: domain.TypedAddress{
		Line1:      "Avenida Paulista, 1000",
		Line2:      "Sala 101",
		City:       "São Paulo",
		Region:     "SP",
		PostalCode: "01310-100",
	},
}

// register saves a consultancy, in a transaction of its own.
func (s stored) register(t *testing.T, typed domain.TypedConsultancy) error {
	t.Helper()
	details, err := domain.NewConsultancyDetails(typed)
	require.NoError(t, err)
	return s.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := s.consultancies.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		return s.consultancies.Save(ctx, domain.RegisterConsultancy(id, details))
	})
}

// row returns the one row a query reads, as a map of its columns.
func (s stored) row(t *testing.T, query string) map[string]any {
	t.Helper()
	rows, err := s.pool.Query(context.Background(), query)
	require.NoError(t, err)
	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToMap)
	require.NoError(t, err)
	return row
}

func TestAConsultancyIsSavedAsItWasTyped(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)

	require.NoError(t, temporary.register(t, fullConsultancy))

	row := temporary.row(t, "SELECT * FROM consultancy")
	assert.Positive(t, row["id"])
	delete(row, "id")
	assert.Equal(t, map[string]any{
		"name":                "Fábrica de Software",
		"legal_name":          "Fábrica de Software Ltda.",
		"country":             "BR",
		"tax_id":              "12.345.678/0001-90",
		"time_zone":           "America/Sao_Paulo",
		"currency":            "BRL",
		"email":               "contato@fabrica.example",
		"telephone":           "+551133334444",
		"website":             "https://fabrica.example",
		"address_line1":       "Avenida Paulista, 1000",
		"address_line2":       "Sala 101",
		"address_city":        "São Paulo",
		"address_region":      "SP",
		"address_postal_code": "01310-100",
		"logo_file_id":        nil,
	}, row)
}

func TestAConsultancyWithNoOptionalFieldStoresNothingInTheirColumns(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	typed := fullConsultancy
	typed.Email, typed.Telephone, typed.Website = "", "", ""
	typed.Address = domain.TypedAddress{City: "São Paulo"}

	require.NoError(t, temporary.register(t, typed))

	row := temporary.row(t, `SELECT email, telephone, website, address_line1, address_line2, address_city,
		address_region, address_postal_code, logo_file_id FROM consultancy`)
	assert.Equal(t, map[string]any{
		"email": nil, "telephone": nil, "website": nil,
		"address_line1": nil, "address_line2": nil, "address_city": "São Paulo",
		"address_region": nil, "address_postal_code": nil, "logo_file_id": nil,
	}, row)
}

// aFile writes the record of a file, as the files area does when an upload
// starts, and returns its identifier.
func (s stored) aFile(t *testing.T, purpose string) domain.FileID {
	t.Helper()
	var id int64
	err := s.pool.QueryRow(context.Background(), `
		INSERT INTO files (id, object_key, purpose, media_type, size, started_at)
		VALUES (nextval('files_id_seq'), md5(random()::text), $1, 'image/png', 1024, now())
		RETURNING id`, purpose).Scan(&id)
	require.NoError(t, err)
	return domain.FileID(id)
}

func TestAConsultancyIsSavedWithTheFileOfItsLogo(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	typed := fullConsultancy
	typed.LogoFileID = temporary.aFile(t, "consultancy_logo")

	require.NoError(t, temporary.register(t, typed))

	assert.Equal(t, int64(typed.LogoFileID), temporary.row(t, "SELECT logo_file_id FROM consultancy")["logo_file_id"])
}

func TestTheDatabaseRefusesALogoThatIsNoFile(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	typed := fullConsultancy
	typed.LogoFileID = 987654
	const foreignKeyViolation = "23503"

	err := temporary.register(t, typed)

	assert.Equal(t, foreignKeyViolation, sqlState(t, err))
}

func TestTheDatabaseRefusesASecondConsultancy(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	require.NoError(t, temporary.register(t, fullConsultancy))
	const uniqueViolation = "23505"

	err := temporary.register(t, fullConsultancy)

	assert.Equal(t, uniqueViolation, sqlState(t, err))
	assert.Equal(t, int64(1), temporary.row(t, "SELECT count(*) FROM consultancy")["count"])
}

func TestTheConsultancyExistsOnceItIsSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ctx := context.Background()
	before, err := temporary.consultancies.Exists(ctx)
	require.NoError(t, err)

	require.NoError(t, temporary.register(t, fullConsultancy))

	after, err := temporary.consultancies.Exists(ctx)
	require.NoError(t, err)
	assert.False(t, before)
	assert.True(t, after)
}

func TestAConsultancySavedInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	details, err := domain.NewConsultancyDetails(fullConsultancy)
	require.NoError(t, err)
	failing := errors.New("a later step failed")

	err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		id, err := temporary.consultancies.NextIdentifier(ctx)
		if err != nil {
			return err
		}
		if err := temporary.consultancies.Save(ctx, domain.RegisterConsultancy(id, details)); err != nil {
			return err
		}
		return failing
	})

	require.ErrorIs(t, err, failing)
	exists, err := temporary.consultancies.Exists(context.Background())
	require.NoError(t, err)
	assert.False(t, exists)
}

// The shell of the app reads the name of the consultancy and the file of its
// logo.
func TestTheConsultancyIsReadWithItsNameAndTheFileOfItsLogo(t *testing.T) {
	t.Parallel()
	withLogo := newStored(t)
	withoutLogo := newStored(t)
	typed := fullConsultancy
	typed.LogoFileID = withLogo.aFile(t, "consultancy_logo")
	require.NoError(t, withLogo.register(t, typed))
	require.NoError(t, withoutLogo.register(t, fullConsultancy))

	read, found, err := withLogo.consultancy.Read(context.Background())
	require.NoError(t, err)
	readWithout, foundWithout, err := withoutLogo.consultancy.Read(context.Background())
	require.NoError(t, err)

	require.True(t, found)
	assert.Equal(t, application.ConsultancyNameAndLogo{Name: "Fábrica de Software", LogoFileID: &typed.LogoFileID}, read)
	require.True(t, foundWithout)
	assert.Equal(t, application.ConsultancyNameAndLogo{Name: "Fábrica de Software", LogoFileID: nil}, readWithout)
}

func TestNoConsultancyIsReadBeforeTheSetupIsConcluded(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)

	_, found, err := temporary.consultancy.Read(context.Background())

	require.NoError(t, err)
	assert.False(t, found)
}

func TestTheStateOfTheSetupFollowsTheSetupCodesAndTheConsultancy(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ctx := context.Background()
	read := func(at time.Time) [2]bool {
		t.Helper()
		state, err := temporary.setupState.Read(ctx, at)
		require.NoError(t, err)
		return [2]bool{state.Concluded, state.SetupCodeWaiting}
	}
	concluded, waiting := true, true

	assert.Equal(t, [2]bool{!concluded, !waiting}, read(noon), "a new install")
	issued := temporary.issue(t, noon)
	assert.Equal(t, [2]bool{!concluded, waiting}, read(noon.Add(59*time.Minute)), "a setup code issued")
	assert.Equal(t, [2]bool{!concluded, !waiting}, read(noon.Add(time.Hour)), "the setup code expired")
	issued.End(noon.Add(10 * time.Minute))
	require.NoError(t, temporary.setupCodes.Save(ctx, issued))
	require.NoError(t, temporary.register(t, fullConsultancy))
	assert.Equal(t, [2]bool{concluded, !waiting}, read(noon.Add(10*time.Minute)), "the setup concluded")
	temporary.issue(t, noon.Add(20*time.Minute))
	assert.Equal(t, [2]bool{concluded, waiting}, read(noon.Add(30*time.Minute)), "a setup code issued after the setup")
}
