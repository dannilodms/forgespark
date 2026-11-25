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

// sentLoginCode queues the e-mail of a login code and delivers it: the
// in-memory sender gives it the identifier `message-<n>`.
func (m *mailroom) sentLoginCode(t *testing.T) domain.EmailID {
	t.Helper()
	id := m.queuedLoginCode(t, "en-US")
	require.NoError(t, m.deliver.Execute(context.Background(), attempt(id)))
	return id
}

func bounceNotice(notice string, id domain.EmailID, messageID string) application.MarkEmailBouncedInput {
	return application.MarkEmailBouncedInput{Notice: notice, EmailID: id, MessageID: messageID}
}

func TestABounceTheSenderReportsMarksTheEmailBounced(t *testing.T) {
	room := newMailroom()
	id := room.sentLoginCode(t)
	room.clock.Advance(3 * time.Minute)

	err := room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", id, "message-1"))

	require.NoError(t, err)
	email := room.emailOf(t, id)
	assert.Equal(t, domain.Bounced, email.Status())
	bouncedAt, bounced := email.BouncedAt()
	assert.True(t, bounced)
	assert.Equal(t, noon.Add(3*time.Minute), bouncedAt)
	sentAt, _ := email.SentAt()
	assert.Equal(t, noon, sentAt, "the bounce keeps when the e-mail was sent")
	assert.Equal(t, []string{"msg_2fKq8"}, room.notices.Kept())
}

func TestANoticeDeliveredAgainChangesNothing(t *testing.T) {
	room := newMailroom()
	first := room.sentLoginCode(t)
	second := room.sentLoginCode(t)
	require.NoError(t, room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", first, "message-1")))

	// The same notice again, naming another e-mail: only its identifier
	// tells it was taken.
	err := room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", second, "message-2"))

	require.NoError(t, err)
	assert.Equal(t, domain.Bounced, room.emailOf(t, first).Status())
	assert.Equal(t, domain.Sent, room.emailOf(t, second).Status())
	assert.Equal(t, []string{"msg_2fKq8"}, room.notices.Kept())
}

func TestANoticeOfAMessageTheEmailWasNotSentAsChangesNothing(t *testing.T) {
	room := newMailroom()
	id := room.sentLoginCode(t)

	err := room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", id, "message-of-another-sending"))

	require.NoError(t, err)
	assert.Equal(t, domain.Sent, room.emailOf(t, id).Status())
	assert.Equal(t, []string{"msg_2fKq8"}, room.notices.Kept(), "a notice that names no e-mail of the system is taken once")
}

func TestANoticeOfAnEmailThatDoesNotExistChangesNothing(t *testing.T) {
	room := newMailroom()

	err := room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", 404, "message-1"))

	require.NoError(t, err)
	assert.Empty(t, room.emails.All())
	assert.Equal(t, []string{"msg_2fKq8"}, room.notices.Kept())
}

func TestANoticeOfAnEmailThatFailedOrAlreadyBouncedChangesNothing(t *testing.T) {
	room := newMailroom()
	failed := room.queuedLoginCode(t, "en-US")
	room.sender.FailWith(errUnreachable)
	require.Error(t, room.deliver.Execute(context.Background(), lastAttempt(failed)))
	room.sender.FailWith(nil)
	bounced := room.sentLoginCode(t)
	require.NoError(t, room.markBounced.Execute(context.Background(), bounceNotice("msg_first", bounced, "message-1")))
	before := room.emails.All()

	room.clock.Advance(time.Hour)
	ofFailed := room.markBounced.Execute(context.Background(), bounceNotice("msg_of_failed", failed, ""))
	ofBounced := room.markBounced.Execute(context.Background(), bounceNotice("msg_second", bounced, "message-1"))

	require.NoError(t, ofFailed)
	require.NoError(t, ofBounced)
	assert.Equal(t, before, room.emails.All())
	assert.Equal(t, []string{"msg_first", "msg_of_failed", "msg_second"}, room.notices.Kept())
}

// The sender reports a bounce after it took the e-mail, which the delivery
// marks sent in the same transaction. A notice that arrives while the e-mail
// is still queued, such as one whose commit failed, waits: the sender
// delivers a notice again until it is taken.
func TestANoticeOfAnEmailStillQueuedFailsAndIsTakenOnceTheEmailIsSent(t *testing.T) {
	room := newMailroom()
	id := room.queuedLoginCode(t, "en-US")

	err := room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", id, "message-1"))

	require.Error(t, err)
	assert.Equal(t, domain.Queued, room.emailOf(t, id).Status())
	assert.Empty(t, room.notices.Kept(), "a notice that failed is not taken")

	require.NoError(t, room.deliver.Execute(context.Background(), attempt(id)))
	require.NoError(t, room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", id, "message-1")))
	assert.Equal(t, domain.Bounced, room.emailOf(t, id).Status())
}

func TestANoticeIsNotTakenWhenTheBounceCannotBeSaved(t *testing.T) {
	room := newMailroom()
	id := room.sentLoginCode(t)
	refused := errors.New("the database refused the write")
	room.emails.FailWith(refused)

	err := room.markBounced.Execute(context.Background(), bounceNotice("msg_2fKq8", id, "message-1"))

	require.ErrorIs(t, err, refused)
	assert.Equal(t, domain.Sent, room.emailOf(t, id).Status())
	assert.Empty(t, room.notices.Kept())
}

func TestANoticeWithNoIdentifierIsRefused(t *testing.T) {
	room := newMailroom()
	id := room.sentLoginCode(t)

	err := room.markBounced.Execute(context.Background(), bounceNotice("", id, "message-1"))

	require.Error(t, err)
	assert.Equal(t, domain.Sent, room.emailOf(t, id).Status())
	assert.Empty(t, room.notices.Kept())
}
