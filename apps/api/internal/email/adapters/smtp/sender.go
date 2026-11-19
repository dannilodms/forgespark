// Package smtp sends e-mail to an SMTP server with go-mail: Mailpit in
// development, the tests and staging (E-mail in CODING_STANDARDS.md). The
// address of a recipient goes to the server and nowhere else: an SMTP server
// repeats it in what it answers, and go-mail in its errors, so a failure is
// told by its step and its codes, never by its text.
package smtp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
)

// timeout bounds the connection to the server and each of its answers. A
// job has a minute, and the worker gives the running jobs the shutdown
// deadline, 20 seconds by default.
const timeout = 15 * time.Second

// Settings is where the SMTP server is and whom the e-mail is from.
type Settings struct {
	// Host and Port are where the SMTP server takes e-mail.
	Host string
	Port int
	// FromAddress and FromName are the sender every e-mail shows.
	FromAddress string
	FromName    string
	// TracerProvider traces each sending. With none, nothing is traced.
	TracerProvider trace.TracerProvider
}

// Sender sends e-mail to one SMTP server. It opens a connection for each
// e-mail.
type Sender struct {
	settings Settings
	tracer   trace.Tracer
}

// NewSender returns the sender of these settings. It connects to nothing
// until an e-mail is sent.
func NewSender(settings Settings) *Sender {
	provider := settings.TracerProvider
	if provider == nil {
		provider = noop.NewTracerProvider()
	}
	return &Sender{settings: settings, tracer: provider.Tracer("github.com/dannilodms/software-factory-manager/apps/api/internal/email/adapters/smtp")}
}

// Send sends the message, with its plain-text part and its HTML part as two
// alternatives, and returns its Message-ID, without the angle brackets.
func (s *Sender) Send(ctx context.Context, message domain.Message) (string, error) {
	ctx, span := s.tracer.Start(ctx, "SMTP send", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		semconv.ServerAddress(s.settings.Host),
		semconv.ServerPort(s.settings.Port),
	))
	defer span.End()

	messageID, err := s.send(ctx, message)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	return messageID, nil
}

func (s *Sender) send(ctx context.Context, message domain.Message) (string, error) {
	written, err := s.write(message)
	if err != nil {
		return "", err
	}
	client, err := mail.NewClient(s.settings.Host,
		mail.WithPort(s.settings.Port),
		// STARTTLS when the server offers it: Mailpit offers none.
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithTimeout(timeout),
	)
	if err != nil {
		return "", fmt.Errorf("prepare the SMTP client: %w", err)
	}
	if err := client.DialAndSendWithContext(ctx, written); err != nil {
		return "", failureOf(err)
	}
	return strings.Trim(written.GetMessageID(), "<>"), nil
}

// write builds the message go-mail sends. go-mail repeats an address it
// refuses, so its errors are replaced.
func (s *Sender) write(message domain.Message) (*mail.Msg, error) {
	written := mail.NewMsg()
	if err := written.FromFormat(s.settings.FromName, s.settings.FromAddress); err != nil {
		return nil, errors.New("write the e-mail: the address of the sender is not one go-mail takes")
	}
	if err := written.To(message.To); err != nil {
		return nil, errors.New("write the e-mail: the address of the recipient is not one go-mail takes")
	}
	written.Subject(message.Subject)
	written.SetMessageID()
	written.SetDate()
	written.SetBodyString(mail.TypeTextPlain, message.Text)
	written.AddAlternativeString(mail.TypeTextHTML, message.HTML)
	return written, nil
}

// failureOf tells why the server did not take the e-mail by the step that
// failed and the codes the server answered with. The text of the answer,
// which can repeat the address of the recipient, and the recipients go-mail
// lists in its error are left out. A failure to connect has no answer of the
// server and is kept as it is.
func failureOf(err error) error {
	var refused *mail.SendError
	if !errors.As(err, &refused) {
		return fmt.Errorf("reach the SMTP server: %w", err)
	}
	return fmt.Errorf("the SMTP server did not take the e-mail, %s: answered %d %s, temporary %t",
		refused.Reason, refused.ErrorCode(), refused.EnhancedStatusCode(), refused.IsTemp())
}
