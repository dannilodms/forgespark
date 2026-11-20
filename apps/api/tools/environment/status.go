package main

import (
	"context"
	"fmt"
	"io"
	"net/netip"
	"strings"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/tools/environment/provider"
)

// tellStatus tells whether the environment is on, since when, at which
// address and host names, and which open pull request has its label.
func tellStatus(ctx context.Context, environment environment, stdout io.Writer) error {
	servers, err := environment.provider()
	if err != nil {
		return err
	}
	name := environment.choices.ServerName
	server, found, err := servers.Find(ctx, name)
	if err != nil {
		return fmt.Errorf("find the server %s at %s: %w", name, environment.choices.Provider, err)
	}
	if !found {
		fmt.Fprintf(stdout, "%s is off.\n", environment.title())
		return tellPullRequests(ctx, environment, stdout)
	}
	fmt.Fprintf(stdout, "%s is on, since %s (%s ago), at %s.\n", environment.title(),
		server.Created.UTC().Format("2006-01-02 15:04 UTC"), age(environment.outside.now().Sub(server.Created)), server.Address)
	if err := tellPullRequests(ctx, environment, stdout); err != nil {
		return err
	}
	return report(stdout, environment, server)
}

// tellPullRequests writes the open pull requests that have the label of the
// environment, or that none has it.
func tellPullRequests(ctx context.Context, environment environment, stdout io.Writer) error {
	labelled, err := environment.outside.github.LabelledPullRequests(ctx, environment.name)
	if err != nil {
		return fmt.Errorf("find the open pull requests with the label %s on GitHub: %w", environment.name, err)
	}
	if len(labelled) == 0 {
		fmt.Fprintf(stdout, "No open pull request has the %s label.\n", environment.name)
	}
	for _, pullRequest := range labelled {
		fmt.Fprintf(stdout, "Pull request #%d has the %s label: %s, %s\n", pullRequest.Number, environment.name, pullRequest.Title, pullRequest.URL)
	}
	return nil
}

// report writes the host names of the server, and how to log in to it.
func report(stdout io.Writer, environment environment, server provider.Server) error {
	names := hostNamesOf(server.Address)
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

// hostNames are the host names of a server. Until the domain is registered
// they are under sslip.io, which answers a name that carries an address with
// that address (provisional, see docs/providers.md).
type hostNames struct {
	app, files, mail string
}

// hostNamesOf writes the address with dashes, as one label of the name.
func hostNamesOf(address netip.Addr) hostNames {
	domain := strings.ReplaceAll(address.String(), ".", "-") + ".sslip.io"
	return hostNames{app: "app." + domain, files: "files." + domain, mail: "mail." + domain}
}

// age writes a duration in hours and minutes.
func age(duration time.Duration) string {
	minutes := int(duration.Minutes())
	return fmt.Sprintf("%dh%02dm", minutes/60, minutes%60)
}
