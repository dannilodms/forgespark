package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/email/application"
	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
)

var errUnreachable = errors.New("dial tcp 127.0.0.1:1025: connection refused")

// queuedLoginCode queues the e-mail of the login code request 42 and tells
// the source of login codes its code, as the area of the login would.
func (m *mailroom) queuedLoginCode(t *testing.T, language string) domain.EmailID {
	t.Helper()
	require.NoError(t, m.queue.Execute(context.Background(), aLoginCode(t, language)))
	m.loginCodes.Give(domain.LoginCodeRequest(42), domain.LoginCodeContent{Code: "482915", Lifetime: 10 * time.Minute})
	saved := m.deliveries.Saved()
	return saved[len(saved)-1]
}

func (m *mailroom) emailOf(t *testing.T, id domain.EmailID) *domain.Email {
	t.Helper()
	email, found := m.emails.Find(id)
	require.True(t, found)
	return &email
}

func attempt(id domain.EmailID) application.DeliverEmailInput {
	return application.DeliverEmailInput{EmailID: id, LastAttempt: false}
}

func lastAttempt(id domain.EmailID) application.DeliverEmailInput {
	return application.DeliverEmailInput{EmailID: id, LastAttempt: true}
}

func TestDeliveringAnEmailSendsItInTheLanguageOfItsRecipientAndMarksItSent(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "pt-BR")
	room.clock.Advance(2 * time.Second)

	err := room.deliver.Execute(context.Background(), attempt(id))

	require.NoError(t, err)
	sent := room.sender.Sent()
	require.Len(t, sent, 1)
	assert.Equal(t, id, sent[0].EmailID)
	assert.Equal(t, "ana.souza@fabrica.example", sent[0].To)
	assert.Equal(t, "Seu código de login", sent[0].Subject)
	assert.Contains(t, sent[0].Text, "482915")
	assert.Contains(t, sent[0].Text, "10 minutos")
	assert.Contains(t, sent[0].HTML, "482915")
	email := room.emailOf(t, id)
	assert.Equal(t, domain.Sent, email.Status())
	messageID, _ := email.MessageID()
	assert.Equal(t, "message-1", messageID)
	sentAt, _ := email.SentAt()
	assert.Equal(t, noon.Add(2*time.Second), sentAt)
}

func TestASecondRunOnAnEmailAlreadySentSendsNothing(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	require.NoError(t, room.deliver.Execute(context.Background(), attempt(id)))
	first := room.emailOf(t, id)

	err := room.deliver.Execute(context.Background(), attempt(id))

	require.NoError(t, err)
	assert.Len(t, room.sender.Sent(), 1)
	assert.Equal(t, first, room.emailOf(t, id))
}

func TestAFailedAttemptLeavesTheEmailQueuedForTheNextOne(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	room.sender.FailWith(errUnreachable)

	err := room.deliver.Execute(context.Background(), attempt(id))

	require.ErrorIs(t, err, errUnreachable)
	assert.Equal(t, domain.Queued, room.emailOf(t, id).Status())

	room.sender.FailWith(nil)
	require.NoError(t, room.deliver.Execute(context.Background(), attempt(id)))
	assert.Len(t, room.sender.Sent(), 1)
	assert.Equal(t, domain.Sent, room.emailOf(t, id).Status())
}

func TestTheLastFailedAttemptMarksTheEmailFailedWithWhatFailed(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	room.sender.FailWith(errUnreachable)
	room.clock.Advance(6 * time.Minute)

	err := room.deliver.Execute(context.Background(), lastAttempt(id))

	require.ErrorIs(t, err, errUnreachable, "the job still fails")
	email := room.emailOf(t, id)
	assert.Equal(t, domain.Failed, email.Status())
	code, failedAt, failed := email.Failure()
	assert.True(t, failed)
	assert.Equal(t, domain.SendingFailed, code)
	assert.Equal(t, noon.Add(6*time.Minute), failedAt)
}

func TestAnEmailThatFailedIsNotSentAgain(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	room.sender.FailWith(errUnreachable)
	require.Error(t, room.deliver.Execute(context.Background(), lastAttempt(id)))
	room.sender.FailWith(nil)

	err := room.deliver.Execute(context.Background(), attempt(id))

	require.NoError(t, err)
	assert.Empty(t, room.sender.Sent())
	assert.Equal(t, domain.Failed, room.emailOf(t, id).Status())
}

func TestAnEmailWhoseContentCannotBeGivenIsNeverSentAndFailsWithItsOwnCode(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	gone := errors.New("the login code request 42 does not exist")
	room.loginCodes.FailWith(gone)

	err := room.deliver.Execute(context.Background(), lastAttempt(id))

	require.ErrorIs(t, err, gone)
	assert.Empty(t, room.sender.Sent())
	code, _, _ := room.emailOf(t, id).Failure()
	assert.Equal(t, domain.ContentUnavailable, code)
}

func TestAnEmailOfATemplateNothingGivesTheContentOfIsNeverSent(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	deliver := application.NewDeliverEmail(room.transactions, room.emails, application.ContentSources{}, room.sender, room.clock)

	err := deliver.Execute(context.Background(), lastAttempt(id))

	require.Error(t, err)
	assert.Empty(t, room.sender.Sent())
	code, _, _ := room.emailOf(t, id).Failure()
	assert.Equal(t, domain.ContentUnavailable, code)
}

// A job cut by the shutdown returns the error of its context, and River runs
// it again later with its attempt given back: even on the last attempt, the
// e-mail stays queued.
func TestARunCutByTheShutdownLeavesTheEmailQueuedEvenOnTheLastAttempt(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	room.sender.FailWith(context.Canceled)

	err := room.deliver.Execute(ctx, lastAttempt(id))

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, domain.Queued, room.emailOf(t, id).Status())
}

// A run whose time ran out is a failure, not a cut: River gives no attempt
// back. On the last attempt the e-mail is marked failed, in a transaction
// whose context is not the one that ran out, as the repository refuses that
// one as the database does.
func TestTheLastAttemptWhoseTimeRanOutMarksTheEmailFailed(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	room.sender.HangUntilTheContextEnds()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := room.deliver.Execute(ctx, lastAttempt(id))

	require.ErrorIs(t, err, context.DeadlineExceeded)
	code, _, failed := room.emailOf(t, id).Failure()
	assert.True(t, failed, "the e-mail stays queued with no job to send it")
	assert.Equal(t, domain.SendingFailed, code)
}

// Before the last attempt, a run whose time ran out leaves the e-mail queued
// for the next one, as any failure does.
func TestAnAttemptWhoseTimeRanOutLeavesTheEmailQueuedForTheNextOne(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")
	room.sender.HangUntilTheContextEnds()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := room.deliver.Execute(ctx, attempt(id))

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, domain.Queued, room.emailOf(t, id).Status())
}

func TestAJobOfAnEmailThatDoesNotExistFails(t *testing.T) {
	room := newMailroom()

	err := room.deliver.Execute(context.Background(), lastAttempt(99))

	assert.Error(t, err)
	assert.Empty(t, room.sender.Sent())
}
