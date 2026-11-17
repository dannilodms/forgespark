// Package report holds what a guard says about a file that breaks a rule.
package report

import (
	"cmp"
	"fmt"
	"slices"
)

// Finding is one broken rule: where it is and what to do about it.
type Finding struct {
	// File is the path from the directory the guard was given, with forward
	// slashes.
	File string
	// Line is zero when the finding is about the whole file.
	Line    int
	Message string
}

func (f Finding) String() string {
	if f.Line == 0 {
		return fmt.Sprintf("%s: %s", f.File, f.Message)
	}
	return fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Message)
}

// Sort orders the findings by file, then by line, so a guard says the same
// thing on every run.
func Sort(findings []Finding) {
	slices.SortStableFunc(findings, func(a, b Finding) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line))
	})
}
