// Package rest maps a thing to its API type. Its path is the one of an
// adapter package, so the mapping check judges the types it declares: a
// struct literal that leaves a field out is rejected.
package rest

import "github.com/dannilodms/forgespark/apps/api/tools/guard/testdata/linter/internal/things/domain"

// Body is the API type of a thing.
type Body struct {
	Name      string
	LegalName string
}

// Complete maps every field, which is accepted.
func Complete(thing domain.Thing) Body {
	return Body{Name: thing.Name, LegalName: thing.LegalName}
}

// Incomplete leaves LegalName out.
func Incomplete(thing domain.Thing) Body {
	return Body{Name: thing.Name}
}

// Failed returns the zero value beside an error, which is accepted.
func Failed(err error) (Body, error) {
	return Body{}, err
}

// Renamed builds a type of the domain and leaves a field out. The check
// judges the types of adapter packages only, so this is accepted.
func Renamed(name string) domain.Thing {
	return domain.Thing{Name: name}
}
