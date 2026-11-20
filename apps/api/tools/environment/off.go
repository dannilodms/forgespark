package main

import (
	"context"
	"fmt"
	"io"
)

// turnOff destroys the server of the environment, with its address and its
// deploy key at the provider, and removes the local copy of the key. When
// the environment is off already it succeeds and says so.
func turnOff(ctx context.Context, environment environment, stdout io.Writer) error {
	servers, err := environment.provider()
	if err != nil {
		return err
	}
	name := environment.choices.ServerName
	server, found, err := servers.Find(ctx, name)
	if err != nil {
		return fmt.Errorf("find the server %s at %s: %w", name, environment.choices.Provider, err)
	}
	// Even with no server: a turn-on cut short may have left its deploy key
	// at the provider.
	if err := servers.Destroy(ctx, name); err != nil {
		return fmt.Errorf("destroy the server %s at %s: %w", name, environment.choices.Provider, err)
	}
	if err := environment.localCopy().remove(); err != nil {
		return fmt.Errorf("remove the local copy of the deploy key: %w", err)
	}
	// Ticket 02 of .scratch/staging: the values come out of GitHub here.
	if !found {
		fmt.Fprintf(stdout, "%s was off already: there was no server to destroy.\n", environment.title())
		return nil
	}
	fmt.Fprintf(stdout, "%s is off: the server %s at %s was destroyed at %s, with its address and its deploy key.\n",
		environment.title(), name, server.Address, environment.choices.Provider)
	return nil
}
