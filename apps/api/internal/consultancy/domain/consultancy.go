package domain

import (
	"errors"
	"strings"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// The longest text each field of the consultancy and of a user takes, in
// characters. A tax ID and a national ID are free text, with no rule of any
// country: the limit is the only one.
const (
	NameMaxLength        = 200
	LegalNameMaxLength   = 200
	TaxIDMaxLength       = 50
	NationalIDMaxLength  = 50
	AddressLineMaxLength = 200
	CityMaxLength        = 100
	RegionMaxLength      = 100
	PostalCodeMaxLength  = 20
)

// ConsultancyID identifies the consultancy.
type ConsultancyID int64

// NewConsultancyID reads an identifier that came from the sequence of the
// consultancy.
func NewConsultancyID(value int64) (ConsultancyID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of the consultancy is a positive number")
	}
	return ConsultancyID(value), nil
}

// FileID identifies a file of the files area, which this area holds by its
// identifier: the logo of the consultancy, the photo of a user. The files
// area checks the file; this area only keeps which one it is.
type FileID int64

// NoSuchFile stands for an identifier that names no file, such as one an
// inbound adapter could not decode: no file has it, so the setup answers it
// as a file that does not exist, together with the other fields it checks.
const NoSuchFile FileID = -1

// TypedConsultancy is the consultancy as a person typed it in the setup. An
// optional field left out is empty, and no logo is the file 0.
type TypedConsultancy struct {
	Name       string
	LegalName  string
	Country    string
	TaxID      string
	TimeZone   string
	Currency   string
	Email      string
	Telephone  string
	Website    string
	Address    TypedAddress
	LogoFileID FileID
}

// TypedAddress is an address as a person typed it: two lines, the city, the
// region and the postal code, so that an address of any country fits. Every
// part may be left out, and an address with no part is no address.
type TypedAddress struct {
	Line1      string
	Line2      string
	City       string
	Region     string
	PostalCode string
}

// ConsultancyDetails are the data of the consultancy, each checked against
// its rule. Only NewConsultancyDetails builds them.
type ConsultancyDetails struct {
	name      string
	legalName string
	country   kernel.Country
	taxID     string
	timeZone  TimeZone
	currency  Currency
	email     optional[kernel.Email]
	telephone optional[Telephone]
	website   optional[Website]
	address   optional[Address]
	logo      optional[FileID]
}

// NewConsultancyDetails checks what was typed. The name, the legal name, the
// country, the tax ID, the time zone and the currency are required; the
// e-mail, the telephone, the website, the address and the logo are not. It
// returns a kernel.InvalidFieldsError that names every field breaking a rule.
func NewConsultancyDetails(typed TypedConsultancy) (ConsultancyDetails, error) {
	var invalid kernel.InvalidFieldsError
	var details ConsultancyDetails
	var err error

	details.name, err = RequiredText(typed.Name, NameMaxLength)
	invalid.AddField("name", err)
	details.legalName, err = RequiredText(typed.LegalName, LegalNameMaxLength)
	invalid.AddField("legal_name", err)
	details.country, err = kernel.NewCountry(typed.Country)
	invalid.AddField("country", err)
	details.taxID, err = RequiredText(typed.TaxID, TaxIDMaxLength)
	invalid.AddField("tax_id", err)
	details.timeZone, err = NewTimeZone(typed.TimeZone)
	invalid.AddField("time_zone", err)
	details.currency, err = NewCurrency(typed.Currency)
	invalid.AddField("currency", err)
	details.email, err = readOptional(typed.Email, kernel.NewEmail)
	invalid.AddField("email", err)
	details.telephone, err = readOptional(typed.Telephone, NewTelephone)
	invalid.AddField("telephone", err)
	details.website, err = readOptional(typed.Website, NewWebsite)
	invalid.AddField("website", err)
	details.address, err = readAddress(typed.Address)
	invalid.AddFieldsOf("address", err)
	details.logo = chosenFile(typed.LogoFileID)

	if err := invalid.Err(); err != nil {
		return ConsultancyDetails{}, err
	}
	return details, nil
}

// chosenFile gives the file chosen, and none for the file 0.
func chosenFile(id FileID) optional[FileID] {
	return optional[FileID]{value: id, present: id != 0}
}

// Name returns the name the consultancy is known by.
func (d ConsultancyDetails) Name() string { return d.name }

// LegalName returns the name documents carry.
func (d ConsultancyDetails) LegalName() string { return d.legalName }

// Country returns the country of the consultancy.
func (d ConsultancyDetails) Country() kernel.Country { return d.country }

// TaxID returns the tax ID, as typed.
func (d ConsultancyDetails) TaxID() string { return d.taxID }

// TimeZone returns the time zone of the consultancy.
func (d ConsultancyDetails) TimeZone() TimeZone { return d.timeZone }

// Currency returns the currency of the consultancy.
func (d ConsultancyDetails) Currency() Currency { return d.currency }

// Email returns the e-mail of the consultancy, and false when it has none.
func (d ConsultancyDetails) Email() (kernel.Email, bool) { return d.email.get() }

// Telephone returns the telephone of the consultancy, and false when it has
// none.
func (d ConsultancyDetails) Telephone() (Telephone, bool) { return d.telephone.get() }

// Website returns the website of the consultancy, and false when it has none.
func (d ConsultancyDetails) Website() (Website, bool) { return d.website.get() }

// Address returns the address of the consultancy, and false when it has
// none.
func (d ConsultancyDetails) Address() (Address, bool) { return d.address.get() }

// LogoFileID returns the file of the logo of the consultancy, and false when
// it has none.
func (d ConsultancyDetails) LogoFileID() (FileID, bool) { return d.logo.get() }

// Address is the address of the consultancy. Every part is free text, and
// any of them may be empty.
type Address struct {
	line1      string
	line2      string
	city       string
	region     string
	postalCode string
}

// readAddress checks every part of an address, and gives no address when no
// part was typed.
func readAddress(typed TypedAddress) (optional[Address], error) {
	var invalid kernel.InvalidFieldsError
	var address Address
	var err error

	address.line1, err = OptionalText(typed.Line1, AddressLineMaxLength)
	invalid.AddField("line1", err)
	address.line2, err = OptionalText(typed.Line2, AddressLineMaxLength)
	invalid.AddField("line2", err)
	address.city, err = OptionalText(typed.City, CityMaxLength)
	invalid.AddField("city", err)
	address.region, err = OptionalText(typed.Region, RegionMaxLength)
	invalid.AddField("region", err)
	address.postalCode, err = OptionalText(typed.PostalCode, PostalCodeMaxLength)
	invalid.AddField("postal_code", err)

	if err := invalid.Err(); err != nil {
		return optional[Address]{}, err
	}
	return optional[Address]{value: address, present: address != Address{}}, nil
}

// Line1 returns the first line of the address.
func (a Address) Line1() string { return a.line1 }

// Line2 returns the second line of the address.
func (a Address) Line2() string { return a.line2 }

// City returns the city.
func (a Address) City() string { return a.city }

// Region returns the region: a state, a province, a county.
func (a Address) Region() string { return a.region }

// PostalCode returns the postal code, as typed.
func (a Address) PostalCode() string { return a.postalCode }

// Consultancy is the software company the system is run for. There is
// exactly one: the setup registers it.
type Consultancy struct {
	id      ConsultancyID
	details ConsultancyDetails
}

// RegisterConsultancy creates the consultancy with data already checked.
func RegisterConsultancy(id ConsultancyID, details ConsultancyDetails) *Consultancy {
	return &Consultancy{id: id, details: details}
}

// ID returns the identifier of the consultancy.
func (c *Consultancy) ID() ConsultancyID {
	return c.id
}

// Details returns the data of the consultancy.
func (c *Consultancy) Details() ConsultancyDetails {
	return c.details
}

// optional holds a value that may be left out.
type optional[T any] struct {
	value   T
	present bool
}

func (o optional[T]) get() (T, bool) {
	return o.value, o.present
}

// readOptional reads a value that may be left out with the constructor of its
// type: an empty text is no value, and anything else must follow the rule.
func readOptional[T any](typed string, read func(string) (T, error)) (optional[T], error) {
	if strings.TrimSpace(typed) == "" {
		return optional[T]{}, nil
	}
	value, err := read(typed)
	if err != nil {
		return optional[T]{}, err
	}
	return optional[T]{value: value, present: true}, nil
}
