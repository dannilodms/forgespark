package domain

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

// sampleContent is a content of each template, to write every template in
// every language.
var sampleContent = map[Template]Content{
	LoginCode: LoginCodeContent{Code: "482915", Lifetime: 10 * time.Minute},
}

// textsOf reads the translation file of a language as names of texts, one
// list for each template.
func textsOf(t *testing.T, tag string) map[string][]string {
	t.Helper()
	read, err := translationFiles.ReadFile("locales/" + tag + ".json")
	require.NoError(t, err, "the texts of e-mail in %s", tag)
	var file map[string]map[string]string
	require.NoError(t, json.Unmarshal(read, &file), "locales/%s.json", tag)
	names := map[string][]string{}
	for template, texts := range file {
		names[template] = slices.Sorted(maps.Keys(texts))
	}
	return names
}

// The build fails when a supported language lacks a text of an e-mail, as the
// app's does (E-mail in CODING_STANDARDS.md): each language has the texts of
// the default language, and nothing else.
func TestEveryLanguageHasEveryTextOfEveryTemplate(t *testing.T) {
	expected := textsOf(t, "en-US")
	for _, template := range templates {
		require.Contains(t, expected, string(template), "en-US has no texts for the template %s", template)
	}
	require.Len(t, expected, len(templates), "en-US has the texts of a template that does not exist")

	for _, tag := range kernel.SupportedLanguages() {
		t.Run(tag, func(t *testing.T) {
			assert.Equal(t, expected, textsOf(t, tag), "locales/%s.json lacks a text, or has one en-US does not", tag)
		})
	}
}

func TestEveryTemplateIsWrittenInEverySupportedLanguage(t *testing.T) {
	require.ElementsMatch(t, templates, slices.Collect(maps.Keys(sampleContent)), "a template has no sample content")
	for _, tag := range kernel.SupportedLanguages() {
		t.Run(tag, func(t *testing.T) {
			language, err := kernel.NewLanguage(tag)
			require.NoError(t, err)
			for _, template := range templates {
				requireWritten(t, template, language)
			}
		})
	}
}

func requireWritten(t *testing.T, template Template, language kernel.Language) {
	t.Helper()
	written, err := template.write(language, sampleContent[template])

	require.NoError(t, err, "%s in %s", template, language)
	subject, err := template.subjectIn(language)
	require.NoError(t, err, "the subject of %s in %s", template, language)
	assert.NotEmpty(t, subject, "the subject of %s in %s", template, language)
	assert.NotEmpty(t, written.html, "the HTML part of %s in %s", template, language)
	assert.NotEmpty(t, written.text, "the plain-text part of %s in %s", template, language)
	assert.NotContains(t, written.html+written.text, "<no value>")
}

// A subject the translation file lacks fails the e-mail, as a missing text of
// its parts does, and never leaves it with no subject (E-mail in
// CODING_STANDARDS.md).
func TestASubjectTheTranslationFileLacksFailsTheEmail(t *testing.T) {
	language, err := kernel.NewLanguage("es")
	require.NoError(t, err)
	without := &templateLibrary{texts: map[string]map[Template]map[string]string{
		"es": {LoginCode: {"lifetime": "Funciona durante {{.Minutes}} minutos."}},
	}}
	empty := &templateLibrary{texts: map[string]map[Template]map[string]string{
		"es": {LoginCode: {"subject": ""}},
	}}

	_, missing := without.subjectOf(LoginCode, language)
	_, blank := empty.subjectOf(LoginCode, language)
	_, noLanguage := (&templateLibrary{texts: map[string]map[Template]map[string]string{}}).subjectOf(LoginCode, language)

	assert.Error(t, missing)
	assert.Error(t, blank)
	assert.Error(t, noLanguage)
}
