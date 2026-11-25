package domain

import (
	"strings"
	"unicode/utf8"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// RequiredText reads a text that must be there, such as a name or a tax ID.
// The spaces around it are dropped and nothing else is changed: it is stored
// as typed. Its length is counted in characters. One left empty breaks
// kernel.RequiredError, and one too long kernel.TooLongError.
func RequiredText(typed string, maxLength int) (string, error) {
	text := strings.TrimSpace(typed)
	if text == "" {
		return "", kernel.RequiredError{}
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
		return "", kernel.TooLongError{MaxLength: maxLength}
	}
	return text, nil
}
