package config_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

func validEnvironment() map[string]string {
	return map[string]string{
		"API_LISTEN_ADDRESS":     "127.0.0.1:8080",
		"DATABASE_URL":           "postgres://api:api-password@localhost:5432/software_factory",
		"MIGRATION_DATABASE_URL": "postgres://migration:migration-password@localhost:5432/software_factory",
	}
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
	environment["DATABASE_POOL_MAX_SIZE"] = "4"
	environment["DATABASE_STATEMENT_TIMEOUT"] = "5s"
	environment["DATABASE_IDLE_IN_TRANSACTION_TIMEOUT"] = "2m"
	environment["HEALTH_CHECK_TIMEOUT"] = "500ms"

	loaded, err := config.Load(config.ServeAPI, environment)

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", loaded.ListenAddress.String())
	assert.Equal(t, "postgres://api:api-password@localhost:5432/software_factory", loaded.DatabaseURL.ConnectionString())
	assert.Equal(t, int32(4), loaded.DatabasePoolMaxSize)
	assert.Equal(t, 5*time.Second, loaded.DatabaseStatementTimeout)
	assert.Equal(t, 2*time.Minute, loaded.DatabaseIdleInTransactionTimeout)
	assert.Equal(t, 500*time.Millisecond, loaded.HealthCheckTimeout)
}

func TestLoadFallsBackToTheDefaults(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, validEnvironment())

	require.NoError(t, err)
	assert.Equal(t, int32(10), loaded.DatabasePoolMaxSize)
	assert.Equal(t, 30*time.Second, loaded.DatabaseStatementTimeout)
	assert.Equal(t, time.Minute, loaded.DatabaseIdleInTransactionTimeout)
	assert.Equal(t, 2*time.Second, loaded.HealthCheckTimeout)
}

func TestLoadNamesARequiredVariableThatIsMissing(t *testing.T) {
	cases := []struct {
		command  config.Command
		variable string
	}{
		{config.ServeAPI, "API_LISTEN_ADDRESS"},
		{config.ServeAPI, "DATABASE_URL"},
		{config.Migrate, "MIGRATION_DATABASE_URL"},
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
