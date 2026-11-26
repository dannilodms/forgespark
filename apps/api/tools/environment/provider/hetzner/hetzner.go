// Package hetzner is the adapter of Hetzner Cloud, on hcloud-go. Each server
// gets the deploy public key registered as a key of the project, under the
// server's name, so that Hetzner sends no root password, and the primary IP
// and the volume kept under the same name. Destroying the server removes the
// key and leaves the other two, which Hetzner bills while they exist.
// Everything it creates carries the labels of the environment, and it finds
// things by name and by those labels. It creates no firewall at Hetzner: the
// firewall is inside the server, set by the cloud-init file.
package hetzner

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
	"github.com/dannilodms/forgespark/apps/api/tools/environment/provider"
)

// Endpoint is the address of Hetzner Cloud's API.
const Endpoint = hcloud.Endpoint

// Settings are what the adapter needs.
type Settings struct {
	// Endpoint is the address of the API. The tests give a stand-in's.
	Endpoint string
	// Token is the API token of the project, which is the whole project's
	// key.
	Token config.Secret
	// Labels mark what the adapter creates, and select what it finds.
	Labels map[string]string
	// PollInterval is the pause between two reads of an action that runs.
	// Zero keeps the one of hcloud-go.
	PollInterval time.Duration
}

// Provider reaches the servers of one environment at Hetzner.
type Provider struct {
	client   *hcloud.Client
	selector string
}

var _ provider.Provider = (*Provider)(nil)

// New gives the adapter.
func New(settings Settings) *Provider {
	options := []hcloud.ClientOption{
		hcloud.WithEndpoint(settings.Endpoint),
		hcloud.WithToken(settings.Token.Reveal()),
	}
	if settings.PollInterval > 0 {
		options = append(options, hcloud.WithPollOpts(hcloud.PollOpts{BackoffFunc: hcloud.ConstantBackoff(settings.PollInterval)}))
	}
	return &Provider{client: hcloud.NewClient(options...), selector: selectorOf(settings.Labels)}
}

// selectorOf writes the labels as a selector that matches all of them.
func selectorOf(labels map[string]string) string {
	terms := make([]string, 0, len(labels))
	for _, key := range slices.Sorted(maps.Keys(labels)) {
		terms = append(terms, key+"="+labels[key])
	}
	return strings.Join(terms, ",")
}

// Keep finds the primary IP and the volume with the name and the labels, and
// creates what is missing in the region: an IPv4 address that Hetzner does
// not delete with a server, and a volume that Hetzner formats with ext4.
func (p *Provider) Keep(ctx context.Context, keeping provider.Keeping) (provider.Kept, error) {
	address, created, err := p.keepAddress(ctx, keeping)
	if err != nil {
		return provider.Kept{}, err
	}
	volume, err := p.keepVolume(ctx, keeping)
	if err != nil {
		return provider.Kept{}, err
	}
	ip, valid := netip.AddrFromSlice(address.IP)
	if !valid {
		return provider.Kept{}, fmt.Errorf("the address %s has no IPv4 address", address.Name)
	}
	return provider.Kept{Address: ip.Unmap(), VolumeDevice: volume.LinuxDevice, NewAddress: created}, nil
}

// keepAddress gives the kept primary IP, created when there is none, and
// whether it was created.
func (p *Provider) keepAddress(ctx context.Context, keeping provider.Keeping) (*hcloud.PrimaryIP, bool, error) {
	address, found, err := p.keptAddress(ctx, keeping.Name)
	if err != nil || found {
		return address, false, err
	}
	result, _, err := p.client.PrimaryIP.Create(ctx, hcloud.PrimaryIPCreateOpts{
		Name:         keeping.Name,
		Type:         hcloud.PrimaryIPTypeIPv4,
		AssigneeType: "server",
		Location:     keeping.Region,
		AutoDelete:   new(false),
		Labels:       keeping.Labels,
	})
	if err != nil {
		return nil, false, fmt.Errorf("create the address %s: %w", keeping.Name, named(err))
	}
	if err := p.waitFor(ctx, result.Action); err != nil {
		return nil, false, fmt.Errorf("wait until the address %s is created: %w", keeping.Name, err)
	}
	return result.PrimaryIP, true, nil
}

// keepVolume gives the kept volume, created when there is none.
func (p *Provider) keepVolume(ctx context.Context, keeping provider.Keeping) (*hcloud.Volume, error) {
	volume, found, err := p.keptVolume(ctx, keeping.Name)
	if err != nil || found {
		return volume, err
	}
	result, _, err := p.client.Volume.Create(ctx, hcloud.VolumeCreateOpts{
		Name:     keeping.Name,
		Size:     keeping.VolumeSize,
		Location: &hcloud.Location{Name: keeping.Region},
		Labels:   keeping.Labels,
		Format:   new(hcloud.VolumeFormatExt4),
	})
	if err != nil {
		return nil, fmt.Errorf("create the volume %s: %w", keeping.Name, named(err))
	}
	if err := p.waitFor(ctx, append([]*hcloud.Action{result.Action}, result.NextActions...)...); err != nil {
		return nil, fmt.Errorf("wait until the volume %s is created: %w", keeping.Name, err)
	}
	return result.Volume, nil
}

// keptAddress gives the primary IP with the name and the labels.
func (p *Provider) keptAddress(ctx context.Context, name string) (*hcloud.PrimaryIP, bool, error) {
	addresses, err := p.client.PrimaryIP.AllWithOpts(ctx, hcloud.PrimaryIPListOpts{ListOpts: hcloud.ListOpts{LabelSelector: p.selector}, Name: name})
	if err != nil {
		return nil, false, fmt.Errorf("list the addresses: %w", err)
	}
	if len(addresses) == 0 {
		return nil, false, nil
	}
	return addresses[0], true, nil
}

// keptVolume gives the volume with the name and the labels.
func (p *Provider) keptVolume(ctx context.Context, name string) (*hcloud.Volume, bool, error) {
	volumes, err := p.client.Volume.AllWithOpts(ctx, hcloud.VolumeListOpts{ListOpts: hcloud.ListOpts{LabelSelector: p.selector}, Name: name})
	if err != nil {
		return nil, false, fmt.Errorf("list the volumes: %w", err)
	}
	if len(volumes) == 0 {
		return nil, false, nil
	}
	return volumes[0], true, nil
}

// waitFor waits until the actions end, leaving out the ones Hetzner did not
// start.
func (p *Provider) waitFor(ctx context.Context, actions ...*hcloud.Action) error {
	started := slices.DeleteFunc(actions, func(action *hcloud.Action) bool { return action == nil })
	if len(started) == 0 {
		return nil
	}
	return p.client.Action.WaitFor(ctx, started...)
}

// Create finds the primary IP and the volume kept under the name of the
// server, registers the deploy key, creates the server with the three and
// waits until Hetzner says the server runs. A key or a server with that
// name already is provider.ErrNameTaken. When the server cannot be created,
// the key is removed again.
func (p *Provider) Create(ctx context.Context, specification provider.Specification) (provider.Server, error) {
	address, found, err := p.keptAddress(ctx, specification.Name)
	if err != nil {
		return provider.Server{}, err
	}
	if !found {
		return provider.Server{}, fmt.Errorf("hetzner keeps no address named %s for the server", specification.Name)
	}
	volume, found, err := p.keptVolume(ctx, specification.Name)
	if err != nil {
		return provider.Server{}, err
	}
	if !found {
		return provider.Server{}, fmt.Errorf("hetzner keeps no volume named %s for the server", specification.Name)
	}
	key, _, err := p.client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      specification.Name,
		PublicKey: specification.DeployPublicKey,
		Labels:    specification.Labels,
	})
	if err != nil {
		return provider.Server{}, fmt.Errorf("register the deploy key: %w", named(err))
	}
	result, _, err := p.client.Server.Create(ctx, hcloud.ServerCreateOpts{
		Name:             specification.Name,
		ServerType:       &hcloud.ServerType{Name: specification.Size},
		Image:            &hcloud.Image{Name: specification.SystemImage},
		Location:         &hcloud.Location{Name: specification.Region},
		SSHKeys:          []*hcloud.SSHKey{key},
		UserData:         specification.CloudInit.Reveal(),
		StartAfterCreate: new(true),
		Labels:           specification.Labels,
		PublicNet:        &hcloud.ServerCreatePublicNet{EnableIPv4: true, IPv4: address, EnableIPv6: false},
		// The cloud-init file mounts the volume where Caddy keeps its data.
		Volumes:   []*hcloud.Volume{volume},
		Automount: new(false),
	})
	if err != nil {
		_, removeErr := p.client.SSHKey.Delete(context.WithoutCancel(ctx), key)
		return provider.Server{}, errors.Join(fmt.Errorf("create the server: %w", named(err)), removeErr)
	}
	actions := append([]*hcloud.Action{result.Action}, result.NextActions...)
	if err := p.client.Action.WaitFor(ctx, actions...); err != nil {
		return provider.Server{}, fmt.Errorf("wait until the server runs: %w", err)
	}
	created, _, err := p.client.Server.GetByID(ctx, result.Server.ID)
	if err != nil {
		return provider.Server{}, fmt.Errorf("read the server: %w", err)
	}
	if created == nil {
		return provider.Server{}, errors.New("read the server: Hetzner no longer has it")
	}
	return serverOf(created)
}

// named turns Hetzner's refusal of a name in use into
// provider.ErrNameTaken.
func named(err error) error {
	if hcloud.IsError(err, hcloud.ErrorCodeUniquenessError) {
		return fmt.Errorf("%w: %w", provider.ErrNameTaken, err)
	}
	return err
}

// Find gives the server with that name and the labels.
func (p *Provider) Find(ctx context.Context, name string) (provider.Server, bool, error) {
	servers, err := p.servers(ctx, name)
	if err != nil {
		return provider.Server{}, false, err
	}
	if len(servers) == 0 {
		return provider.Server{}, false, nil
	}
	server, err := serverOf(servers[0])
	return server, err == nil, err
}

// Destroy deletes the server with that name and the labels, waits until it
// is gone, and removes the deploy key registered under the name. Hetzner
// takes the kept primary IP and volume off the server, and keeps them.
func (p *Provider) Destroy(ctx context.Context, name string) error {
	servers, err := p.servers(ctx, name)
	if err != nil {
		return err
	}
	var failed error
	for _, server := range servers {
		failed = errors.Join(failed, p.deleteServer(ctx, server))
	}
	keys, err := p.client.SSHKey.AllWithOpts(ctx, hcloud.SSHKeyListOpts{ListOpts: hcloud.ListOpts{LabelSelector: p.selector}, Name: name})
	if err != nil {
		return errors.Join(failed, fmt.Errorf("find the deploy key: %w", err))
	}
	for _, key := range keys {
		failed = errors.Join(failed, p.deleteKey(ctx, key))
	}
	return failed
}

func (p *Provider) servers(ctx context.Context, name string) ([]*hcloud.Server, error) {
	servers, err := p.client.Server.AllWithOpts(ctx, hcloud.ServerListOpts{ListOpts: hcloud.ListOpts{LabelSelector: p.selector}, Name: name})
	if err != nil {
		return nil, fmt.Errorf("list the servers: %w", err)
	}
	return servers, nil
}

// deleteServer deletes a server and waits until it is gone.
func (p *Provider) deleteServer(ctx context.Context, server *hcloud.Server) error {
	result, _, err := p.client.Server.DeleteWithResult(ctx, server)
	if err != nil {
		return fmt.Errorf("delete the server %s: %w", server.Name, err)
	}
	if err := p.client.Action.WaitFor(ctx, result.Action); err != nil {
		return fmt.Errorf("wait until the server %s is deleted: %w", server.Name, err)
	}
	return nil
}

func (p *Provider) deleteKey(ctx context.Context, key *hcloud.SSHKey) error {
	if _, err := p.client.SSHKey.Delete(ctx, key); err != nil {
		return fmt.Errorf("remove the deploy key %s: %w", key.Name, err)
	}
	return nil
}

// serverOf reads what the tool needs of a server.
func serverOf(server *hcloud.Server) (provider.Server, error) {
	address, valid := netip.AddrFromSlice(server.PublicNet.IPv4.IP)
	if !valid {
		return provider.Server{}, fmt.Errorf("the server %s has no IPv4 address", server.Name)
	}
	return provider.Server{Name: server.Name, Address: address.Unmap(), Created: server.Created}, nil
}
