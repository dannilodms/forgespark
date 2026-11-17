// Package kernel holds the value objects more than one domain area uses: the
// supported language, the country, the currency, the time zone, the e-mail,
// the telephone and the website, and the rules a field of an input breaks.
// It imports the standard library only.
package kernel

import (
	"errors"
	"strings"
)

// BrokenRule is a rule a value breaks: an error with a stable code, whose text
// is in English and safe to answer with. One with parameters also has
// `Params() map[string]any`.
type BrokenRule interface {
	error
	Code() string
}

// FieldError is one field of an input and the rule it breaks. The field is
// named in snake_case, with a dot between an object and its field, as the
// input is structured: `consultancy.address.city`.
type FieldError struct {
	Field string
	Rule  BrokenRule
}

// InvalidFieldsError lists the fields of an input that break a rule, each
// with the rule it breaks. A use case validates every field before it saves
// anything, so a caller learns of every field to correct at once.
type InvalidFieldsError struct {
	Fields []FieldError
}

func (e InvalidFieldsError) Error() string {
	broken := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		broken = append(broken, field.Field+": "+field.Rule.Error())
	}
	return "invalid fields: " + strings.Join(broken, "; ")
}

// Code returns the stable code of a request whose fields break a rule.
func (InvalidFieldsError) Code() string {
	return "validation_failed"
}

// Check records the field when err is a broken rule, and nothing when err is
// nil. An error that is not a broken rule is recorded as one that names no
// rule of its own.
func (e *InvalidFieldsError) Check(field string, err error) {
	if err == nil {
		return
	}
	var broken BrokenRule
	if !errors.As(err, &broken) {
		broken = unnamedRule{err}
	}
	e.Fields = append(e.Fields, FieldError{Field: field, Rule: broken})
}

// Include records the fields an object broke, each named as a field of that
// object: `country` becomes `consultancy.country`. An error that lists no
// fields is recorded as the object itself breaking a rule.
func (e *InvalidFieldsError) Include(object string, err error) {
	var invalid InvalidFieldsError
	if !errors.As(err, &invalid) {
		e.Check(object, err)
		return
	}
	for _, field := range invalid.Fields {
		e.Fields = append(e.Fields, FieldError{Field: object + "." + field.Field, Rule: field.Rule})
	}
}

// Err returns the error when a field broke a rule, and nil when none did.
func (e InvalidFieldsError) Err() error {
	if len(e.Fields) == 0 {
		return nil
	}
	return e
}

// unnamedRule stands for an error that is not a broken rule. No operation
// declares its code, so an inbound adapter answers it as a failure of the
// system and never as a field to correct.
type unnamedRule struct {
	error
}

func (unnamedRule) Code() string {
	return "unnamed_rule"
}
