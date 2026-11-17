package domain

import (
	"errors"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// UserID identifies a user.
type UserID int64

// NewUserID reads an identifier that came from the sequence of users.
func NewUserID(value int64) (UserID, error) {
	if value <= 0 {
		return 0, errors.New("the identifier of a user is a positive number")
	}
	return UserID(value), nil
}

// TypedUser is a user as a person typed them: in the setup, the first
// administrator. An optional field left out is empty.
type TypedUser struct {
	Name              string
	Email             string
	Country           string
	NationalID        string
	Telephone         string
	PreferredLanguage string
}

// UserDetails are the data of a user, each checked against its rule. Only
// NewUserDetails builds them.
type UserDetails struct {
	name              string
	email             kernel.Email
	country           kernel.Country
	nationalID        string
	telephone         optional[kernel.Telephone]
	preferredLanguage kernel.Language
}

// NewUserDetails checks what was typed. The name, the e-mail, the country,
// the national ID and the preferred language are required; the telephone is
// not. It returns a kernel.InvalidFieldsError that names every field breaking
// a rule.
func NewUserDetails(typed TypedUser) (UserDetails, error) {
	var invalid kernel.InvalidFieldsError
	var details UserDetails
	var err error

	details.name, err = kernel.RequiredText(typed.Name, NameMaxLength)
	invalid.Check("name", err)
	details.email, err = kernel.NewEmail(typed.Email)
	invalid.Check("email", err)
	details.country, err = kernel.NewCountry(typed.Country)
	invalid.Check("country", err)
	details.nationalID, err = kernel.RequiredText(typed.NationalID, NationalIDMaxLength)
	invalid.Check("national_id", err)
	details.telephone, err = readOptional(typed.Telephone, kernel.NewTelephone)
	invalid.Check("telephone", err)
	details.preferredLanguage, err = kernel.NewLanguage(typed.PreferredLanguage)
	invalid.Check("preferred_language", err)

	if err := invalid.Err(); err != nil {
		return UserDetails{}, err
	}
	return details, nil
}

// Name returns the name of the user.
func (d UserDetails) Name() string { return d.name }

// Email returns the e-mail the user logs in with, in lowercase.
func (d UserDetails) Email() kernel.Email { return d.email }

// Country returns the country of the user, which the national ID is read
// with.
func (d UserDetails) Country() kernel.Country { return d.country }

// NationalID returns the national ID, as typed.
func (d UserDetails) NationalID() string { return d.nationalID }

// Telephone returns the telephone of the user, and false when they have
// none.
func (d UserDetails) Telephone() (kernel.Telephone, bool) { return d.telephone.get() }

// PreferredLanguage returns the language the user reads the system in.
func (d UserDetails) PreferredLanguage() kernel.Language { return d.preferredLanguage }

// User is a person registered in the system, who logs in with their e-mail.
// A user has at least one role. A deactivated user cannot log in and keeps
// their record: no user is removed.
type User struct {
	id          UserID
	details     UserDetails
	roles       []RoleID
	deactivated bool
}

// RegisterUser creates a user with data already checked and the roles they
// receive. A user is never without a role.
func RegisterUser(id UserID, details UserDetails, roles ...RoleID) (*User, error) {
	if len(roles) == 0 {
		return nil, errors.New("a user has at least one role")
	}
	return &User{id: id, details: details, roles: slices.Clone(roles)}, nil
}

// ID returns the identifier of the user.
func (u *User) ID() UserID {
	return u.id
}

// Details returns the data of the user.
func (u *User) Details() UserDetails {
	return u.details
}

// Roles returns the roles the user has.
func (u *User) Roles() []RoleID {
	return slices.Clone(u.roles)
}

// Deactivated tells whether the user was deactivated.
func (u *User) Deactivated() bool {
	return u.deactivated
}
