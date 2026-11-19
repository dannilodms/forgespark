package kernel_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

func TestAKeyIsWordsOfLowercaseLettersAndDigitsJoinedByOneUnderscore(t *testing.T) {
	for _, text := range []string{"local", "google", "login_code", "invalid_grant", "azure_ad2", strings.Repeat("a", kernel.KeyMaxLength)} {
		_, err := kernel.NewKey(text)
		assert.NoError(t, err, text)
	}
	for _, text := range []string{"", "Local", "2fa", "_local", "local_", "local__ad", "local-ad", "local ad", "ana@fabrica.example", strings.Repeat("a", kernel.KeyMaxLength+1)} {
		_, err := kernel.NewKey(text)
		assert.Error(t, err, text)
	}
}

// The way in of a login is the login code, or the key of the login provider
// it went through, which is never the way in of the login code.
func TestTheWayInOfALoginIsTheLoginCodeOrTheKeyOfAProvider(t *testing.T) {
	code, err := kernel.NewWayIn("login_code")
	require.NoError(t, err)
	assert.Equal(t, kernel.LoginCodeWayIn, code)
	local, err := kernel.NewProviderWayIn("local")
	require.NoError(t, err)
	assert.Equal(t, kernel.WayIn("local"), local)

	_, notAKey := kernel.NewWayIn("Local Provider")
	_, theLoginCode := kernel.NewProviderWayIn("login_code")

	assert.Error(t, notAKey)
	assert.Error(t, theLoginCode, "a provider took the way in of the login code")
}
