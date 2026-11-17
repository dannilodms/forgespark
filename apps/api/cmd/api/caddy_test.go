package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database/databasetest"
)

// The API behind Caddy, as development runs it (Delivery and Health in
// CODING_STANDARDS.md). Each test starts a container of its own from the
// Caddy image of compose.yaml, with the Caddyfile of development, in front of
// an API and a stand-in for Vite that listen on the machine. The image is
// built by `task services`, which `task api:test` runs first.

const (
	// caddyImage is the image of the Caddy service of compose.yaml.
	caddyImage = "software-factory-manager-caddy:development"
	// caddyfile is the Caddyfile of development. The test runs in the
	// directory of its package: the root of the repository is four up.
	caddyfile = "../../../../deploy/Caddyfile"
	// theApp is what the stand-in for Vite answers to every request.
	theApp = "the app"
	// wrongSetupCode is a setup code that opens nothing.
	wrongSetupCode = `{"code":"ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ"}`
)

// site is Caddy serving the app and the API at one origin.
type site struct {
	url    string
	client *http.Client
}

// reply is what the site answered.
type reply struct {
	received
	header http.Header
}

// machineHost is the interface where what Caddy reaches listens: the host of
// API_LISTEN_ADDRESS, as the API of development does. Docker Desktop reaches
// 127.0.0.1 of the machine; Docker Engine on Linux needs 0.0.0.0.
func machineHost(t *testing.T) string {
	t.Helper()
	address, found := os.LookupEnv("API_LISTEN_ADDRESS")
	require.True(t, found, "API_LISTEN_ADDRESS is not set: run the tests with `task api:test`, which reads .env.example")
	host, _, err := net.SplitHostPort(address)
	require.NoError(t, err)
	return host
}

// onTheMachine gives the variables of an API on a temporary database, listening
// where Caddy reaches it.
func onTheMachine(t *testing.T, temporary databasetest.Database) map[string]string {
	t.Helper()
	environment := environmentOf(temporary)
	environment["API_LISTEN_ADDRESS"] = net.JoinHostPort(machineHost(t), "0")
	return environment
}

// docker runs a command of the Docker CLI and returns what it printed.
func docker(t *testing.T, arguments ...string) string {
	t.Helper()
	var errorOutput bytes.Buffer
	command := exec.CommandContext(t.Context(), "docker", arguments...)
	command.Stderr = &errorOutput
	printed, err := command.Output()
	require.NoError(t, err, "docker %s: %s", strings.Join(arguments, " "), errorOutput.String())
	return strings.TrimSpace(string(printed))
}

// standInForVite listens where Vite would and answers theApp to everything.
// It returns its port.
func standInForVite(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", net.JoinHostPort(machineHost(t), "0"))
	require.NoError(t, err)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, theApp)
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	return port
}

// behindCaddy starts Caddy in front of the API at apiURL and of a stand-in for
// Vite, and waits until it answers.
func behindCaddy(t *testing.T, apiURL string) site {
	t.Helper()
	_, err := exec.CommandContext(t.Context(), "docker", "image", "inspect", caddyImage).Output()
	require.NoError(t, err, "the image %s is not built: run `task services`", caddyImage)
	api, err := url.Parse(apiURL)
	require.NoError(t, err)
	file, err := filepath.Abs(caddyfile)
	require.NoError(t, err)
	// Docker Desktop creates a directory where a file to mount is missing.
	require.FileExists(t, file)

	container := docker(t, "run", "--detach", "--rm",
		"--label", "software-factory-manager.test=caddy",
		"--publish", "127.0.0.1::80",
		"--add-host", "host.docker.internal:host-gateway",
		"--mount", "type=bind,readonly,source="+file+",target=/etc/caddy/Caddyfile",
		"--env", "OTEL_TRACES_EXPORTER=none",
		"--env", "API_UPSTREAM=host.docker.internal:"+api.Port(),
		"--env", "WEB_UPSTREAM=host.docker.internal:"+standInForVite(t),
		caddyImage)
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "--force", container).Run()
	})

	published := strings.Fields(docker(t, "port", container, "80/tcp"))
	require.NotEmpty(t, published)
	caddy := site{
		url: "http://" + published[0],
		client: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
	require.Eventually(t, func() bool {
		request, err := http.NewRequest(http.MethodGet, caddy.url+"/", nil)
		require.NoError(t, err)
		request.Host = "localhost"
		response, err := caddy.client.Do(request)
		if err != nil {
			return false
		}
		_ = response.Body.Close()
		return true
	}, 30*time.Second, 50*time.Millisecond, "Caddy did not answer: %s", docker(t, "logs", container))
	return caddy
}

// send sends a request to the site at the address a browser opens, and keeps
// the path as it is written, dot segments included.
func (s site) send(t *testing.T, method, path, body string) reply {
	t.Helper()
	request, err := http.NewRequest(method, s.url+path, strings.NewReader(body))
	require.NoError(t, err)
	request.Host = "localhost"
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	answer := reply{
		received: received{
			status:    response.StatusCode,
			mediaType: response.Header.Get("Content-Type"),
			allow:     response.Header.Get("Allow"),
			body:      string(raw),
		},
		header: response.Header,
	}
	_ = json.Unmarshal(raw, &answer.fields)
	return answer
}

// committedSchemas reads the schemas of the committed spec, the contract the
// app is generated from.
func committedSchemas(t *testing.T) huma.Registry {
	t.Helper()
	content, err := os.ReadFile(committedSpec)
	require.NoError(t, err)
	var spec struct {
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(content, &spec))
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)
	for name, schema := range spec.Components.Schemas {
		registry.Map()[name] = schemaOf(t, schema)
	}
	return registry
}

func schemaOf(t *testing.T, raw any) *huma.Schema {
	t.Helper()
	encoded, err := json.Marshal(raw)
	require.NoError(t, err)
	schema := &huma.Schema{}
	require.NoError(t, json.Unmarshal(encoded, schema))
	return schema
}

// assertIsAProblem checks a body against the schema Problem of the committed
// spec.
func assertIsAProblem(t *testing.T, body string) {
	t.Helper()
	registry := committedSchemas(t)
	require.Contains(t, registry.Map(), "Problem")
	var value any
	require.NoError(t, json.Unmarshal([]byte(body), &value))
	result := &huma.ValidateResult{}

	huma.Validate(registry, registry.Map()["Problem"], huma.NewPathBuffer([]byte{}, 0), huma.ModeReadFromServer, value, result)

	assert.Empty(t, result.Errors, "the body does not follow the schema Problem of %s: %s", committedSpec, body)
}

func TestThroughCaddyTheAppAndTheAPIShareOneOrigin(t *testing.T) {
	t.Parallel()
	caddy := behindCaddy(t, startAPI(t, onTheMachine(t, databasetest.Migrated(t))))

	app := caddy.send(t, http.MethodGet, "/setup", "")
	api := caddy.send(t, http.MethodGet, "/api/setup/unknown", "")

	assert.Equal(t, http.StatusOK, app.status)
	assert.Equal(t, theApp, app.body)
	assert.Equal(t, http.StatusNotFound, api.status, api.body)
	assert.Equal(t, "application/problem+json", api.mediaType)
	assert.Equal(t, "not_found", api.fields["code"])
	assert.Equal(t, "no-store", api.header.Get("Cache-Control"))
}

func TestCaddyAnswers429InTheErrorFormatPastTheLimitOfTheSetup(t *testing.T) {
	t.Parallel()
	caddy := behindCaddy(t, startAPI(t, onTheMachine(t, databasetest.Migrated(t))))

	// Ten a minute pass untouched: the API answers each.
	for range 10 {
		within := caddy.send(t, http.MethodPost, "/api/setup/open", wrongSetupCode)
		requireSetupCodeInvalid(t, within.received)
		assert.Equal(t, "no-store", within.header.Get("Cache-Control"))
	}
	over := caddy.send(t, http.MethodPost, "/api/setup/open", wrongSetupCode)

	require.Equal(t, http.StatusTooManyRequests, over.status, over.body)
	assert.Equal(t, "application/problem+json", over.mediaType)
	assert.Equal(t, "no-store", over.header.Get("Cache-Control"))
	assert.NotEmpty(t, over.header.Get("Retry-After"))
	traceID, _ := over.fields["trace_id"].(string)
	assert.Regexp(t, `^[0-9a-f]{32}$`, traceID, "Caddy starts the trace of the request")
	assert.JSONEq(t, `{
		"title": "Too Many Requests",
		"status": 429,
		"code": "too_many_requests",
		"params": {},
		"trace_id": "`+traceID+`"
	}`, over.body)
	assertIsAProblem(t, over.body)
	// The POST operations of the setup share the limit.
	for _, path := range []string{"/api/setup/uploads", "/api/setup/conclude", "/api/setup/administrators"} {
		assert.Equal(t, http.StatusTooManyRequests, caddy.send(t, http.MethodPost, path, wrongSetupCode).status, "POST %s", path)
	}
}

func TestTheReadOfTheSetupAndEveryOtherPathHaveNoLimit(t *testing.T) {
	t.Parallel()
	caddy := behindCaddy(t, startAPI(t, onTheMachine(t, databasetest.Migrated(t))))
	// Past the limit of the setup first, so that what follows comes from an
	// address that is over it.
	for range 11 {
		caddy.send(t, http.MethodPost, "/api/setup/open", wrongSetupCode)
	}
	require.Equal(t, http.StatusTooManyRequests, caddy.send(t, http.MethodPost, "/api/setup/open", wrongSetupCode).status)

	for range 15 {
		assert.NotEqual(t, http.StatusTooManyRequests, caddy.send(t, http.MethodGet, "/api/setup", "").status)
		assert.NotEqual(t, http.StatusTooManyRequests, caddy.send(t, http.MethodPost, "/api/sessions", "{}").status)
		assert.NotEqual(t, http.StatusTooManyRequests, caddy.send(t, http.MethodGet, "/setup", "").status)
	}
}

// composeNetwork is the network of the services of compose.yaml, where
// Garage is reached by the name of its service.
const composeNetwork = "software-factory-manager_default"

// filesBehindCaddy starts Caddy with the Caddyfile of development on the
// network of the services, so that the host name of the files leads to the
// Garage of development, as it does in compose.yaml. It returns the address
// Caddy listens on.
func filesBehindCaddy(t *testing.T) string {
	t.Helper()
	file, err := filepath.Abs(caddyfile)
	require.NoError(t, err)
	require.FileExists(t, file)
	container := docker(t, "run", "--detach", "--rm",
		"--label", "software-factory-manager.test=caddy",
		"--network", composeNetwork,
		"--publish", "127.0.0.1::80",
		"--mount", "type=bind,readonly,source="+file+",target=/etc/caddy/Caddyfile",
		"--env", "OTEL_TRACES_EXPORTER=none",
		caddyImage)
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "--force", container).Run()
	})
	published := strings.Fields(docker(t, "port", container, "80/tcp"))
	require.NotEmpty(t, published)
	require.Eventually(t, func() bool {
		response, err := http.Get("http://" + published[0] + "/")
		if err != nil {
			return false
		}
		_ = response.Body.Close()
		return true
	}, 30*time.Second, 50*time.Millisecond, "Caddy did not answer: %s", docker(t, "logs", container))
	return published[0]
}

// throughCaddy sends a request to Caddy as a browser on the machine sends it
// to the URL: with the host name of the URL as its Host.
func throughCaddy(t *testing.T, caddy, method, signedURL string, headers map[string]string, body []byte) *http.Response {
	t.Helper()
	target, err := url.Parse(signedURL)
	require.NoError(t, err)
	host := target.Host
	target.Host = caddy
	request, err := http.NewRequest(method, target.String(), bytes.NewReader(body))
	require.NoError(t, err)
	request.Host = host
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, response.Body)
	_ = response.Body.Close()
	return response
}

// The browser sends an image to the URL the API signed, at the host name of
// the files. Caddy forwards it to Garage with the Host the signature covers,
// and Garage answers the cross-origin questions from the rule of its bucket.
func TestThroughCaddyTheHostNameOfTheFilesLeadsToTheStorage(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	_, code := issueSetupCodeIn(t, temporary)
	api := startAPI(t, environmentOf(temporary))
	started := startUpload(t, api, uploadBody(code, "consultancy_logo", "image/png", 2048))
	require.Equal(t, http.StatusCreated, started.status, started.body)
	signedURL, _ := started.fields["upload_url"].(string)
	caddy := filesBehindCaddy(t)
	appOrigin := os.Getenv("APP_ORIGIN")

	fromTheApp := throughCaddy(t, caddy, http.MethodOptions, signedURL, map[string]string{
		"Origin": appOrigin, "Access-Control-Request-Method": http.MethodPut, "Access-Control-Request-Headers": "content-type",
	}, nil)
	fromElsewhere := throughCaddy(t, caddy, http.MethodOptions, signedURL, map[string]string{
		"Origin": "http://elsewhere.localhost", "Access-Control-Request-Method": http.MethodPut,
	}, nil)
	sent := throughCaddy(t, caddy, http.MethodPut, signedURL, map[string]string{
		"Origin": appOrigin, "Content-Type": "image/png",
	}, aPNG(2048))

	assert.Equal(t, http.StatusOK, fromTheApp.StatusCode)
	assert.Equal(t, appOrigin, fromTheApp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusForbidden, fromElsewhere.StatusCode, "the storage lets another origin send a file")
	assert.Equal(t, http.StatusOK, sent.StatusCode, "the storage refused the image sent through Caddy")
	assert.Equal(t, appOrigin, sent.Header.Get("Access-Control-Allow-Origin"))
}

func TestLivenessAndReadinessCannotBeReachedThroughCaddy(t *testing.T) {
	t.Parallel()
	caddy := behindCaddy(t, startAPI(t, onTheMachine(t, databasetest.Migrated(t))))

	for _, path := range []string{"/livez", "/readyz", "/api/../readyz", "/api/%2e%2e/livez", "/api/..%2freadyz"} {
		answer := caddy.send(t, http.MethodGet, path, "")

		assert.Equal(t, theApp, answer.body, "%s reached something other than the app: %d", path, answer.status)
	}
}

func TestUpAnswersTheStatusOfReadinessWithNoBody(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		environment func(t *testing.T) map[string]string
		status      int
	}{
		"when the API is ready": {
			environment: func(t *testing.T) map[string]string { return onTheMachine(t, databasetest.Migrated(t)) },
			status:      http.StatusOK,
		},
		"when PostgreSQL does not answer": {
			environment: func(t *testing.T) map[string]string {
				return withFileStorage(map[string]string{
					"API_LISTEN_ADDRESS": net.JoinHostPort(machineHost(t), "0"),
					// Port 1 of the machine has no PostgreSQL behind it.
					"DATABASE_URL": "postgres://api:api@127.0.0.1:1/software_factory?sslmode=disable",
				})
			},
			status: http.StatusServiceUnavailable,
		},
	}
	for name, ready := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			caddy := behindCaddy(t, startAPI(t, ready.environment(t)))

			up := caddy.send(t, http.MethodGet, "/up", "")

			assert.Equal(t, ready.status, up.status)
			assert.Empty(t, up.body)
		})
	}
}

func TestUpAnswersAFailureWhenTheAPIIsStopped(t *testing.T) {
	t.Parallel()
	api := launchAPI(t, onTheMachine(t, databasetest.Migrated(t)))
	caddy := behindCaddy(t, api.url)
	api.signal()
	require.Equal(t, 0, <-api.status, "the API stopped with: %s", api.logs)

	up := caddy.send(t, http.MethodGet, "/up", "")

	assert.GreaterOrEqual(t, up.status, http.StatusInternalServerError)
	assert.Empty(t, up.body)
}
