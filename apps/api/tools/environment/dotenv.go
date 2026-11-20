package main

import (
	"bytes"
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

// appendTo writes lines at the end of a `.env`, in the line endings the file
// uses, and leaves every line before them as it was. A file that does not
// exist is created, readable by its owner alone.
func appendTo(path, lines string) error {
	content, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	newline := "\n"
	if bytes.Contains(content, []byte("\r\n")) {
		newline = "\r\n"
	}
	lines = strings.ReplaceAll(lines, "\n", newline)
	if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
		lines = newline + lines
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	_, err = file.WriteString(lines)
	return errors.Join(err, file.Close())
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
