package kernel

import (
	"slices"
	"strings"
)

// Country is a country, by its ISO 3166-1 alpha-2 code: `BR`.
type Country struct {
	code string
}

// NewCountry reads the code of a country. It accepts the codes of the list
// the backend carries, in upper case.
func NewCountry(code string) (Country, error) {
	if err := listed(code, countries, CountryUnknownError{}); err != nil {
		return Country{}, err
	}
	return Country{code: code}, nil
}

// String returns the code of the country.
func (c Country) String() string {
	return c.code
}

// Countries returns the codes of the countries the backend accepts, which
// the app offers (`task api:lists`).
func Countries() []string {
	return append([]string(nil), countries...)
}

// CountryUnknownError is a code that is not a country of ISO 3166-1.
type CountryUnknownError struct{}

func (CountryUnknownError) Error() string {
	return "the country is not an ISO 3166-1 alpha-2 code"
}

// Code returns the stable code an inbound adapter answers with.
func (CountryUnknownError) Code() string {
	return "country_unknown"
}

// listed refuses a value that is empty, as a required field, and one that is
// not on the list, with the rule of that list.
func listed(value string, list []string, unknown BrokenRule) error {
	if strings.TrimSpace(value) == "" {
		return RequiredError{}
	}
	if _, found := slices.BinarySearch(list, value); !found {
		return unknown
	}
	return nil
}
