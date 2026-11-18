package versions

import (
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/software-factory-manager/apps/api/tools/guard/report"
)

// taskfile judges a Taskfile, in the variables under `vars`, at the top or in
// a target: an image a target runs, such as a linter of the workflows, is a
// variable whose name ends in _IMAGE, judged as any image; a tool a target
// downloads takes its version from a variable whose name ends in _VERSION.
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
	findings := judgeToolVersions(variables, func(name string) bool { return strings.HasSuffix(name, "_VERSION") })
	for _, read := range fields(variables) {
		findings = append(findings, judgeTaskImage(read)...)
	}
	return findings
}

// judgeTaskImage judges a variable that names an image.
func judgeTaskImage(read field) []report.Finding {
	if !strings.HasSuffix(read.key, "_IMAGE") {
		return nil
	}
	return judgeImage(read.value)
}
