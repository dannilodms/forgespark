package health_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/platform/health"
)

// answer mounts liveness and readiness on a server and returns the status and
// the body of one request.
func answer(t *testing.T, ready health.Check, timeout time.Duration, path string) (int, string) {
	t.Helper()
	mux := http.NewServeMux()
	health.Mount(mux, ready, timeout)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	response, err := http.Get(server.URL + path)
	require.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return response.StatusCode, string(body)
}

func TestLivenessAnswers200AndTouchesNoPart(t *testing.T) {
	touched := false
	ready := func(context.Context) error {
		touched = true
		return errors.New("PostgreSQL is down")
	}

	status, body := answer(t, ready, time.Second, "/livez")

	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, body)
	assert.False(t, touched, "liveness ran the readiness check")
}

func TestReadinessAnswers200WhenTheCheckPasses(t *testing.T) {
	ready := func(context.Context) error { return nil }

	status, body := answer(t, ready, time.Second, "/readyz")

	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, body)
}

func TestReadinessAnswers503WithNoBodyWhenTheCheckFails(t *testing.T) {
	ready := func(context.Context) error { return errors.New("PostgreSQL is down") }

	status, body := answer(t, ready, time.Second, "/readyz")

	assert.Equal(t, http.StatusServiceUnavailable, status)
	assert.Empty(t, body)
}

func TestReadinessAnswers503WhenTheCheckOutlastsItsTimeout(t *testing.T) {
	ready := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	started := time.Now()
	status, _ := answer(t, ready, 50*time.Millisecond, "/readyz")

	assert.Equal(t, http.StatusServiceUnavailable, status)
	assert.Less(t, time.Since(started), 5*time.Second)
}
