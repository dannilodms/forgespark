package versions

import (
	"regexp"
	"strings"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

// directive is how a line of a module file that holds a version is judged.
type directive struct {
	exact  *regexp.Regexp
	advice string
}

// directives lists the lines of a module file that hold a version. A `tool`
// line holds none: the version of a tool is the one its module is required
// at.
var directives = map[string]directive{
	"go": {
		exact:  regexp.MustCompile(`^\d+\.\d+\.\d+$`),
		advice: "write the three numbers of the Go release",
	},
	"toolchain": {
		exact:  regexp.MustCompile(`^go\d+\.\d+\.\d+$`),
		advice: "write it as go1.2.3",
	},
	"require": {
		exact:  regexp.MustCompile(`^v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+incompatible)?$`),
		advice: "write the version of one release, as v1.2.3",
	},
}

// goModule judges `go.mod` and any other module file, such as the one that
// holds the linter: the release of Go, the toolchain, and each required
// module, which is where the version of a library and of a tool lives.
func goModule(content []byte) ([]report.Finding, error) {
	var findings []report.Finding
	block := ""
	for number, line := range strings.Split(string(content), "\n") {
		code, _, _ := strings.Cut(line, "//")
		words := strings.Fields(code)
		findings = append(findings, judgeModuleLine(number+1, words, block)...)
		block = blockAfter(words, block)
	}
	return findings, nil
}

// blockAfter follows the parentheses of a module file: it returns the
// directive whose block is open after the line.
func blockAfter(words []string, block string) string {
	if len(words) == 2 && words[1] == "(" {
		return words[0]
	}
	if len(words) == 1 && words[0] == ")" {
		return ""
	}
	return block
}

func judgeModuleLine(line int, words []string, block string) []report.Finding {
	// A line inside a block is read as if it started with the directive.
	if block != "" {
		words = append([]string{block}, words...)
	}
	if len(words) < 2 || words[len(words)-1] == "(" || words[len(words)-1] == ")" {
		return nil
	}
	rule, holdsVersion := directives[words[0]]
	version := words[len(words)-1]
	if !holdsVersion || rule.exact.MatchString(version) {
		return nil
	}
	return []report.Finding{inexact(line, words[len(words)-2], version, rule.advice)}
}
