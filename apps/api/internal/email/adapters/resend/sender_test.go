package resend_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/resend"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// These tests send to a simulated Resend, which answers as the documentation
// of its API says. The tests against Resend itself are in
// against_resend_test.go and run on a schedule only.

const (
	apiKey    = "re_simulated_key_8XzP"
	recipient = "ana.souza@fabrica.example"
	resendID  = "4ef8a1c2-0b7d-4f3e-9a51-6d2c8e0f1b37"
)

// simulatedResend is a stand-in for the API of Resend: it keeps what each
// request carried and answers what the test tells it.
type simulatedResend struct {
	*httptest.Server

	mu       sync.Mutex
	requests []received
	status   int
	answer   string
}

// received is what one request carried.
type received struct {
	method string
	path   string
	header http.Header
	body   map[string]any
}

func newSimulatedResend(t *testing.T) *simulatedResend {
	t.Helper()
	simulated := &simulatedResend{
		Server:   nil,
		mu:       sync.Mutex{},
		requests: nil,
		status:   http.StatusOK,
		answer:   `{"id":"` + resendID + `"}`,
	}
	simulated.Server = httptest.NewServer(http.HandlerFunc(simulated.serve))
	t.Cleanup(simulated.Close)
	return simulated
}

func (s *simulatedResend) serve(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)
	s.mu.Lock()
	s.requests = append(s.requests, received{method: r.Method, path: r.URL.Path, header: r.Header.Clone(), body: body})
	status, answer := s.status, s.answer
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, answer)
}

// answers makes every later request get this status and body.
func (s *simulatedResend) answers(status int, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status, s.answer = status, body
}

func (s *simulatedResend) received(t *testing.T) received {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	require.Len(t, s.requests, 1)
	return s.requests[0]
}

func settingsOf(url string, provider trace.TracerProvider) resend.Settings {
	return resend.Settings{
		URL:            url,
		APIKey:         config.NewSecret(apiKey),
		FromAddress:    "no-reply@fabrica.example",
		FromName:       "Fábrica Exemplo",
		TracerProvider: provider,
	}
}

func aMessage() domain.Message {
	return domain.Message{
		EmailID: 37,
		To:      recipient,
		Subject: "Seu código de login",
		HTML:    `<!DOCTYPE html><html lang="pt-BR"><body><p><b>482915</b></p></body></html>`,
		Text:    "Digite este código:\n\n    482915\n",
	}
}

func TestTheSenderSendsTheMessageWithTheIdentifierOfTheRecordAsTheIdempotencyKeyAndATag(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedResend(t)

	messageID, err := resend.NewSender(settingsOf(simulated.URL, nil)).Send(context.Background(), aMessage())

	require.NoError(t, err)
	assert.Equal(t, resendID, messageID, "the identifier Resend answered is the one stored")
	request := simulated.received(t)
	assert.Equal(t, http.MethodPost, request.method)
	assert.Equal(t, "/emails", request.path)
	assert.Equal(t, "Bearer "+apiKey, request.header.Get("Authorization"))
	assert.Equal(t, "application/json", request.header.Get("Content-Type"))
	assert.Equal(t, "email-37", request.header.Get("Idempotency-Key"))
	assert.Empty(t, request.header.Get("Traceparent"), "no trace leaves for the e-mail service")
	assert.Equal(t, map[string]any{
		"from":    "=?utf-8?q?F=C3=A1brica_Exemplo?= <no-reply@fabrica.example>",
		"to":      []any{recipient},
		"subject": "Seu código de login",
		"html":    `<!DOCTYPE html><html lang="pt-BR"><body><p><b>482915</b></p></body></html>`,
		"text":    "Digite este código:\n\n    482915\n",
		"tags":    []any{map[string]any{"name": "email_id", "value": "37"}},
	}, request.body)
}

func TestTheNameOfTheSenderInPlainLettersIsQuoted(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedResend(t)
	settings := settingsOf(simulated.URL, nil)
	settings.FromName = "My Software Factory Manager"

	_, err := resend.NewSender(settings).Send(context.Background(), aMessage())

	require.NoError(t, err)
	assert.Equal(t, `"My Software Factory Manager" <no-reply@fabrica.example>`, simulated.received(t).body["from"])
}

// Resend answers a failure with its status, a name of the failure and a
// message, which may quote what it refused, the address of the recipient
// included. The failure keeps the status and the name only.
func TestAFailureIsToldByItsStatusAndNameAndNeverByTheMessageOfResend(t *testing.T) {
	t.Parallel()
	cases := []struct {
		status int
		name   string
	}{
		{http.StatusUnprocessableEntity, "validation_error"},
		{http.StatusUnauthorized, "missing_api_key"},
		{http.StatusForbidden, "validation_error"},
		{http.StatusConflict, "concurrent_idempotent_requests"},
		{http.StatusConflict, "invalid_idempotent_request"},
		{http.StatusTooManyRequests, "rate_limit_exceeded"},
		{http.StatusInternalServerError, "application_error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			simulated := newSimulatedResend(t)
			simulated.answers(c.status, fmt.Sprintf(`{"statusCode":%d,"name":%q,"message":"The address %s was refused."}`, c.status, c.name, recipient))

			messageID, err := resend.NewSender(settingsOf(simulated.URL, nil)).Send(context.Background(), aMessage())

			require.Error(t, err)
			assert.Empty(t, messageID)
			assert.Equal(t, fmt.Sprintf("the API of Resend did not take the e-mail: answered %d %s", c.status, c.name), err.Error())
		})
	}
}

func TestAnAnswerThatNamesNoKnownFailureIsToldByItsStatus(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		`<html>Bad gateway for ` + recipient + `</html>`,
		`{"name":"refused ` + recipient + `"}`,
		``,
	} {
		simulated := newSimulatedResend(t)
		simulated.answers(http.StatusBadGateway, body)

		_, err := resend.NewSender(settingsOf(simulated.URL, nil)).Send(context.Background(), aMessage())

		require.Error(t, err)
		assert.Equal(t, "the API of Resend did not take the e-mail: answered 502", err.Error())
	}
}

func TestAnAcceptedAnswerWithNoIdentifierIsAFailure(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{}`, `{"id":""}`, `not json`} {
		simulated := newSimulatedResend(t)
		simulated.answers(http.StatusOK, body)

		messageID, err := resend.NewSender(settingsOf(simulated.URL, nil)).Send(context.Background(), aMessage())

		require.Error(t, err, body)
		assert.Empty(t, messageID)
		assert.Contains(t, err.Error(), "the API of Resend took the e-mail and gave it no identifier")
	}
}

func TestAnUnreachableResendIsAFailureThatNamesNoAddress(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	unreachable := "http://" + listener.Addr().String()
	require.NoError(t, listener.Close())

	_, err = resend.NewSender(settingsOf(unreachable, nil)).Send(context.Background(), aMessage())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reach Resend")
	assert.NotContains(t, err.Error(), recipient)
}

func TestASendingCutByItsContextStops(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	t.Cleanup(slow.Close)
	t.Cleanup(func() { close(release) })
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := resend.NewSender(settingsOf(slow.URL, nil)).Send(ctx, aMessage())

	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestTheSpanOfASendingNamesResendAndTheStatusAndNeverTheAddress(t *testing.T) {
	t.Parallel()
	simulated := newSimulatedResend(t)
	simulated.answers(http.StatusUnprocessableEntity,
		`{"statusCode":422,"name":"validation_error","message":"Invalid to: `+recipient+`"}`)
	spans := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))

	_, err := resend.NewSender(settingsOf(simulated.URL, provider)).Send(context.Background(), aMessage())

	require.Error(t, err)
	ended := spans.Ended()
	require.Len(t, ended, 1)
	assert.Equal(t, "Resend send", ended[0].Name())
	assert.Equal(t, trace.SpanKindClient, ended[0].SpanKind())
	assert.Equal(t, codes.Error, ended[0].Status().Code)
	recorded := fmt.Sprint(ended[0].Attributes(), ended[0].Events(), ended[0].Status())
	assert.Contains(t, recorded, "127.0.0.1")
	assert.Contains(t, recorded, "422")
	assert.NotContains(t, recorded, recipient)
	assert.NotContains(t, recorded, apiKey)
}
