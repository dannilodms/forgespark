// Package silenced turns a rule off without naming it or saying why, which
// nolintlint rejects.
package silenced

import "os"

// Create makes a directory and does not look at what went wrong.
func Create(name string) {
	os.Mkdir(name, 0o755) //nolint
}
