package kernel

import (
	"strings"
	"unicode/utf8"
)

// RequiredError is a field that is required and was left empty.
type RequiredError struct{}

func (RequiredError) Error() string {
	return "the field is required"
}

// Code returns the stable code an inbound adapter answers with.
func (RequiredError) Code() string {
	return "required"
}

// TooLongError is a text longer than its field allows.
type TooLongError struct {
	MaxLength int
}

func (TooLongError) Error() string {
	return "the text is longer than the field allows"
}

// Code returns the stable code an inbound adapter answers with.
func (TooLongError) Code() string {
	return "too_long"
}

// Params gives the longest text the field allows, in characters.
func (e TooLongError) Params() map[string]any {
	return map[string]any{"max_length": e.MaxLength}
}

// RequiredText reads a text that must be there, such as a name or a tax ID.
// The spaces around it are dropped and nothing else is changed: it is stored
// as typed. Its length is counted in characters.
func RequiredText(typed string, maxLength int) (string, error) {
	text := strings.TrimSpace(typed)
	if text == "" {
		return "", RequiredError{}
	}
	return limited(text, maxLength)
}

// OptionalText reads a text that may be left out. An empty text, or one of
// spaces only, is no text.
func OptionalText(typed string, maxLength int) (string, error) {
	return limited(strings.TrimSpace(typed), maxLength)
}

func limited(text string, maxLength int) (string, error) {
	if utf8.RuneCountInString(text) > maxLength {
		return "", TooLongError{MaxLength: maxLength}
	}
	return text, nil
}
