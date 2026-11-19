package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

func TestAMalformedEmailTelephoneOrWebsiteIsNamedAndNothingIsSaved(t *testing.T) {
	malformed := map[string]brokenField{
		"an e-mail with no at sign": {
			field: "administrator.email", code: "email_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Email = "ana.souza.fabrica.example" },
		},
		"an e-mail with a name": {
			field: "administrator.email", code: "email_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Email = "Ana <ana@fabrica.example>" },
		},
		"an e-mail of the consultancy with two at signs": {
			field: "consultancy.email", code: "email_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Email = "contato@@fabrica.example" },
		},
		"a telephone with no plus sign": {
			field: "consultancy.telephone", code: "telephone_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Telephone = "11 3333 4444" },
		},
		"a telephone with a hyphen": {
			field: "administrator.telephone", code: "telephone_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Telephone = "+55 11 98888-7777" },
		},
		"a telephone of 6 digits": {
			field: "administrator.telephone", code: "telephone_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Telephone = "+123456" },
		},
		"a telephone of 16 digits": {
			field: "administrator.telephone", code: "telephone_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Telephone = "+1234567890123456" },
		},
		"a website of another scheme": {
			field: "consultancy.website", code: "website_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Website = "ftp://fabrica.example" },
		},
		"a website with no scheme": {
			field: "consultancy.website", code: "website_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Website = "fabrica.example" },
		},
		"a website with no host": {
			field: "consultancy.website", code: "website_invalid",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Website = "https:///path" },
		},
	}
	requireEachFieldRefused(t, malformed)
}

// over is a text one character over the limit. Each ã is two bytes and one
// character.
func over(limit int) string {
	return strings.Repeat("ã", limit+1)
}

func TestATextOverItsLimitIsNamedAndNothingIsSaved(t *testing.T) {
	tooLong := map[string]brokenField{
		"the name of the consultancy": {
			field: "consultancy.name", code: "too_long",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.Name = over(domain.NameMaxLength) },
		},
		"the legal name": {
			field: "consultancy.legal_name", code: "too_long",
			change: func(input *application.ConcludeSetupInput) {
				input.Consultancy.LegalName = over(domain.LegalNameMaxLength)
			},
		},
		"the tax ID": {
			field: "consultancy.tax_id", code: "too_long",
			change: func(input *application.ConcludeSetupInput) { input.Consultancy.TaxID = over(domain.TaxIDMaxLength) },
		},
		"the first line of the address": {
			field: "consultancy.address.line1", code: "too_long",
			change: func(input *application.ConcludeSetupInput) {
				input.Consultancy.Address.Line1 = over(domain.AddressLineMaxLength)
			},
		},
		"the postal code": {
			field: "consultancy.address.postal_code", code: "too_long",
			change: func(input *application.ConcludeSetupInput) {
				input.Consultancy.Address.PostalCode = over(domain.PostalCodeMaxLength)
			},
		},
		"the name of the administrator": {
			field: "administrator.name", code: "too_long",
			change: func(input *application.ConcludeSetupInput) { input.Administrator.Name = over(domain.NameMaxLength) },
		},
		"the national ID": {
			field: "administrator.national_id", code: "too_long",
			change: func(input *application.ConcludeSetupInput) {
				input.Administrator.NationalID = over(domain.NationalIDMaxLength)
			},
		},
		"the e-mail": {
			field: "administrator.email", code: "too_long",
			change: func(input *application.ConcludeSetupInput) {
				input.Administrator.Email = strings.Repeat("a", kernel.EmailMaxLength) + "@fabrica.example"
			},
		},
	}
	requireEachFieldRefused(t, tooLong)
}

func TestATextAtItsLimitIsAccepted(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	atTheLimit := strings.Repeat("ã", domain.TaxIDMaxLength)
	input.Consultancy.TaxID = atTheLimit
	input.Administrator.NationalID = strings.Repeat("ã", domain.NationalIDMaxLength)

	err := installed.concludeSetup.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, atTheLimit, installed.consultancies.All()[0].Details().TaxID())
}

func TestATextOverItsLimitSaysTheLimit(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.TaxID = strings.Repeat("1", domain.TaxIDMaxLength+1)

	err := installed.concludeSetup.Execute(context.Background(), input)

	var invalid kernel.InvalidFieldsError
	require.ErrorAs(t, err, &invalid)
	require.Len(t, invalid.Fields, 1)
	var tooLong kernel.TooLongError
	require.ErrorAs(t, invalid.Fields[0].Rule, &tooLong)
	assert.Equal(t, map[string]any{"max_length": domain.TaxIDMaxLength}, tooLong.Params())
}

func TestEveryFieldThatBreaksARuleIsNamedAtOnce(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.Country = "ZZ"
	input.Consultancy.Address.City = over(domain.CityMaxLength)
	input.Administrator.Email = "not an e-mail"
	input.Administrator.Name = ""

	err := installed.concludeSetup.Execute(context.Background(), input)

	assert.Equal(t, map[string]string{
		"consultancy.country":      "country_unknown",
		"consultancy.address.city": "too_long",
		"administrator.email":      "email_invalid",
		"administrator.name":       "required",
	}, brokenFields(t, err))
	requireNothingSaved(t, installed)
}

func TestTheFieldsAreCheckedBeforeTheSetupCode(t *testing.T) {
	installed := issued(t)
	input := typedSetup("ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ")
	input.Consultancy.Name = ""

	err := installed.concludeSetup.Execute(context.Background(), input)

	assert.Equal(t, map[string]string{"consultancy.name": "required"}, brokenFields(t, err))
	requireNothingSaved(t, installed)
}

func TestAnEmailTypedInUpperCaseIsStoredInLowercase(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.Email = "Contato@Fabrica.Example"
	input.Administrator.Email = "  ANA.SOUZA@FABRICA.EXAMPLE "

	err := installed.concludeSetup.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "ana.souza@fabrica.example", installed.users.All()[0].Details().Email().String())
	email, _ := installed.consultancies.All()[0].Details().Email()
	assert.Equal(t, "contato@fabrica.example", email.String())
}

func TestTheSetupConcludesWithTheOptionalFieldsLeftOut(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.Email = ""
	input.Consultancy.Telephone = ""
	input.Consultancy.Website = " "
	input.Consultancy.Address = domain.TypedAddress{}
	input.Administrator.Telephone = ""

	err := installed.concludeSetup.Execute(context.Background(), input)

	require.NoError(t, err)
	consultancy := installed.consultancies.All()[0].Details()
	_, hasEmail := consultancy.Email()
	_, hasTelephone := consultancy.Telephone()
	_, hasWebsite := consultancy.Website()
	_, hasAddress := consultancy.Address()
	assert.False(t, hasEmail || hasTelephone || hasWebsite || hasAddress, "an optional field left out was saved")
	_, hasTelephone = installed.users.All()[0].Details().Telephone()
	assert.False(t, hasTelephone)
}

func TestAnAddressKeepsThePartsTypedAndLeavesTheOthersEmpty(t *testing.T) {
	installed := issued(t)
	input := typedSetup(firstCode)
	input.Consultancy.Address = domain.TypedAddress{City: " Lisboa ", PostalCode: "1100-148"}

	err := installed.concludeSetup.Execute(context.Background(), input)

	require.NoError(t, err)
	address, hasAddress := installed.consultancies.All()[0].Details().Address()
	require.True(t, hasAddress)
	assert.Equal(t, "Lisboa", address.City())
	assert.Equal(t, "1100-148", address.PostalCode())
	assert.Empty(t, address.Line1())
	assert.Empty(t, address.Line2())
	assert.Empty(t, address.Region())
}
