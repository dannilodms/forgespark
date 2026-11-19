package kernel_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// The rule of an e-mail address, which the e-mail of a user, the recipient of
// an e-mail and the addresses of the configuration share: one address and
// nothing around it, no name and no angle brackets, 254 characters at most.
func TestAnEmailAddressIsOneAddressAndNothingAroundIt(t *testing.T) {
	for _, text := range []string{"ana.souza@fabrica.example", "Seed+Admin@Fabrica.example", strings.Repeat("a", 238) + "@fabrica.example"} {
		assert.True(t, kernel.IsEmailAddress(text), text)
	}
	for _, text := range []string{
		"", "ana.souza", "@fabrica.example", " ana.souza@fabrica.example", "Ana <ana.souza@fabrica.example>",
		"ana@fabrica.example, bruno@fabrica.example", strings.Repeat("a", 239) + "@fabrica.example",
	} {
		assert.False(t, kernel.IsEmailAddress(text), text)
	}
}

// The e-mail of a person is read with the spaces around it dropped, and in
// lowercase, so the way it was typed does not matter where it is compared.
func TestAnEmailIsReadInLowercaseWithoutTheSpacesAroundIt(t *testing.T) {
	email, err := kernel.NewEmail("  Ana.Souza@Fabrica.EXAMPLE ")

	require.NoError(t, err)
	assert.Equal(t, "ana.souza@fabrica.example", email.String())
	_, empty := kernel.NewEmail("  ")
	_, long := kernel.NewEmail(strings.Repeat("a", 239) + "@fabrica.example")
	_, invalid := kernel.NewEmail("Ana <ana.souza@fabrica.example>")
	assert.ErrorIs(t, empty, kernel.RequiredError{})
	assert.ErrorIs(t, long, kernel.TooLongError{MaxLength: kernel.EmailMaxLength})
	assert.ErrorIs(t, invalid, kernel.EmailInvalidError{})
}
