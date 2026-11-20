package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/dannilodms/software-factory-manager/apps/api/tools/environment/sshclient"
)

// The tests of the committed files of deploy/ that the tool reads: what the
// first real turn-on would otherwise be the first to find.

// dockerPackages are the packages of Docker that the cloud-init file
// installs from Docker's repository, each at the version it names.
var dockerPackages = []string{"containerd.io", "docker-ce-cli", "docker-ce", "docker-compose-plugin"}

// committedCloudInit is the part of the cloud-init file these tests read.
type committedCloudInit struct {
	Packages []any    `yaml:"packages"`
	RunCmd   []string `yaml:"runcmd"`
}

func readCommittedCloudInit(t *testing.T) committedCloudInit {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repository, "deploy", "cloud-init.yaml"))
	require.NoError(t, err)
	var file committedCloudInit
	require.NoError(t, yaml.Unmarshal(content, &file), "the committed file, before the tool fills it, is YAML")
	return file
}

// pinned gives the version of each package, empty for one with none.
func pinned(t *testing.T, file committedCloudInit) map[string]string {
	t.Helper()
	versions := map[string]string{}
	for _, entry := range file.Packages {
		name, version := packageOf(t, entry)
		versions[name] = version
	}
	return versions
}

// packageOf reads an entry of the packages: a name, or [name, version].
func packageOf(t *testing.T, entry any) (string, string) {
	t.Helper()
	pair, isPair := entry.([]any)
	if !isPair {
		return fmt.Sprint(entry), ""
	}
	require.Len(t, pair, 2, "a package with a version is [name, version]")
	return fmt.Sprint(pair[0]), fmt.Sprint(pair[1])
}

func TestTheCloudInitFileInstallsDockerAtExactVersionsBuiltForTheSystemImageOfEveryEnvironment(t *testing.T) {
	t.Parallel()
	versions := pinned(t, readCommittedCloudInit(t))
	files, err := filepath.Glob(filepath.Join(repository, "deploy", "server.*.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	for _, file := range files {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "server."), ".yaml")
		assertDockerBuiltForTheSystemImageOf(t, name, versions)
	}
}

func assertDockerBuiltForTheSystemImageOf(t *testing.T, name string, versions map[string]string) {
	t.Helper()
	chosen, err := loadChoices(repository, name)
	require.NoError(t, err)
	release := regexp.MustCompile(`^ubuntu-(\d\d\.\d\d)$`).FindStringSubmatch(chosen.SystemImage)
	require.NotNil(t, release, "the system image of %s names a release of Ubuntu", name)
	exact := `^(\d+:)?\d+\.\d+\.\d+-\d+~ubuntu\.` + regexp.QuoteMeta(release[1]) + `~[a-z]+$`
	for _, docker := range dockerPackages {
		assert.Regexp(t, exact, versions[docker],
			"%s is installed at an exact version built for Ubuntu %s, the system image of %s", docker, release[1], name)
	}
}

func TestTheCloudInitFileStopsAtTheFirstFailureAndSignalsLastThatTheServerIsPrepared(t *testing.T) {
	t.Parallel()
	steps := readCommittedCloudInit(t).RunCmd

	require.NotEmpty(t, steps)
	assert.Equal(t, "set -eu", steps[0], "the steps are one script, which stops at the first one that fails")
	assert.Equal(t, "touch "+sshclient.PreparedSignal, steps[len(steps)-1], "the last step writes the signal readiness waits for")
	assert.True(t, slices.Contains(steps, "ufw --force enable"))
	assert.Less(t, slices.Index(steps, "ufw allow 22/tcp"), slices.Index(steps, "ufw --force enable"), "SSH is let in before the firewall starts")
}

func TestTheRenderedCloudInitFileFitsTheUserDataOfAServer(t *testing.T) {
	t.Parallel()
	keys, err := newServerKeys("staging")
	require.NoError(t, err)

	rendered, err := renderCloudInit(repository, keys)

	require.NoError(t, err)
	assert.Less(t, len(rendered.Reveal()), 32*1024, "Hetzner takes 32 KiB of user data")
	assert.NotContains(t, rendered.Reveal(), "{{")
}

func TestEveryStagingValueIsOfANameOfEnvExample(t *testing.T) {
	t.Parallel()
	names, err := namesIn(filepath.Join(repository, ".env.example"))
	require.NoError(t, err)

	for name := range stagingValues {
		assert.Contains(t, names, name, "the staging value of %s is of no name of .env.example: remove it", name)
	}
}

// required finds a variable that Compose requires: ${NAME:?message}.
var required = regexp.MustCompile(`\$\{([A-Z][A-Z0-9_]*):\?`)

func TestEveryVariableTheComposeFilesOfStagingRequireHasAValueOnEveryPublish(t *testing.T) {
	t.Parallel()
	world := aWorld(t)
	world.readyToPublish()
	world.succeeds("publish", "staging")
	values := world.variablesOnServer()

	for _, file := range []string{"compose.staging.yaml", "compose.base.yaml"} {
		assertEveryRequiredVariableHasAValue(t, file, values)
	}
}

func assertEveryRequiredVariableHasAValue(t *testing.T, file string, values map[string]string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repository, "deploy", file))
	require.NoError(t, err)
	for _, found := range required.FindAllStringSubmatch(string(content), -1) {
		assert.NotEmpty(t, values[found[1]], "deploy/%s requires %s, which the publish leaves empty", file, found[1])
	}
}

func TestTheComposeFileOfStagingDeclaresTheVolumesThePublishRemovesAndTheOnesItKeeps(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile(filepath.Join(repository, "deploy", "compose.staging.yaml"))
	require.NoError(t, err)
	var file struct {
		Name    string         `yaml:"name"`
		Volumes map[string]any `yaml:"volumes"`
	}
	require.NoError(t, yaml.Unmarshal(content, &file))

	assert.Equal(t, "software-factory-manager-staging", file.Name)
	assert.ElementsMatch(t, append(slices.Clone(recreatedVolumes), "caddy-data", "caddy-config"), slices.Collect(maps.Keys(file.Volumes)),
		"the publish removes the data of PostgreSQL and Garage and keeps the certificates of Caddy")
}

func TestTheCaddyfileOfStagingAsksForTheUserOfMailpitThatTheCommentNames(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile(filepath.Join(repository, "deploy", "Caddyfile.staging"))
	require.NoError(t, err)

	assert.Regexp(t, `\n\s+`+mailpitUser+` \{\$MAILPIT_PASSWORD_HASH\}\n`, string(content))
}
