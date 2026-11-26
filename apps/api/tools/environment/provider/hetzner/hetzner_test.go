package hetzner_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
	"github.com/dannilodms/forgespark/apps/api/tools/environment/provider"
	"github.com/dannilodms/forgespark/apps/api/tools/environment/provider/hetzner"
)

// These tests reach a simulated Hetzner Cloud, which keeps its SSH keys,
// servers, primary IPs, volumes and actions in memory and answers as the reference
// of Hetzner's API says (https://docs.hetzner.cloud/reference/cloud). No
// test calls Hetzner itself: the first real turn-on does, by hand.

const (
	token     = "hetzner-token-of-the-simulation"
	deployKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIK4m2tR5i7Wq9xD0aB3cE6fG8hJ1kL2mN3oP4qR5sT6u"
	created   = "2026-11-19T14:00:00Z"
)

// request is what one request carried.
type request struct {
	method        string
	path          string
	query         string
	authorization string
	body          map[string]any
}

// simulatedHetzner is a stand-in for the API of Hetzner Cloud.
type simulatedHetzner struct {
	*httptest.Server

	mu       sync.Mutex
	requests []request
	nextID   int64
	keys     map[int64]map[string]any
	servers  map[int64]map[string]any
	// addresses are the primary IPs, by id, and volumes the volumes.
	addresses map[int64]map[string]any
	volumes   map[int64]map[string]any
	// refuseServers answers every creation of a server with this error.
	refuseServers *apiError
	// failStarts makes every start of a server end in an error.
	failStarts bool
	actions    map[int64]string
}

type apiError struct {
	status       int
	code, detail string
}

func newSimulatedHetzner(t *testing.T) *simulatedHetzner {
	t.Helper()
	simulated := &simulatedHetzner{
		nextID:    100,
		keys:      map[int64]map[string]any{},
		servers:   map[int64]map[string]any{},
		addresses: map[int64]map[string]any{},
		volumes:   map[int64]map[string]any{},
		actions:   map[int64]string{},
	}
	simulated.Server = httptest.NewServer(http.HandlerFunc(simulated.serve))
	t.Cleanup(simulated.Close)
	return simulated
}

func (s *simulatedHetzner) adapter() *hetzner.Provider {
	return hetzner.New(hetzner.Settings{
		Endpoint:     s.URL + "/v1",
		Token:        config.NewSecret(token),
		Labels:       map[string]string{"environment": "staging"},
		PollInterval: time.Millisecond,
	})
}

func (s *simulatedHetzner) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	s.mu.Lock()
	defer s.mu.Unlock()
	path := strings.TrimPrefix(r.URL.Path, "/v1")
	s.requests = append(s.requests, request{
		method: r.Method, path: path, query: r.URL.RawQuery,
		authorization: r.Header.Get("Authorization"), body: body,
	})
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer "+token {
		answerError(w, apiError{http.StatusUnauthorized, "unauthorized", "unable to authenticate"})
		return
	}
	resource, id, _ := strings.Cut(strings.TrimPrefix(path, "/"), "/")
	switch r.Method + " " + resource {
	case "POST ssh_keys":
		s.createKey(w, body)
	case "GET ssh_keys":
		s.list(w, "ssh_keys", s.keys, r)
	case "DELETE ssh_keys":
		s.deleteKey(w, id)
	case "POST servers":
		s.createServer(w, body)
	case "GET servers":
		s.getServers(w, r, id)
	case "DELETE servers":
		s.deleteServer(w, id)
	case "GET actions":
		s.listActions(w, r)
	case "POST primary_ips":
		s.createAddress(w, body)
	case "GET primary_ips":
		s.list(w, "primary_ips", s.addresses, r)
	case "POST volumes":
		s.createVolume(w, body)
	case "GET volumes":
		s.list(w, "volumes", s.volumes, r)
	default:
		answerError(w, apiError{http.StatusNotFound, "not_found", "no such route"})
	}
}

func answer(w http.ResponseWriter, status int, body any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func answerError(w http.ResponseWriter, failure apiError) {
	answer(w, failure.status, map[string]any{"error": map[string]any{"code": failure.code, "message": failure.detail, "details": map[string]any{}}})
}

func (s *simulatedHetzner) id() int64 {
	s.nextID++
	return s.nextID
}

// action starts an action, which is running until it is read.
func (s *simulatedHetzner) action(command string, resource int64, kind string) map[string]any {
	id := s.id()
	s.actions[id] = command
	return map[string]any{
		"id": id, "command": command, "status": "running", "progress": 0, "started": created,
		"finished": nil, "error": nil, "resources": []any{map[string]any{"id": resource, "type": kind}},
	}
}

func named(resources map[int64]map[string]any, name any) bool {
	return slices.ContainsFunc(slices.Collect(maps.Values(resources)), func(resource map[string]any) bool { return resource["name"] == name })
}

func (s *simulatedHetzner) createKey(w http.ResponseWriter, body map[string]any) {
	if named(s.keys, body["name"]) {
		answerError(w, apiError{http.StatusConflict, "uniqueness_error", "SSH key with the same name already exists"})
		return
	}
	id := s.id()
	key := map[string]any{
		"id": id, "name": body["name"], "public_key": body["public_key"], "labels": body["labels"],
		"fingerprint": "b7:2f:30:a0:2f:6c:58:6c:21:04:58:61:ba:06:3b:2f", "created": created,
	}
	s.keys[id] = key
	answer(w, http.StatusCreated, map[string]any{"ssh_key": key})
}

func (s *simulatedHetzner) deleteKey(w http.ResponseWriter, id string) {
	number, _ := strconv.ParseInt(id, 10, 64)
	if _, found := s.keys[number]; !found {
		answerError(w, apiError{http.StatusNotFound, "not_found", "SSH key not found"})
		return
	}
	delete(s.keys, number)
	w.WriteHeader(http.StatusNoContent)
}

func (s *simulatedHetzner) createServer(w http.ResponseWriter, body map[string]any) {
	if s.refuseServers != nil {
		answerError(w, *s.refuseServers)
		return
	}
	if named(s.servers, body["name"]) {
		answerError(w, apiError{http.StatusConflict, "uniqueness_error", "server name is already used"})
		return
	}
	publicNet, _ := body["public_net"].(map[string]any)
	addressID, _ := publicNet["ipv4"].(float64)
	address, found := s.addresses[int64(addressID)]
	if !found || address["assignee_id"] != nil {
		answerError(w, apiError{http.StatusUnprocessableEntity, "invalid_input", "primary IP not found or assigned"})
		return
	}
	id := s.id()
	address["assignee_id"] = id
	volumes, _ := body["volumes"].([]any)
	for _, each := range volumes {
		volumeID, _ := each.(float64)
		s.volumes[int64(volumeID)]["server"] = id
	}
	server := map[string]any{
		"id": id, "name": body["name"], "status": "initializing", "created": created,
		"public_net": map[string]any{
			"ipv4":         map[string]any{"id": address["id"], "ip": address["ip"], "blocked": false, "dns_ptr": "static.7.100.51.198.clients.your-server.de"},
			"ipv6":         nil,
			"floating_ips": []any{},
			"firewalls":    []any{},
		},
		"private_net": []any{},
		"server_type": map[string]any{"id": 114, "name": body["server_type"]},
		"location":    map[string]any{"id": 1, "name": body["location"]},
		"image":       map[string]any{"id": 387894169, "type": "system", "name": body["image"]},
		"iso":         nil, "rescue_enabled": false, "locked": false, "backup_window": nil,
		"outgoing_traffic": 0, "ingoing_traffic": 0, "included_traffic": 21990232555520,
		"protection": map[string]any{"delete": false, "rebuild": false},
		"labels":     body["labels"], "primary_disk_size": 40, "volumes": volumes, "load_balancers": []any{},
	}
	s.servers[id] = server
	answer(w, http.StatusCreated, map[string]any{
		"server":        server,
		"action":        s.action("create_server", id, "server"),
		"next_actions":  []any{s.action("start_server", id, "server")},
		"root_password": nil,
	})
	server["status"] = "running"
}

// createAddress creates a primary IP that no server has yet.
func (s *simulatedHetzner) createAddress(w http.ResponseWriter, body map[string]any) {
	if named(s.addresses, body["name"]) {
		answerError(w, apiError{http.StatusConflict, "uniqueness_error", "primary IP with the same name already exists"})
		return
	}
	id := s.id()
	address := map[string]any{
		"id": id, "ip": "198.51.100.7", "type": body["type"], "name": body["name"], "auto_delete": body["auto_delete"],
		"assignee_id": nil, "assignee_type": body["assignee_type"], "blocked": false, "created": created, "labels": body["labels"],
		"protection": map[string]any{"delete": false}, "dns_ptr": []any{}, "location": map[string]any{"id": 1, "name": body["location"]},
	}
	s.addresses[id] = address
	answer(w, http.StatusCreated, map[string]any{"primary_ip": address, "action": nil})
}

// createVolume creates a volume that no server has yet, which is being
// formatted until its action is read.
func (s *simulatedHetzner) createVolume(w http.ResponseWriter, body map[string]any) {
	if named(s.volumes, body["name"]) {
		answerError(w, apiError{http.StatusConflict, "uniqueness_error", "volume with the same name already exists"})
		return
	}
	id := s.id()
	volume := map[string]any{
		"id": id, "name": body["name"], "size": body["size"], "server": nil, "status": "creating", "created": created,
		"labels": body["labels"], "format": body["format"], "linux_device": fmt.Sprintf("/dev/disk/by-id/scsi-0HC_Volume_%d", id),
		"protection": map[string]any{"delete": false}, "location": map[string]any{"id": 1, "name": body["location"]},
	}
	s.volumes[id] = volume
	answer(w, http.StatusCreated, map[string]any{"volume": volume, "action": s.action("create_volume", id, "volume"), "next_actions": []any{}})
	volume["status"] = "available"
}

func (s *simulatedHetzner) getServers(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		s.list(w, "servers", s.servers, r)
		return
	}
	number, _ := strconv.ParseInt(id, 10, 64)
	server, found := s.servers[number]
	if !found {
		answerError(w, apiError{http.StatusNotFound, "not_found", "server not found"})
		return
	}
	answer(w, http.StatusOK, map[string]any{"server": server})
}

func (s *simulatedHetzner) deleteServer(w http.ResponseWriter, id string) {
	number, _ := strconv.ParseInt(id, 10, 64)
	if _, found := s.servers[number]; !found {
		answerError(w, apiError{http.StatusNotFound, "not_found", "server not found"})
		return
	}
	delete(s.servers, number)
	takeOff(s.addresses, "assignee_id", number)
	takeOff(s.volumes, "server", number)
	answer(w, http.StatusOK, map[string]any{"action": s.action("delete_server", number, "server")})
}

// takeOff takes the server off every resource that names it in that field,
// as Hetzner does with the primary IP and the volumes of a server it
// deletes.
func takeOff(resources map[int64]map[string]any, field string, server int64) {
	for _, resource := range resources {
		resource[field] = cleared(resource[field], server)
	}
}

// cleared gives nothing for the server, and any other value as it is.
func cleared(value any, server int64) any {
	if value == server {
		return nil
	}
	return value
}

// list answers a list, filtered by name and by a label selector of
// key=value terms, in one page.
func (s *simulatedHetzner) list(w http.ResponseWriter, kind string, resources map[int64]map[string]any, r *http.Request) {
	query := r.URL.Query()
	found := []any{}
	for _, id := range slices.Sorted(maps.Keys(resources)) {
		found = append(found, matching(resources[id], query.Get("name"), query.Get("label_selector"))...)
	}
	answer(w, http.StatusOK, map[string]any{
		kind:   found,
		"meta": map[string]any{"pagination": map[string]any{"page": 1, "per_page": 50, "previous_page": nil, "next_page": nil, "last_page": 1, "total_entries": len(found)}},
	})
}

// matching gives the resource when it has the name, if one is asked for,
// and every label of the selector.
func matching(resource map[string]any, name, selector string) []any {
	if name != "" && resource["name"] != name {
		return nil
	}
	values, _ := resource["labels"].(map[string]any)
	missing := slices.ContainsFunc(strings.Split(selector, ","), func(term string) bool {
		key, value, _ := strings.Cut(term, "=")
		return term != "" && values[key] != value
	})
	if missing {
		return nil
	}
	return []any{resource}
}

// listActions answers the actions asked for by id, each finished now.
func (s *simulatedHetzner) listActions(w http.ResponseWriter, r *http.Request) {
	found := []any{}
	for _, id := range r.URL.Query()["id"] {
		found = append(found, s.finished(id)...)
	}
	answer(w, http.StatusOK, map[string]any{
		"actions": found,
		"meta":    map[string]any{"pagination": map[string]any{"page": 1, "per_page": 50, "previous_page": nil, "next_page": nil, "last_page": 1, "total_entries": len(found)}},
	})
}

// finished gives the action with that id, which has ended now: in success,
// or in an error for the start of a server when the test asks for it.
func (s *simulatedHetzner) finished(id string) []any {
	number, _ := strconv.ParseInt(id, 10, 64)
	command, known := s.actions[number]
	if !known {
		return nil
	}
	action := map[string]any{"id": number, "command": command, "status": "success", "progress": 100, "started": created, "finished": created, "error": nil, "resources": []any{}}
	if s.failStarts && command == "start_server" {
		action["status"] = "error"
		action["error"] = map[string]any{"code": "action_failed", "message": "Action failed"}
	}
	return []any{action}
}

// sent gives the requests of that method and path.
func (s *simulatedHetzner) sent(method, path string) []request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.DeleteFunc(slices.Clone(s.requests), func(each request) bool { return each.method != method || each.path != path })
}

func (s *simulatedHetzner) counts() (keys, servers, addresses, volumes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.keys), len(s.servers), len(s.addresses), len(s.volumes)
}

// idOf gives the id of the resource with that name.
func (s *simulatedHetzner) idOf(resources map[int64]map[string]any, name string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return firstID(resources, func(resource map[string]any) bool { return resource["name"] == name })
}

// actionOf gives the id of the action of that command.
func (s *simulatedHetzner) actionOf(command string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return firstID(s.actions, func(each string) bool { return each == command })
}

// firstID gives the lowest id whose value matches, or 0 when none does.
func firstID[V any](values map[int64]V, matches func(V) bool) int64 {
	ids := slices.Sorted(maps.Keys(values))
	at := slices.IndexFunc(ids, func(id int64) bool { return matches(values[id]) })
	if at < 0 {
		return 0
	}
	return ids[at]
}

// waitedFor tells whether the adapter read the action until it ended.
func (s *simulatedHetzner) waitedFor(id int64) bool {
	return slices.ContainsFunc(s.sent(http.MethodGet, "/actions"), func(each request) bool {
		return slices.Contains(strings.Split(each.query, "&"), fmt.Sprintf("id=%d", id))
	})
}

func aKeeping() provider.Keeping {
	return provider.Keeping{Name: "staging", Region: "fsn1", VolumeSize: 10, Labels: map[string]string{"environment": "staging"}}
}

// kept makes sure Hetzner keeps the address and the volume of staging.
func kept(t *testing.T, adapter *hetzner.Provider) provider.Kept {
	t.Helper()
	kept, err := adapter.Keep(context.Background(), aKeeping())
	require.NoError(t, err)
	return kept
}

func aSpecification() provider.Specification {
	return provider.Specification{
		Name:            "staging",
		Size:            "cx23",
		Region:          "fsn1",
		SystemImage:     "ubuntu-26.04",
		CloudInit:       config.NewSecret("#cloud-config\nssh_pwauth: false\n"),
		DeployPublicKey: deployKey,
		Labels:          map[string]string{"environment": "staging"},
	}
}

func TestKeepCreatesTheAddressAndTheVolumeTheFirstTime(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)

	kept, err := simulated.adapter().Keep(context.Background(), aKeeping())

	require.NoError(t, err)
	volume := simulated.idOf(simulated.volumes, "staging")
	assert.Equal(t, provider.Kept{
		Address:      netip.MustParseAddr("198.51.100.7"),
		VolumeDevice: fmt.Sprintf("/dev/disk/by-id/scsi-0HC_Volume_%d", volume),
		NewAddress:   true,
	}, kept)
	addresses := simulated.sent(http.MethodPost, "/primary_ips")
	require.Len(t, addresses, 1)
	assert.Equal(t, map[string]any{
		"name": "staging", "type": "ipv4", "assignee_type": "server", "location": "fsn1", "auto_delete": false,
		"labels": map[string]any{"environment": "staging"},
	}, addresses[0].body, "an IPv4 address of the region, which Hetzner keeps when a server is deleted")
	volumes := simulated.sent(http.MethodPost, "/volumes")
	require.Len(t, volumes, 1)
	assert.Equal(t, map[string]any{
		"name": "staging", "size": float64(10), "location": "fsn1", "format": "ext4",
		"labels": map[string]any{"environment": "staging"},
	}, volumes[0].body, "a volume of the region that Hetzner formats with ext4")
	assert.True(t, simulated.waitedFor(simulated.actionOf("create_volume")), "the volume is given once it is created")
}

func TestKeepGivesWhatHetznerKeepsAlreadyAndCreatesNothing(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	first := kept(t, adapter)

	again, err := adapter.Keep(context.Background(), aKeeping())

	require.NoError(t, err)
	assert.Equal(t, first.Address, again.Address)
	assert.Equal(t, first.VolumeDevice, again.VolumeDevice)
	assert.False(t, again.NewAddress, "the DNS records point to it already")
	assert.Len(t, simulated.sent(http.MethodPost, "/primary_ips"), 1)
	assert.Len(t, simulated.sent(http.MethodPost, "/volumes"), 1)
	lists := simulated.sent(http.MethodGet, "/primary_ips")
	require.NotEmpty(t, lists)
	assert.Contains(t, lists[len(lists)-1].query, "label_selector=environment%3Dstaging")
	assert.Contains(t, lists[len(lists)-1].query, "name=staging")
}

func TestCreateRegistersTheDeployKeyAndCreatesTheServerWithItTheAddressAndTheVolume(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	kept(t, adapter)

	server, err := adapter.Create(context.Background(), aSpecification())

	require.NoError(t, err)
	assert.Equal(t, provider.Server{
		Name: "staging", Address: netip.MustParseAddr("198.51.100.7"), Created: time.Date(2026, 11, 19, 14, 0, 0, 0, time.UTC),
	}, server)
	keys := simulated.sent(http.MethodPost, "/ssh_keys")
	require.Len(t, keys, 1)
	assert.Equal(t, "Bearer "+token, keys[0].authorization)
	assert.Equal(t, map[string]any{
		"name": "staging", "public_key": deployKey, "labels": map[string]any{"environment": "staging"},
	}, keys[0].body)
	servers := simulated.sent(http.MethodPost, "/servers")
	require.Len(t, servers, 1)
	assert.Equal(t, map[string]any{
		"name":               "staging",
		"server_type":        "cx23",
		"image":              "ubuntu-26.04",
		"location":           "fsn1",
		"ssh_keys":           []any{float64(simulated.idOf(simulated.keys, "staging"))},
		"user_data":          "#cloud-config\nssh_pwauth: false\n",
		"labels":             map[string]any{"environment": "staging"},
		"start_after_create": true,
		"public_net": map[string]any{
			"enable_ipv4": true, "enable_ipv6": false, "ipv4": float64(simulated.idOf(simulated.addresses, "staging")),
		},
		"volumes":   []any{float64(simulated.idOf(simulated.volumes, "staging"))},
		"automount": false,
	}, servers[0].body, "the kept address and volume, no IPv6, the deploy key so that no root password is e-mailed, and no firewall; the cloud-init file mounts the volume")
}

func TestCreateWithNothingKeptCreatesNothing(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)

	_, err := simulated.adapter().Create(context.Background(), aSpecification())

	require.ErrorContains(t, err, "hetzner keeps no address named staging for the server")
	keys, servers, _, _ := simulated.counts()
	assert.Zero(t, keys)
	assert.Zero(t, servers)
}

func TestCreateReturnsOnceHetznerSaysTheServerRuns(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	kept(t, adapter)

	_, err := adapter.Create(context.Background(), aSpecification())

	require.NoError(t, err)
	assert.True(t, simulated.waitedFor(simulated.actionOf("create_server")), "the creation of the server is waited for")
	assert.True(t, simulated.waitedFor(simulated.actionOf("start_server")), "and its start")
}

func TestAServerThatFailsToStartIsAnErrorAndStaysForTheCallerToDestroy(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	simulated.failStarts = true
	adapter := simulated.adapter()
	kept(t, adapter)

	_, err := adapter.Create(context.Background(), aSpecification())

	require.ErrorContains(t, err, "wait until the server runs: Action failed")
	require.NoError(t, adapter.Destroy(context.Background(), "staging"))
}

func TestAServerThatCannotBeCreatedLeavesNoKeyBehind(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	simulated.refuseServers = &apiError{http.StatusUnprocessableEntity, "invalid_input", "image ubuntu-26.04 not found"}
	adapter := simulated.adapter()
	kept(t, adapter)

	_, err := adapter.Create(context.Background(), aSpecification())

	require.ErrorContains(t, err, "create the server: image ubuntu-26.04 not found")
	keys, servers, addresses, volumes := simulated.counts()
	assert.Zero(t, keys)
	assert.Zero(t, servers)
	assert.Equal(t, 1, addresses, "the kept address stays")
	assert.Equal(t, 1, volumes, "and so does the kept volume")
}

func TestANameInUseIsTheNameTaken(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	kept(t, adapter)
	_, err := adapter.Create(context.Background(), aSpecification())
	require.NoError(t, err)

	_, err = adapter.Create(context.Background(), aSpecification())

	require.ErrorIs(t, err, provider.ErrNameTaken)
	keys, servers, _, _ := simulated.counts()
	assert.Equal(t, 1, keys, "the key of the server that exists is left as it is")
	assert.Equal(t, 1, servers)
}

func TestFindGivesTheServerWithTheNameAndTheLabelsOfTheEnvironment(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	kept(t, adapter)
	_, err := adapter.Create(context.Background(), aSpecification())
	require.NoError(t, err)

	server, found, err := adapter.Find(context.Background(), "staging")

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, netip.MustParseAddr("198.51.100.7"), server.Address)
	lists := simulated.sent(http.MethodGet, "/servers")
	require.NotEmpty(t, lists)
	assert.Contains(t, lists[len(lists)-1].query, "label_selector=environment%3Dstaging")
	assert.Contains(t, lists[len(lists)-1].query, "name=staging")
}

func TestFindDoesNotGiveAServerOfTheNameWithoutTheLabels(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	production := hetzner.New(hetzner.Settings{
		Endpoint: simulated.URL + "/v1", Token: config.NewSecret(token), Labels: map[string]string{"environment": "production"}, PollInterval: time.Millisecond,
	})
	keeping := aKeeping()
	keeping.Labels = map[string]string{"environment": "production"}
	_, err := production.Keep(context.Background(), keeping)
	require.NoError(t, err)
	specification := aSpecification()
	specification.Labels = map[string]string{"environment": "production"}
	_, err = production.Create(context.Background(), specification)
	require.NoError(t, err)

	_, found, err := simulated.adapter().Find(context.Background(), "staging")

	require.NoError(t, err)
	assert.False(t, found)
}

func TestDestroyLeavesOnlyTheKeptAddressAndVolumeAtHetzner(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	kept(t, adapter)
	_, err := adapter.Create(context.Background(), aSpecification())
	require.NoError(t, err)

	require.NoError(t, adapter.Destroy(context.Background(), "staging"))

	keys, servers, addresses, volumes := simulated.counts()
	assert.Zero(t, keys, "the deploy key goes with the server")
	assert.Zero(t, servers)
	assert.Equal(t, 1, addresses, "the address stays, for the next server")
	assert.Equal(t, 1, volumes, "and so does the volume")
	_, found, err := adapter.Find(context.Background(), "staging")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestFindKeptGivesTheAddressAndTheVolumeLeftAfterTheServerIsDestroyed(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	kept(t, adapter)
	_, err := adapter.Create(context.Background(), aSpecification())
	require.NoError(t, err)
	require.NoError(t, adapter.Destroy(context.Background(), "staging"))

	holdings, err := adapter.FindKept(context.Background(), "staging")

	require.NoError(t, err)
	assert.Equal(t, provider.Holdings{Address: netip.MustParseAddr("198.51.100.7"), VolumeSize: 10}, holdings)
	lists := simulated.sent(http.MethodGet, "/volumes")
	require.NotEmpty(t, lists)
	assert.Contains(t, lists[len(lists)-1].query, "label_selector=environment%3Dstaging")
	assert.Contains(t, lists[len(lists)-1].query, "name=staging")
}

func TestFindKeptWithNothingKeptGivesNothingAndCreatesNothing(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)

	holdings, err := simulated.adapter().FindKept(context.Background(), "staging")

	require.NoError(t, err)
	assert.Equal(t, provider.Holdings{}, holdings)
	assert.False(t, holdings.Address.IsValid())
	assert.Empty(t, simulated.sent(http.MethodPost, "/primary_ips"))
	assert.Empty(t, simulated.sent(http.MethodPost, "/volumes"))
}

func TestTheNextServerGetsTheSameAddressAndVolume(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	first := kept(t, adapter)
	_, err := adapter.Create(context.Background(), aSpecification())
	require.NoError(t, err)
	require.NoError(t, adapter.Destroy(context.Background(), "staging"))

	again := kept(t, adapter)
	server, err := adapter.Create(context.Background(), aSpecification())

	require.NoError(t, err)
	assert.Equal(t, first.Address, server.Address)
	assert.Equal(t, first.VolumeDevice, again.VolumeDevice)
	servers := simulated.sent(http.MethodPost, "/servers")
	require.Len(t, servers, 2)
	assert.Equal(t, servers[0].body["volumes"], servers[1].body["volumes"])
}

func TestDestroyingWhatDoesNotExistSucceedsAndRemovesAKeyLeftBehind(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := simulated.adapter()
	simulated.mu.Lock()
	simulated.keys[1] = map[string]any{"id": 1, "name": "staging", "public_key": deployKey, "labels": map[string]any{"environment": "staging"}, "fingerprint": "x", "created": created}
	simulated.mu.Unlock()

	require.NoError(t, adapter.Destroy(context.Background(), "staging"))

	keys, _, _, _ := simulated.counts()
	assert.Zero(t, keys)
}

func TestAWrongTokenIsAnErrorThatDoesNotCarryTheToken(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedHetzner(t)
	adapter := hetzner.New(hetzner.Settings{Endpoint: simulated.URL + "/v1", Token: config.NewSecret("a-revoked-token"), Labels: map[string]string{"environment": "staging"}})

	_, _, err := adapter.Find(context.Background(), "staging")

	require.ErrorContains(t, err, "unable to authenticate")
	assert.NotContains(t, err.Error(), "a-revoked-token")
}
