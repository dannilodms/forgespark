package domain

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"slices"
	"sync"
	texttemplate "text/template"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// Template is the kind of an e-mail: what it says. Each template has an HTML
// part and a plain-text part, `templates/<template>.html` and
// `templates/<template>.txt`, and its texts in the translation file of every
// supported language, `locales/<language>.json`. The format is in E-mail in
// CODING_STANDARDS.md.
type Template string

// The templates of the e-mails the system sends.
const (
	// LoginCode is the e-mail that carries a login code.
	LoginCode Template = "login_code"
)

// templates lists every template. A new one is a line here, its two parts, its
// texts in every translation file and a value in the CHECK constraint of the
// table of e-mails.
var templates = []Template{LoginCode}

func (t Template) known() bool {
	return slices.Contains(templates, t)
}

// Content is what one e-mail says beyond the texts of its template: the
// values the template writes in. It is known only when the e-mail is
// delivered, and the area that asked for the e-mail gives it, so a value that
// must be stored nowhere, such as a login code, is in no table.
type Content interface {
	// Template is the template the content is written in.
	Template() Template
}

// LoginCodeContent is the content of the e-mail of a login code: the code and
// how long it works. The area of the login computes both.
type LoginCodeContent struct {
	Code     string
	Lifetime time.Duration
}

// Template is LoginCode.
func (LoginCodeContent) Template() Template { return LoginCode }

// Minutes is how long the code works, in whole minutes, as the texts write it.
func (c LoginCodeContent) Minutes() int {
	return int(c.Lifetime / time.Minute)
}

//go:embed templates/*.html templates/*.txt
var templateFiles embed.FS

//go:embed locales/*.json
var translationFiles embed.FS

// written is the two parts of an e-mail, in one language.
type written struct {
	html string
	text string
}

// data is what the two parts of a template read: the language, the texts of
// the template in that language, already filled with the content, and the
// content itself.
type data struct {
	Language string
	Text     map[string]string
	Content  Content
}

// subjectIn returns the subject of the e-mails of the template in the
// language. A subject takes no value from the content: it is written in the
// e-mail record when the e-mail is queued, before the content exists. A
// subject the language lacks is an error, never an e-mail with no subject.
func (t Template) subjectIn(language kernel.Language) (string, error) {
	read, err := library()
	if err != nil {
		return "", err
	}
	return read.subjectOf(t, language)
}

// subjectOf returns the subject of the template in the language, as its
// translation file writes it.
func (l *templateLibrary) subjectOf(template Template, language kernel.Language) (string, error) {
	subject := l.texts[language.String()][template]["subject"]
	if subject == "" {
		return "", fmt.Errorf("no subject of %s in %s", template, language)
	}
	return subject, nil
}

// write writes the two parts of the template in the language, with the
// content. A text the language lacks, or a value the content lacks, is an
// error, never an empty space in an e-mail.
func (t Template) write(language kernel.Language, content Content) (written, error) {
	read, err := library()
	if err != nil {
		return written{}, err
	}
	texts, err := read.textsIn(language, t, content)
	if err != nil {
		return written{}, err
	}
	filled := data{Language: language.String(), Text: texts, Content: content}

	var html, text bytes.Buffer
	if err := read.html[t].Execute(&html, filled); err != nil {
		return written{}, fmt.Errorf("write the HTML part of %s in %s: %w", t, language, err)
	}
	if err := read.text[t].Execute(&text, filled); err != nil {
		return written{}, fmt.Errorf("write the plain-text part of %s in %s: %w", t, language, err)
	}
	return written{html: html.String(), text: text.String()}, nil
}

// templateLibrary is every template and every text, read from the files the
// binary carries.
type templateLibrary struct {
	html map[Template]*htmltemplate.Template
	text map[Template]*texttemplate.Template
	// texts holds, for each language and each template, the text of each
	// name, as written in its translation file.
	texts map[string]map[Template]map[string]string
}

// library reads the files once, the first time an e-mail is written. A file
// that cannot be read fails every e-mail, and the tests of this package.
var library = sync.OnceValues(readLibrary)

func readLibrary() (*templateLibrary, error) {
	read := &templateLibrary{
		html:  map[Template]*htmltemplate.Template{},
		text:  map[Template]*texttemplate.Template{},
		texts: map[string]map[Template]map[string]string{},
	}
	var failed error
	for _, template := range templates {
		failed = errors.Join(failed, read.readParts(template))
	}
	for _, tag := range kernel.SupportedLanguages() {
		failed = errors.Join(failed, read.readTexts(tag))
	}
	if failed != nil {
		return nil, failed
	}
	return read, nil
}

func (l *templateLibrary) readParts(template Template) error {
	html, err := htmltemplate.ParseFS(templateFiles, "templates/"+string(template)+".html")
	if err != nil {
		return fmt.Errorf("read the HTML part of %s: %w", template, err)
	}
	text, err := texttemplate.ParseFS(templateFiles, "templates/"+string(template)+".txt")
	if err != nil {
		return fmt.Errorf("read the plain-text part of %s: %w", template, err)
	}
	l.html[template] = html.Option("missingkey=error")
	l.text[template] = text.Option("missingkey=error")
	return nil
}

func (l *templateLibrary) readTexts(tag string) error {
	file, err := translationFiles.ReadFile("locales/" + tag + ".json")
	if err != nil {
		return fmt.Errorf("read the texts of e-mail in %s: %w", tag, err)
	}
	var texts map[Template]map[string]string
	if err := json.Unmarshal(file, &texts); err != nil {
		return fmt.Errorf("read the texts of e-mail in %s: %w", tag, err)
	}
	l.texts[tag] = texts
	return nil
}

// textsIn fills each text of the template in the language with the content.
// A text is a template of Go's text/template over the content, so
// `{{.Minutes}}` writes a value of it.
func (l *templateLibrary) textsIn(language kernel.Language, template Template, content Content) (map[string]string, error) {
	texts, found := l.texts[language.String()][template]
	if !found {
		return nil, fmt.Errorf("no texts of %s in %s", template, language)
	}
	filled := make(map[string]string, len(texts))
	var failed error
	for name, text := range texts {
		written, err := fill(name, text, content)
		filled[name] = written
		failed = errors.Join(failed, err)
	}
	if failed != nil {
		return nil, fmt.Errorf("write the texts of %s in %s: %w", template, language, failed)
	}
	return filled, nil
}

func fill(name, text string, content Content) (string, error) {
	parsed, err := texttemplate.New(name).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", fmt.Errorf("read the text %s: %w", name, err)
	}
	var written bytes.Buffer
	if err := parsed.Execute(&written, content); err != nil {
		return "", fmt.Errorf("write the text %s: %w", name, err)
	}
	return written.String(), nil
}
