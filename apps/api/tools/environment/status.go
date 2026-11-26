package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dannilodms/forgespark/apps/api/tools/environment/provider"
)

// tellStatus tells whether the environment is on, since when, at which
// address and host names, and which open pull request has its label. When
// it is off, it also tells what the provider keeps for it, which bills
// while nothing runs.
func tellStatus(ctx context.Context, environment environment, stdout io.Writer) error {
	theProvider, err := environment.provider()
	if err != nil {
		return err
	}
	name := environment.choices.ServerName
	server, found, err := theProvider.Find(ctx, name)
	if err != nil {
		return fmt.Errorf("find the server %s at %s: %w", name, environment.choices.Provider, err)
	}
	if !found {
		return tellOff(ctx, environment, theProvider, stdout)
	}
	fmt.Fprintf(stdout, "%s is on, since %s (%s ago), at %s.\n", environment.title(),
		server.Created.UTC().Format("2006-01-02 15:04 UTC"), age(environment.outside.now().Sub(server.Created)), server.Address)
	if err := tellPullRequests(ctx, environment, stdout); err != nil {
		return err
	}
	return report(stdout, environment)
}

// tellOff tells that the environment is off, what the provider keeps for
// it, and which open pull request has its label.
func tellOff(ctx context.Context, environment environment, theProvider provider.Provider, stdout io.Writer) error {
	fmt.Fprintf(stdout, "%s is off.\n", environment.title())
	if err := tellKept(ctx, environment, theProvider, stdout); err != nil {
		return err
	}
	return tellPullRequests(ctx, environment, stdout)
}

// tellKept writes what the provider keeps under the name of the
// environment, the address and the volume, which it bills until they are
// deleted.
func tellKept(ctx context.Context, environment environment, theProvider provider.Provider, stdout io.Writer) error {
	name := environment.choices.ServerName
	holdings, err := theProvider.FindKept(ctx, name)
	if err != nil {
		return fmt.Errorf("find what %s keeps for %s: %w", environment.choices.Provider, name, err)
	}
	kept := keptParts(holdings)
	if len(kept) == 0 {
		fmt.Fprintf(stdout, "Nothing is kept at %s: the next turn-on creates an address and a volume.\n", environment.choices.Provider)
		return nil
	}
	fmt.Fprintf(stdout, "Kept at %s for the next turn-on, and billed until deleted: %s.\n", environment.choices.Provider, strings.Join(kept, " and "))
	return nil
}

// keptParts names each thing the provider keeps.
func keptParts(holdings provider.Holdings) []string {
	var parts []string
	if holdings.Address.IsValid() {
		parts = append(parts, "the address "+holdings.Address.String())
	}
	if holdings.VolumeSize > 0 {
		parts = append(parts, fmt.Sprintf("a volume of %d GB", holdings.VolumeSize))
	}
	return parts
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

// age writes a duration in hours and minutes.
func age(duration time.Duration) string {
	minutes := int(duration.Minutes())
	return fmt.Sprintf("%dh%02dm", minutes/60, minutes%60)
}
