// Command guard runs the checks of the build that no linter has. `code`
// judges the Go module: the dependency rule, `else` and the level of
// nesting. `versions` judges the repository: every version it pins is exact.
//
// It prints one line for each finding, `file:line: what to do`, and ends
// with status 1 when there is one. Rules: Enforcement in the backend rules
// and in Versions of CODING_STANDARDS.md.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/controlflow"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/imports"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/versions"
)

const usage = `Usage: guard <command> [directory]

Commands:
  code [module]            Check the dependency rule, else and nesting in the
                           Go module of the directory
  versions [repository]    Check that every version the repository pins is
                           exact

The directory is the current one when it is left out.
`

// check judges a directory and returns what breaks a rule in it.
type check func(directory string) ([]report.Finding, error)

// commands maps the first argument to the checks it runs.
var commands = map[string][]check{
	"code":     {imports.Check, controlflow.Check},
	"versions": {versions.Check},
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one command and returns the exit status of the process: 0
// when nothing breaks a rule, 1 when something does, and 2 when the command
// was not understood or a check could not read what it judges.
func run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 || len(arguments) > 2 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	checks, known := commands[arguments[0]]
	if !known {
		fmt.Fprintf(stderr, "guard: unknown command %q\n\n%s", arguments[0], usage)
		return 2
	}
	directory := "."
	if len(arguments) == 2 {
		directory = arguments[1]
	}
	findings, err := runChecks(checks, directory)
	if err != nil {
		fmt.Fprintf(stderr, "guard: %v\n", err)
		return 2
	}
	if len(findings) == 0 {
		return 0
	}
	for _, finding := range findings {
		fmt.Fprintln(stdout, finding)
	}
	fmt.Fprintf(stderr, "guard: %s\n", count(len(findings)))
	return 1
}

// runChecks runs each check on the directory and puts their findings in one
// list, ordered by file and by line.
func runChecks(checks []check, directory string) ([]report.Finding, error) {
	var findings []report.Finding
	var unread error
	for _, check := range checks {
		found, err := check(directory)
		unread = errors.Join(unread, err)
		findings = append(findings, found...)
	}
	if unread != nil {
		return nil, unread
	}
	report.Sort(findings)
	return findings, nil
}

func count(findings int) string {
	if findings == 1 {
		return "1 finding"
	}
	return fmt.Sprintf("%d findings", findings)
}
