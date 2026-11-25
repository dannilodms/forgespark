package config_test

import (
	"fmt"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
)

// withLocalLoginProvider is the environment of the API with the local login
// provider turned on.
func withLocalLoginProvider() map[string]string {
	environment := maps.Clone(validEnvironment())
	environment["LOCAL_LOGIN_PROVIDER_KEY"] = "local"
	environment["LOCAL_LOGIN_PROVIDER_DISPLAY_NAME"] = "Local login provider"
	environment["LOCAL_LOGIN_PROVIDER_ISSUER_URL"] = "http://localhost:5556"
	environment["LOCAL_LOGIN_PROVIDER_CLIENT_ID"] = "forgespark"
	environment["LOCAL_LOGIN_PROVIDER_CLIENT_SECRET"] = "a-secret-of-the-client"
	return environment
}

var localLoginProviderVariables = []string{
	"LOCAL_LOGIN_PROVIDER_KEY", "LOCAL_LOGIN_PROVIDER_DISPLAY_NAME", "LOCAL_LOGIN_PROVIDER_ISSUER_URL",
	"LOCAL_LOGIN_PROVIDER_CLIENT_ID", "LOCAL_LOGIN_PROVIDER_CLIENT_SECRET",
}

func TestTheAPIReadsTheLocalLoginProviderWhenItIsOn(t *testing.T) {
	loaded, err := config.Load(config.ServeAPI, withLocalLoginProvider())

	require.NoError(t, err)
	assert.True(t, loaded.LocalLoginProviderIsOn())
	assert.Equal(t, "local", loaded.LocalLoginProviderKey.String())
	assert.Equal(t, "Local login provider", loaded.LocalLoginProviderDisplayName)
	assert.Equal(t, "http://localhost:5556", loaded.LocalLoginProviderIssuerURL.String())
	assert.Equal(t, "forgespark", loaded.LocalLoginProviderClientID)
	assert.Equal(t, "a-secret-of-the-client", loaded.LocalLoginProviderClientSecret.Reveal())
	assert.NotContains(t, fmt.Sprintf("%v %+v %#v", loaded, loaded, loaded), "a-secret-of-the-client")
}

func TestTheLocalLoginProviderIsOffWhenItsVariablesAreEmpty(t *testing.T) {
	environment := withLocalLoginProvider()
	for _, variable := range localLoginProviderVariables {
		environment[variable] = ""
	}

	loaded, err := config.Load(config.ServeAPI, environment)

	require.NoError(t, err)
	assert.False(t, loaded.LocalLoginProviderIsOn())
}

// Turning the provider on takes the five variables: one left out stops the
// API at its start, named.
func TestTheLocalLoginProviderOnNeedsEachOfItsVariables(t *testing.T) {
	for _, variable := range localLoginProviderVariables {
		t.Run(variable, func(t *testing.T) {
			_, err := config.Load(config.ServeAPI, without(withLocalLoginProvider(), variable))

			require.Error(t, err)
			assert.Contains(t, err.Error(), variable+" is required")
		})
	}
}

func TestTheKeyOfTheLocalLoginProviderIsAKeyInSnakeCase(t *testing.T) {
	for _, refused := range []string{"Local", "local-dex", "local__dex", "login_code", "the local provider"} {
		_, err := config.Load(config.ServeAPI, with(withLocalLoginProvider(), "LOCAL_LOGIN_PROVIDER_KEY", refused))

		require.Error(t, err, refused)
		assert.Contains(t, err.Error(), "LOCAL_LOGIN_PROVIDER_KEY is not valid", refused)
		assert.NotContains(t, err.Error(), refused)
	}
}

func TestTheIssuerOfTheLocalLoginProviderIsAnHTTPURL(t *testing.T) {
	_, err := config.Load(config.ServeAPI, with(withLocalLoginProvider(), "LOCAL_LOGIN_PROVIDER_ISSUER_URL", "localhost:5556"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "LOCAL_LOGIN_PROVIDER_ISSUER_URL is not valid")
}

// Only the API logs in through a provider: the other commands need none of
// its variables, and run with them half set.
func TestOnlyTheAPINeedsTheLocalLoginProvider(t *testing.T) {
	halfSet := with(workerEnvironment(), "LOCAL_LOGIN_PROVIDER_KEY", "local")

	_, err := config.Load(config.Worker, halfSet)

	require.NoError(t, err)
}
