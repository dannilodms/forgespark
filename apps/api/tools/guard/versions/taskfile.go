package versions

import (
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/software-factory-manager/apps/api/tools/guard/report"
)

// taskfile judges a Taskfile: a tool a target runs from an image, such as
// the linters of the workflows, takes its version from a variable under
// `vars` whose name ends in _VERSION, at the top or in a target.
func taskfile(content []byte) ([]report.Finding, error) {
	document, err := parse(content)
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	for _, mapping := range mappings(document) {
		findings = append(findings, judgeTaskVariables(child(mapping, "vars"))...)
	}
	return findings, nil
}

func judgeTaskVariables(variables *yaml.Node) []report.Finding {
	return judgeToolVersions(variables, func(name string) bool { return strings.HasSuffix(name, "_VERSION") })
}
