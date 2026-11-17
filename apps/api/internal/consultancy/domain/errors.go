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
