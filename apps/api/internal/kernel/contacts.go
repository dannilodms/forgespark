package kernel

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

// EmailMaxLength is the longest e-mail address there is (RFC 5321).
const EmailMaxLength = 254

// Email is an e-mail address, in lowercase, so that the way it was typed
// does not matter where it is compared.
type Email struct {
	address string
}

// NewEmail reads an e-mail address: one address and nothing around it, no
// name and no angle brackets.
func NewEmail(typed string) (Email, error) {
	address, err := RequiredText(typed, EmailMaxLength)
	if err != nil {
		return Email{}, err
	}
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Name != "" || parsed.Address != address {
		return Email{}, EmailInvalidError{}
	}
	return Email{address: strings.ToLower(address)}, nil
}

// String returns the address.
func (e Email) String() string {
	return e.address
}

// EmailInvalidError is a text that is not an e-mail address.
type EmailInvalidError struct{}

func (EmailInvalidError) Error() string {
	return "the text is not an e-mail address"
}

// Code returns the stable code an inbound adapter answers with.
func (EmailInvalidError) Code() string {
	return "email_invalid"
}

// TelephonePattern is the form of a telephone, as a regular expression: a
// plus sign and 7 to 15 digits (E.164), with spaces anywhere after the plus.
const TelephonePattern = `^\+( *[0-9]){7,15} *$`

var telephoneForm = regexp.MustCompile(TelephonePattern)

// Telephone is a telephone number in international form: a plus sign and its
// digits, with no space. It is checked for its form only, with no rule of any
// country.
type Telephone struct {
	number string
}

// NewTelephone reads a telephone number. The spaces typed in it are dropped.
func NewTelephone(typed string) (Telephone, error) {
	number := strings.TrimSpace(typed)
	if number == "" {
		return Telephone{}, RequiredError{}
	}
	if !telephoneForm.MatchString(number) {
		return Telephone{}, TelephoneInvalidError{}
	}
	return Telephone{number: strings.ReplaceAll(number, " ", "")}, nil
}

// String returns the number, with no space.
func (t Telephone) String() string {
	return t.number
}

// TelephoneInvalidError is a text that is not a telephone in international
// form.
type TelephoneInvalidError struct{}

func (TelephoneInvalidError) Error() string {
	return "the text is not a telephone in international form: a plus sign and 7 to 15 digits"
}

// Code returns the stable code an inbound adapter answers with.
func (TelephoneInvalidError) Code() string {
	return "telephone_invalid"
}

// WebsiteMaxLength is the longest address of a website the system keeps.
const WebsiteMaxLength = 2048

// Website is the address of a website: an http or https address with a host.
type Website struct {
	address string
}

// NewWebsite reads the address of a website. It is kept as typed.
func NewWebsite(typed string) (Website, error) {
	address, err := RequiredText(typed, WebsiteMaxLength)
	if err != nil {
		return Website{}, err
	}
	parsed, err := url.Parse(address)
	if err != nil || !isWebAddress(address, parsed) {
		return Website{}, WebsiteInvalidError{}
	}
	return Website{address: address}, nil
}

// isWebAddress tells whether the text is an http or https address with a
// host, and no space and no user in it.
func isWebAddress(address string, parsed *url.URL) bool {
	scheme := strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://")
	return scheme && parsed.Host != "" && parsed.User == nil && !strings.ContainsAny(address, " \t\n")
}

// String returns the address.
func (w Website) String() string {
	return w.address
}

// WebsiteInvalidError is a text that is not an http or https address.
type WebsiteInvalidError struct{}

func (WebsiteInvalidError) Error() string {
	return "the text is not an http or https address"
}

// Code returns the stable code an inbound adapter answers with.
func (WebsiteInvalidError) Code() string {
	return "website_invalid"
}
