package oidc_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/oidc"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/oidc/oidctest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// These tests log in through the Dex of development, the local login
// provider, at LOCAL_LOGIN_PROVIDER_ISSUER_URL, with its static client, and
// drive its login page over HTTP as a browser does. Dex sends the browser back
// to the callback of the API at APP_ORIGIN, which is the one address its
// client accepts; the tests read the code there and go no further.

func required(t *testing.T, variable string) string {
	t.Helper()
	value := os.Getenv(variable)
	if value == "" {
		t.Fatalf("%s is not set: run the tests with `task api:test`, which reads .env.example", variable)
	}
	return value
}

// callbackOf is the address Dex sends the browser back to.
func callbackOf(t *testing.T) string {
	t.Helper()
	return required(t, "APP_ORIGIN") + "/api/session/login_providers/" + required(t, "LOCAL_LOGIN_PROVIDER_KEY") + "/callback"
}

// settingsOfDex are the settings of the local login provider of
// development, the failures of the provider written to the logs.
func settingsOfDex(t *testing.T, logs *bytes.Buffer) oidc.Settings {
	t.Helper()
	return oidc.Settings{
		Key:            required(t, "LOCAL_LOGIN_PROVIDER_KEY"),
		DisplayName:    "Local login provider",
		IssuerURL:      required(t, "LOCAL_LOGIN_PROVIDER_ISSUER_URL"),
		ClientID:       required(t, "LOCAL_LOGIN_PROVIDER_CLIENT_ID"),
		ClientSecret:   config.NewSecret(required(t, "LOCAL_LOGIN_PROVIDER_CLIENT_SECRET")),
		RedirectURL:    callbackOf(t),
		TracerProvider: nil,
		Logger:         slog.New(slog.NewJSONHandler(logs, nil)),
	}
}

func providerOf(t *testing.T, settings oidc.Settings) *oidc.LoginProvider {
	t.Helper()
	provider, err := oidc.NewLoginProvider(settings)
	require.NoError(t, err)
	return provider
}

func started(t *testing.T, provider *oidc.LoginProvider) domain.ProviderLogin {
	t.Helper()
	login, err := domain.StartProviderLogin(provider.Key(), domain.RootPath, rand.Reader)
	require.NoError(t, err)
	return login
}

func aBrowser(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &http.Client{Jar: jar, Timeout: 10 * time.Second}
}

// loggedInAtDex sends a browser to the provider for the login, logs the
// person in at Dex, and returns the code and the state Dex sent the browser
// back with.
func loggedInAtDex(t *testing.T, provider *oidc.LoginProvider, login domain.ProviderLogin, person oidctest.Person) (code, state string) {
	t.Helper()
	address, err := provider.AuthorizationURL(context.Background(), login)
	require.NoError(t, err)
	back := oidctest.LogIn(t, aBrowser(t), address, person, callbackOf(t))
	return back.Query().Get("code"), back.Query().Get("state")
}

// withVerifier is the login with another verifier, as a browser that did not
// start it would have.
func withVerifier(t *testing.T, login domain.ProviderLogin, verifier, nonce string) domain.ProviderLogin {
	t.Helper()
	other, err := domain.RestoreProviderLogin(domain.StoredProviderLogin{
		Provider: login.Provider().String(),
		State:    login.State(),
		Nonce:    nonce,
		Verifier: verifier,
		ReturnTo: login.ReturnTo().String(),
	})
	require.NoError(t, err)
	return other
}

// The browser is sent to the authorization endpoint the issuer names, with
// the client, the callback, the scopes of an e-mail, the state, the nonce and
// the S256 challenge of the verifier, never the verifier.
func TestTheBrowserIsSentToTheProviderWithTheStateTheNonceAndTheChallengeOfTheVerifier(t *testing.T) {
	t.Parallel()
	provider := providerOf(t, settingsOfDex(t, &bytes.Buffer{}))
	login := started(t, provider)

	address, err := provider.AuthorizationURL(context.Background(), login)

	require.NoError(t, err)
	sent, err := url.Parse(address)
	require.NoError(t, err)
	assert.Equal(t, required(t, "LOCAL_LOGIN_PROVIDER_ISSUER_URL")+"/auth", sent.Scheme+"://"+sent.Host+sent.Path)
	query := sent.Query()
	challenge := sha256.Sum256([]byte(login.Verifier()))
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(challenge[:]), query.Get("code_challenge"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
	assert.Equal(t, login.State(), query.Get("state"))
	assert.Equal(t, login.Nonce(), query.Get("nonce"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, required(t, "LOCAL_LOGIN_PROVIDER_CLIENT_ID"), query.Get("client_id"))
	assert.Equal(t, callbackOf(t), query.Get("redirect_uri"))
	assert.ElementsMatch(t, []string{"openid", "email"}, strings.Fields(query.Get("scope")))
	assert.NotContains(t, address, login.Verifier())
}

func TestALoginAtDexGivesTheAccountAndItsVerifiedEmail(t *testing.T) {
	t.Parallel()
	provider := providerOf(t, settingsOfDex(t, &bytes.Buffer{}))
	login := started(t, provider)
	code, state := loggedInAtDex(t, provider, login, oidctest.Administrator)

	identity, err := provider.Identify(context.Background(), code, login)

	require.NoError(t, err)
	assert.Equal(t, login.State(), state, "Dex sent the browser back with another state")
	assert.NotEmpty(t, identity.Subject)
	assert.Equal(t, "seed+admin@fabrica.example", identity.VerifiedEmail)
}

// The subject is the account: the same each time the person logs in, and
// another for another person.
func TestEachLoginOfAPersonGivesTheSubjectOfTheirAccount(t *testing.T) {
	t.Parallel()
	provider := providerOf(t, settingsOfDex(t, &bytes.Buffer{}))
	subjectOf := func(person oidctest.Person) string {
		login := started(t, provider)
		code, _ := loggedInAtDex(t, provider, login, person)
		identity, err := provider.Identify(context.Background(), code, login)
		require.NoError(t, err)
		return identity.Subject
	}

	first, second, developer := subjectOf(oidctest.Administrator), subjectOf(oidctest.Administrator), subjectOf(oidctest.Developer)

	assert.Equal(t, first, second)
	assert.NotEqual(t, first, developer)
}

// A code read by someone else is no use without the verifier, which never
// left the system.
func TestACodeExchangedWithAnotherVerifierIsRefused(t *testing.T) {
	t.Parallel()
	logs := &bytes.Buffer{}
	provider := providerOf(t, settingsOfDex(t, logs))
	login := started(t, provider)
	code, _ := loggedInAtDex(t, provider, login, oidctest.Administrator)
	other := started(t, provider)

	_, err := provider.Identify(context.Background(), code, withVerifier(t, login, other.Verifier(), login.Nonce()))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_grant")
	assert.Contains(t, logs.String(), "a login through a provider failed")
	assert.NotContains(t, logs.String(), code)
	assert.NotContains(t, logs.String(), other.Verifier())
}

// An ID token is accepted only with the nonce of the login that asked for
// it: a token of another login, replayed, logs nobody in.
func TestAnIDTokenWithAnotherNonceIsRefused(t *testing.T) {
	t.Parallel()
	provider := providerOf(t, settingsOfDex(t, &bytes.Buffer{}))
	login := started(t, provider)
	code, _ := loggedInAtDex(t, provider, login, oidctest.Administrator)
	other := started(t, provider)

	_, err := provider.Identify(context.Background(), code, withVerifier(t, login, login.Verifier(), other.Nonce()))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce")
}

func TestACodeIsExchangedOnce(t *testing.T) {
	t.Parallel()
	provider := providerOf(t, settingsOfDex(t, &bytes.Buffer{}))
	login := started(t, provider)
	code, _ := loggedInAtDex(t, provider, login, oidctest.Administrator)
	_, err := provider.Identify(context.Background(), code, login)
	require.NoError(t, err)

	_, again := provider.Identify(context.Background(), code, login)

	assert.Error(t, again)
}

// A secret the provider does not know is told by the answer of the provider,
// and never written to the logs.
func TestAWrongSecretOfTheClientIsRefusedAndNotLogged(t *testing.T) {
	t.Parallel()
	logs := &bytes.Buffer{}
	settings := settingsOfDex(t, logs)
	settings.ClientSecret = config.NewSecret("a-secret-dex-does-not-know")
	provider := providerOf(t, settings)
	login := started(t, provider)
	code, _ := loggedInAtDex(t, provider, login, oidctest.Administrator)

	_, err := provider.Identify(context.Background(), code, login)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, logs.String(), `"login_provider":"local"`)
	assert.NotContains(t, logs.String(), "a-secret-dex-does-not-know")
	assert.NotContains(t, logs.String(), code)
}

// A provider that does not answer starts no login, and the failure is
// logged, naming the provider. It is discovered again on the next login.
func TestAProviderThatCannotBeReachedFailsAndIsLogged(t *testing.T) {
	t.Parallel()
	logs := &bytes.Buffer{}
	settings := settingsOfDex(t, logs)
	// Nothing answers on port 1 of the machine.
	settings.IssuerURL = "http://127.0.0.1:1"
	provider := providerOf(t, settings)

	_, err := provider.AuthorizationURL(context.Background(), started(t, provider))
	_, again := provider.AuthorizationURL(context.Background(), started(t, provider))

	require.Error(t, err)
	require.Error(t, again)
	assert.Equal(t, 2, strings.Count(logs.String(), "a login through a provider failed"))
	assert.Contains(t, logs.String(), `"login_provider":"local"`)
}

func TestAProviderWithAKeyThatIsNotOneIsRefused(t *testing.T) {
	t.Parallel()
	settings := settingsOfDex(t, &bytes.Buffer{})
	settings.Key = "Local Provider"

	_, err := oidc.NewLoginProvider(settings)

	assert.Error(t, err)
}

// standInIssuer is an OpenID Connect provider of the test, which answers
// every code with an ID token of these claims, signed with a key of its own:
// Dex says every e-mail it knows is verified, so a token that does not say so
// comes from here.
func standInIssuer(t *testing.T, claims map[string]any) *httptest.Server {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "the-key"}}, nil)
	require.NoError(t, err)
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	mux.HandleFunc("GET /.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{
			"issuer":                                server.URL,
			"authorization_endpoint":                server.URL + "/auth",
			"token_endpoint":                        server.URL + "/token",
			"jwks_uri":                              server.URL + "/keys",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("GET /keys", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "the-key", Algorithm: "RS256", Use: "sig"}}})
	})
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, _ *http.Request) {
		payload := map[string]any{"iss": server.URL, "aud": "the-client", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
		for name, value := range claims {
			payload[name] = value
		}
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		signed, err := signer.Sign(body)
		require.NoError(t, err)
		token, err := signed.CompactSerialize()
		require.NoError(t, err)
		writeJSON(t, w, map[string]any{"access_token": "an-access-token", "token_type": "Bearer", "id_token": token})
	})
	return server
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(body))
}

func standInProvider(t *testing.T, issuer *httptest.Server) *oidc.LoginProvider {
	t.Helper()
	return providerOf(t, oidc.Settings{
		Key:            "stand_in",
		DisplayName:    "A stand-in",
		IssuerURL:      issuer.URL,
		ClientID:       "the-client",
		ClientSecret:   config.NewSecret("the-secret"),
		RedirectURL:    "http://localhost/api/session/login_providers/stand_in/callback",
		TracerProvider: nil,
		Logger:         nil,
	})
}

func TestTheEmailOfATokenThatDoesNotSayItIsVerifiedIsNotGiven(t *testing.T) {
	t.Parallel()
	cases := map[string]map[string]any{
		"not verified":        {"email": "ana.souza@fabrica.example", "email_verified": false},
		"no word on it":       {"email": "ana.souza@fabrica.example"},
		"verified, no e-mail": {"email_verified": true},
	}
	for name, claims := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			issuer := standInIssuer(t, claims)
			provider := standInProvider(t, issuer)
			login := started(t, provider)
			claims["sub"] = "the-account"
			claims["nonce"] = login.Nonce()

			identity, err := provider.Identify(context.Background(), "a-code", login)

			require.NoError(t, err)
			assert.Equal(t, "the-account", identity.Subject)
			assert.Empty(t, identity.VerifiedEmail)
		})
	}
}

func TestTheEmailOfATokenThatSaysItIsVerifiedIsGiven(t *testing.T) {
	t.Parallel()
	claims := map[string]any{"sub": "the-account", "email": "Ana.Souza@Fabrica.example", "email_verified": true}
	issuer := standInIssuer(t, claims)
	provider := standInProvider(t, issuer)
	login := started(t, provider)
	claims["nonce"] = login.Nonce()

	identity, err := provider.Identify(context.Background(), "a-code", login)

	require.NoError(t, err)
	assert.Equal(t, "Ana.Souza@Fabrica.example", identity.VerifiedEmail)
}

// A token signed for another client, or by another issuer, is not valid for
// this login.
func TestATokenForAnotherClientOrFromAnotherIssuerIsRefused(t *testing.T) {
	t.Parallel()
	for name, claim := range map[string]map[string]any{
		"another client": {"aud": "another-client"},
		"another issuer": {"iss": "https://elsewhere.example"},
		"expired":        {"exp": time.Now().Add(-time.Hour).Unix()},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			claims := map[string]any{"sub": "the-account", "email": "ana.souza@fabrica.example", "email_verified": true}
			issuer := standInIssuer(t, claims)
			provider := standInProvider(t, issuer)
			login := started(t, provider)
			claims["nonce"] = login.Nonce()
			for claimName, value := range claim {
				claims[claimName] = value
			}

			_, err := provider.Identify(context.Background(), "a-code", login)

			assert.Error(t, err)
		})
	}
}
