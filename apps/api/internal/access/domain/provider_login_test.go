package domain_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

func TestTheKeyOfALoginProviderIsAKeyInSnakeCase(t *testing.T) {
	for _, accepted := range []string{"local", "google", "microsoft_entra", "github2", strings.Repeat("a", 64)} {
		key, err := domain.NewLoginProviderKey(accepted)

		require.NoError(t, err, accepted)
		assert.Equal(t, accepted, key.String())
		assert.Equal(t, domain.WayIn(accepted), key.WayIn())
	}
}

func TestAKeyThatIsNotInSnakeCaseOrIsTheWayInOfTheLoginCodeIsRefused(t *testing.T) {
	for _, refused := range []string{"", "Local", "local-provider", "_local", "local_", "local__dex", "2local", "lo cal", strings.Repeat("a", 65), "login_code"} {
		_, err := domain.NewLoginProviderKey(refused)

		assert.Error(t, err, refused)
	}
}

func TestAPathOfTheAppIsKeptAsTheReturnPath(t *testing.T) {
	for _, path := range []string{"/", "/projects", "/projects/prj_Xk9fQ2mB", "/projects?tab=files&page=2", "/login"} {
		assert.Equal(t, path, domain.NewReturnPath(path).String(), path)
	}
}

// A path to return to comes from the address the browser was given: a value
// that would send the browser to another site, or to the API, returns it to
// the root of the app.
func TestAnythingButAPathOfTheAppReturnsToTheRoot(t *testing.T) {
	for _, refused := range []string{
		"", "projects", "https://elsewhere.example/", "//elsewhere.example/projects", "/\\elsewhere.example",
		"/projects\\x", "/api/session", "/api", "/projects\r\nLocation: x", "/pro jects", "javascript:alert(1)",
		"/" + strings.Repeat("a", 2048),
	} {
		assert.Equal(t, "/", domain.NewReturnPath(refused).String(), "%q", refused)
	}
	assert.Equal(t, domain.RootPath, domain.NewReturnPath("https://elsewhere.example"))
}

// The state, the nonce and the verifier are each 32 bytes of the random
// source, in base64url without padding: 43 characters, as RFC 7636 asks of a
// verifier.
func TestALoginThroughAProviderDrawsItsStateNonceAndVerifierFromTheRandomSource(t *testing.T) {
	local := keyOf(t, "local")

	login, err := domain.StartProviderLogin(local, domain.NewReturnPath("/projects"), bytes.NewReader(bytesFrom(0, 96)))

	require.NoError(t, err)
	assert.Equal(t, local, login.Provider())
	assert.Equal(t, "/projects", login.ReturnTo().String())
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(bytesFrom(0, 32)), login.State())
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(bytesFrom(32, 32)), login.Nonce())
	assert.Len(t, login.Verifier(), 43)
	assert.NotEqual(t, login.State(), login.Nonce())
}

func TestALoginThroughAProviderFailsWhenTheRandomSourceFails(t *testing.T) {
	_, err := domain.StartProviderLogin(keyOf(t, "local"), domain.RootPath, bytes.NewReader(make([]byte, 40)))

	assert.Error(t, err)
}

// The callback is accepted only for the provider the login started with, and
// with the state it was given.
func TestALoginThroughAProviderAcceptsOnlyItsProviderAndItsState(t *testing.T) {
	login, err := domain.StartProviderLogin(keyOf(t, "local"), domain.RootPath, bytes.NewReader(bytesFrom(0, 96)))
	require.NoError(t, err)

	assert.True(t, login.Accepts(keyOf(t, "local"), login.State()))
	assert.False(t, login.Accepts(keyOf(t, "google"), login.State()))
	assert.False(t, login.Accepts(keyOf(t, "local"), login.State()[:42]))
	assert.False(t, login.Accepts(keyOf(t, "local"), ""))
	assert.False(t, login.Accepts(keyOf(t, "local"), login.Nonce()))
}

func TestALoginThroughAProviderIsRestoredAsItWasKept(t *testing.T) {
	started, err := domain.StartProviderLogin(keyOf(t, "local"), domain.NewReturnPath("/projects"), bytes.NewReader(bytesFrom(7, 96)))
	require.NoError(t, err)

	restored, err := domain.RestoreProviderLogin(domain.StoredProviderLogin{
		Provider: "local",
		State:    started.State(),
		Nonce:    started.Nonce(),
		Verifier: started.Verifier(),
		ReturnTo: "/projects",
	})

	require.NoError(t, err)
	assert.Equal(t, started, restored)
}

func TestALoginThroughAProviderKeptForAnotherKeyIsNotRestored(t *testing.T) {
	_, err := domain.RestoreProviderLogin(domain.StoredProviderLogin{Provider: "Not A Key", State: "s", Nonce: "n", Verifier: "v"})

	assert.Error(t, err)
}

func TestTheSubjectOfAnAccountAtAProviderIsOneTo255Characters(t *testing.T) {
	subject, err := domain.NewProviderSubject("CiQ1ZDBmNmExZS0zYzJiLTRmN2EtOWUxZC0wYTFiMmMzZDRlMDESBWxvY2Fs")
	require.NoError(t, err)
	assert.Equal(t, "CiQ1ZDBmNmExZS0zYzJiLTRmN2EtOWUxZC0wYTFiMmMzZDRlMDESBWxvY2Fs", subject.String())

	_, err = domain.NewProviderSubject(strings.Repeat("s", 255))
	require.NoError(t, err)

	for _, refused := range []string{"", strings.Repeat("s", 256)} {
		_, err := domain.NewProviderSubject(refused)
		assert.Error(t, err)
	}
}

func TestAProviderLinkHoldsTheAccountAndTheUserItWasLinkedTo(t *testing.T) {
	subject, err := domain.NewProviderSubject("the-account")
	require.NoError(t, err)

	link := domain.LinkProvider(keyOf(t, "local"), subject, 7, noon)

	assert.Equal(t, keyOf(t, "local"), link.Provider())
	assert.Equal(t, subject, link.Subject())
	assert.Equal(t, domain.UserID(7), link.User())
	assert.Equal(t, noon, link.LinkedAt())

	restored, err := domain.RestoreProviderLink(domain.StoredProviderLink{Provider: "local", Subject: "the-account", User: 7, LinkedAt: noon})
	require.NoError(t, err)
	assert.Equal(t, link, restored)
}

func TestAProviderLinkStoredWithAnUnreadableKeyOrSubjectIsNotRestored(t *testing.T) {
	for _, stored := range []domain.StoredProviderLink{
		{Provider: "Not A Key", Subject: "the-account", User: 7, LinkedAt: noon},
		{Provider: "local", Subject: "", User: 7, LinkedAt: noon},
		{Provider: "local", Subject: "the-account", User: 0, LinkedAt: noon},
	} {
		_, err := domain.RestoreProviderLink(stored)

		assert.Error(t, err, "%+v", stored)
	}
}

func keyOf(t *testing.T, key string) domain.LoginProviderKey {
	t.Helper()
	held, err := domain.NewLoginProviderKey(key)
	require.NoError(t, err)
	return held
}

func bytesFrom(first byte, size int) []byte {
	drawn := make([]byte, size)
	for i := range drawn {
		drawn[i] = first + byte(i)
	}
	return drawn
}
