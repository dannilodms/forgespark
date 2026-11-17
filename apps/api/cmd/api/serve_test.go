package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// servedAPI is the command that serves the API, running as the process
// would, until the test tells it to stop.
type servedAPI struct {
	// url is the address the API listens on.
	url string
	// signal starts the shutdown, as the first SIGTERM does.
	signal context.CancelFunc
	// status receives the exit status of the command.
	status <-chan int
	// logs is the error output of the command, where its logs go.
	logs *output
}

// launchAPI runs the command that serves the API and waits until it listens.
// The test stops it with signal and reads status.
func launchAPI(t *testing.T, environment map[string]string) servedAPI {
	t.Helper()
	shutdown, signal := context.WithCancel(context.Background())
	var stdout output
	logs := &output{}
	status := make(chan int, 1)
	go func() { status <- run(shutdown, nil, environment, &stdout, logs) }()
	t.Cleanup(signal)

	started := regexp.MustCompile(`"msg":"the API started".*"address":"([^"]+)"`)
	var address string
	require.Eventually(t, func() bool {
		found := started.FindStringSubmatch(logs.String())
		if found == nil {
			return false
		}
		address = found[1]
		return true
	}, 10*time.Second, 10*time.Millisecond, "the API did not start: %s", logs)
	return servedAPI{url: "http://" + address, signal: signal, status: status, logs: logs}
}

// lines reads the logs written so far as one JSON object for each line, and
// fails the test on a line that is not one.
func (a servedAPI) lines(t *testing.T) []map[string]any {
	t.Helper()
	lines := []map[string]any{}
	for line := range strings.Lines(a.logs.String()) {
		var object map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &object), "not a JSON object: %s", line)
		lines = append(lines, object)
	}
	return lines
}

// messages lists the message of each line of the logs written so far.
func (a servedAPI) messages(t *testing.T) []string {
	t.Helper()
	messages := []string{}
	for _, line := range a.lines(t) {
		messages = append(messages, line["msg"].(string))
	}
	return messages
}

// answer is what a client received.
type answer struct {
	status int
	err    error
}

// caller is a client that keeps to one connection of its own, as a monitor
// or a proxy does.
type caller struct {
	client *http.Client
}

func newCaller() caller {
	return caller{client: &http.Client{Transport: &http.Transport{}}}
}

// get sends one request and gives its answer when it arrives.
func (c caller) get(url string, headers ...string) <-chan answer {
	answered := make(chan answer, 1)
	go func() { answered <- c.send(http.MethodGet, url, headers, "") }()
	return answered
}

// send sends one request and waits for its answer. The headers are names
// followed by their values.
func (c caller) send(method, url string, headers []string, body string) answer {
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return answer{err: err}
	}
	for index := 1; index < len(headers); index += 2 {
		request.Header.Set(headers[index-1], headers[index])
	}
	response, err := c.client.Do(request)
	if err != nil {
		return answer{err: err}
	}
	defer response.Body.Close()
	_, err = io.Copy(io.Discard, response.Body)
	return answer{status: response.StatusCode, err: err}
}

// requireRefused waits until the API takes no new connection.
func requireRefused(t *testing.T, apiURL string) {
	t.Helper()
	address := strings.TrimPrefix(apiURL, "http://")
	require.Eventually(t, func() bool {
		connection, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			return true
		}
		_ = connection.Close()
		return false
	}, 20*time.Second, 10*time.Millisecond, "the API still takes a new connection")
}

// stallingPostgres stands between the API and PostgreSQL. While it stalls,
// it holds back what the two say to each other, so that a request that
// reaches the database stays in flight for as long as the test needs.
type stallingPostgres struct {
	// url is the connection URL of the API role, through this proxy.
	url      string
	upstream string
	listener net.Listener

	mu      sync.Mutex
	flowing chan struct{}
	held    atomic.Int64
}

func stallablePostgres(t *testing.T, connectionURL string) *stallingPostgres {
	t.Helper()
	parsed, err := url.Parse(connectionURL)
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	proxy := &stallingPostgres{upstream: parsed.Host, listener: listener, flowing: make(chan struct{})}
	close(proxy.flowing)
	parsed.Host = listener.Addr().String()
	proxy.url = parsed.String()
	go proxy.accept()
	t.Cleanup(func() {
		proxy.resume()
		_ = listener.Close()
	})
	return proxy
}

// stall holds back everything from now on.
func (p *stallingPostgres) stall() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.held.Store(0)
	p.flowing = make(chan struct{})
}

// resume lets through what was held back and what comes after.
func (p *stallingPostgres) resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case <-p.flowing:
	default:
		close(p.flowing)
	}
}

// requireHolding waits until the proxy holds something back: a query that
// PostgreSQL has not seen yet.
func (p *stallingPostgres) requireHolding(t *testing.T) {
	t.Helper()
	require.Eventually(t, func() bool { return p.held.Load() > 0 }, 10*time.Second, 5*time.Millisecond,
		"nothing reached PostgreSQL while it stalled")
}

func (p *stallingPostgres) accept() {
	for p.acceptOne() {
	}
}

func (p *stallingPostgres) acceptOne() bool {
	connection, err := p.listener.Accept()
	if err != nil {
		return false
	}
	go p.forward(connection)
	return true
}

func (p *stallingPostgres) forward(client net.Conn) {
	server, err := net.Dial("tcp", p.upstream)
	if err != nil {
		_ = client.Close()
		return
	}
	go p.pipe(server, client)
	p.pipe(client, server)
}

func (p *stallingPostgres) pipe(to, from net.Conn) {
	defer to.Close()
	defer from.Close()
	buffer := make([]byte, 32*1024)
	for p.relay(to, from, buffer) {
	}
}

// relay passes on one read, after the stall ends. It tells whether the
// connection is still open.
func (p *stallingPostgres) relay(to, from net.Conn, buffer []byte) bool {
	read, err := from.Read(buffer)
	if err != nil {
		return false
	}
	p.mu.Lock()
	flowing := p.flowing
	p.mu.Unlock()
	select {
	case <-flowing:
	default:
		p.held.Add(1)
		<-flowing
	}
	_, err = to.Write(buffer[:read])
	return err == nil
}
