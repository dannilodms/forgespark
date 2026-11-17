package main

import (
	"maps"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// refusedField is a change to the body of concluding the setup that breaks
// the rule of one field, and the entry of `errors` that names it.
type refusedField struct {
	change   func(body map[string]any)
	location string
	code     string
}

// without removes a property of the object at a dotted path.
func without(path, property string) func(body map[string]any) {
	return func(body map[string]any) { delete(object(body, path), property) }
}

// with sets a property of the object at a dotted path.
func with(path, property string, value any) func(body map[string]any) {
	return func(body map[string]any) { object(body, path)[property] = value }
}

func TestEachFieldThatBreaksARuleAnswers422NamingItAndSavesNothing(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	before := everyRow(t, asOwner(t, temporary))
	required := strings.Fields("consultancy.name consultancy.legal_name consultancy.country consultancy.tax_id " +
		"consultancy.time_zone consultancy.currency administrator.name administrator.email administrator.country " +
		"administrator.national_id administrator.preferred_language")
	refused := map[string]refusedField{}
	for _, field := range required {
		path, property, _ := strings.Cut(field, ".")
		refused[field+" left out"] = refusedField{without(path, property), "body." + field, "required"}
		refused[field+" empty"] = refusedField{with(path, property, ""), "body." + field, "required"}
	}
	over := func(limit int) string { return strings.Repeat("ã", limit+1) }
	maps.Copy(refused, map[string]refusedField{
		"a name of spaces only":       {with("consultancy", "name", "   "), "body.consultancy.name", "required"},
		"a national ID of spaces":     {with("administrator", "national_id", " \t "), "body.administrator.national_id", "required"},
		"a country outside its list":  {with("consultancy", "country", "ZZ"), "body.consultancy.country", "country_unknown"},
		"a country of the person":     {with("administrator", "country", "UK"), "body.administrator.country", "country_unknown"},
		"a currency outside its list": {with("consultancy", "currency", "XAU"), "body.consultancy.currency", "currency_unknown"},
		"a time zone outside its list": {
			with("consultancy", "time_zone", "America/Atlantis"), "body.consultancy.time_zone", "time_zone_unknown",
		},
		"a language outside its list": {
			with("administrator", "preferred_language", "fr"), "body.administrator.preferred_language", "invalid",
		},
		"a malformed e-mail":        {with("administrator", "email", "ana.souza.fabrica.example"), "body.administrator.email", "invalid"},
		"an e-mail with a name":     {with("consultancy", "email", "Fábrica <contato@fabrica.example>"), "body.consultancy.email", "invalid"},
		"a telephone with no plus":  {with("consultancy", "telephone", "11 3333 4444"), "body.consultancy.telephone", "invalid"},
		"a telephone too short":     {with("administrator", "telephone", "+123456"), "body.administrator.telephone", "invalid"},
		"a website of another kind": {with("consultancy", "website", "ftp://fabrica.example"), "body.consultancy.website", "invalid"},
		"a website with a user":     {with("consultancy", "website", "https://ana@fabrica.example"), "body.consultancy.website", "invalid"},
		"a tax ID over its limit": {
			with("consultancy", "tax_id", over(domain.TaxIDMaxLength)), "body.consultancy.tax_id", "invalid",
		},
		"a national ID over its limit": {
			with("administrator", "national_id", over(domain.NationalIDMaxLength)), "body.administrator.national_id", "invalid",
		},
		"a name over its limit": {
			with("administrator", "name", over(domain.NameMaxLength)), "body.administrator.name", "invalid",
		},
		"a postal code over its limit": {
			with("consultancy.address", "postal_code", over(domain.PostalCodeMaxLength)), "body.consultancy.address.postal_code", "invalid",
		},
		"an unknown property": {with("consultancy", "nickname", "Fábrica"), "body.consultancy.nickname", "invalid"},
		"no setup code":       {func(body map[string]any) { delete(body, "code") }, "body.code", "required"},
	})

	for name, field := range refused {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := setupBody(code)
			field.change(body)

			answer := concludeSetup(t, api, body)

			require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
			assert.Equal(t, "application/problem+json", answer.mediaType)
			assert.Equal(t, "validation_failed", answer.fields["code"])
			assert.Regexp(t, `^[0-9a-f]{32}$`, answer.fields["trace_id"])
			fields, _ := answer.fields["errors"].([]any)
			require.Len(t, fields, 1, answer.body)
			named, _ := fields[0].(map[string]any)
			assert.Equal(t, field.location, named["location"])
			assert.Equal(t, field.code, named["code"])
			assert.Equal(t, before, everyRow(t, asOwner(t, temporary)), "a refused setup changed what is stored")
		})
	}
}

func TestARuleOfTheDomainNamesTheFieldWithTheCodeOfTheRuleAndItsParameters(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	body := setupBody(code)
	object(body, "consultancy")["country"] = "ZZ"
	object(body, "consultancy")["currency"] = "XAU"

	answer := concludeSetup(t, api, body)

	require.Equal(t, http.StatusUnprocessableEntity, answer.status, answer.body)
	traceID, _ := answer.fields["trace_id"].(string)
	assert.JSONEq(t, `{
		"title": "Unprocessable Entity",
		"status": 422,
		"detail": "validation failed",
		"code": "validation_failed",
		"params": {},
		"trace_id": "`+traceID+`",
		"errors": [
			{"location": "body.consultancy.country", "code": "country_unknown", "params": {}, "message": "the country is not an ISO 3166-1 alpha-2 code"},
			{"location": "body.consultancy.currency", "code": "currency_unknown", "params": {}, "message": "the currency is not the ISO 4217 code of a currency in circulation"}
		]
	}`, answer.body)
}
