package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/tools/environment/provider"
	"github.com/dannilodms/software-factory-manager/apps/api/tools/environment/provider/providertest"
	"github.com/dannilodms/software-factory-manager/apps/api/tools/environment/sshclient"
)

// The commands run as the process would, on a checkout written into a
// temporary directory with the committed choices and cloud-init file of this
// repository, against a provider and a readiness kept in memory.

// repository is the root of this repository, from the directory of the
// package.
const repository = "../../../.."

// token is the token of the provider the tests give the commands.
const token = "hetzner-token-only-the-tests-know"

// committed are the files of the repository the commands read.
var committed = []string{"deploy/server.staging.yaml", "deploy/cloud-init.yaml"}

// clock stands still until the test moves it.
type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

// fakeReadiness answers what the test tells it and keeps the servers it
// was asked to wait for.
type fakeReadiness struct {
	err     error
	targets []sshclient.Target
}

func (r *fakeReadiness) Wait(_ context.Context, target sshclient.Target) error {
	r.targets = append(r.targets, target)
	return r.err
}

// world is what the commands of one test reach.
type world struct {
	t         *testing.T
	directory string
	provider  *providertest.Provider
	readiness *fakeReadiness
	clock     *clock
	variables map[string]string
	// tokens are the tokens the provider was reached with.
	tokens []config.Secret
	labels []map[string]string
}

// aWorld writes a main checkout of the repository and runs the commands in
// it, with the token in the variables of the process.
func aWorld(t *testing.T) *world {
	t.Helper()
	root := t.TempDir()
	copyCommitted(t, root)
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o700))
	now := &clock{now: time.Date(2026, 11, 19, 14, 0, 0, 0, time.UTC)}
	return &world{
		t:         t,
		directory: root,
		provider:  providertest.NewProvider(now.Now),
		readiness: &fakeReadiness{},
		clock:     now,
		variables: map[string]string{"HETZNER_TOKEN": token},
	}
}

func copyCommitted(t *testing.T, root string) {
	t.Helper()
	for _, file := range committed {
		content, err := os.ReadFile(filepath.Join(repository, file))
		require.NoError(t, err)
		target := filepath.Join(root, filepath.FromSlash(file))
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o700))
		require.NoError(t, os.WriteFile(target, content, 0o600))
	}
}

func (w *world) outside() outside {
	return outside{
		directory: w.directory,
		connect: func(_ string, token config.Secret, labels map[string]string) provider.Provider {
			w.tokens = append(w.tokens, token)
			w.labels = append(w.labels, labels)
			return w.provider
		},
		readiness: w.readiness,
		now:       w.clock.Now,
	}
}

// result is what one command printed, and its exit status.
type result struct {
	status         int
	stdout, stderr string
}

func (w *world) run(arguments ...string) result {
	w.t.Helper()
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), arguments, w.variables, w.outside(), &stdout, &stderr)
	return result{status: status, stdout: stdout.String(), stderr: stderr.String()}
}

// succeeds runs a command that must succeed and print no error.
func (w *world) succeeds(arguments ...string) string {
	w.t.Helper()
	ran := w.run(arguments...)
	require.Equal(w.t, 0, ran.status, ran.stderr)
	require.Empty(w.t, ran.stderr)
	return ran.stdout
}

// localCopy is the folder where the main checkout keeps the keys of
// staging.
func (w *world) localCopy() string {
	return filepath.Join(w.directory, ".servers", "staging")
}

// specification is what the server of staging was created from.
func (w *world) specification() provider.Specification {
	w.t.Helper()
	specification, found := w.provider.SpecificationOf("staging")
	require.True(w.t, found, "the server of staging exists")
	return specification
}
