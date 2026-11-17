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
