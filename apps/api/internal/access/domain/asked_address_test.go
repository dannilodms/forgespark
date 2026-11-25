package domain_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
)

func emailOf(t *testing.T, typed string) kernel.Email {
	t.Helper()
	email, err := kernel.NewEmail(typed)
	require.NoError(t, err)
	return email
}

// The limits count an address by an HMAC-SHA256 of it under a key derived
// from the key of the login code, so their table holds no address. The values
// were computed apart, with Python's hmac: the HMAC of the address under the
// HMAC of the label "the addresses the limits of the login code count" under
// the key.
func TestAnAddressIsCountedByItsHMACUnderAKeyDerivedFromTheKeyOfTheLoginCode(t *testing.T) {
	cases := map[string]struct {
		keyStart byte
		expected string
	}{
		"the bytes 0 to 31": {keyStart: 0, expected: "f06a44585fc04c7c7febdc07bc726e70cf62ae20e19495513f71babf717df84b"},
		"the bytes 1 to 32": {keyStart: 1, expected: "057ecf588430bd425302619633913228fd1925bd150b9ce51e567013d1495e58"},
	}
	for name, example := range cases {
		t.Run(name, func(t *testing.T) {
			asked := keyFrom(t, example.keyStart).AskedAddress(emailOf(t, "ana.souza@fabrica.example"))

			assert.Equal(t, example.expected, hex.EncodeToString(asked.Digest()))
		})
	}
}

// An address is counted the same however it was typed, since the e-mail is
// read in lowercase, and apart from any other address.
func TestAnAddressIsCountedTheSameHoweverItWasTyped(t *testing.T) {
	key := keyFrom(t, 0)

	typed := key.AskedAddress(emailOf(t, "  Ana.Souza@Fabrica.EXAMPLE "))
	another := key.AskedAddress(emailOf(t, "bruno.lima@fabrica.example"))

	assert.Equal(t, key.AskedAddress(emailOf(t, "ana.souza@fabrica.example")), typed)
	assert.NotEqual(t, typed, another)
}

// The key of the login code is a secret: it prints as a placeholder in every
// format, a log line included.
func TestTheKeyOfTheLoginCodePrintsAsAPlaceholder(t *testing.T) {
	key := keyFrom(t, 0xa0)
	encoded, err := json.Marshal(map[string]any{"key": key})
	require.NoError(t, err)
	var line strings.Builder
	slog.New(slog.NewJSONHandler(&line, nil)).Info("built", slog.Any("key", key))

	for _, printed := range []string{fmt.Sprint(key), fmt.Sprintf("%v %+v %#v %s %x", key, key, key, key, key), string(encoded), line.String()} {
		assert.NotContains(t, printed, "a0a1a2")
		assert.NotContains(t, printed, "160 161 162")
		assert.Contains(t, printed, "[secret]")
	}
}
