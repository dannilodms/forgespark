package kernel

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A value is looked up in its list by binary search, which needs the list
// sorted and each value once. A list edited by hand must stay so.
func TestEachListIsSortedWithEachValueOnce(t *testing.T) {
	lists := map[string][]string{
		"countries":           countries,
		"supported languages": supportedLanguages,
	}
	for name, list := range lists {
		t.Run(name, func(t *testing.T) {
			assert.True(t, slices.IsSorted(list), "the list is not sorted")
			assert.Len(t, slices.Compact(slices.Clone(list)), len(list), "a value is listed twice")
		})
	}
	assert.Len(t, countries, 249)
}

func TestTheListsAcceptWhatTheirSourcesListAndNothingElse(t *testing.T) {
	accepted := map[string]error{
		"Brazil":                  errorOf(NewCountry("BR")),
		"Antarctica":              errorOf(NewCountry("AQ")),
		"Brazilian Portuguese":    errorOf(NewLanguage("pt-BR")),
		"English of the US":       errorOf(NewLanguage("en-US")),
		"Spanish with no country": errorOf(NewLanguage("es")),
	}
	for name, err := range accepted {
		assert.NoError(t, err, name)
	}

	refused := map[string]error{
		"Kosovo, which ISO 3166-1 does not assign": errorOf(NewCountry("XK")),
		"the United Kingdom by its common name":    errorOf(NewCountry("UK")),
		"Portuguese with no country":               errorOf(NewLanguage("pt")),
	}
	for name, err := range refused {
		assert.Error(t, err, name)
	}
}

func errorOf[T any](_ T, err error) error {
	return err
}
