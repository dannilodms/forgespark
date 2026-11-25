// Package imports checks the dependency rule of the backend: what a package
// under `internal` may import, by the place it has in the layout.
//
// The rule is read from the path of a package, never from a list of areas,
// so an area that is created tomorrow is judged with no change here.
package imports

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/dannilodms/forgespark/apps/api/tools/guard/gosource"
	"github.com/dannilodms/forgespark/apps/api/tools/guard/report"
)

const (
	pointer       = " (Dependency rule in CODING_STANDARDS.md)"
	layoutMessage = "is outside the layout: a package under internal lives in kernel, in platform, " +
		"or in <area>/domain, <area>/application or <area>/adapters/<technology> (Layout in CODING_STANDARDS.md)"
)

// layer is the part of the layout a package belongs to.
type layer int

const (
	kernel layer = iota
	platform
	domain
	application
	adapter
)

// rules says, for each layer, what the table of the dependency rule allows.
// It is the text a finding shows.
var rules = map[layer]string{
	kernel:      "kernel imports the standard library only",
	platform:    "platform imports third-party libraries and kernel, and no domain area",
	domain:      "a domain package imports the standard library, kernel and the domain of its own area",
	application: "an application package imports the standard library, the domain of its own area, kernel, platform and the application of other areas",
	adapter:     "an adapter package imports the application and the domain of its own area, kernel, platform and third-party libraries, and no other adapter package",
}

// place is where a package sits in the layout. The area is empty for kernel
// and platform, and the technology is set for an adapter only.
type place struct {
	layer      layer
	area       string
	technology string
}

// Check reads the imports of every Go file under root/internal, test files
// included, and returns one finding for each import the dependency rule
// forbids and one for each directory that is outside the layout. root is the
// directory of go.mod.
func Check(root string) ([]report.Finding, error) {
	module, err := modulePath(root)
	if err != nil {
		return nil, err
	}
	files, err := gosource.Files(filepath.Join(root, "internal"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	var unread error
	for _, file := range files {
		found, err := checkFile(root, module, "internal/"+file)
		unread = errors.Join(unread, err)
		findings = append(findings, found...)
	}
	if unread != nil {
		return nil, unread
	}
	report.Sort(findings)
	// A directory outside the layout is reported by each of its files, and
	// listed once.
	return slices.Compact(findings), nil
}

// placeOf reads the place of a package from its directory under internal.
func placeOf(directory string) (place, bool) {
	parts := strings.Split(directory, "/")
	if parts[0] == "kernel" {
		return place{layer: kernel}, true
	}
	if parts[0] == "platform" {
		return place{layer: platform}, true
	}
	if len(parts) < 2 {
		return place{}, false
	}
	area := parts[0]
	if parts[1] == "domain" {
		return place{layer: domain, area: area}, true
	}
	if parts[1] == "application" {
		return place{layer: application, area: area}, true
	}
	if parts[1] == "adapters" && len(parts) > 2 {
		return place{layer: adapter, area: area, technology: parts[2]}, true
	}
	return place{}, false
}

// checkFile judges the imports of one file, named by its path from the root
// of the module.
func checkFile(root, module, file string) ([]report.Finding, error) {
	directory := path.Dir(file)
	from, known := placeOf(strings.TrimPrefix(directory, "internal/"))
	if !known {
		return []report.Finding{{File: directory, Message: layoutMessage}}, nil
	}
	positions := token.NewFileSet()
	parsed, err := parser.ParseFile(positions, filepath.Join(root, filepath.FromSlash(file)), nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	isTest := strings.HasSuffix(file, "_test.go")
	var findings []report.Finding
	for _, spec := range parsed.Imports {
		// The parser accepted the path, so it is a valid string.
		imported, _ := strconv.Unquote(spec.Path.Value)
		line := positions.Position(spec.Pos()).Line
		findings = append(findings, judgeImport(file, line, from, isTest, module, imported)...)
	}
	return findings, nil
}

func judgeImport(file string, line int, from place, isTest bool, module, imported string) []report.Finding {
	if allows(from, isTest, module, imported) {
		return nil
	}
	return []report.Finding{{
		File:    file,
		Line:    line,
		Message: fmt.Sprintf("imports %q: %s%s", imported, rules[from.layer], pointer),
	}}
}

// allows tells whether a package at from may import the path.
func allows(from place, isTest bool, module, imported string) bool {
	inside, isInternal := strings.CutPrefix(imported, module+"/internal/")
	if isInternal {
		to, known := placeOf(inside)
		return known && allowsInternal(from, to)
	}
	if isStandardLibrary(imported) {
		return true
	}
	// Everything else is a library: a third-party module, or a package of
	// this module outside internal, such as the migrations. A test may use
	// one in any layer: the assertion library, the temporary database.
	return isTest || from.layer == adapter || from.layer == platform
}

// allowsInternal is the table of the dependency rule, for an import of a
// package under internal.
func allowsInternal(from, to place) bool {
	if to.layer == kernel {
		return true
	}
	sameArea := from.area == to.area
	switch from.layer {
	case platform:
		return to.layer == platform
	case domain:
		return to.layer == domain && sameArea
	case application:
		return to.layer == platform || to.layer == application || (to.layer == domain && sameArea)
	case adapter:
		return to.layer == platform || (sameArea && allowsInsideArea(from, to))
	}
	return false
}

// allowsInsideArea tells whether an adapter may import a package of its own
// area: its application, its domain, and the packages of its own technology.
func allowsInsideArea(from, to place) bool {
	return to.layer != adapter || to.technology == from.technology
}

// isStandardLibrary tells a package of the standard library from a module
// path, whose first element is a host name and so holds a dot.
func isStandardLibrary(imported string) bool {
	first, _, _ := strings.Cut(imported, "/")
	return !strings.Contains(first, ".")
}

// modulePath reads the path of the module from go.mod.
func modulePath(root string) (string, error) {
	file := filepath.Join(root, "go.mod")
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	found := moduleLine.FindSubmatch(content)
	if found == nil {
		return "", fmt.Errorf("%s names no module", file)
	}
	return string(found[1]), nil
}

var moduleLine = regexp.MustCompile(`(?m)^module\s+(\S+)`)
