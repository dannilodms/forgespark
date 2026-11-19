// Package oidc logs a person in through a login provider that issues an ID
// token, by OpenID Connect (Login in CODING_STANDARDS.md, ADR-0011): the
// discovery of the provider from its issuer, the authorization code flow with
// PKCE, a state and a nonce, and the validation of the ID token with
// `coreos/go-oidc/v3`. The e-mail is taken from the token only when the token
// says it is verified. The local login provider is one, pointed at Dex; a
// provider such as Google is another set of settings.
//
// Nothing the login carries reaches a log or a span: no code, no token, no
// verifier, no secret and no e-mail. A failure is logged by the step and the
// answer of the provider.
package oidc

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	openid "github.com/coreos/go-oidc/v3/oidc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"golang.org/x/oauth2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// timeout bounds each call to the provider, from the connection to the last
// byte of the answer: well inside the drain deadline of the API.
const timeout = 10 * time.Second

// scopes are what the login asks the provider for: the ID token, with the
// e-mail of the account.
var scopes = []string{openid.ScopeOpenID, "email"}

// Settings are a login provider, as configuration gives it.
type Settings struct {
	// Key names the provider in its paths, its links and the audit log.
	Key string
	// DisplayName is what the login screen shows on its button.
	DisplayName string
	// IssuerURL is the issuer of the provider, which its discovery document
	// is read from and its ID tokens must name.
	IssuerURL string
	// ClientID and ClientSecret are the client the system is at the
	// provider.
	ClientID     string
	ClientSecret config.Secret
	// RedirectURL is the callback of the provider in the API, at the origin
	// of the app, which the provider sends the browser back to.
	RedirectURL string
	// TracerProvider traces each call to the provider. With none, nothing is
	// traced.
	TracerProvider trace.TracerProvider
	// Logger writes a failure of the provider. With none, nothing is
	// written.
	Logger *slog.Logger
}

// LoginProvider is one login provider of OpenID Connect. It discovers the
// provider at the first login through it, and again at the next one while the
// provider cannot be reached, so the API starts whether the provider answers
// or not.
type LoginProvider struct {
	key      domain.LoginProviderKey
	settings Settings
	client   *http.Client
	logger   *slog.Logger

	mu         sync.Mutex
	discovered *discovered
}

var _ application.LoginProvider = (*LoginProvider)(nil)

// discovered is what the discovery of the provider gave: its endpoints, and
// the validation of its ID tokens with its keys.
type discovered struct {
	config   oauth2.Config
	verifier *openid.IDTokenVerifier
}

// NewLoginProvider returns the provider of these settings. It calls nothing:
// the provider is discovered at the first login through it.
func NewLoginProvider(settings Settings) (*LoginProvider, error) {
	key, err := domain.NewLoginProviderKey(settings.Key)
	if err != nil {
		return nil, err
	}
	provider := settings.TracerProvider
	if provider == nil {
		provider = noop.NewTracerProvider()
	}
	logger := settings.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	tracer := provider.Tracer("github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/oidc")
	return &LoginProvider{
		key:      key,
		settings: settings,
		client: &http.Client{
			Timeout:   timeout,
			Transport: tracedTransport{base: http.DefaultTransport, tracer: tracer, key: key},
		},
		logger:     logger,
		mu:         sync.Mutex{},
		discovered: nil,
	}, nil
}

// Key returns the key of the provider.
func (p *LoginProvider) Key() domain.LoginProviderKey {
	return p.key
}

// DisplayName returns the name the login screen shows.
func (p *LoginProvider) DisplayName() string {
	return p.settings.DisplayName
}

// AuthorizationURL returns the authorization endpoint of the provider with
// the client, the callback, the scopes, the state, the nonce and the S256
// challenge of the verifier.
func (p *LoginProvider) AuthorizationURL(ctx context.Context, login domain.ProviderLogin) (string, error) {
	found, err := p.discover(ctx)
	if err != nil {
		p.tell(ctx, err)
		return "", err
	}
	return found.config.AuthCodeURL(login.State(), openid.Nonce(login.Nonce()), oauth2.S256ChallengeOption(login.Verifier())), nil
}

// Identify exchanges the code with the verifier, validates the ID token the
// provider answers (its signature, its issuer, its client and its expiry)
// and its nonce, and returns its subject, and its e-mail when the token says
// the e-mail is verified.
func (p *LoginProvider) Identify(ctx context.Context, code string, login domain.ProviderLogin) (application.ProviderIdentity, error) {
	identity, err := p.identify(ctx, code, login)
	if err != nil {
		p.tell(ctx, err)
		return application.ProviderIdentity{}, err
	}
	return identity, nil
}

// claims are what the login reads of an ID token besides its subject.
type claims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (p *LoginProvider) identify(ctx context.Context, code string, login domain.ProviderLogin) (application.ProviderIdentity, error) {
	found, err := p.discover(ctx)
	if err != nil {
		return application.ProviderIdentity{}, err
	}
	ctx = openid.ClientContext(ctx, p.client)
	token, err := found.config.Exchange(ctx, code, oauth2.VerifierOption(login.Verifier()))
	if err != nil {
		return application.ProviderIdentity{}, exchangeFailure(err)
	}
	raw, answered := token.Extra("id_token").(string)
	if !answered || raw == "" {
		return application.ProviderIdentity{}, errors.New("the provider exchanged the code and answered no ID token")
	}
	idToken, err := found.verifier.Verify(ctx, raw)
	if err != nil {
		return application.ProviderIdentity{}, fmt.Errorf("the ID token is not valid: %s", brief(err))
	}
	if subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(login.Nonce())) != 1 {
		return application.ProviderIdentity{}, errors.New("the ID token carries the nonce of another login")
	}
	var read claims
	if err := idToken.Claims(&read); err != nil {
		return application.ProviderIdentity{}, errors.New("the claims of the ID token cannot be read")
	}
	identity := application.ProviderIdentity{Subject: idToken.Subject, VerifiedEmail: ""}
	if read.EmailVerified {
		identity.VerifiedEmail = read.Email
	}
	return identity, nil
}

// discover reads the discovery document of the provider once, and again
// while it fails.
func (p *LoginProvider) discover(ctx context.Context) (*discovered, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.discovered != nil {
		return p.discovered, nil
	}
	provider, err := openid.NewProvider(openid.ClientContext(ctx, p.client), p.settings.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover the provider at its issuer: %s", brief(err))
	}
	p.discovered = &discovered{
		config: oauth2.Config{
			ClientID:     p.settings.ClientID,
			ClientSecret: p.settings.ClientSecret.Reveal(),
			Endpoint:     provider.Endpoint(),
			RedirectURL:  p.settings.RedirectURL,
			Scopes:       scopes,
		},
		verifier: provider.Verifier(&openid.Config{ClientID: p.settings.ClientID}),
	}
	return p.discovered, nil
}

// tell writes the failure of a login through the provider: the step and the
// answer of the provider, never what the login carried.
func (p *LoginProvider) tell(ctx context.Context, err error) {
	p.logger.LogAttrs(ctx, slog.LevelWarn, "a login through a provider failed",
		slog.String("login_provider", p.key.String()), slog.String("error", err.Error()))
}

// errorCodeForm is the form of the code of a failure OAuth defines, such as
// `invalid_grant`. Anything else the provider answers is not repeated.
var errorCodeForm = regexp.MustCompile(`^[a-z][a-z_]{0,63}$`)

// exchangeFailure tells why the exchange of a code failed: the status and
// the code of the answer of the provider, or that it could not be reached.
// The description the provider wrote, and the body of its answer, are left
// out.
func exchangeFailure(err error) error {
	var refused *oauth2.RetrieveError
	if errors.As(err, &refused) {
		return fmt.Errorf("the provider refused the code: answered %d %s", statusOf(refused), codeOf(refused.ErrorCode))
	}
	var unreached *url.Error
	if errors.As(err, &unreached) {
		return fmt.Errorf("reach the provider to exchange the code: %s", brief(unreached.Err))
	}
	return errors.New("the provider exchanged the code and its answer cannot be read")
}

func statusOf(refused *oauth2.RetrieveError) int {
	if refused.Response == nil {
		return 0
	}
	return refused.Response.StatusCode
}

func codeOf(code string) string {
	if !errorCodeForm.MatchString(code) {
		return "with no code"
	}
	return code
}

// briefLength is the most of the text of a failure a line carries.
const briefLength = 200

// brief keeps the first line of the text of a failure, and its first 200
// characters: a library may write the whole answer of the provider there.
func brief(err error) string {
	first, _, _ := strings.Cut(err.Error(), "\n")
	if len(first) > briefLength {
		return first[:briefLength] + "..."
	}
	return first
}

// tracedTransport opens a span for each call to the provider: its method,
// the host it went to and the status of the answer, never its path, its
// query, its headers or its body. The trace of the request does not go with
// the call: an outbound call carries it to the agents of the system only
// (Telemetry in CODING_STANDARDS.md).
type tracedTransport struct {
	base   http.RoundTripper
	tracer trace.Tracer
	key    domain.LoginProviderKey
}

func (t tracedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.Start(request.Context(), "Login provider "+request.Method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.ServerAddress(request.URL.Hostname()),
			semconv.HTTPRequestMethodKey.String(request.Method),
			attribute.String("login_provider", t.key.String()),
		))
	defer span.End()

	response, err := t.base.RoundTrip(request.WithContext(ctx))
	if err != nil {
		span.SetStatus(codes.Error, "the provider could not be reached")
		return nil, err
	}
	span.SetAttributes(semconv.HTTPResponseStatusCode(response.StatusCode))
	if response.StatusCode >= http.StatusBadRequest {
		span.SetStatus(codes.Error, "")
	}
	return response, nil
}
