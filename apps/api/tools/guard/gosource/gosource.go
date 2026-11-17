// Package gosource finds the Go files a guard judges.
package gosource

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Files lists the Go files under root, as paths from root with forward
// slashes, in the order of the directory tree. It leaves out what the Go
// tool leaves out of a build: `testdata`, `vendor`, and a directory whose
// name starts with a dot or an underscore.
func Files(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return enter(root, path, entry.Name())
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	return files, err
}

// enter tells the walk whether to read a directory.
func enter(root, path, name string) error {
	if path == root {
		return nil
	}
	if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return filepath.SkipDir
	}
	return nil
}
