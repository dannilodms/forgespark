package versions

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

// launchers are the words after which a command line names a package it
// downloads and runs: `npx name@1.2.3` and `pnpm dlx name@1.2.3`.
var launchers = map[string]bool{
	"npx": true,
	"dlx": true,
}

// exactSpec is a package with one version, with or without a scope.
var exactSpec = regexp.MustCompile(`^(@[^/@\s]+/)?[^/@\s]+@` + npmVersion + `$`)

// mcpServers judges `.mcp.json`: a server that is downloaded when it starts
// names its package with an exact version.
func mcpServers(content []byte) ([]report.Finding, error) {
	document, err := parse(content)
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	for _, server := range fields(child(document, "mcpServers")) {
		findings = append(findings, judgeServer(server)...)
	}
	return findings, nil
}

// judgeServer reads the command line of a server. A server that downloads
// nothing, such as one reached at an address, holds no version.
func judgeServer(server field) []report.Finding {
	words := append(texts(child(server.value, "command")), texts(child(server.value, "args"))...)
	launcher := slices.IndexFunc(words, isLauncher)
	if launcher < 0 {
		return nil
	}
	arguments := words[launcher+1:]
	named := slices.IndexFunc(arguments, func(word *yaml.Node) bool { return !strings.HasPrefix(word.Value, "-") })
	if named < 0 {
		return []report.Finding{{
			Line:    server.line,
			Message: fmt.Sprintf("MCP server %q downloads a package and names none%s", server.key, pointer),
		}}
	}
	downloaded := arguments[named]
	if exactSpec.MatchString(downloaded.Value) {
		return nil
	}
	return []report.Finding{{
		Line:    downloaded.Line,
		Message: fmt.Sprintf("package %q names no exact version: write it as name@1.2.3%s", downloaded.Value, pointer),
	}}
}

// isLauncher tells whether a word of a command line is a launcher, whatever
// its directory and its extension on Windows.
func isLauncher(word *yaml.Node) bool {
	name := strings.ToLower(path.Base(strings.ReplaceAll(word.Value, `\`, "/")))
	name = strings.TrimSuffix(strings.TrimSuffix(name, ".cmd"), ".exe")
	return launchers[name]
}
