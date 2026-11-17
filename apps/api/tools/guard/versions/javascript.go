package versions

import (
	"regexp"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/software-factory-manager/apps/api/tools/guard/report"
)

// npmVersion is one version of a package: three numbers, and the name of a
// pre-release when the package is adopted before its stable release.
const npmVersion = `\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?`

const npmAdvice = "write one version, with no range and no tag"

// exactPackage is one version, or another package at one version
// (`npm:name@1.2.3`).
var exactPackage = regexp.MustCompile(`^(npm:.+@)?` + npmVersion + `$`)

// isExactDependency tells whether what a dependency is set to names one
// version. `catalog:` points at the catalog, which is judged where it is.
func isExactDependency(version string) bool {
	return strings.HasPrefix(version, "catalog:") || exactPackage.MatchString(version)
}

// pnpmWorkspace judges the catalog, the named catalogs and the overrides.
func pnpmWorkspace(content []byte) ([]report.Finding, error) {
	document, err := parse(content)
	if err != nil {
		return nil, err
	}
	sections := []*yaml.Node{child(document, "catalog"), child(document, "overrides")}
	for _, named := range fields(child(document, "catalogs")) {
		sections = append(sections, named.value)
	}
	return judgeSections(sections), nil
}

// packageManifest judges what a `package.json` sets its dependencies to, and
// the runtime and the package manager it names.
func packageManifest(content []byte) ([]report.Finding, error) {
	document, err := parse(content)
	if err != nil {
		return nil, err
	}
	findings := judgeSections([]*yaml.Node{
		child(document, "dependencies"),
		child(document, "devDependencies"),
		child(document, "optionalDependencies"),
	})
	for _, engine := range mappings(child(document, "devEngines")) {
		findings = append(findings, judgeEngine(engine)...)
	}
	return append(findings, judgePackageManager(child(document, "packageManager"))...), nil
}

func judgeSections(sections []*yaml.Node) []report.Finding {
	var findings []report.Finding
	for _, section := range sections {
		findings = append(findings, judgeDependencies(section)...)
	}
	return findings
}

func judgeDependencies(section *yaml.Node) []report.Finding {
	var findings []report.Finding
	for _, dependency := range fields(section) {
		findings = append(findings, judgeDependency(dependency)...)
	}
	return findings
}

func judgeDependency(dependency field) []report.Finding {
	if !isText(dependency.value) || isExactDependency(dependency.value.Value) {
		return nil
	}
	return []report.Finding{inexact(dependency.value.Line, dependency.key, dependency.value.Value, npmAdvice)}
}

// judgeEngine judges the `version` of an entry of `devEngines`.
func judgeEngine(engine *yaml.Node) []report.Finding {
	name, version := child(engine, "name"), child(engine, "version")
	if !isText(name) || !isText(version) || exactPackage.MatchString(version.Value) {
		return nil
	}
	return []report.Finding{inexact(version.Line, name.Value, version.Value, npmAdvice)}
}

// judgePackageManager judges the `packageManager` field: `pnpm@1.2.3`.
func judgePackageManager(node *yaml.Node) []report.Finding {
	if !isText(node) {
		return nil
	}
	name, version, _ := strings.Cut(node.Value, "@")
	if exactPackage.MatchString(version) {
		return nil
	}
	return []report.Finding{inexact(node.Line, name, version, npmAdvice)}
}
