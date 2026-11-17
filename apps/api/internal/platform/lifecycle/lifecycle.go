// Package lifecycle owns the life of the process: the signals that end it,
// the shutdown notice, and the server that stops without cutting a request
// short. No domain area handles a signal. Rules: Shutdown in
// CODING_STANDARDS.md.
package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	// readHeaderTimeout is how long a caller has to send the headers of a
	// request.
	readHeaderTimeout = 10 * time.Second
	// cutGrace is how long a request that was cut has to notice it, before
	// its connection is closed under it.
	cutGrace = 2 * time.Second
	// landingCheck is how often the drain looks for requests still in flight.
	landingCheck = 5 * time.Millisecond
)

// Signals returns the channel on which the signals that end the process
// arrive: SIGTERM and SIGINT.
func Signals() <-chan os.Signal {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	return signals
}

// ShutdownNotice returns the shutdown notice of the process: a context that
// ends when the first signal arrives. A second signal exits at once, with a
// status other than zero.
//
// The notice only tells that the shutdown started. A command that serves
// never derives the context of a request or of a query from it, because the
// work in flight must go on after the signal. A stream watches its Done
// channel to close when the shutdown starts.
func ShutdownNotice(signals <-chan os.Signal, exit func(status int)) context.Context {
	notice, start := context.WithCancel(context.Background())
	go func() {
		<-signals
		start()
		<-signals
		exit(1)
	}()
	return notice
}

// Serve answers requests on the listener until the shutdown notice, then
// stops without cutting work short.
//
// From the notice on, new connections are refused and the ones already open
// are still answered. The requests in flight have the drain deadline to
// finish. At the deadline the context of each remaining request is
// cancelled, which rolls back its transaction, and a log line says how many
// were cut.
//
// It returns nil when every request finished, and an error when one was cut
// or when the server could not take connections.
func Serve(shutdown context.Context, listener net.Listener, handler http.Handler, drainDeadline time.Duration, logger *slog.Logger) error {
	// The requests outlive the notice: only the drain deadline cancels them.
	requests, cut := context.WithCancel(context.WithoutCancel(shutdown))
	defer cut()
	flying := &inFlight{}
	server := &http.Server{
		Handler:           flying.count(handler),
		ReadHeaderTimeout: readHeaderTimeout,
		BaseContext:       func(net.Listener) context.Context { return requests },
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	var failure error
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		failure = server.Serve(listener)
	}()

	select {
	case <-stopped:
		return fmt.Errorf("take connections: %w", failure)
	case <-shutdown.Done():
	}

	// Closing the listener refuses new connections and leaves the open ones
	// alone, which http.Server.Shutdown would close as soon as they idle.
	_ = listener.Close()
	<-stopped
	deadline, release := context.WithTimeout(context.Background(), drainDeadline)
	defer release()
	if flying.landed(deadline) && server.Shutdown(deadline) == nil {
		return nil
	}

	left := flying.now()
	cut()
	grace, release := context.WithTimeout(context.Background(), cutGrace)
	defer release()
	flying.landed(grace)
	_ = server.Close()
	if left == 0 {
		return nil
	}
	logger.Error("requests cut by the shutdown", slog.Int64("count", left))
	return fmt.Errorf("requests cut by the shutdown: %d", left)
}

// inFlight counts the requests a handler is answering.
type inFlight struct {
	requests atomic.Int64
}

func (f *inFlight) count(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests.Add(1)
		defer f.requests.Add(-1)
		next.ServeHTTP(w, r)
	})
}

func (f *inFlight) now() int64 {
	return f.requests.Load()
}

// landed waits until no request is in flight. It tells whether that happened
// before the context ended.
func (f *inFlight) landed(ctx context.Context) bool {
	check := time.NewTicker(landingCheck)
	defer check.Stop()
	for f.now() > 0 && waited(ctx, check) {
	}
	return f.now() == 0
}

// waited blocks until the next tick, and tells whether the context is still
// open.
func waited(ctx context.Context, check *time.Ticker) bool {
	select {
	case <-ctx.Done():
		return false
	case <-check.C:
		return true
	}
}
