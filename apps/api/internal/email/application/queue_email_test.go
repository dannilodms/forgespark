package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/application/applicationtest"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// mailroom is the e-mail area on its in-memory implementations: the two use
// cases, and what they write to.
type mailroom struct {
	clock        *applicationtest.Clock
	transactions *applicationtest.Transactions
	emails       *applicationtest.EmailRepository
	deliveries   *applicationtest.DeliveryJobs
	loginCodes   *applicationtest.ContentSource
	sender       *applicationtest.Sender
	queue        *application.QueueEmail
	deliver      *application.DeliverEmail
}

func newMailroom() *mailroom {
	clock := applicationtest.NewClock(noon)
	emails := applicationtest.NewEmailRepository()
	deliveries := applicationtest.NewDeliveryJobs()
	transactions := applicationtest.NewTransactions(emails, deliveries)
	loginCodes := applicationtest.NewContentSource()
	sender := applicationtest.NewSender()
	contents := application.ContentSources{application.LoginCode: loginCodes}
	return &mailroom{
		clock:        clock,
		transactions: transactions,
		emails:       emails,
		deliveries:   deliveries,
		loginCodes:   loginCodes,
		sender:       sender,
		queue:        application.NewQueueEmail(transactions, emails, deliveries, clock),
		deliver:      application.NewDeliverEmail(transactions, emails, contents, sender, clock),
	}
}

func languageOf(t *testing.T, tag string) kernel.Language {
	t.Helper()
	language, err := kernel.NewLanguage(tag)
	require.NoError(t, err)
	return language
}

// aLoginCode is the e-mail of the login code request 42 to Ana, in her
// language.
func aLoginCode(t *testing.T, language string) application.QueueEmailInput {
	t.Helper()
	return application.QueueEmailInput{
		Recipient: "ana.souza@fabrica.example",
		Template:  application.LoginCode,
		Language:  languageOf(t, language),
		Reference: application.LoginCodeRequest(42),
	}
}

func (m *mailroom) requireNothingSaved(t *testing.T) {
	t.Helper()
	assert.Empty(t, m.emails.All(), "an e-mail was recorded")
	assert.Empty(t, m.deliveries.Saved(), "a job was saved")
}

func TestQueueingAnEmailSavesItsRecordAndTheJobThatDeliversItTogether(t *testing.T) {
	room := newMailroom()

	err := room.queue.Execute(context.Background(), aLoginCode(t, "es"))

	require.NoError(t, err)
	emails := room.emails.All()
	require.Len(t, emails, 1)
	email := emails[0]
	assert.Equal(t, "ana.souza@fabrica.example", email.Recipient().String())
	assert.Equal(t, "Tu código de inicio de sesión", email.Subject())
	assert.Equal(t, domain.LoginCode, email.Template())
	assert.Equal(t, "es", email.Language().String())
	assert.Equal(t, domain.LoginCodeRequest(42), email.Reference())
	assert.Equal(t, domain.Queued, email.Status())
	assert.Equal(t, noon, email.QueuedAt())
	assert.Equal(t, []domain.EmailID{email.ID()}, room.deliveries.Saved(), "the job names the e-mail by its identifier")
}

func TestNothingIsSavedWhenTheJobCannotBeSaved(t *testing.T) {
	room := newMailroom()
	refused := errors.New("the jobs table refused the row")
	room.deliveries.FailWith(refused)

	err := room.queue.Execute(context.Background(), aLoginCode(t, "en-US"))

	require.ErrorIs(t, err, refused)
	room.requireNothingSaved(t)
}

func TestNothingIsSavedWhenTheRecordCannotBeSaved(t *testing.T) {
	room := newMailroom()
	refused := errors.New("the table of e-mails refused the row")
	room.emails.FailWith(refused)

	err := room.queue.Execute(context.Background(), aLoginCode(t, "en-US"))

	require.ErrorIs(t, err, refused)
	room.requireNothingSaved(t)
}

// Another area queues the e-mail inside its own transaction: when that
// transaction fails afterwards, neither the record nor the job is kept.
func TestAnEmailQueuedInTheTransactionOfTheCallerIsUndoneWithIt(t *testing.T) {
	room := newMailroom()
	broken := errors.New("a later rule of the caller broke")

	err := room.transactions.Run(context.Background(), func(ctx context.Context) error {
		require.NoError(t, room.queue.Execute(ctx, aLoginCode(t, "en-US")))
		return broken
	})

	require.ErrorIs(t, err, broken)
	room.requireNothingSaved(t)
}

func TestARecipientThatIsNotAnAddressIsRefusedAndNothingIsSaved(t *testing.T) {
	room := newMailroom()
	input := aLoginCode(t, "en-US")
	input.Recipient = "Ana Souza <ana.souza@fabrica.example>"

	err := room.queue.Execute(context.Background(), input)

	require.ErrorIs(t, err, application.RecipientInvalidError{})
	assert.NotContains(t, err.Error(), "ana.souza@fabrica.example")
	room.requireNothingSaved(t)
}
