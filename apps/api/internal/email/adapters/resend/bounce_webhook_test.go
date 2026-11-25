package resend_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/email/adapters/resend"
	"github.com/dannilodms/forgespark/apps/api/internal/email/adapters/resend/resendtest"
	"github.com/dannilodms/forgespark/apps/api/internal/email/application"
	"github.com/dannilodms/forgespark/apps/api/internal/email/application/applicationtest"
	"github.com/dannilodms/forgespark/apps/api/internal/email/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/kernel"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/restapi"
)

var noon = time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)

// webhook is the webhook of the bounces on the in-memory e-mail area, served
// over HTTP, with one e-mail Resend took.
type webhook struct {
	url     string
	emails  *applicationtest.EmailRepository
	notices *applicationtest.BounceNotices
	clock   *applicationtest.Clock
	sentID  domain.EmailID
}

func newWebhook(t *testing.T) webhook {
	t.Helper()
	clock := applicationtest.NewClock(noon)
	emails := applicationtest.NewEmailRepository()
	notices := applicationtest.NewBounceNotices()
	transactions := applicationtest.NewTransactions(emails, notices)
	markBounced := application.NewMarkEmailBounced(transactions, emails, notices, clock)
	secret, err := config.NewWebhookSecret(resendtest.Secret)
	require.NoError(t, err)

	mux := http.NewServeMux()
	restapi.New(mux, restapi.Limits{BodyBytes: 1 << 20, BodyReadTimeout: 5 * time.Second})
	resend.NewBounceWebhook(secret, restapi.Limits{BodyBytes: 1 << 20, BodyReadTimeout: 5 * time.Second}, markBounced, clock).Mount(mux)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	hook := webhook{url: server.URL + resend.BounceWebhookPath, emails: emails, notices: notices, clock: clock, sentID: 0}
	hook.sentID = hook.sent(t, 37, resendID)
	return hook
}

// sent stores an e-mail Resend took, with the identifier it gave it.
func (w webhook) sent(t *testing.T, id domain.EmailID, resendID string) domain.EmailID {
	t.Helper()
	to, err := domain.NewRecipient(recipient)
	require.NoError(t, err)
	english, err := kernel.NewLanguage("en-US")
	require.NoError(t, err)
	email, err := domain.Queue(id, to, domain.LoginCode, english, domain.LoginCodeRequest(42), noon)
	require.NoError(t, err)
	require.NoError(t, email.MarkSent(resendID, noon))
	require.NoError(t, w.emails.Save(context.Background(), email))
	return id
}

func (w webhook) status(t *testing.T, id domain.EmailID) domain.Status {
	t.Helper()
	email, found := w.emails.Find(id)
	require.True(t, found)
	return email.Status()
}

// delivered is what the webhook answered.
type delivered struct {
	status    int
	mediaType string
	problem   map[string]any
}

// deliver sends a notice to the webhook with these headers, as Resend does.
func (w webhook) deliver(t *testing.T, headers []string, body string) delivered {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, w.url, strings.NewReader(body))
	require.NoError(t, err)
	for index := 1; index < len(headers); index += 2 {
		request.Header.Set(headers[index-1], headers[index])
	}
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	answer := delivered{status: response.StatusCode, mediaType: response.Header.Get("Content-Type"), problem: nil}
	if len(raw) > 0 {
		require.NoError(t, json.Unmarshal(raw, &answer.problem), "%s", raw)
	}
	return answer
}

// signed sends a notice signed with the secret of the webhook, sent now.
func (w webhook) signed(t *testing.T, notice, body string) delivered {
	t.Helper()
	return w.deliver(t, resendtest.Headers(resendtest.Secret, notice, w.clock.Now(), body), body)
}

func requireRefusedSignature(t *testing.T, answer delivered) {
	t.Helper()
	require.Equal(t, http.StatusUnauthorized, answer.status)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Equal(t, "unauthorized", answer.problem["code"])
	assert.Equal(t, "the signature of the notice is not valid", answer.problem["detail"])
}

func TestASignedBounceMarksTheEmailBounced(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	hook.clock.Advance(4 * time.Minute)

	answer := hook.signed(t, "msg_2fKq8bX1", resendtest.Bounce(resendID, 37, recipient))

	require.Equal(t, http.StatusNoContent, answer.status, answer.problem)
	email, _ := hook.emails.Find(hook.sentID)
	assert.Equal(t, domain.Bounced, email.Status())
	bouncedAt, _ := email.BouncedAt()
	assert.Equal(t, noon.Add(4*time.Minute), bouncedAt)
	assert.Equal(t, []string{"msg_2fKq8bX1"}, hook.notices.Kept())
}

func TestANoticeDeliveredAgainIsAnsweredAndChangesNothing(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	other := hook.sent(t, 38, "9b2d7e10-other")
	require.Equal(t, http.StatusNoContent, hook.signed(t, "msg_2fKq8bX1", resendtest.Bounce(resendID, 37, recipient)).status)

	again := hook.signed(t, "msg_2fKq8bX1", resendtest.Bounce("9b2d7e10-other", 38, recipient))

	assert.Equal(t, http.StatusNoContent, again.status, "Resend stops delivering a notice that was answered")
	assert.Equal(t, domain.Sent, hook.status(t, other))
}

func TestANoticeSignedWithAnotherSecretIsRefused(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	another := "whsec_" + base64.StdEncoding.EncodeToString([]byte("another key, which Resend has not"))
	body := resendtest.Bounce(resendID, 37, recipient)

	answer := hook.deliver(t, resendtest.Headers(another, "msg_2fKq8bX1", noon, body), body)

	requireRefusedSignature(t, answer)
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))
	assert.Empty(t, hook.notices.Kept())
}

func TestANoticeChangedAfterItWasSignedIsRefused(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	signedBody := resendtest.Event("email.delivered", resendID, map[string]string{"email_id": "37"}, recipient)
	changedBody := resendtest.Bounce(resendID, 37, recipient)

	answer := hook.deliver(t, resendtest.Headers(resendtest.Secret, "msg_2fKq8bX1", noon, signedBody), changedBody)

	requireRefusedSignature(t, answer)
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))
}

func TestANoticeWithAnotherIdentifierOrTimeThanTheOnesSignedIsRefused(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	body := resendtest.Bounce(resendID, 37, recipient)
	headers := resendtest.Headers(resendtest.Secret, "msg_2fKq8bX1", noon, body)

	otherNotice := hook.deliver(t, withHeader(headers, "svix-id", "msg_other"), body)
	otherTime := hook.deliver(t, withHeader(headers, "svix-timestamp", strconv.FormatInt(noon.Unix()+1, 10)), body)

	requireRefusedSignature(t, otherNotice)
	requireRefusedSignature(t, otherTime)
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))
}

func TestANoticeWithoutTheHeadersOfItsSignatureIsRefused(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	body := resendtest.Bounce(resendID, 37, recipient)
	headers := resendtest.Headers(resendtest.Secret, "msg_2fKq8bX1", noon, body)

	for _, missing := range []string{"svix-id", "svix-timestamp", "svix-signature"} {
		answer := hook.deliver(t, withHeader(headers, missing, ""), body)

		requireRefusedSignature(t, answer)
	}
	requireRefusedSignature(t, hook.deliver(t, withHeader(headers, "svix-timestamp", "noon"), body))
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))
}

// A signed notice sent long before or after now could be one recorded and
// sent again. Svix's own libraries allow five minutes either way.
func TestANoticeSignedMoreThanFiveMinutesFromNowIsRefused(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	body := resendtest.Bounce(resendID, 37, recipient)

	for _, sentAt := range []time.Time{noon.Add(-5*time.Minute - time.Second), noon.Add(5*time.Minute + time.Second)} {
		answer := hook.deliver(t, resendtest.Headers(resendtest.Secret, "msg_2fKq8bX1", sentAt, body), body)

		requireRefusedSignature(t, answer)
	}
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))

	inTime := hook.deliver(t, resendtest.Headers(resendtest.Secret, "msg_2fKq8bX1", noon.Add(-5*time.Minute), body), body)
	assert.Equal(t, http.StatusNoContent, inTime.status)
}

// While a secret is being replaced, Svix signs with both, separated by a
// space, each after the version of its scheme.
func TestANoticeIsTakenWhenOneOfItsSignaturesIsRight(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	body := resendtest.Bounce(resendID, 37, recipient)
	timestamp := strconv.FormatInt(noon.Unix(), 10)
	right := resendtest.Signature(resendtest.Secret, "msg_2fKq8bX1", timestamp, body)
	headers := resendtest.Headers(resendtest.Secret, "msg_2fKq8bX1", noon, body)

	onlyAnotherScheme := hook.deliver(t, withHeader(headers, "svix-signature", "v2,"+right), body)
	several := hook.deliver(t, withHeader(headers, "svix-signature", "v1,c29tZXRoaW5nIGVsc2U= v1,not-base64 v1,"+right), body)

	requireRefusedSignature(t, onlyAnotherScheme)
	assert.Equal(t, http.StatusNoContent, several.status)
	assert.Equal(t, domain.Bounced, hook.status(t, hook.sentID))
}

func TestAnEventThatIsNotABounceIsAnsweredAndChangesNothing(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)

	for _, eventType := range []string{"email.delivered", "email.complained", "email.opened"} {
		body := resendtest.Event(eventType, resendID, map[string]string{"email_id": "37"}, recipient)

		answer := hook.signed(t, "msg_"+eventType, body)

		assert.Equal(t, http.StatusNoContent, answer.status, eventType)
	}
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))
	assert.Empty(t, hook.notices.Kept())
}

// An e-mail this system did not send carries no tag of a record, or one that
// is not a number.
func TestABounceOfAnEmailWithNoRecordOfTheSystemIsAnsweredAndChangesNothing(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)

	for notice, tags := range map[string]map[string]string{
		"msg_no_tags":    nil,
		"msg_other_tags": {"category": "newsletter"},
		"msg_not_number": {"email_id": "ana"},
		"msg_zero":       {"email_id": "0"},
	} {
		answer := hook.signed(t, notice, resendtest.Event("email.bounced", resendID, tags, recipient))

		assert.Equal(t, http.StatusNoContent, answer.status, notice)
	}
	assert.Equal(t, domain.Sent, hook.status(t, hook.sentID))
	assert.Empty(t, hook.notices.Kept())
}

// Resend's documentation shows the tags of an event as an object; the tags of
// a sending are a list of names and values. Both are read.
func TestTheTagsOfABounceAreReadAsAListToo(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	body := `{"type":"email.bounced","created_at":"2026-11-18T12:00:01.126Z","data":{"email_id":"` + resendID +
		`","to":["` + recipient + `"],"tags":[{"name":"category","value":"login"},{"name":"email_id","value":"37"}]}}`

	answer := hook.signed(t, "msg_2fKq8bX1", body)

	assert.Equal(t, http.StatusNoContent, answer.status)
	assert.Equal(t, domain.Bounced, hook.status(t, hook.sentID))
}

func TestASignedNoticeThatIsNotJSONIsRefused(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)

	answer := hook.signed(t, "msg_2fKq8bX1", "type=email.bounced")

	require.Equal(t, http.StatusBadRequest, answer.status)
	assert.Equal(t, "bad_request", answer.problem["code"])
}

func TestABounceThatCannotBeTakenNowFailsSoResendDeliversItAgain(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)
	to, err := domain.NewRecipient(recipient)
	require.NoError(t, err)
	english, err := kernel.NewLanguage("en-US")
	require.NoError(t, err)
	queued, err := domain.Queue(39, to, domain.LoginCode, english, domain.LoginCodeRequest(43), noon)
	require.NoError(t, err)
	require.NoError(t, hook.emails.Save(context.Background(), queued))

	answer := hook.signed(t, "msg_2fKq8bX1", resendtest.Bounce(resendID, 39, recipient))

	require.Equal(t, http.StatusInternalServerError, answer.status)
	assert.Equal(t, "internal_server_error", answer.problem["code"])
	assert.Empty(t, hook.notices.Kept())
}

func TestTheWebhookTakesPostOnly(t *testing.T) {
	t.Parallel()
	hook := newWebhook(t)

	response, err := http.Get(hook.url)
	require.NoError(t, err)
	defer response.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode)
	assert.Equal(t, "POST", response.Header.Get("Allow"))
}

// withHeader gives the headers with the value of one of them replaced: an
// empty value is no value.
func withHeader(headers []string, name, value string) []string {
	changed := slices.Clone(headers)
	changed[slices.Index(changed, name)+1] = value
	return changed
}
