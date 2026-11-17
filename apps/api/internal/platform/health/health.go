// Package health answers liveness and readiness. Both live outside /api and
// outside the OpenAPI spec, need no authentication and answer with the status
// code alone. Rules: Health in CODING_STANDARDS.md.
package health

import (
	"context"
	"net/http"
	"time"
)

// Check tells whether the essential parts answer. A nil error means ready.
type Check func(ctx context.Context) error

// Mount adds GET /livez and GET /readyz to the mux. Liveness touches no part.
// Readiness runs the check under the timeout and answers 503 when it fails or
// does not end in time.
func Mount(mux *http.ServeMux, ready Check, timeout time.Duration) {
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		if err := ready(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
