package smtp_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/dannilodms/forgespark/apps/api/internal/email/adapters/smtp"
	"github.com/dannilodms/forgespark/apps/api/internal/email/adapters/smtp/smtptest"
	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
)

// These tests send to the Mailpit of development, at EMAIL_SMTP_ADDRESS, and
// read each e-mail back through its API, at MAILPIT_PORT. Each one sends to
// an address of its own, so they run together and beside development.

func required(t *testing.T, variable string) string {
	t.Helper()
	value := os.Getenv(variable)
	if value == "" {
		t.Fatalf("%s is not set: run the tests with `task api:test`, which reads .env.example", variable)
	}
	return value
}

func settingsOf(t *testing.T, address string) smtp.Settings {
	t.Helper()
	host, port, err := net.SplitHostPort(address)
	require.NoError(t, err)
	number, err := strconv.Atoi(port)
	require.NoError(t, err)
	return smtp.Settings{
		Host:           host,
		Port:           number,
		FromAddress:    "no-reply@fabrica.example",
		FromName:       "Fábrica Exemplo",
		TracerProvider: nil,
	}
}

// aRecipient is an address no other test sends to.
func aRecipient() string {
	return "ana+" + rand.Text()[:12] + "@fabrica.example"
}

func aMessage(to string) domain.Message {
	return domain.Message{
		EmailID: 7,
		To:      to,
		Subject: "Seu código de login",
		HTML:    `<!DOCTYPE html><html lang="pt-BR"><body><p>Digite este código:</p><p><b>482915</b></p></body></html>`,
		Text:    "Digite este código:\n\n    482915\n",
	}
}

// mailpitMessage is what the API of Mailpit answers about one e-mail.
type mailpitMessage struct {
	MessageID string
	From      struct{ Name, Address string }
	To        []struct{ Name, Address string }
	Subject   string
	Text      string
	HTML      string
}

// readBack finds the e-mail of this Message-ID in Mailpit and reads it.
func readBack(t *testing.T, messageID string) mailpitMessage {
	t.Helper()
	api := "http://localhost:" + required(t, "MAILPIT_PORT") + "/api/v1"
	var found struct {
		Messages []struct{ ID string }
	}
	require.Eventually(t, func() bool {
		getJSON(t, api+"/search?query="+url.QueryEscape(`message-id:"`+messageID+`"`), &found)
		return len(found.Messages) == 1
	}, 10*time.Second, 50*time.Millisecond, "Mailpit has no e-mail with the Message-ID %s", messageID)
	var message mailpitMessage
	getJSON(t, api+"/message/"+found.Messages[0].ID, &message)
	return message
}

func getJSON(t *testing.T, address string, into any) {
	t.Helper()
	response, err := http.Get(address)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode, address)
	require.NoError(t, json.NewDecoder(response.Body).Decode(into))
}

func TestAnEmailSentArrivesWithBothPartsAndTheRightRecipient(t *testing.T) {
	t.Parallel()
	sender := smtp.NewSender(settingsOf(t, required(t, "EMAIL_SMTP_ADDRESS")))
	to := aRecipient()

	messageID, err := sender.Send(context.Background(), aMessage(to))

	require.NoError(t, err)
	require.NotEmpty(t, messageID)
	assert.NotContains(t, messageID, "<", "the Message-ID is kept without its angle brackets")
	received := readBack(t, messageID)
	assert.Equal(t, messageID, received.MessageID)
	require.Len(t, received.To, 1)
	assert.Equal(t, to, received.To[0].Address)
	assert.Equal(t, "no-reply@fabrica.example", received.From.Address)
	assert.Equal(t, "Fábrica Exemplo", received.From.Name)
	assert.Equal(t, "Seu código de login", received.Subject)
	assert.Contains(t, received.HTML, "<b>482915</b>")
	assert.Contains(t, received.Text, "Digite este código:")
	assert.Contains(t, received.Text, "482915")
	assert.NotContains(t, received.Text, "<b>", "the plain-text part is its own")
}

// An SMTP server repeats the address it refuses, and go-mail lists it in its
// error: neither reaches the error of the sender or its span.
func TestARefusedRecipientFailsWithTheCodesOfTheServerAndNeverTheAddress(t *testing.T) {
	t.Parallel()
	server := smtptest.NewRefusingServer(t)
	spans := tracetest.NewSpanRecorder()
	settings := settingsOf(t, server.Address())
	settings.TracerProvider = sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans))
	to := aRecipient()

	_, err := smtp.NewSender(settings).Send(context.Background(), aMessage(to))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "RCPT TO")
	assert.Contains(t, err.Error(), "550 5.1.1")
	assert.NotContains(t, err.Error(), to)
	ended := spans.Ended()
	require.Len(t, ended, 1)
	assert.Equal(t, "SMTP send", ended[0].Name())
	assert.Equal(t, codes.Error, ended[0].Status().Code)
	recorded := fmt.Sprint(ended[0].Attributes(), ended[0].Events(), ended[0].Status())
	assert.Contains(t, recorded, "127.0.0.1")
	assert.NotContains(t, recorded, to)
}

func TestAServerThatCannotBeReachedFails(t *testing.T) {
	t.Parallel()
	// Port 1 of the machine has no SMTP server behind it.
	sender := smtp.NewSender(settingsOf(t, "127.0.0.1:1"))

	_, err := sender.Send(context.Background(), aMessage(aRecipient()))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reach the SMTP server")
}

func TestARecipientGoMailDoesNotTakeFailsWithoutRepeatingIt(t *testing.T) {
	t.Parallel()
	sender := smtp.NewSender(settingsOf(t, required(t, "EMAIL_SMTP_ADDRESS")))
	refused := "ana souza@@fabrica.example"

	_, err := sender.Send(context.Background(), aMessage(refused))

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "ana souza")
}
