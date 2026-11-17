package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/health"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// serveAPI wires the API and serves it until the context ends. It never
// changes the schema: readiness tells when a migration is pending.
func serveAPI(ctx context.Context, environment map[string]string, _, stderr io.Writer) error {
	settings, err := config.Load(config.ServeAPI, environment)
	if err != nil {
		return err
	}

	pool, err := database.NewPool(database.PoolSettings{
		ConnectionURL:            settings.DatabaseURL.ConnectionString(),
		MaxSize:                  settings.DatabasePoolMaxSize,
		StatementTimeout:         settings.DatabaseStatementTimeout,
		IdleInTransactionTimeout: settings.DatabaseIdleInTransactionTimeout,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	mux := http.NewServeMux()
	ready := func(ctx context.Context) error {
		return database.CheckSchema(ctx, pool, migrations.Files)
	}
	health.Mount(mux, ready, settings.HealthCheckTimeout)

	listener, err := net.Listen("tcp", settings.ListenAddress.String())
	if err != nil {
		return fmt.Errorf("listen on API_LISTEN_ADDRESS: %w", err)
	}
	fmt.Fprintf(stderr, "api: listening on %s\n", listener.Addr())

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	return serveUntilDone(ctx, server, listener)
}

// serveUntilDone serves until the context ends, then lets the requests in
// flight finish.
func serveUntilDone(ctx context.Context, server *http.Server, listener net.Listener) error {
	failed := make(chan error, 1)
	go func() { failed <- server.Serve(listener) }()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}
	if err := server.Shutdown(context.WithoutCancel(ctx)); err != nil {
		return err
	}
	if err := <-failed; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
