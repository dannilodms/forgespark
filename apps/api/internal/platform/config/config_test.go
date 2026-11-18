package config_test

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

func validEnvironment() map[string]string {
	return map[string]string{
		"API_LISTEN_ADDRESS":             "127.0.0.1:8080",
		"DATABASE_URL":                   "postgres://api:api-password@localhost:5432/software_factory",
		"MIGRATION_DATABASE_URL":         "postgres://migration:migration-password@localhost:5432/software_factory",
		"APP_ORIGIN":                     "https://factory.example.com",
		"FILE_STORAGE_URL":               "http://garage:3900",
		"FILE_STORAGE_PUBLIC_URL":        "https://files.factory.example.com",
		"FILE_STORAGE_REGION":            "garage",
		"FILE_STORAGE_BUCKET":            "factory-files",
		"FILE_STORAGE_ACCESS_KEY_ID":     "GK0123456789abcdef01234567",
		"FILE_STORAGE_SECRET_ACCESS_KEY": "storage-secret-access-key",
	}
}

func TestLoadReadsWhereTheFileStorageIsAndTheKeyToSignWith(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, validEnvironment())

	require.NoError(t, err)
	assert.Equal(t, "http://garage:3900", loaded.FileStorageURL.String())
	assert.Equal(t, "https://files.factory.example.com", loaded.FileStoragePublicURL.String())
	assert.Equal(t, "garage", loaded.FileStorageRegion)
	assert.Equal(t, "factory-files", loaded.FileStorageBucket)
	assert.Equal(t, "GK0123456789abcdef01234567", loaded.FileStorageAccessKeyID)
	assert.Equal(t, "storage-secret-access-key", loaded.FileStorageSecretAccessKey.Reveal())
}

func TestALoadedConfigurationNeverPrintsTheSecretKeyOfTheFileStorage(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, validEnvironment())
	require.NoError(t, err)

	for _, printed := range []string{
		fmt.Sprint(loaded), fmt.Sprintf("%+v", loaded), fmt.Sprintf("%#v", loaded), loaded.FileStorageSecretAccessKey.String(),
	} {
		assert.NotContains(t, printed, "storage-secret-access-key")
	}
}

func TestOnlyTheAPINeedsTheFileStorage(t *testing.T) {
	environment := validEnvironment()
	maps.DeleteFunc(environment, func(variable, _ string) bool { return strings.HasPrefix(variable, "FILE_STORAGE_") })

	_, migrate := config.Load(config.Migrate, environment)
	_, issue := config.Load(config.IssueSetupCode, environment)

	assert.NoError(t, migrate)
	assert.NoError(t, issue)
}

func without(environment map[string]string, variable string) map[string]string {
	delete(environment, variable)
	return environment
}

func with(environment map[string]string, variable, value string) map[string]string {
	environment[variable] = value
	return environment
}

func TestLoadReadsTheEnvironmentIntoTypedValues(t *testing.T) {
	environment := validEnvironment()
	environment["SHUTDOWN_DRAIN_DEADLINE"] = "5s"
	environment["LOG_LEVEL"] = "debug"
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = "http://collector:4318"
	environment["DATABASE_POOL_MAX_SIZE"] = "4"
	environment["DATABASE_STATEMENT_TIMEOUT"] = "5s"
	environment["DATABASE_IDLE_IN_TRANSACTION_TIMEOUT"] = "2m"
	environment["HEALTH_CHECK_TIMEOUT"] = "500ms"

	loaded, err := config.Load(config.ServeAPI, environment)

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", loaded.ListenAddress.String())
	assert.Equal(t, 5*time.Second, loaded.ShutdownDrainDeadline)
	assert.Equal(t, slog.LevelDebug, loaded.LogLevel)
	assert.Equal(t, "http://collector:4318", loaded.TelemetryExportURL.String())
	assert.Equal(t, "postgres://api:api-password@localhost:5432/software_factory", loaded.DatabaseURL.ConnectionString())
	assert.Equal(t, int32(4), loaded.DatabasePoolMaxSize)
	assert.Equal(t, 5*time.Second, loaded.DatabaseStatementTimeout)
	assert.Equal(t, 2*time.Minute, loaded.DatabaseIdleInTransactionTimeout)
	assert.Equal(t, 500*time.Millisecond, loaded.HealthCheckTimeout)
}

func TestLoadFallsBackToTheDefaults(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, validEnvironment())

	require.NoError(t, err)
	assert.Equal(t, 20*time.Second, loaded.ShutdownDrainDeadline)
	assert.Equal(t, slog.LevelInfo, loaded.LogLevel)
	assert.False(t, loaded.TelemetryExportURL.IsSet(), "telemetry is exported to no address by default")
	assert.Equal(t, int32(10), loaded.DatabasePoolMaxSize)
	assert.Equal(t, 30*time.Second, loaded.DatabaseStatementTimeout)
	assert.Equal(t, time.Minute, loaded.DatabaseIdleInTransactionTimeout)
	assert.Equal(t, 2*time.Second, loaded.HealthCheckTimeout)
}

func TestAnEmptyExportAddressIsNoAddress(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, with(validEnvironment(), "OTEL_EXPORTER_OTLP_ENDPOINT", ""))

	require.NoError(t, err)
	assert.False(t, loaded.TelemetryExportURL.IsSet())
}

func TestLoadNamesARequiredVariableThatIsMissing(t *testing.T) {
	cases := []struct {
		command  config.Command
		variable string
	}{
		{config.ServeAPI, "API_LISTEN_ADDRESS"},
		{config.ServeAPI, "DATABASE_URL"},
		{config.ServeAPI, "FILE_STORAGE_URL"},
		{config.ServeAPI, "FILE_STORAGE_PUBLIC_URL"},
		{config.ServeAPI, "FILE_STORAGE_REGION"},
		{config.ServeAPI, "FILE_STORAGE_BUCKET"},
		{config.ServeAPI, "FILE_STORAGE_ACCESS_KEY_ID"},
		{config.ServeAPI, "FILE_STORAGE_SECRET_ACCESS_KEY"},
		{config.Migrate, "MIGRATION_DATABASE_URL"},
		{config.IssueSetupCode, "DATABASE_URL"},
		{config.IssueSetupCode, "APP_ORIGIN"},
	}
	for _, c := range cases {
		t.Run(string(c.command)+" without "+c.variable, func(t *testing.T) {
			_, err := config.Load(c.command, without(validEnvironment(), c.variable))

			require.EqualError(t, err, "configuration: "+c.variable+" is required")
		})
	}
}

func TestLoadTreatsAnEmptyRequiredVariableAsMissing(t *testing.T) {
	_, err := config.Load(config.ServeAPI, with(validEnvironment(), "DATABASE_URL", ""))

	require.EqualError(t, err, "configuration: DATABASE_URL is required")
}

func TestEachCommandRequiresOnlyTheDatabaseRoleItUses(t *testing.T) {
	_, err := config.Load(config.ServeAPI, without(validEnvironment(), "MIGRATION_DATABASE_URL"))
	require.NoError(t, err)

	_, err = config.Load(config.Migrate, map[string]string{
		"MIGRATION_DATABASE_URL": "postgres://migration:migration-password@localhost:5432/software_factory",
	})
	require.NoError(t, err)
}

func TestLoadNamesAnInvalidVariableAndNeverShowsItsValue(t *testing.T) {
	cases := []struct {
		variable string
		value    string
		line     string
	}{
		{"API_LISTEN_ADDRESS", "no-port-here", "configuration: API_LISTEN_ADDRESS is not valid: expected host:port"},
		{"SHUTDOWN_DRAIN_DEADLINE", "twenty", "configuration: SHUTDOWN_DRAIN_DEADLINE is not valid: expected a duration such as 30s"},
		{"SHUTDOWN_DRAIN_DEADLINE", "0s", "configuration: SHUTDOWN_DRAIN_DEADLINE is not valid: expected more than zero"},
		{"LOG_LEVEL", "chatty", "configuration: LOG_LEVEL is not valid: expected debug, info, warn or error"},
		{"OTEL_EXPORTER_OTLP_ENDPOINT", "collector:4318", "configuration: OTEL_EXPORTER_OTLP_ENDPOINT is not valid: expected an http or https URL"},
		{"OTEL_EXPORTER_OTLP_ENDPOINT", "grpc://collector:4317", "configuration: OTEL_EXPORTER_OTLP_ENDPOINT is not valid: expected an http or https URL"},
		{"DATABASE_URL", "mysql://api:hunter2@localhost/db", "configuration: DATABASE_URL is not valid: expected a PostgreSQL connection URL"},
		{"DATABASE_URL", "postgres://api:hunter2@localhost:port/db", "configuration: DATABASE_URL is not valid: expected a PostgreSQL connection URL"},
		{"DATABASE_URL", "host=localhost password=hunter2", "configuration: DATABASE_URL is not valid: expected a PostgreSQL connection URL"},
		{"DATABASE_POOL_MAX_SIZE", "plenty", "configuration: DATABASE_POOL_MAX_SIZE is not valid: expected a whole number"},
		{"DATABASE_POOL_MAX_SIZE", "0", "configuration: DATABASE_POOL_MAX_SIZE is not valid: expected at least 1"},
		{"DATABASE_STATEMENT_TIMEOUT", "45", "configuration: DATABASE_STATEMENT_TIMEOUT is not valid: expected a duration such as 30s"},
		{"DATABASE_STATEMENT_TIMEOUT", "-5s", "configuration: DATABASE_STATEMENT_TIMEOUT is not valid: expected more than zero"},
		{"DATABASE_IDLE_IN_TRANSACTION_TIMEOUT", "0s", "configuration: DATABASE_IDLE_IN_TRANSACTION_TIMEOUT is not valid: expected more than zero"},
		{"HEALTH_CHECK_TIMEOUT", "soon", "configuration: HEALTH_CHECK_TIMEOUT is not valid: expected a duration such as 30s"},
		{"HEALTH_CHECK_TIMEOUT", "0s", "configuration: HEALTH_CHECK_TIMEOUT is not valid: expected more than zero"},
		{"APP_ORIGIN", "factory.example.com", "configuration: APP_ORIGIN is not valid: expected an origin such as https://example.com"},
		{"APP_ORIGIN", "ftp://factory.example.com", "configuration: APP_ORIGIN is not valid: expected an origin such as https://example.com"},
		{"APP_ORIGIN", "https://factory.example.com/app", "configuration: APP_ORIGIN is not valid: expected an origin such as https://example.com"},
		{"APP_ORIGIN", "https://factory.example.com?next=setup", "configuration: APP_ORIGIN is not valid: expected an origin such as https://example.com"},
		{"APP_ORIGIN", "https://someone:hunter2@factory.example.com", "configuration: APP_ORIGIN is not valid: expected an origin such as https://example.com"},
		{"FILE_STORAGE_URL", "garage:3900", "configuration: FILE_STORAGE_URL is not valid: expected an origin such as https://example.com"},
		{"FILE_STORAGE_URL", "http://garage:3900/bucket", "configuration: FILE_STORAGE_URL is not valid: expected an origin such as https://example.com"},
		{"FILE_STORAGE_PUBLIC_URL", "files.localhost", "configuration: FILE_STORAGE_PUBLIC_URL is not valid: expected an origin such as https://example.com"},
	}
	for _, c := range cases {
		t.Run(c.variable+"="+c.value, func(t *testing.T) {
			_, err := config.Load(config.ServeAPI, with(validEnvironment(), c.variable, c.value))

			require.EqualError(t, err, c.line)
			assert.NotContains(t, err.Error(), c.value)
		})
	}
}

func TestLoadReportsEveryProblemInOneLine(t *testing.T) {
	environment := without(validEnvironment(), "API_LISTEN_ADDRESS")
	environment["DATABASE_POOL_MAX_SIZE"] = "plenty"

	_, err := config.Load(config.ServeAPI, environment)

	require.EqualError(t, err, "configuration: API_LISTEN_ADDRESS is required; DATABASE_POOL_MAX_SIZE is not valid: expected a whole number")
}

func TestAnInvalidValueOfAVariableTheCommandDoesNotUseStillStopsIt(t *testing.T) {
	_, err := config.Load(config.Migrate, map[string]string{
		"MIGRATION_DATABASE_URL": "postgres://migration:migration-password@localhost:5432/software_factory",
		"DATABASE_POOL_MAX_SIZE": "plenty",
	})

	require.EqualError(t, err, "configuration: DATABASE_POOL_MAX_SIZE is not valid: expected a whole number")
}

func TestALoadedConfigurationNeverPrintsAPassword(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, validEnvironment())
	require.NoError(t, err)

	for _, printed := range []string{fmt.Sprint(loaded), fmt.Sprintf("%+v", loaded), fmt.Sprintf("%#v", loaded.DatabaseURL.String())} {
		assert.NotContains(t, printed, "api-password")
		assert.NotContains(t, printed, "migration-password")
	}
}

func TestTheOriginOfTheAppIsReadWithoutATrailingSlash(t *testing.T) {
	for _, value := range []string{"https://factory.example.com", "https://factory.example.com/", "http://localhost:5173"} {
		loaded, err := config.Load(config.IssueSetupCode, with(validEnvironment(), "APP_ORIGIN", value))

		require.NoError(t, err)
		assert.Equal(t, strings.TrimSuffix(value, "/"), loaded.AppOrigin.String())
	}
}

func TestIssueSetupCodeNeedsNeitherTheListenAddressNorTheMigrationRole(t *testing.T) {
	_, err := config.Load(config.IssueSetupCode, map[string]string{
		"DATABASE_URL": "postgres://api:api-password@localhost:5432/software_factory",
		"APP_ORIGIN":   "https://factory.example.com",
	})

	require.NoError(t, err)
}

// The configuration is the one package that reads the environment: the
// command that starts and the support of the tests ask it.
func TestTheEnvironmentIsTheVariablesOfTheProcess(t *testing.T) {
	t.Setenv("CONFIG_TEST_VARIABLE", "a value=with an equals sign")

	read := config.Environment()

	assert.Equal(t, "a value=with an equals sign", read["CONFIG_TEST_VARIABLE"])
}
