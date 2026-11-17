package domain

// SetupCodeInvalidError is the answer to a setup code that opens nothing: one
// that is wrong, one that expired and one that was ended. The three get the
// same answer, so that nobody learns from it whether a code ever existed.
type SetupCodeInvalidError struct{}

func (SetupCodeInvalidError) Error() string {
	return "the setup code is wrong, expired or was ended"
}

// Code returns the stable code an inbound adapter answers with.
func (SetupCodeInvalidError) Code() string {
	return "setup_code_invalid"
}

// SetupAlreadyConcludedError is the answer to concluding the first setup when
// the consultancy already exists: someone concluded it first, or the setup
// code opens the registration of a new administrator.
type SetupAlreadyConcludedError struct{}

func (SetupAlreadyConcludedError) Error() string {
	return "the setup is already concluded"
}

// Code returns the stable code an inbound adapter answers with.
func (SetupAlreadyConcludedError) Code() string {
	return "setup_already_concluded"
}

// SetupNotConcludedError is the answer to registering a new administrator
// with a setup code while the consultancy does not exist: the setup code
// opens the first setup instead.
type SetupNotConcludedError struct{}

func (SetupNotConcludedError) Error() string {
	return "the setup is not concluded"
}

// Code returns the stable code an inbound adapter answers with.
func (SetupNotConcludedError) Code() string {
	return "setup_not_concluded"
}

// EmailInUseError is the answer to registering a user with an e-mail that
// belongs to a user already, deactivated or not: an e-mail is unique among
// users, and the setup changes no user it did not create.
type EmailInUseError struct{}

func (EmailInUseError) Error() string {
	return "the e-mail belongs to a user"
}

// Code returns the stable code an inbound adapter answers with.
func (EmailInUseError) Code() string {
	return "email_in_use"
}
