package versions

import (
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

var (
	pinnedAction = regexp.MustCompile(`^[^@\s]+@[0-9a-f]{40}$`)
	toolVersion  = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)
)

// workflow judges a workflow, or any other YAML file under `.github`: the
// actions it uses, the versions of the tools it installs, the images it
// runs and the system of its runners. A key is judged wherever it is, so a
// composite action and a reusable workflow are covered too.
func workflow(content []byte) ([]report.Finding, error) {
	document, err := parse(content)
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	for _, mapping := range mappings(document) {
		findings = append(findings, judgeWorkflowKeys(mapping)...)
	}
	return findings, nil
}

func judgeWorkflowKeys(mapping *yaml.Node) []report.Finding {
	var findings []report.Finding
	for _, read := range fields(mapping) {
		findings = append(findings, judgeWorkflowKey(read)...)
	}
	return findings
}

func judgeWorkflowKey(read field) []report.Finding {
	switch read.key {
	case "uses":
		return judgeAction(read.value)
	case "env":
		// A tool a step installs takes its version from a variable whose
		// name ends in _VERSION.
		return judgeToolVersions(read.value, func(name string) bool { return strings.HasSuffix(name, "_VERSION") })
	case "with":
		return judgeToolVersions(read.value, func(name string) bool { return name == "version" || strings.HasSuffix(name, "-version") })
	case "runs-on":
		return judgeRunners(read.value)
	case "image", "container":
		return judgeImage(read.value)
	}
	return nil
}

// judgeAction judges what a `uses` names: an action or a reusable workflow
// by its full commit hash, an image by its exact version. A path of the
// repository has no version.
func judgeAction(node *yaml.Node) []report.Finding {
	if !isText(node) || strings.HasPrefix(node.Value, "./") || pinnedAction.MatchString(node.Value) {
		return nil
	}
	image, isImage := strings.CutPrefix(node.Value, "docker://")
	if isImage && isExactImage(image) {
		return nil
	}
	if isImage {
		return []report.Finding{imageFinding(node.Line, image)}
	}
	return []report.Finding{{
		Line:    node.Line,
		Message: fmt.Sprintf("action %q is not pinned: write its full commit hash after the @%s", node.Value, pointer),
	}}
}

// judgeToolVersions judges the values of a mapping whose key names a
// version.
func judgeToolVersions(node *yaml.Node, namesVersion func(name string) bool) []report.Finding {
	var findings []report.Finding
	for _, read := range fields(node) {
		findings = append(findings, judgeToolVersion(read, namesVersion(read.key))...)
	}
	return findings
}

func judgeToolVersion(read field, namesVersion bool) []report.Finding {
	if !namesVersion || !isText(read.value) || toolVersion.MatchString(read.value.Value) {
		return nil
	}
	return []report.Finding{inexact(read.value.Line, read.key, read.value.Value, "write its three numbers")}
}

// judgeRunners rejects a runner label that follows the newest system.
func judgeRunners(node *yaml.Node) []report.Finding {
	var findings []report.Finding
	for _, label := range texts(node) {
		findings = append(findings, judgeRunner(label)...)
	}
	return findings
}

func judgeRunner(label *yaml.Node) []report.Finding {
	if !strings.HasSuffix(label.Value, "-latest") {
		return nil
	}
	return []report.Finding{{
		Line:    label.Line,
		Message: fmt.Sprintf("runner %q is a floating tag: name the version of the system, as ubuntu-24.04 does%s", label.Value, pointer),
	}}
}
