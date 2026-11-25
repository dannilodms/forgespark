package application_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auditapplication "github.com/dannilodms/forgespark/apps/api/internal/audit/application"
	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/application"
	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// typedSetup is what a person types in the setup, every field filled in,
// with the code they were given.
func typedSetup(code string) application.ConcludeSetupInput {
	return application.ConcludeSetupInput{
		Code: code,
		Consultancy: domain.TypedConsultancy{
			Name:      "Fábrica de Software",
			LegalName: "Fábrica de Software Ltda.",
			Country:   "BR",
			TaxID:     "12.345.678/0001-90",
			TimeZone:  "America/Sao_Paulo",
			Currency:  "BRL",
			Email:     "contato@fabrica.example",
			Telephone: "+55 11 3333 4444",
			Website:   "https://fabrica.example",
			Address: domain.TypedAddress{
				Line1:      "Avenida Paulista, 1000",
				Line2:      "Sala 101",
				City:       "São Paulo",
				Region:     "SP",
				PostalCode: "01310-100",
			},
		},
		Administrator: domain.TypedUser{
			Name:              "Ana Souza",
			Email:             "ana.souza@fabrica.example",
			Country:           "BR",
			NationalID:        "123.456.789-09",
			Telephone:         "+55 11 98888 7777",
			PreferredLanguage: "pt-BR",
		},
	}
}

// issued is an installation with one setup code issued at noon, the first
// code, and the clock ten minutes later.
func issued(t *testing.T) installation {
	t.Helper()
	installed := newInstallation()
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	installed.clock.Advance(10 * time.Minute)
	return installed
}

func TestConcludingTheSetupRegistersTheConsultancyAsItWasTyped(t *testing.T) {
	installed := issued(t)

	err := installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode))

	require.NoError(t, err)
	consultancies := installed.consultancies.All()
	require.Len(t, consultancies, 1)
	details := consultancies[0].Details()
	assert.Equal(t, "Fábrica de Software", details.Name())
	assert.Equal(t, "Fábrica de Software Ltda.", details.LegalName())
	assert.Equal(t, "BR", details.Country().String())
	assert.Equal(t, "12.345.678/0001-90", details.TaxID(), "the tax ID is stored as typed")
	assert.Equal(t, "America/Sao_Paulo", details.TimeZone().String())
	assert.Equal(t, "BRL", details.Currency().String())
	email, hasEmail := details.Email()
	assert.True(t, hasEmail)
	assert.Equal(t, "contato@fabrica.example", email.String())
	telephone, hasTelephone := details.Telephone()
	assert.True(t, hasTelephone)
	assert.Equal(t, "+551133334444", telephone.String(), "the telephone is stored without spaces")
	website, hasWebsite := details.Website()
	assert.True(t, hasWebsite)
	assert.Equal(t, "https://fabrica.example", website.String())
	address, hasAddress := details.Address()
	assert.True(t, hasAddress)
	assert.Equal(t, "Avenida Paulista, 1000", address.Line1())
	assert.Equal(t, "Sala 101", address.Line2())
	assert.Equal(t, "São Paulo", address.City())
	assert.Equal(t, "SP", address.Region())
	assert.Equal(t, "01310-100", address.PostalCode())
}

func TestConcludingTheSetupRegistersTheAdministratorWithTheAdministratorRole(t *testing.T) {
	installed := issued(t)

	err := installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode))

	require.NoError(t, err)
	users := installed.users.All()
	require.Len(t, users, 1)
	details := users[0].Details()
	assert.Equal(t, "Ana Souza", details.Name())
	assert.Equal(t, "ana.souza@fabrica.example", details.Email().String())
	assert.Equal(t, "BR", details.Country().String())
	assert.Equal(t, "123.456.789-09", details.NationalID(), "the national ID is stored as typed")
	telephone, hasTelephone := details.Telephone()
	assert.True(t, hasTelephone)
	assert.Equal(t, "+5511988887777", telephone.String())
	assert.Equal(t, "pt-BR", details.PreferredLanguage().String())
	assert.False(t, users[0].Deactivated())
	administratorRole := roleWithKey(t, installed, domain.AdministratorRoleKey)
	assert.Equal(t, []domain.RoleID{administratorRole.ID()}, users[0].Roles())
}

func TestConcludingTheSetupCreatesTheDeveloperRoleWhichNoUserHolds(t *testing.T) {
	installed := issued(t)

	err := installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode))

	require.NoError(t, err)
	roles := installed.roles.All()
	require.Len(t, roles, 2)
	developerRole := roleWithKey(t, installed, domain.DeveloperRoleKey)
	for _, user := range installed.users.All() {
		assert.NotContains(t, user.Roles(), developerRole.ID())
	}
}

func roleWithKey(t *testing.T, installed installation, key domain.RoleKey) domain.Role {
	t.Helper()
	roles := installed.roles.All()
	position := slices.IndexFunc(roles, func(role domain.Role) bool { return role.Key() == key })
	require.GreaterOrEqual(t, position, 0, "no role has the key %s", key)
	return roles[position]
}

func TestConcludingTheSetupEndsTheSetupCodeAndRecordsTheConclusion(t *testing.T) {
	installed := issued(t)

	err := installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode))

	require.NoError(t, err)
	setupCodes := installed.setupCodes.All()
	require.Len(t, setupCodes, 1)
	endedAt, ended := setupCodes[0].EndedAt()
	assert.True(t, ended, "the setup code still works")
	assert.Equal(t, noon.Add(10*time.Minute), endedAt)
	users := installed.users.All()
	require.Len(t, users, 1)
	entries := installed.auditLog.Entries()
	require.Len(t, entries, 2)
	assert.Equal(t, auditapplication.SetupConcluded, entries[1].Operation())
	assert.Equal(t, auditapplication.SetupCodeUsed(int64(setupCodes[0].ID())), entries[1].Actor())
	assert.Equal(t, auditapplication.User(int64(users[0].ID())), entries[1].Subject())
	assert.Equal(t, noon.Add(10*time.Minute), entries[1].OccurredAt())
}

// brokenFields returns, for each field the error names, the code of the rule
// it breaks.
func brokenFields(t *testing.T, err error) map[string]string {
	t.Helper()
	var invalid kernel.InvalidFieldsError
	require.ErrorAs(t, err, &invalid)
	broken := map[string]string{}
	for _, field := range invalid.Fields {
		broken[field.Field] = field.Rule.Code()
	}
	return broken
}

// requireNothingSaved checks that a setup that was refused saved nothing and
// left its setup code working.
func requireNothingSaved(t *testing.T, installed installation) {
	t.Helper()
	assert.Empty(t, installed.consultancies.All(), "a consultancy was saved")
	assert.Empty(t, installed.users.All(), "a user was saved")
	assert.Empty(t, installed.roles.All(), "a role was saved")
	assert.Len(t, installed.auditLog.Entries(), 1, "the conclusion was recorded")
	_, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: firstCode})
	assert.NoError(t, err, "the setup code stopped working")
}

func TestEachRequiredFieldLeftOutIsNamedAndNothingIsSaved(t *testing.T) {
	leftOut := map[string]func(input *application.ConcludeSetupInput){
		"consultancy.name":                 func(input *application.ConcludeSetupInput) { input.Consultancy.Name = "" },
		"consultancy.legal_name":           func(input *application.ConcludeSetupInput) { input.Consultancy.LegalName = "" },
		"consultancy.country":              func(input *application.ConcludeSetupInput) { input.Consultancy.Country = "" },
		"consultancy.tax_id":               func(input *application.ConcludeSetupInput) { input.Consultancy.TaxID = "" },
		"consultancy.time_zone":            func(input *application.ConcludeSetupInput) { input.Consultancy.TimeZone = "" },
		"consultancy.currency":             func(input *application.ConcludeSetupInput) { input.Consultancy.Currency = "" },
		"administrator.name":               func(input *application.ConcludeSetupInput) { input.Administrator.Name = "" },
		"administrator.email":              func(input *application.ConcludeSetupInput) { input.Administrator.Email = "" },
		"administrator.country":            func(input *application.ConcludeSetupInput) { input.Administrator.Country = "" },
		"administrator.national_id":        func(input *application.ConcludeSetupInput) { input.Administrator.NationalID = "" },
		"administrator.preferred_language": func(input *application.ConcludeSetupInput) { input.Administrator.PreferredLanguage = "" },
	}
	for field, leave := range leftOut {
		t.Run(field, func(t *testing.T) {
			installed := issued(t)
			input := typedSetup(firstCode)
			leave(&input)

			err := installed.concludeSetup.Execute(context.Background(), input)

			assert.Equal(t, map[string]string{field: "required"}, brokenFields(t, err))
			requireNothingSaved(t, installed)
		})
	}
}

func TestARequiredFieldOfSpacesOnlyIsLeftOut(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.Name = "   "
	input.Administrator.NationalID = " \t "

	err := installed.concludeSetup.Execute(context.Background(), input)

	assert.Equal(t, map[string]string{"consultancy.name": "required", "administrator.national_id": "required"}, brokenFields(t, err))
	requireNothingSaved(t, installed)
}

func TestAValueOutsideItsListIsNamedAndNothingIsSaved(t *testing.T) {
	outside := map[string]brokenField{
		"a country that does not exist": {
			field: "consultancy.country", code: "country_unknown",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Country = "ZZ" },
		},
		"a country in lower case": {
			field: "consultancy.country", code: "country_unknown",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Country = "br" },
		},
		"a code that is not a currency in circulation": {
			field: "consultancy.currency", code: "currency_unknown",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Currency = "XAU" },
		},
		"a time zone that does not exist": {
			field: "consultancy.time_zone", code: "time_zone_unknown",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.TimeZone = "America/Atlantis" },
		},
		"a country that is not an ISO code": {
			field: "administrator.country", code: "country_unknown",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Country = "UK" },
		},
		"a language the system does not offer": {
			field: "administrator.preferred_language", code: "language_unsupported",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.PreferredLanguage = "fr" },
		},
	}
	requireEachFieldRefused(t, outside)
}

// brokenField is a change to a typed setup that breaks the rule of one field.
type brokenField struct {
	field  string
	code   string
	change func(input *application.ConcludeSetupInput)
}

// requireEachFieldRefused concludes a setup with each change and checks that
// it names the field, with its rule, and saves nothing.
func requireEachFieldRefused(t *testing.T, changes map[string]brokenField) {
	t.Helper()
	for name, broken := range changes {
		t.Run(name, func(t *testing.T) {
			installed := issued(t)
			input := typedSetup(firstCode)
			broken.change(&input)

			err := installed.concludeSetup.Execute(context.Background(), input)

			assert.Equal(t, map[string]string{broken.field: broken.code}, brokenFields(t, err))
			requireNothingSaved(t, installed)
		})
	}
}
