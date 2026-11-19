package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	accessoidc "github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/oidc"
	accessrest "github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/rest"
	accessapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/access/application"
	accessdomain "github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	filess3 "github.com/dannilodms/software-factory-manager/apps/api/internal/files/adapters/s3"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/health"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/jobs"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/lifecycle"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/restapi"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/session"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/telemetry"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// errShuttingDown is why readiness fails from the signal on.
var errShuttingDown = errors.New("the API is shutting down")

// fileStorageOf is the file storage the configuration names, whose calls are
// traced. The API and the worker reach it the same way.
func fileStorageOf(settings config.Config, provider trace.TracerProvider) *filess3.Storage {
	return filess3.NewStorage(filess3.Settings{
		URL:             settings.FileStorageURL.String(),
		PublicURL:       settings.FileStoragePublicURL.String(),
		Region:          settings.FileStorageRegion,
		Bucket:          settings.FileStorageBucket,
		AccessKeyID:     settings.FileStorageAccessKeyID,
		SecretAccessKey: settings.FileStorageSecretAccessKey,
		TracerProvider:  provider,
	})
}

// loginCodeKeyOf is the key of the login code the configuration holds, in the
// type of the access area, which prints as a placeholder as the type of the
// configuration does: its bytes pass from one secret type to the other, and
// only the HMACs of the access area read them (Configuration in
// CODING_STANDARDS.md).
func loginCodeKeyOf(settings config.Config) (accessdomain.LoginCodeKey, error) {
	key, err := accessdomain.NewLoginCodeKey(settings.LoginCodeKey.Reveal())
	if err != nil {
		return accessdomain.LoginCodeKey{}, fmt.Errorf("LOGIN_CODE_KEY: %w", err)
	}
	return key, nil
}

// loginProvidersOf are the login providers the configuration turns on: the
// local one, when its variables are set, which a line of the log names (Login
// in CODING_STANDARDS.md). Its callback is at the origin of the app.
func loginProvidersOf(settings config.Config, provider trace.TracerProvider, logger *slog.Logger) ([]accessapplication.LoginProvider, error) {
	if !settings.LocalLoginProviderIsOn() {
		return nil, nil
	}
	key := settings.LocalLoginProviderKey.String()
	local, err := accessoidc.NewLoginProvider(accessoidc.Settings{
		Key:            key,
		DisplayName:    settings.LocalLoginProviderDisplayName,
		IssuerURL:      settings.LocalLoginProviderIssuerURL.String(),
		ClientID:       settings.LocalLoginProviderClientID,
		ClientSecret:   settings.LocalLoginProviderClientSecret,
		RedirectURL:    settings.AppOrigin.String() + accessrest.CallbackPath(key),
		TracerProvider: provider,
		Logger:         logger,
	})
	if err != nil {
		return nil, fmt.Errorf("the local login provider: %w", err)
	}
	logger.Info("the local login provider is on",
		slog.String("login_provider", key),
		slog.String("display_name", settings.LocalLoginProviderDisplayName),
		slog.String("issuer", settings.LocalLoginProviderIssuerURL.String()))
	return []accessapplication.LoginProvider{local}, nil
}

// serveAPI serves the API until the shutdown notice, then stops without
// cutting a request short. Its logs are JSON lines on the error output: the
// start and the stop of the process, and what Telemetry in
// CODING_STANDARDS.md lists. A configuration that is not valid is the one
// plain line, written before there is a logger.
func serveAPI(shutdown context.Context, environment map[string]string, _, stderr io.Writer) error {
	settings, err := config.Load(config.ServeAPI, environment)
	if err != nil {
		return err
	}

	logger := telemetry.NewLogger(stderr, settings.LogLevel)
	telemetry.RouteTheStandardLogger(logger)
	if err := serve(shutdown, settings, logger); err != nil {
		logger.Error("the API stopped", slog.String("error", err.Error()))
		return errLogged
	}
	logger.Info("the API stopped")
	return nil
}

// serve wires the API and serves it. It never changes the schema: readiness
// tells when a migration is pending.
//
// When it returns, the pool is closed and then the last spans are sent, in
// that order.
func serve(shutdown context.Context, settings config.Config, logger *slog.Logger) error {
	tracing, pool, stop, err := tracedDatabase(settings, logger)
	if err != nil {
		return err
	}
	defer stop()
	saver, err := jobs.NewSaver(pool, tracing.Provider, logger)
	if err != nil {
		return err
	}
	sessions := session.New(pool, session.Settings{
		Secure:    settings.AppOrigin.IsHTTPS(),
		AppOrigin: settings.AppOrigin,
	})

	mux := http.NewServeMux()
	ready := func(ctx context.Context) error {
		if shutdown.Err() != nil {
			return errShuttingDown
		}
		return database.CheckSchema(ctx, pool, migrations.Files)
	}
	health.Mount(mux, ready, settings.HealthCheckTimeout)
	providers, err := loginProvidersOf(settings, tracing.Provider, logger)
	if err != nil {
		return err
	}
	loginCodeKey, err := loginCodeKeyOf(settings)
	if err != nil {
		return err
	}
	_, err = mountREST(mux, served{
		pool:           pool,
		storage:        fileStorageOf(settings, tracing.Provider),
		saver:          saver,
		sessions:       sessions,
		loginProviders: providers,
		loginCodeKey:   loginCodeKey,
	})
	if err != nil {
		return err
	}
	mountBounceWebhook(mux, pool, settings)

	listener, err := net.Listen("tcp", settings.ListenAddress.String())
	if err != nil {
		return fmt.Errorf("listen on API_LISTEN_ADDRESS: %w", err)
	}
	logger.Info("the API started", slog.String("address", listener.Addr().String()))

	// Every request under the API carries the session of its browser, and
	// one that would change state with its cookie from another site is
	// refused before the session is read. Liveness and readiness carry none,
	// and record no span: a probe every few seconds would fill the traces.
	handler := telemetry.Requests(sessions.Under(restapi.Prefix, mux), logger, tracing.Provider, health.Paths()...)
	return lifecycle.Serve(shutdown, listener, handler, settings.ShutdownDrainDeadline, logger)
}
