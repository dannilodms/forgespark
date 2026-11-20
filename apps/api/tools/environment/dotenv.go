package main

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// valueIn reads the value of one variable from a file written as Task reads
// a `.env`: `NAME=value` on a line, `export ` before it allowed, `#` starting
// a comment, and the value in quotes or not. The last line that sets the
// name wins. A file that does not exist sets nothing.
func valueIn(path, name string) (string, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	value := ""
	for line := range strings.Lines(string(content)) {
		value = assignedIn(line, name, value)
	}
	return value, nil
}

// assignedIn gives the value the line sets for the name, or the one before
// it when the line sets another name or none.
func assignedIn(line, name, before string) string {
	line = strings.TrimPrefix(strings.TrimSpace(line), "export ")
	key, value, found := strings.Cut(line, "=")
	if !found || strings.TrimSpace(key) != name {
		return before
	}
	return unquoted(strings.TrimSpace(value))
}

// unquoted takes the value out of its quotes, or a comment off a value with
// none.
func unquoted(value string) string {
	if value == "" {
		return ""
	}
	quote := value[:1]
	if quote != `"` && quote != `'` {
		uncommented, _, _ := strings.Cut(value, " #")
		return strings.TrimSpace(uncommented)
	}
	inside, _, _ := strings.Cut(value[1:], quote)
	return inside
}
