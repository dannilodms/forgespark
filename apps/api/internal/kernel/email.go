package kernel

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

// EmailMaxLength is the longest e-mail address there is (RFC 5321).
const EmailMaxLength = 254

// Email is an e-mail address, in lowercase, so that the way it was typed
// does not matter where it is compared. The consultancy area keeps it for a
// user and for the consultancy, and the access area reads the one typed at
// the login with the same rule.
type Email struct {
	address string
}

// NewEmail reads an e-mail address: one address and nothing around it, no
// name and no angle brackets. The spaces around it are dropped. One left
// empty breaks RequiredError, and one longer than EmailMaxLength breaks
// TooLongError.
func NewEmail(typed string) (Email, error) {
	address := strings.TrimSpace(typed)
	if address == "" {
		return Email{}, RequiredError{}
	}
	if utf8.RuneCountInString(address) > EmailMaxLength {
		return Email{}, TooLongError{MaxLength: EmailMaxLength}
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
