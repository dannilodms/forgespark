package config

import (
	"os"
	"strings"
)

// Environment reads the variables of the process as a map, which is what
// Load takes. It is the one place that reads the environment: the command
// that starts reads it once, and the support of the tests asks it for the
// databases the tests create.
func Environment() map[string]string {
	variables := map[string]string{}
	for _, entry := range os.Environ() {
		name, value, _ := strings.Cut(entry, "=")
		variables[name] = value
	}
	return variables
}
