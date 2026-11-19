package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

func languageOf(t *testing.T, tag string) kernel.Language {
	t.Helper()
	language, err := kernel.NewLanguage(tag)
	require.NoError(t, err)
	return language
}

func queued(t *testing.T, language string) *domain.Email {
	t.Helper()
	recipient, err := domain.NewRecipient("ana.souza@fabrica.example")
	require.NoError(t, err)
	email, err := domain.Queue(7, recipient, domain.LoginCode, languageOf(t, language), domain.LoginCodeRequest(42), noon)
	require.NoError(t, err)
	return email
}

func TestAnEmailIsQueuedWithTheSubjectOfItsTemplateInTheLanguageOfItsRecipient(t *testing.T) {
	email := queued(t, "pt-BR")

	assert.Equal(t, domain.EmailID(7), email.ID())
	assert.Equal(t, "ana.souza@fabrica.example", email.Recipient().String())
	assert.Equal(t, "Seu código de login", email.Subject())
	assert.Equal(t, domain.LoginCode, email.Template())
	assert.Equal(t, "pt-BR", email.Language().String())
	assert.Equal(t, domain.LoginCodeRequestKind, email.Reference().Kind())
	assert.Equal(t, int64(42), email.Reference().Identifier())
	assert.Equal(t, domain.Queued, email.Status())
	assert.True(t, email.IsQueued())
	assert.Equal(t, noon, email.QueuedAt())
	_, sent := email.SentAt()
	assert.False(t, sent)
	_, _, failed := email.Failure()
	assert.False(t, failed)
}

func TestTheEmailOfALoginCodeSaysTheCodeHowLongItWorksAndToIgnoreItWhenNotAskedFor(t *testing.T) {
	cases := map[string][]string{
		"en-US": {"Your login code", "10 minutes", "If you did not ask for this code, ignore this e-mail."},
		"pt-BR": {"Seu código de login", "10 minutos", "Se você não pediu este código, ignore este e-mail."},
		"es":    {"Tu código de inicio de sesión", "10 minutos", "Si no pediste este código, ignora este correo electrónico."},
	}
	for tag, expected := range cases {
		t.Run(tag, func(t *testing.T) {
			email := queued(t, tag)

			message, err := email.Message(domain.LoginCodeContent{Code: "482915", Lifetime: 10 * time.Minute})

			require.NoError(t, err)
			assert.Equal(t, domain.EmailID(7), message.EmailID)
			assert.Equal(t, "ana.souza@fabrica.example", message.To)
			assert.Equal(t, expected[0], message.Subject)
			for _, part := range []string{message.Text, message.HTML} {
				assert.Contains(t, part, "482915")
				assert.Contains(t, part, expected[1])
				assert.Contains(t, part, expected[2])
			}
			assert.Contains(t, message.HTML, `<html lang="`+tag+`">`)
		})
	}
}

func TestTheHTMLPartEscapesTheContent(t *testing.T) {
	email := queued(t, "en-US")

	message, err := email.Message(domain.LoginCodeContent{Code: "<b>48</b>", Lifetime: 10 * time.Minute})

	require.NoError(t, err)
	assert.Contains(t, message.HTML, "&lt;b&gt;48&lt;/b&gt;")
	assert.NotContains(t, message.HTML, "<b>48</b>")
	assert.Contains(t, message.Text, "<b>48</b>", "the plain-text part is plain text")
}

func TestAnEmailIsMarkedSentOnceWithTheIdentifierTheSenderGaveIt(t *testing.T) {
	email := queued(t, "en-US")
	later := noon.Add(3 * time.Second)

	require.NoError(t, email.MarkSent("message-1@fabrica.example", later))

	assert.Equal(t, domain.Sent, email.Status())
	assert.False(t, email.IsQueued())
	messageID, sent := email.MessageID()
	assert.True(t, sent)
	assert.Equal(t, "message-1@fabrica.example", messageID)
	sentAt, _ := email.SentAt()
	assert.Equal(t, later, sentAt)
	assert.ErrorIs(t, email.MarkSent("message-2@fabrica.example", later),
		domain.OperationNotAllowedError{Operation: domain.MarkSentOperation, Status: domain.Sent})
	assert.ErrorIs(t, email.MarkFailed(domain.SendingFailed, later),
		domain.OperationNotAllowedError{Operation: domain.MarkFailedOperation, Status: domain.Sent})
}

func TestAnEmailIsMarkedFailedWithWhyAndCanThenBeNeitherSentNorFailedAgain(t *testing.T) {
	email := queued(t, "en-US")
	later := noon.Add(6 * time.Minute)

	require.NoError(t, email.MarkFailed(domain.SendingFailed, later))

	assert.Equal(t, domain.Failed, email.Status())
	failure, failedAt, failed := email.Failure()
	assert.True(t, failed)
	assert.Equal(t, domain.SendingFailed, failure)
	assert.Equal(t, later, failedAt)
	var refused domain.OperationNotAllowedError
	require.ErrorAs(t, email.MarkSent("message-1@fabrica.example", later), &refused)
	assert.Equal(t, "email_operation_not_allowed", refused.Code())
	assert.Equal(t, map[string]any{"operation": "mark_sent", "status": "failed"}, refused.Params())
	assert.Error(t, email.MarkFailed(domain.SendingFailed, later))
}

func TestASentEmailIsMarkedBouncedOnceForTheMessageItWasSentAs(t *testing.T) {
	email := queued(t, "en-US")
	require.NoError(t, email.MarkSent("4ef8a1c2-resend", noon))
	later := noon.Add(2 * time.Minute)

	require.ErrorIs(t, email.MarkBounced("another-message", later), domain.BounceOfAnotherMessageError{})
	assert.Equal(t, domain.Sent, email.Status(), "a bounce of another message changes nothing")
	require.NoError(t, email.MarkBounced("4ef8a1c2-resend", later))

	assert.Equal(t, domain.Bounced, email.Status())
	bouncedAt, bounced := email.BouncedAt()
	assert.True(t, bounced)
	assert.Equal(t, later, bouncedAt)
	messageID, _ := email.MessageID()
	assert.Equal(t, "4ef8a1c2-resend", messageID)
	assert.ErrorIs(t, email.MarkBounced("4ef8a1c2-resend", later),
		domain.OperationNotAllowedError{Operation: domain.MarkBouncedOperation, Status: domain.Bounced})
}

func TestOnlyASentEmailCanBounce(t *testing.T) {
	stillQueued := queued(t, "en-US")
	failed := queued(t, "en-US")
	require.NoError(t, failed.MarkFailed(domain.SendingFailed, noon))

	assert.ErrorIs(t, stillQueued.MarkBounced("", noon),
		domain.OperationNotAllowedError{Operation: domain.MarkBouncedOperation, Status: domain.Queued})
	assert.ErrorIs(t, failed.MarkBounced("", noon),
		domain.OperationNotAllowedError{Operation: domain.MarkBouncedOperation, Status: domain.Failed})
	_, bounced := stillQueued.BouncedAt()
	assert.False(t, bounced)
}

func TestAMessageIsWrittenOnlyWithTheContentOfItsTemplate(t *testing.T) {
	email := queued(t, "en-US")

	_, err := email.Message(otherContent{})

	assert.Error(t, err)
}

type otherContent struct{}

func (otherContent) Template() domain.Template { return "some_other_template" }

func TestARecipientIsOneAddressAndItsErrorNeverRepeatsIt(t *testing.T) {
	for _, address := range []string{
		"ana.souza",
		"Ana Souza <ana.souza@fabrica.example>",
		"ana.souza@fabrica.example, bruno@fabrica.example",
		" ana.souza@fabrica.example",
	} {
		_, err := domain.NewRecipient(address)

		require.ErrorIs(t, err, domain.RecipientInvalidError{}, address)
		assert.NotContains(t, err.Error(), address)
	}
	_, err := domain.NewRecipient("")
	assert.ErrorIs(t, err, domain.RecipientInvalidError{})
}

func TestAnEmailIsQueuedOnlyWithAKnownTemplateALanguageAndAReference(t *testing.T) {
	recipient, err := domain.NewRecipient("ana.souza@fabrica.example")
	require.NoError(t, err)
	english := languageOf(t, "en-US")

	_, unknownTemplate := domain.Queue(1, recipient, "newsletter", english, domain.LoginCodeRequest(1), noon)
	_, noLanguage := domain.Queue(1, recipient, domain.LoginCode, kernel.Language{}, domain.LoginCodeRequest(1), noon)
	_, noReference := domain.Queue(1, recipient, domain.LoginCode, english, domain.LoginCodeRequest(0), noon)
	_, unknownReference := domain.Queue(1, recipient, domain.LoginCode, english, domain.RestoreReference("invoice", 1), noon)

	assert.Error(t, unknownTemplate)
	assert.Error(t, noLanguage)
	assert.Error(t, noReference)
	assert.Error(t, unknownReference)
}
