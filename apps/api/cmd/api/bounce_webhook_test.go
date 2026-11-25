package main

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/forgespark/apps/api/internal/email/adapters/resend/resendtest"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/database/databasetest"
)

// The webhook of the bounces through the binary: present only with its
// secret, refusing a wrong signature, taking a notice once. Its other cases
// are tested on the adapter, in email/adapters/resend.

const webhookPath = "/api/resend/bounces"

// withBounceWebhook gives the API the secret of the webhook of Resend.
func withBounceWebhook(environment map[string]string) map[string]string {
	environment["RESEND_WEBHOOK_SECRET"] = resendtest.Secret
	return environment
}

// sentThroughResend records an e-mail Resend took, with the identifier it
// gave it, as the worker would.
func sentThroughResend(t *testing.T, owner *pgx.Conn, id int64, resendID string) {
	t.Helper()
	_, err := owner.Exec(context.Background(), `
		INSERT INTO emails (id, recipient, subject, template, language, reference_kind, reference_id, status,
		                    queued_at, message_id, sent_at)
		VALUES ($1, 'ana.souza@fabrica.example', 'Your login code', 'login_code', 'en-US', 'login_code_request', 42, 'sent',
		        now(), $2, now())`, id, resendID)
	require.NoError(t, err)
}

func statusOfEmail(t *testing.T, owner *pgx.Conn, id int64) string {
	t.Helper()
	var status string
	require.NoError(t, owner.QueryRow(context.Background(), "SELECT status FROM emails WHERE id = $1", id).Scan(&status))
	return status
}

func noticesKept(t *testing.T, owner *pgx.Conn) []string {
	t.Helper()
	rows, err := owner.Query(context.Background(), "SELECT id FROM bounce_notices ORDER BY id")
	require.NoError(t, err)
	kept, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	return kept
}

// deliverNotice sends a notice to the webhook as Resend does, with the
// headers of its signature, and no cookie.
func deliverNotice(t *testing.T, api string, headers []string, body string) received {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, api+webhookPath, strings.NewReader(body))
	require.NoError(t, err)
	for index := 1; index < len(headers); index += 2 {
		request.Header.Set(headers[index-1], headers[index])
	}
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	return received{status: response.StatusCode, mediaType: response.Header.Get("Content-Type"), body: string(raw)}
}

// signedNotice sends a bounce of the e-mail, signed with the secret of the
// webhook now.
func signedNotice(t *testing.T, api, notice, resendID string, emailID int64) received {
	t.Helper()
	body := resendtest.Bounce(resendID, emailID, "ana.souza@fabrica.example")
	return deliverNotice(t, api, resendtest.Headers(resendtest.Secret, notice, time.Now(), body), body)
}

func TestWithoutItsSecretTheBounceWebhookIsAnUnknownPath(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	owner := asOwner(t, temporary)
	sentThroughResend(t, owner, 37, "4ef8a1c2-resend")
	api := startAPI(t, environmentOf(temporary))

	answer := signedNotice(t, api, "msg_2fKq8bX1", "4ef8a1c2-resend", 37)

	require.Equal(t, http.StatusNotFound, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Contains(t, answer.body, `"code":"not_found"`)
	assert.Equal(t, "sent", statusOfEmail(t, owner, 37))
	assert.Empty(t, noticesKept(t, owner))
}

func TestTheBounceWebhookRefusesANoticeWithAWrongSignature(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	owner := asOwner(t, temporary)
	sentThroughResend(t, owner, 37, "4ef8a1c2-resend")
	api := startAPI(t, withBounceWebhook(environmentOf(temporary)))
	body := resendtest.Bounce("4ef8a1c2-resend", 37, "ana.souza@fabrica.example")
	forged := "whsec_" + base64.StdEncoding.EncodeToString([]byte("a key that is not the webhook's"))

	answer := deliverNotice(t, api, resendtest.Headers(forged, "msg_2fKq8bX1", time.Now(), body), body)

	require.Equal(t, http.StatusUnauthorized, answer.status, answer.body)
	assert.Equal(t, "application/problem+json", answer.mediaType)
	assert.Contains(t, answer.body, `"code":"unauthorized"`)
	assert.Contains(t, answer.body, `"detail":"the signature of the notice is not valid"`)
	assert.Equal(t, "sent", statusOfEmail(t, owner, 37))
	assert.Empty(t, noticesKept(t, owner))
}

func TestTheBounceWebhookMarksTheEmailBouncedAndTakesARepeatedNoticeOnce(t *testing.T) {
	t.Parallel()
	temporary := databasetest.Migrated(t)
	owner := asOwner(t, temporary)
	sentThroughResend(t, owner, 37, "4ef8a1c2-resend")
	sentThroughResend(t, owner, 38, "9b2d7e10-resend")
	api := startAPI(t, withBounceWebhook(environmentOf(temporary)))

	first := signedNotice(t, api, "msg_2fKq8bX1", "4ef8a1c2-resend", 37)
	// The same notice again, naming the other e-mail: only its identifier
	// tells it was taken.
	again := signedNotice(t, api, "msg_2fKq8bX1", "9b2d7e10-resend", 38)

	require.Equal(t, http.StatusNoContent, first.status, first.body)
	require.Equal(t, http.StatusNoContent, again.status, again.body)
	assert.Equal(t, "bounced", statusOfEmail(t, owner, 37))
	assert.Equal(t, "sent", statusOfEmail(t, owner, 38))
	assert.Equal(t, []string{"msg_2fKq8bX1"}, noticesKept(t, owner))
}

// The webhook is a handler of its own: the spec, from which the client of
// the app is generated, does not list it.
func TestTheSpecLeavesTheBounceWebhookOut(t *testing.T) {
	t.Parallel()

	for path := range generatedSpec(t).Paths {
		assert.NotContains(t, path, "resend")
		assert.NotContains(t, path, "bounce")
	}
}
