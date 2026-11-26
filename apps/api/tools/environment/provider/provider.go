// Package provider names what the tool asks of the company that rents it
// its servers: keep what the servers of an environment share from one to
// the next, tell what it keeps, create a server from a specification, find
// one by its name and destroy one by its name. Hetzner is the one adapter today, in the package
// hetzner; providertest keeps servers in memory for the tests of the
// commands.
package provider

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
)

// ErrNameTaken is the answer of Create when the provider already has a
// server, or a key, with the name of the new one: another turn-on may be
// creating it at this moment, so nothing of it is destroyed.
var ErrNameTaken = errors.New("the provider already has a server or a key with this name")

// Provider keeps what the servers of one environment share, and creates,
// finds and destroys them.
type Provider interface {
	// Keep makes sure the provider keeps, under that name, the IPv4 address
	// and the volume that every server of the environment gets, and gives
	// them. It creates only what is missing, and gives what exists as it is.
	Keep(ctx context.Context, keeping Keeping) (Kept, error)
	// FindKept gives what the provider keeps under that name, the address
	// and the volume, each only if it exists, and creates nothing.
	FindKept(ctx context.Context, name string) (Holdings, error)
	// Create creates a server, with the address and the volume kept under
	// its name, and returns once the provider says it is running. On a
	// failure it leaves nothing of the server behind that it knows of;
	// Destroy removes the rest.
	Create(ctx context.Context, specification Specification) (Server, error)
	// Find tells whether a server with that name exists, and gives it.
	Find(ctx context.Context, name string) (Server, bool, error)
	// Destroy destroys the server with that name and the deploy key
	// registered with it, and leaves what Keep kept. Destroying what does
	// not exist succeeds.
	Destroy(ctx context.Context, name string) error
}

// Keeping is what Keep keeps, and where.
type Keeping struct {
	// Name names the address and the volume, as it names the servers they
	// go to.
	Name string
	// Region is where they are, which is where the servers run: the
	// provider gives a server only an address and a volume of its region.
	Region string
	// VolumeSize is the size of a new volume, in gigabytes.
	VolumeSize int
	// Labels mark them, as they mark the servers.
	Labels map[string]string
}

// Kept is what the provider keeps for an environment from one server to the
// next. Turning the environment off destroys neither.
type Kept struct {
	// Address is the IPv4 address that every server gets, to which the DNS
	// records of the host names of the environment point.
	Address netip.Addr
	// VolumeDevice is the path of the volume on a server it is attached to,
	// a block device that has a file system already.
	VolumeDevice string
	// NewAddress tells that Keep created the address: the DNS records point
	// to the one before, and are set again by hand.
	NewAddress bool
}

// Holdings is what the provider keeps under the name of an environment, as
// FindKept reads it. The provider bills both while they exist, also while
// the environment is off.
type Holdings struct {
	// Address is the kept IPv4 address. It is the zero netip.Addr, which is
	// not valid, when none is kept.
	Address netip.Addr
	// VolumeSize is the size of the kept volume, in gigabytes. It is zero
	// when none is kept.
	VolumeSize int
}

// Specification is what a new server is made of.
type Specification struct {
	Name        string
	Size        string
	Region      string
	SystemImage string
	// CloudInit is the user data of the server, the cloud-init file with the
	// keys of this server. It holds the private host key, so it prints as a
	// placeholder.
	CloudInit config.Secret
	// DeployPublicKey is the public key of the deploy user, as a line of
	// authorized keys.
	DeployPublicKey string
	// Labels mark what the provider creates for the environment.
	Labels map[string]string
}

// Server is a server that exists at the provider.
type Server struct {
	Name string
	// Address is its public IPv4 address.
	Address netip.Addr
	// Created is when the provider created it.
	Created time.Time
}
