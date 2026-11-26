package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
	"github.com/dannilodms/forgespark/apps/api/tools/environment/provider"
)

const (
	// readinessDeadline is how long a new server has to answer SSH and finish
	// cloud-init.
	readinessDeadline = 10 * time.Minute
	// undoDeadline is how long destroying a failed turn-on may take.
	undoDeadline = 5 * time.Minute
	// deployUser is the user that logs in to a server, the one the
	// cloud-init file creates.
	deployUser = "deploy"
	sshPort    = "22"
)

// turnOn creates the server of the environment, waits until it is ready
// and gives the GitHub Environment what reaches it. When the server exists
// already, it checks that it answers, gives GitHub its values again and
// creates nothing.
func turnOn(ctx context.Context, environment environment, stdout io.Writer) error {
	theProvider, err := environment.provider()
	if err != nil {
		return err
	}
	mailpitPassword, err := environment.mailpitPassword(stdout)
	if err != nil {
		return err
	}
	if err := environment.prepareGitHub(ctx); err != nil {
		return err
	}
	name := environment.choices.ServerName
	server, found, err := theProvider.Find(ctx, name)
	if err != nil {
		return fmt.Errorf("find the server %s at %s: %w", name, environment.choices.Provider, err)
	}
	if found {
		return confirmOn(ctx, environment, server, mailpitPassword, stdout)
	}
	return create(ctx, environment, theProvider, mailpitPassword, stdout)
}

// confirmOn checks that a server that exists answers, with the keys the
// main checkout kept when it was created, and gives GitHub its values
// again.
func confirmOn(ctx context.Context, environment environment, server provider.Server, mailpitPassword config.Secret, stdout io.Writer) error {
	keys, kept, err := environment.localCopy().read()
	if err != nil {
		return fmt.Errorf("read the local copy of the deploy key: %w", err)
	}
	if !kept {
		return fmt.Errorf("%s is on at %s, but %s keeps no copy of its deploy key: turn %s off and on again",
			environment.name, server.Address, environment.localCopy().folder, environment.name)
	}
	fmt.Fprintf(stdout, "%s is on already, at %s. Checking that it answers.\n", environment.title(), server.Address)
	if err := environment.waitUntilReady(ctx, server.Address, keys); err != nil {
		return fmt.Errorf("%s is on at %s, but does not answer: %w", environment.name, server.Address, err)
	}
	fmt.Fprintln(stdout, "It answers: nothing was created.")
	values := serverValues{address: server.Address, keys: keys, turnedOn: server.Created}
	if err := environment.writeValues(ctx, values, mailpitPassword, stdout); err != nil {
		return fmt.Errorf("%s is on at %s, but GitHub does not have its values (%w): turn %s on again", environment.name, server.Address, err, environment.name)
	}
	return report(stdout, environment)
}

// create makes sure the provider keeps the address and the volume of the
// environment, creates the server with them and new keys, waits until it is
// ready, keeps the keys and gives GitHub its values. A failure after the
// server was asked for destroys it, and leaves the address and the volume.
func create(ctx context.Context, environment environment, theProvider provider.Provider, mailpitPassword config.Secret, stdout io.Writer) error {
	chosen := environment.choices
	keys, err := newServerKeys(environment.name)
	if err != nil {
		return fmt.Errorf("generate the keys of the server: %w", err)
	}
	fmt.Fprintf(stdout, "Turning %s on: creating the server %s at %s, a %s with %s in %s.\n",
		environment.name, chosen.ServerName, chosen.Provider, chosen.Size, chosen.SystemImage, chosen.Region)
	kept, err := theProvider.Keep(ctx, provider.Keeping{
		Name:       chosen.ServerName,
		Region:     chosen.Region,
		VolumeSize: chosen.VolumeSize,
		Labels:     environment.labels(),
	})
	if err != nil {
		return fmt.Errorf("keep the address and the volume of %s at %s: %w", environment.name, chosen.Provider, err)
	}
	if kept.NewAddress {
		names := environment.hostNames()
		fmt.Fprintf(stdout, "There was no address kept for %s at %s, so a new one is kept now: %s. "+
			"Set the A records of %s, %s and %s to it (docs/providers.md, Domain and DNS): until then the host names do not reach the server.\n",
			environment.name, chosen.Provider, kept.Address, names.app, names.files, names.mail)
	}
	cloudInit, err := renderCloudInit(environment.checkout.root, keys, kept.VolumeDevice)
	if err != nil {
		return err
	}
	specification := provider.Specification{
		Name:            chosen.ServerName,
		Size:            chosen.Size,
		Region:          chosen.Region,
		SystemImage:     chosen.SystemImage,
		CloudInit:       cloudInit,
		DeployPublicKey: keys.deployPublic,
		Labels:          environment.labels(),
	}
	server, err := theProvider.Create(ctx, specification)
	if errors.Is(err, provider.ErrNameTaken) {
		return fmt.Errorf("create the server %s: %w. Another turn-on may be creating it now; when none is, turn %s off and on again",
			chosen.ServerName, err, environment.name)
	}
	if err != nil {
		return environment.undo(ctx, theProvider, fmt.Errorf("create the server %s: %w", chosen.ServerName, err))
	}
	fmt.Fprintf(stdout, "The server is at %s. Waiting until it answers SSH with its host key and cloud-init has prepared it, for up to %v.\n",
		server.Address, readinessDeadline)
	login := keys.login()
	if err := environment.waitUntilReady(ctx, server.Address, login); err != nil {
		return environment.undo(ctx, theProvider, fmt.Errorf("wait until the server is ready: %w", err))
	}
	if err := environment.localCopy().save(server.Address, login); err != nil {
		return environment.undo(ctx, theProvider, fmt.Errorf("keep the local copy of the deploy key: %w", err))
	}
	values := serverValues{address: server.Address, keys: login, turnedOn: server.Created}
	if err := environment.writeValues(ctx, values, mailpitPassword, stdout); err != nil {
		return environment.undo(ctx, theProvider, err)
	}
	fmt.Fprintf(stdout, "%s is on.\n", environment.title())
	return report(stdout, environment)
}

// waitUntilReady waits until the server at that address answers SSH with
// the host key and cloud-init has prepared it.
func (e environment) waitUntilReady(ctx context.Context, address netip.Addr, keys loginKeys) error {
	ctx, cancel := context.WithTimeout(ctx, readinessDeadline)
	defer cancel()
	return e.outside.readiness.Wait(ctx, keys.at(address))
}

// undo destroys what a failed turn-on created, with the deploy key at the
// provider, takes its values out of GitHub and removes the local copy, then
// reports the failure. The address and the volume stay, for the next
// turn-on. It runs to its end even when a signal cut the
// turn-on short.
func (e environment) undo(ctx context.Context, theProvider provider.Provider, cause error) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), undoDeadline)
	defer cancel()
	failed := errors.Join(
		failedTo("destroy the server", theProvider.Destroy(ctx, e.choices.ServerName)),
		e.deleteValues(ctx),
		failedTo("remove the local copy of the deploy key", e.localCopy().remove()),
	)
	if failed != nil {
		return fmt.Errorf("%w; undoing it failed too (%w): turn %s off", cause, failed, e.name)
	}
	return fmt.Errorf("%w; the server was destroyed, and nothing of it is left", cause)
}
