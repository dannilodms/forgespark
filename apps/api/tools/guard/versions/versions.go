// Package versions checks that every version the repository pins is exact:
// no range, no `latest`, no floating tag. It reads each home of a version
// that Versions in CODING_STANDARDS.md lists and that a file can be judged
// for: the Go module files, the Compose files, the Dockerfiles, the
// workflows, the Taskfile, the pnpm catalog with its package manifest, and
// `.mcp.json`.
//
// A file is recognised by its name, wherever it is, so a new Dockerfile or a
// new Compose file is judged with no change here.
package versions

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

const pointer = " (Versions in CODING_STANDARDS.md)"

// judge reads one file and returns what is not exact in it. The file of a
// finding is filled in by the caller.
type judge func(content []byte) ([]report.Finding, error)

var (
	composeFile = regexp.MustCompile(`^(docker-)?compose(\..+)?\.ya?ml$`)
	dockerFile  = regexp.MustCompile(`^(Dockerfile|Containerfile)(\..+)?$|\.Dockerfile$`)
	yamlFile    = regexp.MustCompile(`\.ya?ml$`)
)

// judgeOf picks the judge of a file from its path, or nil when the file
// holds no version.
func judgeOf(file string) judge {
	name := path.Base(file)
	switch {
	case name == "go.mod" || strings.HasSuffix(name, ".mod"):
		return goModule
	case composeFile.MatchString(name):
		return compose
	case dockerFile.MatchString(name) && !strings.HasSuffix(name, ".dockerignore"):
		return dockerfile
	case strings.HasPrefix(file, ".github/") && yamlFile.MatchString(name):
		return workflow
	case name == "pnpm-workspace.yaml":
		return pnpmWorkspace
	case name == "package.json":
		return packageManifest
	case name == ".mcp.json":
		return mcpServers
	case name == "Taskfile.yml" || name == "Taskfile.yaml":
		return taskfile
	}
	return nil
}

// Check reads the files under root, the root of the repository, and returns
// one finding for each version that is not exact.
func Check(root string) ([]report.Finding, error) {
	var findings []report.Finding
	err := filepath.WalkDir(root, func(location string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return enter(root, location, entry.Name())
		}
		found, err := checkFile(root, location)
		findings = append(findings, found...)
		return err
	})
	if err != nil {
		return nil, err
	}
	report.Sort(findings)
	return findings, nil
}

// enter tells the walk whether to read a directory. It leaves out what is
// installed or copied in, the fixtures of tests, and the hidden directories
// other than `.github`.
func enter(root, location, name string) error {
	if location == root || name == ".github" {
		return nil
	}
	if name == "node_modules" || name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
		return filepath.SkipDir
	}
	return nil
}

func checkFile(root, location string) ([]report.Finding, error) {
	relative, err := filepath.Rel(root, location)
	if err != nil {
		return nil, err
	}
	file := filepath.ToSlash(relative)
	judge := judgeOf(file)
	if judge == nil {
		return nil, nil
	}
	content, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}
	findings, err := judge(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	for index := range findings {
		findings[index].File = file
	}
	return findings, nil
}

// inexact is the finding for a named thing whose version is not exact.
func inexact(line int, name, version, advice string) report.Finding {
	return report.Finding{
		Line:    line,
		Message: fmt.Sprintf("%s is %q, not an exact version: %s%s", name, version, advice, pointer),
	}
}
