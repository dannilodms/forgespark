// Package resend is the e-mail of production through Resend (E-mail in
// CODING_STANDARDS.md). Outbound, it sends an e-mail through Resend's API,
// with the identifier of its record as the idempotency key and as a tag.
// Inbound, it takes the notices of a bounce Resend sends to the API, signed
// with Svix.
//
// It is a small client of its own on the standard library: the API is one
// call, and a library would carry Resend's messages, which may quote the
// address of a recipient, into the errors. The address goes to Resend and
// nowhere else.
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/config"
)

// URL is where the API of Resend answers.
const URL = "https://api.resend.com"

// timeout bounds a sending, from the connection to the last byte of the
// answer. A job has a minute, and the worker gives the running jobs the
// shutdown deadline, 20 seconds by default.
const timeout = 15 * time.Second

// answerLimit is the most of an answer the sender reads: an identifier, or
// the name of a failure.
const answerLimit = 64 << 10

// EmailTag is the name of the tag that carries the identifier of the e-mail
// record. The notice of a bounce names the e-mail by it.
const EmailTag = "email_id"

// Settings is where Resend is, the key to call it with and whom the e-mail
// is from.
type Settings struct {
	// URL is where the API of Resend answers: URL, or a stand-in in the
	// tests.
	URL string
	// APIKey is the key of the Resend account, sent as a bearer token.
	APIKey config.Secret
	// FromAddress and FromName are the sender every e-mail shows.
	FromAddress string
	FromName    string
	// TracerProvider traces each sending. With none, nothing is traced.
	TracerProvider trace.TracerProvider
}

// Sender sends e-mail through the API of Resend.
type Sender struct {
	settings Settings
	client   *http.Client
	tracer   trace.Tracer
	server   string
}

// NewSender returns the sender of these settings. It calls nothing until an
// e-mail is sent.
func NewSender(settings Settings) *Sender {
	provider := settings.TracerProvider
	if provider == nil {
		provider = noop.NewTracerProvider()
	}
	server := ""
	if parsed, err := url.Parse(settings.URL); err == nil {
		server = parsed.Hostname()
	}
	return &Sender{
		settings: settings,
		client:   &http.Client{Timeout: timeout},
		tracer:   provider.Tracer("github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/resend"),
		server:   server,
	}
}

// Send sends the message and returns the identifier Resend gave it. Resend
// sends one e-mail for an idempotency key in 24 hours: a second sending of
// the same record answers the identifier of the first.
func (s *Sender) Send(ctx context.Context, message domain.Message) (string, error) {
	ctx, span := s.tracer.Start(ctx, "Resend send", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		semconv.ServerAddress(s.server),
		semconv.HTTPRequestMethodKey.String(http.MethodPost),
	))
	defer span.End()

	messageID, status, err := s.send(ctx, message)
	if status != 0 {
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetAttributes(attribute.String("resend.email_id", messageID))
	return messageID, nil
}

// sendRequest is the body of POST /emails.
type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
	Tags    []tag    `json:"tags"`
}

type tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// sent is the answer of Resend when it took the e-mail.
type sent struct {
	ID string `json:"id"`
}

// refusal is the answer of Resend when it did not. Its message is not read.
type refusal struct {
	Name string `json:"name"`
}

func (s *Sender) send(ctx context.Context, message domain.Message) (string, int, error) {
	request, err := s.request(ctx, message)
	if err != nil {
		return "", 0, err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return "", 0, fmt.Errorf("reach Resend: %w", err)
	}
	defer response.Body.Close()
	answer, err := io.ReadAll(io.LimitReader(response.Body, answerLimit))
	if err != nil {
		return "", response.StatusCode, fmt.Errorf("read the answer of Resend: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", response.StatusCode, refusalOf(response.StatusCode, answer)
	}
	var took sent
	if err := json.Unmarshal(answer, &took); err != nil || took.ID == "" {
		return "", response.StatusCode, errors.New("the API of Resend took the e-mail and gave it no identifier")
	}
	return took.ID, response.StatusCode, nil
}

// request writes the call that sends the message. The trace of the job does
// not go with it: an outbound call carries it to the agents of the system
// only (Telemetry in CODING_STANDARDS.md).
func (s *Sender) request(ctx context.Context, message domain.Message) (*http.Request, error) {
	identifier := strconv.FormatInt(int64(message.EmailID), 10)
	from := mail.Address{Name: s.settings.FromName, Address: s.settings.FromAddress}
	body, err := json.Marshal(sendRequest{
		From:    from.String(),
		To:      []string{message.To},
		Subject: message.Subject,
		HTML:    message.HTML,
		Text:    message.Text,
		Tags:    []tag{{Name: EmailTag, Value: identifier}},
	})
	if err != nil {
		return nil, errors.New("write the call to Resend")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.settings.URL+"/emails", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("write the call to Resend: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+s.settings.APIKey.Reveal())
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "email-"+identifier)
	return request, nil
}

// refusalOf tells why Resend did not take the e-mail by the status and the
// name of the failure, such as `validation_error`, which is a key
// (kernel.Key): anything else is not repeated. Its message, which may quote
// the address of the recipient, is left out.
func refusalOf(status int, answer []byte) error {
	var refused refusal
	if json.Unmarshal(answer, &refused) != nil || !isKey(refused.Name) {
		return fmt.Errorf("the API of Resend did not take the e-mail: answered %d", status)
	}
	return fmt.Errorf("the API of Resend did not take the e-mail: answered %d %s", status, refused.Name)
}

func isKey(text string) bool {
	_, err := kernel.NewKey(text)
	return err == nil
}
