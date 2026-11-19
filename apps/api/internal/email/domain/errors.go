package domain

// RecipientInvalidError is a recipient that is not an e-mail address. It
// never repeats the address.
type RecipientInvalidError struct{}

func (RecipientInvalidError) Error() string {
	return "the recipient is not an e-mail address"
}

// Code returns the stable code an inbound adapter answers with.
func (RecipientInvalidError) Code() string {
	return "recipient_invalid"
}

// OperationNotAllowedError is an operation the status of an e-mail does not
// allow, such as marking failed an e-mail that was sent.
type OperationNotAllowedError struct {
	Operation Operation
	Status    Status
}

func (e OperationNotAllowedError) Error() string {
	return "an e-mail that is " + string(e.Status) + " cannot take the operation " + string(e.Operation)
}

// Code returns the stable code an inbound adapter answers with.
func (OperationNotAllowedError) Code() string {
	return "email_operation_not_allowed"
}

// Params gives the operation and the status that refused it.
func (e OperationNotAllowedError) Params() map[string]any {
	return map[string]any{"operation": string(e.Operation), "status": string(e.Status)}
}
