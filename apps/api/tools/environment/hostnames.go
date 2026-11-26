package main

import (
	"fmt"
	"io"
)

// report writes the host names of the server, and how to log in to it.
func report(stdout io.Writer, environment environment) error {
	names := environment.hostNames()
	fmt.Fprintf(stdout, "  App:    https://%s\n  Files:  https://%s\n  Mail:   https://%s\n", names.app, names.files, names.mail)
	kept, err := environment.localCopy().kept()
	if err != nil {
		return fmt.Errorf("look for the local copy of the deploy key: %w", err)
	}
	if !kept {
		fmt.Fprintln(stdout, "  SSH:    the main checkout keeps no copy of the deploy key")
		return nil
	}
	fmt.Fprintf(stdout, "  SSH:    %s\n", environment.localCopy().sshCommand())
	return nil
}

// hostNames are the host names of an environment, under the domain of its
// choices. Their DNS records point to the address the provider keeps for
// it, so they stay the same from one server to the next (docs/providers.md).
type hostNames struct {
	app, files, mail string
}

// hostNames gives the host names of the environment.
func (e environment) hostNames() hostNames {
	domain := e.choices.Domain
	return hostNames{app: "app." + domain, files: "files." + domain, mail: "mail." + domain}
}
