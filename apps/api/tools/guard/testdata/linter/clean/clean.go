// Package clean is what the linter accepts.
package clean

import (
	"fmt"
	"io"
	"os"
)

// Size returns the size of a file. The error of the deferred Close and the
// one of a print are not checked, as the configuration allows.
func Size(name string, output io.Writer) (int64, error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	information, err := file.Stat()
	if err != nil {
		return 0, err
	}
	fmt.Fprintln(output, information.Name())
	return information.Size(), nil
}
