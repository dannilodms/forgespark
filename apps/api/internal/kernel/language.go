package kernel

// supportedLanguages are the languages the system offers, by their BCP 47
// tag: the single list of ADR-0005. English is the default language. The
// contract carries it as the values of `preferred_language`, and the app
// reads it from the client generated from the contract.
//
// Adding a language is a tag here, `task api:spec` and `task web:generate`,
// its name in `apps/web/src/shared/locales/languages.ts` and its file in
// every `locales` folder of the app; `task check` fails until each is done.
// It is never a change to the database.
var supportedLanguages = []string{"en-US", "es", "pt-BR"}

// Language is a supported language: one the system offers for its
// interface, its messages and its configuration content.
type Language struct {
	tag string
}

// NewLanguage reads the tag of a supported language, written as the list
// writes it.
func NewLanguage(tag string) (Language, error) {
	if err := listed(tag, supportedLanguages, LanguageUnsupportedError{}); err != nil {
		return Language{}, err
	}
	return Language{tag: tag}, nil
}

// SupportedLanguages returns the tags of the supported languages.
func SupportedLanguages() []string {
	return append([]string(nil), supportedLanguages...)
}

// String returns the tag of the language.
func (l Language) String() string {
	return l.tag
}

// LanguageUnsupportedError is a language the system does not offer.
type LanguageUnsupportedError struct{}

func (LanguageUnsupportedError) Error() string {
	return "the language is not a supported language"
}

// Code returns the stable code an inbound adapter answers with.
func (LanguageUnsupportedError) Code() string {
	return "language_unsupported"
}
