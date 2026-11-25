package restapi

import (
	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// Language is a supported language in a body, the preferred language of a
// user. Its schema lists the supported languages of kernel, so the generated
// client knows them, and asks at least one character, so an empty one is
// answered as a required field left empty.
type Language string

// Schema lists the supported languages.
func (Language) Schema(huma.Registry) *huma.Schema {
	tags := kernel.SupportedLanguages()
	enum := make([]any, 0, len(tags))
	for _, tag := range tags {
		enum = append(enum, tag)
	}
	atLeastOne := 1
	return &huma.Schema{Type: huma.TypeString, MinLength: &atLeastOne, Enum: enum}
}
