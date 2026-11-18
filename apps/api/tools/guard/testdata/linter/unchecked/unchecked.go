// Package unchecked drops an error, which errcheck rejects.
package unchecked

import "os"

// Create makes a directory and does not look at what went wrong.
func Create(name string) {
	os.Mkdir(name, 0o755)
}
