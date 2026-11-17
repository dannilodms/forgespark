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

// CountryUnknownError is a code that is not a country of ISO 3166-1.
type CountryUnknownError struct{}

func (CountryUnknownError) Error() string {
	return "the country is not an ISO 3166-1 alpha-2 code"
}

// Code returns the stable code an inbound adapter answers with.
func (CountryUnknownError) Code() string {
	return "country_unknown"
}

// Currency is a currency in circulation, by its ISO 4217 code: `BRL`.
type Currency struct {
	code string
}

// NewCurrency reads the code of a currency. It accepts the codes of the list
// the backend carries, in upper case.
func NewCurrency(code string) (Currency, error) {
	if err := listed(code, currencies, CurrencyUnknownError{}); err != nil {
		return Currency{}, err
	}
	return Currency{code: code}, nil
}

// String returns the code of the currency.
func (c Currency) String() string {
	return c.code
}

// CurrencyUnknownError is a code that is not a currency in circulation.
type CurrencyUnknownError struct{}

func (CurrencyUnknownError) Error() string {
	return "the currency is not the ISO 4217 code of a currency in circulation"
}

// Code returns the stable code an inbound adapter answers with.
func (CurrencyUnknownError) Code() string {
	return "currency_unknown"
}

// TimeZone is a time zone, by its name in the IANA time zone database:
// `America/Sao_Paulo`.
type TimeZone struct {
	name string
}

// NewTimeZone reads the name of a time zone. It accepts the names of the list
// the backend carries, written as the database writes them.
func NewTimeZone(name string) (TimeZone, error) {
	if err := listed(name, timeZones, TimeZoneUnknownError{}); err != nil {
		return TimeZone{}, err
	}
	return TimeZone{name: name}, nil
}

// String returns the name of the time zone.
func (z TimeZone) String() string {
	return z.name
}

// TimeZoneUnknownError is a name that is not in the IANA time zone database.
type TimeZoneUnknownError struct{}

func (TimeZoneUnknownError) Error() string {
	return "the time zone is not a name of the IANA time zone database"
}

// Code returns the stable code an inbound adapter answers with.
func (TimeZoneUnknownError) Code() string {
	return "time_zone_unknown"
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
